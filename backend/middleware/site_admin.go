package middleware

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
)

var siteAdminUIDs map[string]bool

func init() {
	siteAdminUIDs = make(map[string]bool)
	raw := os.Getenv("SITE_ADMIN_UIDS")
	if raw == "" {
		return
	}
	for _, uid := range strings.Split(raw, ",") {
		uid = strings.TrimSpace(uid)
		if uid != "" {
			siteAdminUIDs[uid] = true
		}
	}
}

// IsSiteAdmin returns true if the given Firebase UID is in the SITE_ADMIN_UIDS list.
func IsSiteAdmin(uid string) bool {
	return siteAdminUIDs[uid]
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
