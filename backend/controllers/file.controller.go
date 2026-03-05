package controllers

import (
	"chatcoal/cache"
	"chatcoal/storage"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ServeFile generates a presigned S3 URL for the requested object key and
// redirects the client to it. The presigned URL is valid for 1 hour.
func ServeFile(c *fiber.Ctx) error {
	key := c.Params("*")
	if key == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "missing file key"})
	}

	if !storage.Available() {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "s3 not configured"})
	}

	if cached := cache.GetPresignedURL(key); cached != "" {
		return c.Redirect(cached, fiber.StatusTemporaryRedirect)
	}

	url, err := storage.PresignedGetURL(c.Context(), key, 1*time.Hour)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate download URL"})
	}

	cache.SetPresignedURL(key, url)
	return c.Redirect(url, fiber.StatusTemporaryRedirect)
}
