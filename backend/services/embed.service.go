package services

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"syscall"
	"time"

	"chatcoal/cache"
	"chatcoal/database"
	"chatcoal/models"

	"github.com/gofiber/fiber/v2/log"
	"golang.org/x/net/html"
)

var urlRegex = regexp.MustCompile(`https?://[^\s<>"` + "`" + `\)\]]+`)

const (
	maxURLsPerMessage = 5
	ogFetchTimeout    = 5 * time.Second
	ogMaxBody         = 512 * 1024 // 512 KB
	ogMaxRedirects    = 3
)

// ExtractURLs finds up to maxURLsPerMessage URLs in a message.
func ExtractURLs(content string) []string {
	matches := urlRegex.FindAllString(content, -1)
	// De-duplicate while preserving order
	seen := make(map[string]struct{}, len(matches))
	result := make([]string, 0, len(matches))
	for _, u := range matches {
		// Trim trailing punctuation that's likely not part of the URL
		u = strings.TrimRight(u, ".,;:!?")
		if _, ok := seen[u]; ok {
			continue
		}
		seen[u] = struct{}{}
		result = append(result, u)
		if len(result) >= maxURLsPerMessage {
			break
		}
	}
	return result
}

// ogSSRFSafeDialer returns a dialer that rejects connections to private IPs.
func ogSSRFSafeDialer() *net.Dialer {
	return &net.Dialer{
		Timeout: ogFetchTimeout,
		Control: func(network, address string, c syscall.RawConn) error {
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				return err
			}
			ip := net.ParseIP(host)
			if ip != nil && isPrivateIP(ip) {
				return &net.AddrError{Err: "blocked private IP", Addr: address}
			}
			return nil
		},
	}
}

var ogHTTPClient = &http.Client{
	Timeout: ogFetchTimeout,
	Transport: &http.Transport{
		DialContext:       ogSSRFSafeDialer().DialContext,
		TLSClientConfig:  &tls.Config{MinVersion: tls.VersionTLS12},
		DisableKeepAlives: true,
	},
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= ogMaxRedirects {
			return http.ErrUseLastResponse
		}
		// Check redirect target for SSRF
		host := req.URL.Hostname()
		ips, err := net.LookupIP(host)
		if err == nil {
			for _, ip := range ips {
				if isPrivateIP(ip) {
					return &net.AddrError{Err: "redirect to private IP blocked", Addr: host}
				}
			}
		}
		return nil
	},
}

// FetchOGMetadata fetches OpenGraph metadata from a URL.
func FetchOGMetadata(rawURL string) (*models.LinkEmbed, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	// DNS lookup to check for private IPs before connecting
	ips, err := net.LookupIP(parsed.Hostname())
	if err != nil {
		return nil, err
	}
	for _, ip := range ips {
		if isPrivateIP(ip) {
			return nil, &net.AddrError{Err: "private IP", Addr: parsed.Hostname()}
		}
	}

	if !cache.OGDomainRateLimitOK(parsed.Hostname()) {
		return nil, nil
	}

	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "chatcoal-bot/1.0 (link preview)")
	req.Header.Set("Accept", "text/html")

	resp, err := ogHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, nil
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/html") {
		return nil, nil
	}

	body := io.LimitReader(resp.Body, ogMaxBody)
	return parseOGTags(body, rawURL)
}

// parseOGTags parses OG meta tags from HTML.
func parseOGTags(r io.Reader, rawURL string) (*models.LinkEmbed, error) {
	tokenizer := html.NewTokenizer(r)
	embed := &models.LinkEmbed{URL: rawURL}
	var titleFromTag string
	var inTitle bool

	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			// End of document
			if embed.Title == "" && titleFromTag != "" {
				embed.Title = titleFromTag
			}
			if embed.Title == "" && embed.Description == "" && embed.Image == "" {
				return nil, nil
			}
			return embed, nil

		case html.StartTagToken, html.SelfClosingTagToken:
			tn, hasAttr := tokenizer.TagName()
			tag := string(tn)

			if tag == "title" {
				inTitle = true
				continue
			}

			// Stop parsing at <body> — OG tags are in <head>
			if tag == "body" {
				if embed.Title == "" && titleFromTag != "" {
					embed.Title = titleFromTag
				}
				if embed.Title == "" && embed.Description == "" && embed.Image == "" {
					return nil, nil
				}
				return embed, nil
			}

			if tag != "meta" || !hasAttr {
				continue
			}

			var property, content string
			for {
				key, val, more := tokenizer.TagAttr()
				k := string(key)
				v := string(val)
				if k == "property" || k == "name" {
					property = v
				}
				if k == "content" {
					content = v
				}
				if !more {
					break
				}
			}

			switch property {
			case "og:title":
				embed.Title = content
			case "og:description":
				embed.Description = content
			case "og:image":
				embed.Image = content
			case "og:site_name":
				embed.SiteName = content
			case "description":
				if embed.Description == "" {
					embed.Description = content
				}
			}

		case html.TextToken:
			if inTitle {
				titleFromTag = strings.TrimSpace(string(tokenizer.Text()))
			}

		case html.EndTagToken:
			tn, _ := tokenizer.TagName()
			if string(tn) == "title" {
				inTitle = false
			}
		}
	}
}

// FetchAndStoreEmbeds extracts URLs, fetches OG metadata, stores in DB,
// and calls the broadcast callback.
func FetchAndStoreEmbeds(messageID models.Snowflake, content string, tableName string, broadcastFn func(embeds json.RawMessage)) {
	urls := ExtractURLs(content)
	if len(urls) == 0 {
		return
	}

	var embeds []models.LinkEmbed
	for _, u := range urls {
		// Check cache first
		if cached, ok := cache.GetOGCache(u); ok {
			var embed models.LinkEmbed
			if json.Unmarshal([]byte(cached), &embed) == nil && embed.Title != "" {
				embeds = append(embeds, embed)
			}
			continue
		}

		embed, err := FetchOGMetadata(u)
		if err != nil {
			log.Debugf("[embed] fetch failed for %s: %v", u, err)
			continue
		}
		if embed == nil {
			// Cache empty result to avoid re-fetching
			cache.SetOGCache(u, "{}")
			continue
		}

		// Truncate long fields
		if len(embed.Description) > 300 {
			embed.Description = embed.Description[:300] + "..."
		}
		if len(embed.Title) > 200 {
			embed.Title = embed.Title[:200]
		}

		data, _ := json.Marshal(embed)
		cache.SetOGCache(u, string(data))
		embeds = append(embeds, *embed)
	}

	if len(embeds) == 0 {
		return
	}

	embedJSON, err := json.Marshal(embeds)
	if err != nil {
		return
	}

	if err := database.Database.Table(tableName).
		Where("id = ?", messageID).
		Update("embeds", embedJSON).Error; err != nil {
		log.Errorf("[embed] DB update failed for message %d: %v", messageID, err)
		return
	}

	if broadcastFn != nil {
		broadcastFn(embedJSON)
	}
}
