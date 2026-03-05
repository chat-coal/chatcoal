package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// PerUserRateLimiter returns a rate limiter keyed by Firebase UID.
// Must be placed after FirebaseAuthMiddleware so firebaseUID is set in locals.
func PerUserRateLimiter(max int, expiration time.Duration) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: expiration,
		KeyGenerator: func(c *fiber.Ctx) string {
			if uid, ok := c.Locals("firebaseUID").(string); ok && uid != "" {
				return "uid:" + uid
			}
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "Too many requests"})
		},
	})
}
