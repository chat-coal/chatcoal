package main

import (
	"os"
	"strconv"
	"strings"
	"time"

	"chatcoal/cache"
	"chatcoal/database"
	"chatcoal/metrics"
	"chatcoal/middleware"
	"chatcoal/models"
	"chatcoal/routes"
	"chatcoal/services"
	"chatcoal/storage"
	"chatcoal/ws"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Warn("No .env file found, using environment variables")
	}

	app := fiber.New(fiber.Config{
		BodyLimit:   25 * 1024 * 1024,        // 25MB
		Concurrency: 512 * 1024,              // max goroutines for request handling
		ProxyHeader: fiber.HeaderXForwardedFor, // trust Traefik's forwarded IP
	})

	// HSTS — HTTP→HTTPS redirect is handled by Traefik; this header instructs
	// browsers to always use HTTPS for this origin going forward.
	app.Use(func(c *fiber.Ctx) error {
		c.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		return c.Next()
	})

	corsMiddleware := cors.New(cors.Config{
		AllowOrigins: os.Getenv("APP_ORIGINS") + "," + os.Getenv("APP_DOMAIN"),
		AllowOriginsFunc: func(origin string) bool {
			// Allow any localhost origin for the Electron desktop app,
			// which serves the production build from a dynamic port.
			return strings.HasPrefix(origin, "http://localhost:")
		},
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	})
	app.Use("/api", corsMiddleware)

	// /federation/info must be reachable from any origin (used by remote instances).
	app.Use("/federation/info", cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET",
	}))

	if os.Getenv("APP_DEBUG") == "1" {
		app.Use("/api", logger.New())
	}

	var nodeID int64 = 1
	if s := os.Getenv("APP_SNOWFLAKE_NODE"); s != "" {
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			nodeID = n
		}
	}
	models.InitSnowflake(nodeID)

	if err := database.Connect(); err != nil {
		log.Error("Unable to connect to database")
		return
	}

	services.StartUnreadBatcher()

	if err := cache.Connect(); err != nil {
		log.Warn("Redis cache init failed: ", err)
	}

	if err := storage.Connect(); err != nil {
		log.Warn("S3 storage init failed (using local uploads): ", err)
	}

	if err := middleware.InitFirebase(); err != nil {
		log.Warn("Firebase init failed (auth will not work): ", err)
	}

	if err := services.InitFederationKeys(); err != nil {
		log.Warn("Federation key init failed (federation disabled): ", err)
	}

	// WebSocket hub
	hub := ws.NewHub()
	go hub.Run()
	cache.StartSubscriber(hub)

	// Internal metrics endpoint (not behind CORS or rate limiter)
	metricsToken := os.Getenv("METRICS_TOKEN")
	app.Get("/internal/metrics", func(c *fiber.Ctx) error {
		if metricsToken == "" || c.Get("Authorization") != "Bearer "+metricsToken {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}

		var platform metrics.PlatformStats
		database.Database.Model(&models.Server{}).Count(&platform.TotalServers)
		database.Database.Model(&models.Channel{}).Count(&platform.TotalChannels)

		var voice metrics.VoiceStats
		ch, users, err := services.GetVoiceStats()
		if err != nil {
			log.Warnf("metrics: livekit stats failed: %v", err)
		} else {
			voice.ActiveChannels = ch
			voice.ActiveUsers = users
		}

		return c.JSON(metrics.Take(hub.ShardQueueDepths(), platform, voice))
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("chatcoal API")
	})

	app.Static("/uploads", "./uploads")

	// IP-based rate limit: general DoS protection for all API routes.
	// Skipped in debug mode (no Traefik proxy, all requests share 127.0.0.1).
	// Per-user rate limiting is applied inside authenticated route groups (see api.routes.go).
	if os.Getenv("APP_DEBUG") != "1" {
		app.Use("/api", limiter.New(limiter.Config{
			Max:        100,
			Expiration: 1 * time.Minute,
			KeyGenerator: func(c *fiber.Ctx) string {
				return c.IP()
			},
			LimitReached: func(c *fiber.Ctx) error {
				return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "Too many requests"})
			},
		}))
	}

	routes.SetupRoutes(app)
	routes.SetupWebSocket(app, hub)

	if err := app.Listen(":3000"); err != nil {
		log.Error("Unable to start app ", err)
		return
	}
}
