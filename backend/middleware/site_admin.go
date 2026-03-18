package middleware

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// IsSiteAdmin returns true if the given Firebase UID is in the SITE_ADMIN_UIDS list.
// Reads the env var on every call so it works even when godotenv loads after package init.
func IsSiteAdmin(uid string) bool {
	raw := os.Getenv("SITE_ADMIN_UIDS")
	if raw == "" {
		return false
	}
	for _, admin := range strings.Split(raw, ",") {
		if strings.TrimSpace(admin) == uid {
			return true
		}
	}
	return false
}

// SiteAdminMiddleware returns 403 if the authenticated user is not a site admin.
func SiteAdminMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		uid, _ := c.Locals("firebaseUID").(string)
		if uid == "" || !IsSiteAdmin(uid) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
		}
		return c.Next()
	}
}
