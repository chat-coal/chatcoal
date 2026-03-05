package services

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"chatcoal/cache"
	"chatcoal/database"
	"chatcoal/models"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gofiber/fiber/v2/log"
)

// InstanceInfo is the JSON response from GET /federation/info.
type InstanceInfo struct {
	Domain    string `json:"domain"`
	Name      string `json:"name"`
	PublicKey string `json:"public_key_pem"`
	Version   string `json:"version"`
}

// AssertionClaims are the JWT claims embedded in a federation assertion token.
type AssertionClaims struct {
	jwt.RegisteredClaims
	DisplayName  string `json:"display_name"`
	AvatarURL    string `json:"avatar_url"`
	Nonce        string `json:"nonce"`
	HomeInstance string `json:"-"` // populated from Issuer after verification
}

var (
	instancePrivateKey ed25519.PrivateKey
	instancePublicKey  ed25519.PublicKey
	instanceDomain     string
)

// InitFederationKeys loads or generates the instance's Ed25519 keypair.
// It is a no-op when FEDERATION_DOMAIN is not set.
func InitFederationKeys() error {
	domain := os.Getenv("FEDERATION_DOMAIN")
	if domain == "" {
		return nil
	}
	instanceDomain = domain

	var config models.InstanceConfig
	err := database.Database.First(&config).Error
	if err != nil {
		// Generate a new keypair and persist it.
		pub, priv, genErr := ed25519.GenerateKey(rand.Reader)
		if genErr != nil {
			return fmt.Errorf("federation: failed to generate keypair: %w", genErr)
		}
		privPEM, encErr := encodePrivateKey(priv)
		if encErr != nil {
			return encErr
		}
		pubPEM, encErr := encodePublicKey(pub)
		if encErr != nil {
			return encErr
		}
		config = models.InstanceConfig{
			ID:         1,
			PrivateKey: privPEM,
			PublicKey:  pubPEM,
			Domain:     domain,
		}
		if dbErr := database.Database.Create(&config).Error; dbErr != nil {
			return fmt.Errorf("federation: failed to store keypair: %w", dbErr)
		}
		instancePrivateKey = priv
		instancePublicKey = pub
		return nil
	}

	priv, parseErr := decodePrivateKey(config.PrivateKey)
	if parseErr != nil {
		return parseErr
	}
	pub, parseErr := decodePublicKeyBytes(config.PublicKey)
	if parseErr != nil {
		return parseErr
	}
	instancePrivateKey = priv
	instancePublicKey = pub
	return nil
}

// GetPublicKeyPEM returns this instance's PEM-encoded Ed25519 public key.
func GetPublicKeyPEM() string {
	if instancePublicKey == nil {
		return ""
	}
	s, _ := encodePublicKey(instancePublicKey)
	return s
}

// SignAssertion creates a signed JWT assertion for federated login.
func SignAssertion(username, displayName, avatarURL, visitingDomain, nonce string) (string, error) {
	if instancePrivateKey == nil {
		return "", errors.New("federation: not initialized")
	}
	now := time.Now()
	claims := &AssertionClaims{
		DisplayName: displayName,
		AvatarURL:   avatarURL,
		Nonce:       nonce,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   username,
			Issuer:    instanceDomain,
			Audience:  jwt.ClaimStrings{visitingDomain},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(5 * time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	return token.SignedString(instancePrivateKey)
}

// VerifyAssertion validates a federation assertion JWT and returns its claims.
// expectedVisiting must equal FEDERATION_DOMAIN of the verifying instance.
func VerifyAssertion(tokenStr, expectedVisiting string) (*AssertionClaims, error) {
	claims := &AssertionClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("federation: unexpected signing method %v", t.Header["alg"])
		}
		c, ok := t.Claims.(*AssertionClaims)
		if !ok {
			return nil, errors.New("federation: invalid claims type")
		}
		if c.Issuer == "" {
			return nil, errors.New("federation: missing issuer")
		}
		return fetchFederationPublicKey(c.Issuer)
	}, jwt.WithValidMethods([]string{"EdDSA"}))
	if err != nil {
		return nil, fmt.Errorf("federation: %w", err)
	}
	if !token.Valid {
		return nil, errors.New("federation: token invalid")
	}
	if !claims.VerifyAudience(expectedVisiting, true) {
		return nil, errors.New("federation: token audience mismatch")
	}
	claims.HomeInstance = claims.Issuer
	return claims, nil
}

// fetchFederationPublicKey returns the Ed25519 public key for the given domain,
// checking the DB cache first and falling back to a live fetch.
func fetchFederationPublicKey(domain string) (ed25519.PublicKey, error) {
	var instance models.FederatedInstance
	if err := database.Database.Where("domain = ?", domain).First(&instance).Error; err == nil {
		pub, decErr := decodePublicKeyBytes(instance.PublicKey)
		if decErr == nil {
			return pub, nil
		}
		// Cached key is corrupted — return the error instead of silently
		// falling through to a remote fetch that could serve attacker keys.
		return nil, fmt.Errorf("federation: cached public key for %s is corrupted: %w", domain, decErr)
	}
	info, err := FetchInstanceInfo(domain)
	if err != nil {
		return nil, err
	}
	return decodePublicKeyBytes(info.PublicKey)
}

// safeTransport returns an http.Transport that validates resolved IPs at
// dial-time, preventing DNS rebinding attacks.
func safeTransport() *http.Transport {
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	return &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			addrs, err := net.DefaultResolver.LookupHost(ctx, host)
			if err != nil {
				return nil, err
			}
			for _, a := range addrs {
				ip := net.ParseIP(a)
				if ip != nil && isPrivateIP(ip) {
					return nil, fmt.Errorf("domain resolves to a private/reserved IP")
				}
			}
			// Connect to the first resolved address.
			return dialer.DialContext(ctx, network, net.JoinHostPort(addrs[0], port))
		},
	}
}

// FetchInstanceInfo fetches and caches federation metadata from a remote instance.
func FetchInstanceInfo(domain string) (*InstanceInfo, error) {
	if err := ValidatePublicDomain(domain); err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 10 * time.Second, Transport: safeTransport()}
	resp, err := client.Get("https://" + domain + "/federation/info")
	if err != nil {
		return nil, fmt.Errorf("federation: failed to reach %s: %w", domain, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("federation: /federation/info returned status %d", resp.StatusCode)
	}
	var info InstanceInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("federation: failed to decode instance info: %w", err)
	}

	// Upsert into DB cache.
	var instance models.FederatedInstance
	if err := database.Database.Where("domain = ?", domain).First(&instance).Error; err != nil {
		instance = models.FederatedInstance{Domain: domain, PublicKey: info.PublicKey, Name: info.Name}
		database.Database.Create(&instance)
	} else {
		database.Database.Model(&instance).Updates(map[string]interface{}{
			"public_key": info.PublicKey,
			"name":       info.Name,
		})
	}
	return &info, nil
}

// ValidatePublicDomain checks that a domain string is safe for outbound requests.
func ValidatePublicDomain(domain string) error {
	if domain == "" {
		return errors.New("empty domain")
	}
	if strings.ContainsAny(domain, "/:@") {
		return errors.New("domain must not contain slashes, colons, or @ signs")
	}
	if !strings.Contains(domain, ".") {
		return errors.New("domain must contain a dot")
	}
	addrs, err := net.LookupHost(domain)
	if err != nil {
		return fmt.Errorf("failed to resolve domain %q: %w", domain, err)
	}
	for _, addr := range addrs {
		ip := net.ParseIP(addr)
		if ip != nil && isPrivateIP(ip) {
			return fmt.Errorf("domain %q resolves to a private/reserved IP", domain)
		}
	}
	return nil
}

var privateRanges = []net.IPNet{
	mustCIDR("0.0.0.0/8"),
	mustCIDR("10.0.0.0/8"),
	mustCIDR("100.64.0.0/10"),
	mustCIDR("127.0.0.0/8"),
	mustCIDR("169.254.0.0/16"),
	mustCIDR("172.16.0.0/12"),
	mustCIDR("192.168.0.0/16"),
	mustCIDR("240.0.0.0/4"),
	mustCIDR("::1/128"),
	mustCIDR("fc00::/7"),
	mustCIDR("fe80::/10"),
}

func mustCIDR(s string) net.IPNet {
	_, n, err := net.ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	return *n
}

func isPrivateIP(ip net.IP) bool {
	for _, r := range privateRanges {
		if r.Contains(ip) {
			return true
		}
	}
	return false
}

// CreateFederationSession stores a federation session in Redis and returns the raw token.
func CreateFederationSession(federatedUID string) (string, error) {
	if cache.Redis == nil {
		return "", errors.New("federation: redis not available")
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := hex.EncodeToString(raw)
	h := sha256.Sum256([]byte(token))
	key := "fed:session:" + hex.EncodeToString(h[:])
	if err := cache.Redis.Set(context.Background(), key, federatedUID, 7*24*time.Hour).Err(); err != nil {
		return "", err
	}
	return token, nil
}

// VerifyFederationSession looks up a session token and returns the stored firebaseUID.
func VerifyFederationSession(token string) (string, error) {
	if cache.Redis == nil {
		return "", errors.New("federation: redis not available")
	}
	h := sha256.Sum256([]byte(token))
	key := "fed:session:" + hex.EncodeToString(h[:])
	return cache.Redis.Get(context.Background(), key).Result()
}

// DeleteFederationSession invalidates a federation session token.
func DeleteFederationSession(token string) error {
	if cache.Redis == nil {
		return nil
	}
	h := sha256.Sum256([]byte(token))
	key := "fed:session:" + hex.EncodeToString(h[:])
	return cache.Redis.Del(context.Background(), key).Err()
}

// --- PEM helpers ---

func encodePrivateKey(priv ed25519.PrivateKey) (string, error) {
	b, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return "", err
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: b})), nil
}

func encodePublicKey(pub ed25519.PublicKey) (string, error) {
	b, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return "", err
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: b})), nil
}

func decodePrivateKey(pemStr string) (ed25519.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("federation: failed to decode private key PEM")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	priv, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, errors.New("federation: parsed key is not ed25519")
	}
	return priv, nil
}

func decodePublicKeyBytes(pemStr string) (ed25519.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("federation: failed to decode public key PEM")
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub, ok := key.(ed25519.PublicKey)
	if !ok {
		return nil, errors.New("federation: parsed key is not ed25519")
	}
	return pub, nil
}

// --- Channel Federation ---

// ChannelMessageClaims are the JWT claims for a federated channel message.
type ChannelMessageClaims struct {
	jwt.RegisteredClaims
	FederationID string `json:"federation_id"`
	ChannelName  string `json:"channel_name"`
	Content      string `json:"content"`
	AuthorName   string `json:"author_name"`
	AuthorAvatar string `json:"author_avatar"`
	AuthorFedUID string `json:"author_fed_uid"`
	MessageType  string `json:"message_type"`
}

// GenerateFederationID creates a 32-byte random hex string for channel federation.
func GenerateFederationID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// SignChannelMessage creates a signed JWT for relaying a message to a federated channel.
func SignChannelMessage(federationID, content, authorName, authorAvatar, authorFedUID, messageType string) (string, error) {
	if instancePrivateKey == nil {
		return "", errors.New("federation: not initialized")
	}
	now := time.Now()
	claims := &ChannelMessageClaims{
		FederationID: federationID,
		Content:      content,
		AuthorName:   authorName,
		AuthorAvatar: authorAvatar,
		AuthorFedUID: authorFedUID,
		MessageType:  messageType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    instanceDomain,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(5 * time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	return token.SignedString(instancePrivateKey)
}

// VerifyChannelMessage validates a federated channel message JWT.
func VerifyChannelMessage(tokenStr string) (*ChannelMessageClaims, error) {
	claims := &ChannelMessageClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("federation: unexpected signing method %v", t.Header["alg"])
		}
		c, ok := t.Claims.(*ChannelMessageClaims)
		if !ok {
			return nil, errors.New("federation: invalid claims type")
		}
		if c.Issuer == "" {
			return nil, errors.New("federation: missing issuer")
		}
		return fetchFederationPublicKey(c.Issuer)
	}, jwt.WithValidMethods([]string{"EdDSA"}))
	if err != nil {
		return nil, fmt.Errorf("federation: %w", err)
	}
	if !token.Valid {
		return nil, errors.New("federation: token invalid")
	}
	return claims, nil
}

// EnableChannelFederation sets a federation_id on a channel, making it federable.
func EnableChannelFederation(channelID models.Snowflake) (*models.Channel, error) {
	var ch models.Channel
	if err := database.Database.First(&ch, channelID).Error; err != nil {
		return nil, errors.New("channel not found")
	}
	if ch.FederationID != nil {
		return &ch, nil // already enabled
	}
	fedID := GenerateFederationID()
	ch.FederationID = &fedID
	if err := database.Database.Model(&ch).Update("federation_id", fedID).Error; err != nil {
		return nil, err
	}
	return &ch, nil
}

// DisableChannelFederation removes federation from a channel and deletes all links.
func DisableChannelFederation(channelID models.Snowflake) error {
	database.Database.Where("channel_id = ?", channelID).Delete(&models.FederatedChannelLink{})
	return database.Database.Model(&models.Channel{}).Where("id = ?", channelID).Update("federation_id", nil).Error
}

// CreateChannelLink links a local channel to a remote federated channel.
func CreateChannelLink(channelID models.Snowflake, remoteDomain, remoteFedID string) (*models.FederatedChannelLink, error) {
	link := models.FederatedChannelLink{
		ChannelID:          channelID,
		RemoteDomain:       remoteDomain,
		RemoteFederationID: remoteFedID,
		Active:             true,
	}
	if err := database.Database.Create(&link).Error; err != nil {
		return nil, err
	}
	return &link, nil
}

// DeleteChannelLink removes a channel link.
func DeleteChannelLink(linkID models.Snowflake) error {
	return database.Database.Delete(&models.FederatedChannelLink{}, linkID).Error
}

// GetChannelLinks returns all links for a given channel.
func GetChannelLinks(channelID models.Snowflake) ([]models.FederatedChannelLink, error) {
	var links []models.FederatedChannelLink
	err := database.Database.Where("channel_id = ?", channelID).Find(&links).Error
	return links, err
}

// GetChannelByFederationID finds a channel by its federation_id.
func GetChannelByFederationID(fedID string) (*models.Channel, error) {
	var ch models.Channel
	if err := database.Database.Where("federation_id = ?", fedID).First(&ch).Error; err != nil {
		return nil, err
	}
	return &ch, nil
}

// GetLinksForFederationID returns all active links where the local channel
// has the given federation_id. Used for hub relay.
func GetLinksForFederationID(fedID string) ([]models.FederatedChannelLink, error) {
	ch, err := GetChannelByFederationID(fedID)
	if err != nil {
		return nil, err
	}
	var links []models.FederatedChannelLink
	err = database.Database.Where("channel_id = ? AND active = ?", ch.ID, true).Find(&links).Error
	return links, err
}

// RelayMessageToRemote sends a signed JWT to a remote instance's channel message endpoint.
func RelayMessageToRemote(domain, federationID, signedJWT string) error {
	if err := ValidatePublicDomain(domain); err != nil {
		return err
	}
	body, _ := json.Marshal(map[string]string{"token": signedJWT})
	client := &http.Client{Timeout: 10 * time.Second, Transport: safeTransport()}
	url := "https://" + domain + "/federation/channels/" + federationID + "/messages"
	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("federation relay: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("federation relay: remote returned %d", resp.StatusCode)
	}
	return nil
}

// RelayFederatedMessage signs a message and relays it to hub or subscribers.
// If the channel has links (this instance is linking to a remote hub), send to hub.
// If the channel IS the hub (has federation_id and links from others), relay to subscribers.
func RelayFederatedMessage(ch *models.Channel, message *models.Message, user *models.User) {
	if ch.FederationID == nil || instanceDomain == "" {
		return
	}

	fedUID := user.FirebaseUID
	signedToken, err := SignChannelMessage(
		*ch.FederationID,
		message.Content,
		user.DisplayName,
		user.AvatarURL,
		fedUID,
		message.Type,
	)
	if err != nil {
		log.Errorf("federation: failed to sign message: %v", err)
		return
	}

	// If we have outbound links for this channel, send to those remotes (spoke → hub).
	links, _ := GetChannelLinks(ch.ID)
	for _, link := range links {
		if err := RelayMessageToRemote(link.RemoteDomain, link.RemoteFederationID, signedToken); err != nil {
			log.Errorf("federation: relay to %s failed: %v", link.RemoteDomain, err)
		}
	}

	// If we are the hub (other instances link to our federation_id), relay to all subscribers.
	hubLinks, _ := GetLinksForFederationID(*ch.FederationID)
	for _, link := range hubLinks {
		// Don't relay back to the originating domain.
		if strings.HasPrefix(fedUID, "fed:") && strings.HasSuffix(fedUID, "@"+link.RemoteDomain) {
			continue
		}
		if err := RelayMessageToRemote(link.RemoteDomain, link.RemoteFederationID, signedToken); err != nil {
			log.Errorf("federation: hub relay to %s failed: %v", link.RemoteDomain, err)
		}
	}
}

// FetchRemoteChannelInfo fetches basic info about a remote federated channel.
func FetchRemoteChannelInfo(domain, federationID string) (map[string]string, error) {
	if err := ValidatePublicDomain(domain); err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 10 * time.Second, Transport: safeTransport()}
	resp, err := client.Get("https://" + domain + "/federation/channels/" + federationID + "/info")
	if err != nil {
		return nil, fmt.Errorf("federation: failed to reach %s: %w", domain, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("federation: remote returned %d", resp.StatusCode)
	}
	var info map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	return info, nil
}
