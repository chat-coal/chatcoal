package routes

import (
	"chatcoal/controllers"
	"chatcoal/middleware"
	"time"

	"github.com/gofiber/fiber/v2"
)

// SetupFederationRoutes registers all federation endpoints.
func SetupFederationRoutes(app *fiber.App) {
	// Public discovery endpoint — no auth, no CORS restriction.
	app.Get("/federation/info", controllers.GetFederationInfo)

	// Server-to-server channel federation endpoints (public, rate-limited).
	app.Get("/federation/channels/:federationId/info", controllers.GetFederatedChannelInfo)
	app.Post("/federation/channels/:federationId/messages",
		middleware.PerUserRateLimiter(30, 1*time.Minute), controllers.ReceiveFederatedMessage)

	// API endpoints under /api/federation — tighter rate limit.
	fed := app.Group("/api/federation", middleware.PerUserRateLimiter(5, 1*time.Minute))
	fed.Post("/begin", controllers.BeginFederation)
	fed.Post("/verify", controllers.VerifyFederation)

	// Authorize requires the user to be authenticated on this instance.
	fed.Post("/authorize", middleware.UnifiedAuthMiddleware(), controllers.AuthorizeFederation)
}
