package controllers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/url"
	"os"
	"strings"
	"time"

	"chatcoal/cache"
	"chatcoal/models"
	"chatcoal/services"

	"github.com/gofiber/fiber/v2"
)

// GetFederationInfo handles GET /federation/info.
// Returns this instance's public key and metadata for federation verification.
func GetFederationInfo(c *fiber.Ctx) error {
	domain := os.Getenv("FEDERATION_DOMAIN")
	if domain == "" {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "federation not enabled on this instance"})
	}
	pubKey := services.GetPublicKeyPEM()
	if pubKey == "" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "federation keys not initialized"})
	}
	name := os.Getenv("APP_NAME")
	if name == "" {
		name = domain
	}
	return c.JSON(fiber.Map{
		"domain":         domain,
		"name":           name,
		"public_key_pem": pubKey,
		"version":        "1",
	})
}

// BeginFederation handles POST /api/federation/begin.
// Validates the federated_id, fetches the remote instance's public key,
// stores a nonce, and returns the auth_url to redirect the user to.
func BeginFederation(c *fiber.Ctx) error {
	myDomain := os.Getenv("FEDERATION_DOMAIN")
	if myDomain == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "federation not enabled on this instance"})
	}
	if cache.Redis == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "federation requires Redis"})
	}

	var body struct {
		FederatedID string `json:"federated_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	body.FederatedID = strings.TrimSpace(body.FederatedID)

	// Validate format: must be username@domain (domain must have a dot).
	parts := strings.SplitN(body.FederatedID, "@", 2)
	if len(parts) != 2 || parts[0] == "" || !strings.Contains(parts[1], ".") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid federated ID format — expected username@domain.tld"})
	}
	remoteDomain := parts[1]

	// Check if this domain is allowed by policy.
	if allowed, err := services.CheckInstanceAllowed(remoteDomain); err != nil || !allowed {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "federation with this instance is not allowed"})
	}

	// Fetch remote instance info (also validates domain and caches public key).
	if _, err := services.FetchInstanceInfo(remoteDomain); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "could not reach remote instance"})
	}

	// Generate a cryptographically random nonce.
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal error"})
	}
	nonce := hex.EncodeToString(raw)

	// Store nonce → federated_id mapping with 5-minute TTL.
	nonceKey := "fed:challenge:" + nonce
	if err := cache.Redis.Set(context.Background(), nonceKey, body.FederatedID, 5*time.Minute).Err(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to store challenge"})
	}

	callbackURL := "https://" + myDomain + "/federation/callback"
	authURL := "https://" + remoteDomain + "/federation/authorize" +
		"?visiting=" + url.QueryEscape(myDomain) +
		"&nonce=" + url.QueryEscape(nonce) +
		"&callback=" + url.QueryEscape(callbackURL)

	return c.JSON(fiber.Map{"auth_url": authURL})
}

// AuthorizeFederation handles POST /api/federation/authorize.
// Signs an assertion JWT for the authenticated local user and returns
// the redirect URL for the visiting instance's callback.
// Requires: UnifiedAuthMiddleware (user must be a local Firebase user with a username).
func AuthorizeFederation(c *fiber.Ctx) error {
	firebaseUID, ok := c.Locals("firebaseUID").(string)
	if !ok || firebaseUID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var body struct {
		Visiting string `json:"visiting"`
		Nonce    string `json:"nonce"`
		Callback string `json:"callback"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	body.Visiting = strings.TrimSpace(body.Visiting)
	body.Callback = strings.TrimSpace(body.Callback)
	body.Nonce = strings.TrimSpace(body.Nonce)

	if body.Visiting == "" || body.Nonce == "" || body.Callback == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "visiting, nonce, and callback are required"})
	}

	// Validate callback URL: must be HTTPS and hosted on the visiting domain.
	parsedCallback, err := url.Parse(body.Callback)
	if err != nil || parsedCallback.Scheme != "https" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "callback must be an HTTPS URL"})
	}
	if parsedCallback.Hostname() != body.Visiting {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "callback hostname must match visiting domain"})
	}

	// Only local users (HomeInstance == nil) can act as identity providers.
	user, err := services.GetUserByFirebaseUID(firebaseUID)
	if err != nil || user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}
	if user.HomeInstance != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "federated users cannot authorize identity for other instances"})
	}
	if user.Username == nil || *user.Username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "you must set a username before using federation"})
	}

	// Sign the assertion.
	assertionToken, err := services.SignAssertion(
		*user.Username,
		user.DisplayName,
		user.AvatarURL,
		body.Visiting,
		body.Nonce,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to sign assertion"})
	}

	redirectURL := body.Callback + "?token=" + url.QueryEscape(assertionToken)
	return c.JSON(fiber.Map{"redirect_url": redirectURL})
}

// VerifyFederation handles POST /api/federation/verify.
// Verifies a signed assertion token from a home instance, creates a local
// federated user record, and returns a session token.
func VerifyFederation(c *fiber.Ctx) error {
	myDomain := os.Getenv("FEDERATION_DOMAIN")
	if myDomain == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "federation not enabled on this instance"})
	}
	if cache.Redis == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "federation requires Redis"})
	}

	var body struct {
		Token string `json:"token"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if body.Token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "token is required"})
	}

	// Verify the JWT assertion.
	claims, err := services.VerifyAssertion(body.Token, myDomain)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid assertion"})
	}

	// Check if the home instance is allowed by policy.
	if allowed, pErr := services.CheckInstanceAllowed(claims.HomeInstance); pErr != nil || !allowed {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "federation with this instance is not allowed"})
	}

	// Check the nonce against what was stored in /begin.
	nonceKey := "fed:challenge:" + claims.Nonce
	storedFID, err := cache.Redis.Get(context.Background(), nonceKey).Result()
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid or expired challenge nonce"})
	}

	// Verify the identity matches what /begin was started for.
	expectedFID := claims.Subject + "@" + claims.HomeInstance
	if storedFID != expectedFID {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "identity mismatch"})
	}

	// Consume the nonce (one-time use).
	cache.Redis.Del(context.Background(), nonceKey)

	// Create or update the local federated user record.
	federatedUID := "fed:" + expectedFID
	user, err := services.GetOrCreateFederatedUser(federatedUID, claims.HomeInstance, claims.DisplayName, claims.AvatarURL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create user record"})
	}

	// Issue a federation session token.
	sessionToken, err := services.CreateFederationSession(federatedUID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create session"})
	}

	return c.JSON(fiber.Map{
		"session_token": sessionToken,
		"user":          user,
	})
}

// --- Channel Federation (User-facing) ---

// EnableChannelFederation handles POST /api/servers/:id/channels/:channelId/federation.
func EnableChannelFederation(c *fiber.Ctx) error {
	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	serverID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid server ID"})
	}
	if user == nil || !services.HasPermission(user.ID, serverID, services.PermManageChannels) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Missing permission"})
	}

	channelID, err := parseSnowflakeParam(c, "channelId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid channel ID"})
	}

	ch, err := services.EnableChannelFederation(channelID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(ch)
}

// DisableChannelFederation handles DELETE /api/servers/:id/channels/:channelId/federation.
func DisableChannelFederation(c *fiber.Ctx) error {
	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	serverID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid server ID"})
	}
	if user == nil || !services.HasPermission(user.ID, serverID, services.PermManageChannels) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Missing permission"})
	}

	channelID, err := parseSnowflakeParam(c, "channelId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid channel ID"})
	}

	if err := services.DisableChannelFederation(channelID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// LinkRemoteChannel handles POST /api/servers/:id/channels/:channelId/federation/link.
func LinkRemoteChannel(c *fiber.Ctx) error {
	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	serverID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid server ID"})
	}
	if user == nil || !services.HasPermission(user.ID, serverID, services.PermManageChannels) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Missing permission"})
	}

	channelID, err := parseSnowflakeParam(c, "channelId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid channel ID"})
	}

	var body struct {
		RemoteAddress string `json:"remote_address"` // format: domain/fed/federation_id
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Parse remote_address: "domain/fed/federation_id"
	parts := strings.SplitN(body.RemoteAddress, "/fed/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid remote address format — expected domain/fed/federation_id"})
	}
	remoteDomain := parts[0]
	remoteFedID := parts[1]

	// Check policy
	if allowed, pErr := services.CheckInstanceAllowed(remoteDomain); pErr != nil || !allowed {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "federation with this instance is not allowed"})
	}

	link, err := services.CreateChannelLink(channelID, remoteDomain, remoteFedID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create link"})
	}
	return c.Status(fiber.StatusCreated).JSON(link)
}

// UnlinkRemoteChannel handles DELETE /api/servers/:id/channels/:channelId/federation/link/:linkId.
func UnlinkRemoteChannel(c *fiber.Ctx) error {
	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	serverID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid server ID"})
	}
	if user == nil || !services.HasPermission(user.ID, serverID, services.PermManageChannels) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Missing permission"})
	}

	linkID, err := parseSnowflakeParam(c, "linkId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid link ID"})
	}

	if err := services.DeleteChannelLink(linkID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete link"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// --- Server-to-Server Channel Federation Endpoints ---

// GetFederatedChannelInfo handles GET /federation/channels/:federationId/info.
func GetFederatedChannelInfo(c *fiber.Ctx) error {
	fedID := c.Params("federationId")
	ch, err := services.GetChannelByFederationID(fedID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "channel not found"})
	}
	domain := os.Getenv("FEDERATION_DOMAIN")
	appName := os.Getenv("APP_NAME")
	if appName == "" {
		appName = domain
	}
	return c.JSON(fiber.Map{
		"federation_id": fedID,
		"channel_name":  ch.Name,
		"server_name":   appName,
	})
}

// ReceiveFederatedMessage handles POST /federation/channels/:federationId/messages.
func ReceiveFederatedMessage(c *fiber.Ctx) error {
	fedID := c.Params("federationId")

	var body struct {
		Token string `json:"token"`
	}
	if err := c.BodyParser(&body); err != nil || body.Token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "token is required"})
	}

	// Verify the JWT
	claims, err := services.VerifyChannelMessage(body.Token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token"})
	}

	// Policy check
	if allowed, pErr := services.CheckInstanceAllowed(claims.Issuer); pErr != nil || !allowed {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "federation with this instance is not allowed"})
	}

	// Find local channel
	ch, err := services.GetChannelByFederationID(fedID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "channel not found"})
	}

	// Get or create the federated user
	fedUID := claims.AuthorFedUID
	if fedUID == "" {
		fedUID = "fed:unknown@" + claims.Issuer
	}
	user, err := services.GetOrCreateFederatedUser(fedUID, claims.Issuer, claims.AuthorName, claims.AuthorAvatar)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create federated user"})
	}

	// Create local message
	msgType := claims.MessageType
	if msgType == "" {
		msgType = "user"
	}
	var message *models.Message
	if msgType == "user" {
		message, err = services.CreateMessage(claims.Content, ch.ID, ch.ServerID, user.ID, "", "", 0, 0, 0, nil, nil)
	} else {
		message, err = services.CreateSystemMessage(msgType, claims.Content, ch.ID, ch.ServerID, user.ID)
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create message"})
	}

	// Broadcast via WebSocket locally
	broadcastEvent(ch.ServerID, "message", message)

	// If this is the hub, relay to other subscribers (exclude the sender's domain)
	links, _ := services.GetLinksForFederationID(fedID)
	for _, link := range links {
		if link.RemoteDomain == claims.Issuer {
			continue
		}
		go services.RelayMessageToRemote(link.RemoteDomain, link.RemoteFederationID, body.Token)
	}

	return c.SendStatus(fiber.StatusCreated)
}
