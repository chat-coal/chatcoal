package middleware

import (
	"context"
	"strings"
	"time"

	"chatcoal/cache"
	"chatcoal/services"

	"github.com/gofiber/fiber/v2"
)

// UnifiedAuthMiddleware accepts both Firebase ID tokens and federation session tokens.
// It sets c.Locals("firebaseUID") on success. For Firebase users this is the real UID;
// for federated users it is the synthetic "fed:alice@instance-a.com" string.
func UnifiedAuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing authorization header"})
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid authorization format"})
		}

		// Try federation session first (fast Redis lookup).
		if fid, err := services.VerifyFederationSession(token); err == nil {
			c.Locals("firebaseUID", fid)
			return c.Next()
		}

		// Fall back to Firebase.
		if firebaseAuth == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Firebase not initialized"})
		}

		// Check Redis cache — same logic as FirebaseAuthMiddleware.
		if cache.Redis != nil {
			if val, err := cache.Redis.Get(context.Background(), tokenCacheKey(token)).Result(); err == nil {
				// Cached values: "uid", "anon:uid", or "unverified:uid"
				if rest, ok := strings.CutPrefix(val, "anon:"); ok {
					c.Locals("firebaseUID", rest)
					c.Locals("firebaseIsAnonymous", true)
					return c.Next()
				}
				if rest, ok := strings.CutPrefix(val, "unverified:"); ok {
					c.Locals("firebaseUID", rest)
					c.Locals("firebaseIsAnonymous", false)
					c.Locals("firebaseEmailVerified", false)
					return c.Next()
				}
				c.Locals("firebaseUID", val)
				c.Locals("firebaseIsAnonymous", false)
				c.Locals("firebaseEmailVerified", true)
				return c.Next()
			}
		}

		// Cache miss — full Firebase verification.
		decodedToken, err := firebaseAuth.VerifyIDToken(context.Background(), token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
		}

		isAnon := decodedToken.Firebase.SignInProvider == "anonymous"
		emailVerified := true
		if !isAnon {
			if ev, ok := decodedToken.Claims["email_verified"].(bool); ok {
				emailVerified = ev
			}
		}

		if cache.Redis != nil {
			ttl := time.Until(time.Unix(decodedToken.Expires, 0)) - 5*time.Minute
			if ttl > 0 {
				cacheVal := decodedToken.UID
				if isAnon {
					cacheVal = "anon:" + decodedToken.UID
				} else if !emailVerified {
					cacheVal = "unverified:" + decodedToken.UID
				}
				cache.Redis.Set(context.Background(), tokenCacheKey(token), cacheVal, ttl)
			}
		}

		c.Locals("firebaseUID", decodedToken.UID)
		c.Locals("firebaseToken", decodedToken)
		c.Locals("firebaseIsAnonymous", isAnon)
		c.Locals("firebaseEmailVerified", emailVerified)
		return c.Next()
	}
}

// VerifyTokenUnified verifies a token — federation session first, then Firebase.
// Used by the WebSocket auth handshake.
func VerifyTokenUnified(token string) (string, error) {
	if fid, err := services.VerifyFederationSession(token); err == nil {
		return fid, nil
	}
	return VerifyToken(token)
}
