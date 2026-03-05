package routes

import (
	"chatcoal/controllers"
	"chatcoal/middleware"
	"time"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	SetupFederationRoutes(app)

	api := app.Group("/api")

	// Per-user rate limiter applied to all authenticated groups (60 req/min per user).
	// This runs after UnifiedAuthMiddleware sets firebaseUID in locals.
	perUser := middleware.PerUserRateLimiter(60, 1*time.Minute)

	// File serving (presigned S3 redirect) — no auth required; access control relies on
	// cryptographically random UUID v4 keys (unguessable) in the filename.
	// Browsers never send Authorization headers for image/CSS loads, so auth here would break avatars.
	api.Get("/files/*", controllers.ServeFile)

	// Auth routes
	auth := api.Group("/auth", middleware.UnifiedAuthMiddleware(), perUser)
	auth.Post("/login", controllers.Login)
	auth.Get("/me", controllers.GetMe)
	auth.Put("/profile", controllers.UpdateProfile)
	auth.Delete("/account", controllers.DeleteAccount)
	auth.Get("/check-username", middleware.PerUserRateLimiter(10, 1*time.Minute), controllers.CheckUsername)

	// Server routes
	servers := api.Group("/servers", middleware.UnifiedAuthMiddleware(), perUser)
	servers.Get("/", controllers.GetServers)
	servers.Post("/", controllers.CreateServer)
	servers.Post("/join", controllers.JoinServer)
	servers.Get("/public", controllers.GetPublicServers)
	servers.Post("/:id/join", controllers.JoinPublicServer)
	servers.Put("/:id", controllers.UpdateServer)
	servers.Delete("/:id", controllers.DeleteServer)
	servers.Delete("/:id/leave", controllers.LeaveServer)
	servers.Get("/:id/members", controllers.GetServerMembers)
	servers.Patch("/:id/members/:userId/role", controllers.UpdateMemberRole)
	servers.Delete("/:id/members/:userId", controllers.KickMember)
	servers.Post("/:id/bans/:userId", controllers.BanMember)
	servers.Get("/:id/bans", controllers.GetServerBans)
	servers.Delete("/:id/bans/:userId", controllers.UnbanUser)
	servers.Post("/:id/transfer", controllers.TransferOwnership)
	servers.Get("/:id/invites", controllers.GetInvites)
	servers.Post("/:id/invites", controllers.CreateInvite)
	servers.Delete("/:id/invites/:inviteId", controllers.DeleteInvite)
	servers.Put("/:id/channels/reorder", controllers.ReorderChannels)
	servers.Get("/:id/channels", controllers.GetChannels)
	servers.Post("/:id/channels", controllers.CreateChannel)
	servers.Get("/:id/search", middleware.PerUserRateLimiter(10, 1*time.Minute), controllers.SearchMessages)
	servers.Get("/:id/voice-states", controllers.GetVoiceStates)
	servers.Post("/:id/voice-token", controllers.GetVoiceToken)

	// Invite routes (standalone)
	invites := api.Group("/invites", middleware.UnifiedAuthMiddleware(), perUser)
	invites.Get("/:code", controllers.ResolveInvite)

	// Channel routes
	channels := api.Group("/channels", middleware.UnifiedAuthMiddleware(), perUser)
	channels.Put("/:id", controllers.UpdateChannel)
	channels.Delete("/:id", controllers.DeleteChannel)
	channels.Get("/:id/messages", controllers.GetMessages)
	channels.Post("/:id/messages", controllers.SendMessage)
	channels.Get("/:id/posts", controllers.GetForumPosts)
	channels.Post("/:id/posts", controllers.CreateForumPost)

	// Channel read state
	channels.Put("/:id/read", controllers.MarkChannelAsRead)

	// Channel pins
	channels.Get("/:id/pins", controllers.GetPinnedMessages)

	// Message routes
	messages := api.Group("/messages", middleware.UnifiedAuthMiddleware(), perUser)
	messages.Put("/:id", controllers.EditMessage)
	messages.Delete("/:id", controllers.DeleteMessageHandler)
	messages.Put("/:id/reactions/:emoji", controllers.ToggleMessageReaction)
	messages.Put("/:id/pin", controllers.PinMessageHandler)
	messages.Delete("/:id/pin", controllers.UnpinMessageHandler)

	// User profiles
	users := api.Group("/users", middleware.UnifiedAuthMiddleware(), perUser)
	users.Get("/:id", controllers.GetUserProfile)

	// Forum post routes
	forumPosts := api.Group("/forum-posts", middleware.UnifiedAuthMiddleware(), perUser)
	forumPosts.Get("/:id", controllers.GetForumPostByIDHandler)
	forumPosts.Put("/:id", controllers.EditForumPostHandler)
	forumPosts.Delete("/:id", controllers.DeleteForumPostHandler)
	forumPosts.Get("/:id/messages", controllers.GetForumPostMessages)
	forumPosts.Post("/:id/messages", controllers.SendForumPostMessage)

	// DM routes
	dms := api.Group("/dms", middleware.UnifiedAuthMiddleware(), perUser)
	dms.Get("/", controllers.GetDMChannels)
	dms.Post("/", controllers.CreateOrGetDMChannel)
	dms.Get("/:id/messages", controllers.GetDMMessages)
	dms.Post("/:id/messages", controllers.SendDMMessage)
	dms.Put("/:id/read", controllers.MarkDMAsRead)

	// DM message routes
	dmMessages := api.Group("/dm-messages", middleware.UnifiedAuthMiddleware(), perUser)
	dmMessages.Put("/:id", controllers.EditDMMessage)
	dmMessages.Delete("/:id", controllers.DeleteDMMessage)
	dmMessages.Put("/:id/reactions/:emoji", controllers.ToggleDMMessageReaction)

	// Unread counts
	unread := api.Group("/unread", middleware.UnifiedAuthMiddleware(), perUser)
	unread.Get("/", controllers.GetUnreadCounts)

	// Notification settings (mute)
	notifSettings := api.Group("/notification-settings", middleware.UnifiedAuthMiddleware(), perUser)
	notifSettings.Get("/", controllers.GetNotificationSettings)
	notifSettings.Put("/", controllers.UpdateNotificationSetting)

	// Admin routes (site admins only)
	admin := api.Group("/admin", middleware.UnifiedAuthMiddleware(), middleware.SiteAdminMiddleware(), perUser)
	admin.Get("/federation/policy", controllers.GetFederationPolicy)
	admin.Put("/federation/policy", controllers.UpdateFederationPolicy)
	admin.Post("/federation/instances", controllers.AddInstancePolicy)
	admin.Delete("/federation/instances/:domain", controllers.RemoveInstancePolicy)

	// Channel federation management (requires PermManageChannels, enforced in controller)
	servers.Post("/:id/channels/:channelId/federation", controllers.EnableChannelFederation)
	servers.Delete("/:id/channels/:channelId/federation", controllers.DisableChannelFederation)
	servers.Post("/:id/channels/:channelId/federation/link", controllers.LinkRemoteChannel)
	servers.Delete("/:id/channels/:channelId/federation/link/:linkId", controllers.UnlinkRemoteChannel)
}
