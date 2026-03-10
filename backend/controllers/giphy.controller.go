package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

var giphyClient = &http.Client{Timeout: 5 * time.Second}

// SearchGifs proxies a Giphy search request so the API key stays server-side.
func SearchGifs(c *fiber.Ctx) error {
	apiKey := os.Getenv("GIPHY_API_KEY")
	if apiKey == "" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "GIF support is not configured"})
	}

	q := c.Query("q")
	if q == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "query is required"})
	}

	limit := c.Query("limit", "20")
	offset := c.Query("offset", "0")

	url := fmt.Sprintf("https://api.giphy.com/v1/gifs/search?api_key=%s&q=%s&limit=%s&offset=%s&rating=pg-13&lang=en",
		apiKey, q, limit, offset)

	return proxyGiphy(c, url)
}

// TrendingGifs proxies the Giphy trending endpoint.
func TrendingGifs(c *fiber.Ctx) error {
	apiKey := os.Getenv("GIPHY_API_KEY")
	if apiKey == "" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "GIF support is not configured"})
	}

	limit := c.Query("limit", "20")
	offset := c.Query("offset", "0")

	url := fmt.Sprintf("https://api.giphy.com/v1/gifs/trending?api_key=%s&limit=%s&offset=%s&rating=pg-13",
		apiKey, limit, offset)

	return proxyGiphy(c, url)
}

func proxyGiphy(c *fiber.Ctx, url string) error {
	resp, err := giphyClient.Get(url)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "failed to reach Giphy"})
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "failed to read Giphy response"})
	}

	if resp.StatusCode != http.StatusOK {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "Giphy returned an error"})
	}

	var result json.RawMessage
	if err := json.Unmarshal(body, &result); err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "invalid Giphy response"})
	}

	c.Set("Content-Type", "application/json")
	return c.Send(body)
}
