package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"chatcoal/cache"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/api/option"
)

var firebaseAuth *auth.Client

func InitFirebase() error {
	ctx := context.Background()

	// Try service account file first
	if _, err := os.Stat("firebase-service-account.json"); err == nil {
		opt := option.WithCredentialsFile("firebase-service-account.json")
		app, err := firebase.NewApp(ctx, nil, opt)
		if err != nil {
			return err
		}
		firebaseAuth, err = app.Auth(ctx)
		return err
	}

	// Fall back to project ID from env
	projectID := os.Getenv("FIREBASE_PROJECT_ID")
	if projectID == "" {
		return fmt.Errorf("no firebase credentials: set FIREBASE_PROJECT_ID env var or provide firebase-service-account.json")
	}

	app, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: projectID})
	if err != nil {
		return err
	}
	firebaseAuth, err = app.Auth(ctx)
	return err
}

func tokenCacheKey(token string) string {
	h := sha256.Sum256([]byte(token))
	return "fbtoken:" + hex.EncodeToString(h[:])
}

// verifyAndCache verifies a Firebase ID token, caching the result in Redis.
// Returns the Firebase UID on success.
func verifyAndCache(token string) (string, error) {
	if firebaseAuth == nil {
		return "", fmt.Errorf("firebase not initialized")
	}

	// Check Redis cache first — strip the "anon:" or "unverified:" prefix
	// if present so callers always receive the bare UID.
	if cache.Redis != nil {
		val, err := cache.Redis.Get(context.Background(), tokenCacheKey(token)).Result()
		if err == nil {
			uid := val
			if rest, ok := strings.CutPrefix(uid, "anon:"); ok {
				uid = rest
			} else if rest, ok := strings.CutPrefix(uid, "unverified:"); ok {
				uid = rest
			}
			return uid, nil
		}
	}

	// Cache miss — verify with Firebase
	decodedToken, err := firebaseAuth.VerifyIDToken(context.Background(), token)
	if err != nil {
		return "", err
	}

	// Cache with TTL based on token expiry (cap at 55 minutes to expire before the token).
	// Use the same prefix format as UnifiedAuthMiddleware so both code paths
	// produce a consistent cache entry.
	if cache.Redis != nil {
		ttl := time.Until(time.Unix(decodedToken.Expires, 0)) - 5*time.Minute
		if ttl > 0 {
			cacheVal := decodedToken.UID
			isAnon := decodedToken.Firebase.SignInProvider == "anonymous"
			if isAnon {
				cacheVal = "anon:" + decodedToken.UID
			} else {
				emailVerified := true
				if ev, ok := decodedToken.Claims["email_verified"].(bool); ok {
					emailVerified = ev
				}
				if !emailVerified {
					cacheVal = "unverified:" + decodedToken.UID
				}
			}
			cache.Redis.Set(context.Background(), tokenCacheKey(token), cacheVal, ttl)
		}
	}

	return decodedToken.UID, nil
}

// VerifyToken verifies a Firebase ID token and returns the Firebase UID.
// Used by WebSocket routes to authenticate before upgrade.
func VerifyToken(token string) (string, error) {
	return verifyAndCache(token)
}

// VerifyTokenFull calls Firebase directly (bypasses the Redis cache) and
// returns the full decoded token. Use this for sensitive operations that must
// confirm the token was issued recently.
func VerifyTokenFull(rawToken string) (*auth.Token, error) {
	if firebaseAuth == nil {
		return nil, fmt.Errorf("firebase not initialized")
	}
	return firebaseAuth.VerifyIDToken(context.Background(), rawToken)
}

// DeleteFirebaseUser deletes a Firebase user by UID.
func DeleteFirebaseUser(uid string) error {
	if firebaseAuth == nil {
		return fmt.Errorf("firebase not initialized")
	}
	return firebaseAuth.DeleteUser(context.Background(), uid)
}

func FirebaseAuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if firebaseAuth == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Firebase not initialized"})
		}

		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing authorization header"})
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid authorization format"})
		}

		// Check Redis cache first
		if cache.Redis != nil {
			if val, err := cache.Redis.Get(context.Background(), tokenCacheKey(token)).Result(); err == nil {
				uid, isAnon := strings.CutPrefix(val, "anon:")
				c.Locals("firebaseUID", uid)
				c.Locals("firebaseIsAnonymous", isAnon)
				return c.Next()
			}
		}

		// Cache miss — full verification (also provides decoded token for Login)
		decodedToken, err := firebaseAuth.VerifyIDToken(context.Background(), token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
		}

		if cache.Redis != nil {
			ttl := time.Until(time.Unix(decodedToken.Expires, 0)) - 5*time.Minute
			if ttl > 0 {
				cacheVal := decodedToken.UID
				if decodedToken.Firebase.SignInProvider == "anonymous" {
					cacheVal = "anon:" + decodedToken.UID
				}
				cache.Redis.Set(context.Background(), tokenCacheKey(token), cacheVal, ttl)
			}
		}

		c.Locals("firebaseUID", decodedToken.UID)
		c.Locals("firebaseToken", decodedToken)

		return c.Next()
	}
}
