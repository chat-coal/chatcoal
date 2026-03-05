package controllers

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"chatcoal/models"
	"chatcoal/services"
	"chatcoal/storage"
	"chatcoal/ws"

	"github.com/gofiber/fiber/v2"
)

func GetServers(c *fiber.Ctx) error {
	uid := c.Locals("firebaseUID").(string)
	user, err := services.GetUserByFirebaseUID(uid)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	servers, err := services.GetServersByUserID(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch servers"})
	}

	return c.JSON(servers)
}

func CreateServer(c *fiber.Ctx) error {
	uid := c.Locals("firebaseUID").(string)
	user, err := services.GetUserByFirebaseUID(uid)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	if user.IsRestricted() {
		count, _ := services.CountUserServers(user.ID)
		if count >= 2 {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Verify your email to join more servers"})
		}
	}

	var body struct {
		Name     string `json:"name" validate:"required,min=1,max=100"`
		IsPublic bool   `json:"is_public"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}
	if msg := validateBody(&body); msg != "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": msg})
	}

	server, err := services.CreateServer(body.Name, user.ID, body.IsPublic)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create server"})
	}

	return c.Status(fiber.StatusCreated).JSON(server)
}

func JoinServer(c *fiber.Ctx) error {
	uid := c.Locals("firebaseUID").(string)
	user, err := services.GetUserByFirebaseUID(uid)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	var body struct {
		InviteCode string `json:"invite_code" validate:"required,max=20"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}
	if msg := validateBody(&body); msg != "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": msg})
	}

	// Check if already a member (try new invite table first, then legacy)
	if invite, _ := services.GetInviteByCode(body.InviteCode); invite != nil {
		if services.IsServerMember(user.ID, invite.Server.ID) {
			return c.JSON(&invite.Server)
		}
	} else if server, _ := services.GetServerByInviteCode(body.InviteCode); server != nil {
		if services.IsServerMember(user.ID, server.ID) {
			return c.JSON(server)
		}
	}

	if user.IsRestricted() {
		count, _ := services.CountUserServers(user.ID)
		if count >= 2 {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Verify your email to join more servers"})
		}
	}

	server, err := services.JoinServer(user.ID, body.InviteCode)
	if err != nil {
		if err.Error() == "you are banned from this server" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Invalid invite code"})
	}

	if newMember, err := services.GetServerMember(user.ID, server.ID); err == nil {
		broadcastEvent(server.ID, "member_join", newMember)
	}

	if msg := services.PostSystemAnnouncement(server, "join", "joined the server", user.ID); msg != nil {
		broadcastEvent(server.ID, "message", msg)
	}

	return c.JSON(server)
}

func LeaveServer(c *fiber.Ctx) error {
	uid := c.Locals("firebaseUID").(string)
	user, err := services.GetUserByFirebaseUID(uid)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	serverID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid server ID"})
	}

	server, srvErr := services.GetServerByID(serverID)

	member, err := services.LeaveServer(user.ID, serverID)
	if err != nil {
		if err.Error() == "owner cannot leave server" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	broadcastEvent(serverID, "member_leave", member)

	if srvErr == nil {
		if msg := services.PostSystemAnnouncement(server, "leave", "left the server", user.ID); msg != nil {
			broadcastEvent(serverID, "message", msg)
		}
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func GetServerMembers(c *fiber.Ctx) error {
	serverID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid server ID"})
	}

	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil || !services.IsServerMember(user.ID, serverID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	after := parseSnowflakeQuery(c, "after")
	members, err := services.GetServerMembers(serverID, after)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch members"})
	}

	// Overlay live presence from Hub, respecting invisible status
	onlineUsers := ws.GetHub().GetOnlineUserIDs()
	for i := range members {
		if onlineUsers[members[i].UserID] && members[i].User.Status != "invisible" {
			members[i].User.Status = "online"
		} else {
			members[i].User.Status = "offline"
		}
	}

	return c.JSON(members)
}

func UpdateServer(c *fiber.Ctx) error {
	uid := c.Locals("firebaseUID").(string)
	user, err := services.GetUserByFirebaseUID(uid)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	serverID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid server ID"})
	}

	// Fetch old icon URL before update so we can clean up storage
	oldIconURL := ""
	if existingServer, _ := services.GetServerByID(serverID); existingServer != nil {
		oldIconURL = existingServer.IconURL
	}

	name := c.FormValue("name")
	iconURL := ""
	clearIcon := c.FormValue("clear_icon") == "true"

	file, err := c.FormFile("icon")
	if err == nil && file != nil {
		ext := strings.ToLower(filepath.Ext(file.Filename))
		allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
		if !allowed[ext] {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid file type. Allowed: jpg, png, gif, webp"})
		}
		if err := checkMagicBytes(file); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid file type. Allowed: jpg, png, gif, webp"})
		}

		filename := fmt.Sprintf("%d_%s%s", serverID, stamp(), ext)
		url, uploadErr := uploadFile(c, file, "server-icons", filename)
		if uploadErr != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save icon"})
		}
		iconURL = url
		clearIcon = false // uploading a new file takes precedence
	}

	var isPublicPtr *bool
	if isPublicStr := c.FormValue("is_public"); isPublicStr != "" {
		v := isPublicStr == "true"
		isPublicPtr = &v
	}

	var showJoinLeavePtr *bool
	if sjl := c.FormValue("show_join_leave"); sjl != "" {
		v := sjl == "true"
		showJoinLeavePtr = &v
	}

	var systemChannelIDPtr *models.Snowflake
	if scID := c.FormValue("system_channel_id"); scID != "" {
		id, parseErr := parseSnowflakeString(scID)
		if parseErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid system_channel_id"})
		}
		systemChannelIDPtr = &id
	}

	server, err := services.UpdateServer(serverID, user.ID, name, iconURL, clearIcon, isPublicPtr, showJoinLeavePtr, systemChannelIDPtr)
	if err != nil {
		if err.Error() == "only the owner can update the server" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update server"})
	}

	// Delete old icon from storage after successful update
	if (iconURL != "" || clearIcon) && oldIconURL != "" {
		storage.DeleteFileByURL(oldIconURL)
	}

	broadcastEvent(serverID, "server_update", server)

	return c.JSON(server)
}

func UpdateMemberRole(c *fiber.Ctx) error {
	uid := c.Locals("firebaseUID").(string)
	user, err := services.GetUserByFirebaseUID(uid)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	serverID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid server ID"})
	}

	targetID, err := parseSnowflakeParam(c, "userId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	var body struct {
		Role string `json:"role" validate:"required,oneof=admin member"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}
	if msg := validateBody(&body); msg != "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": msg})
	}

	member, err := services.UpdateMemberRole(user.ID, targetID, serverID, body.Role)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}

	broadcastEvent(serverID, "member_update", member)
	return c.JSON(member)
}

func KickMember(c *fiber.Ctx) error {
	uid := c.Locals("firebaseUID").(string)
	user, err := services.GetUserByFirebaseUID(uid)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	serverID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid server ID"})
	}

	targetID, err := parseSnowflakeParam(c, "userId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	server, _ := services.GetServerByID(serverID)

	member, err := services.KickMember(user.ID, targetID, serverID)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}

	services.RecordAuditLog(serverID, user.ID, "member_kick", &targetID, nil)

	broadcastEvent(serverID, "member_leave", member)

	if server != nil {
		if msg := services.PostSystemAnnouncement(server, "leave", "was removed from the server", targetID); msg != nil {
			broadcastEvent(serverID, "message", msg)
		}
	}

	// Notify the kicked user directly so their client can react
	serverName := ""
	if server != nil {
		serverName = server.Name
	}
	if payload, err := json.Marshal(map[string]interface{}{
		"type": "kicked",
		"data": map[string]interface{}{
			"server_id":   serverID,
			"server_name": serverName,
		},
	}); err == nil {
		ws.GetHub().SendToUserGlobal(targetID, payload)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func BanMember(c *fiber.Ctx) error {
	uid := c.Locals("firebaseUID").(string)
	user, err := services.GetUserByFirebaseUID(uid)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	serverID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid server ID"})
	}

	targetID, err := parseSnowflakeParam(c, "userId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	var body struct {
		Reason *string `json:"reason"`
	}
	c.BodyParser(&body)

	server, _ := services.GetServerByID(serverID)

	member, err := services.BanMember(user.ID, targetID, serverID, body.Reason)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}

	services.RecordAuditLog(serverID, user.ID, "member_ban", &targetID, nil)

	broadcastEvent(serverID, "member_leave", member)

	if server != nil {
		if msg := services.PostSystemAnnouncement(server, "leave", "was banned from the server", targetID); msg != nil {
			broadcastEvent(serverID, "message", msg)
		}
	}

	// Notify the banned user directly so their client can react
	serverName := ""
	if server != nil {
		serverName = server.Name
	}
	if payload, err := json.Marshal(map[string]interface{}{
		"type": "banned",
		"data": map[string]interface{}{
			"server_id":   serverID,
			"server_name": serverName,
		},
	}); err == nil {
		ws.GetHub().SendToUserGlobal(targetID, payload)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func GetServerBans(c *fiber.Ctx) error {
	uid := c.Locals("firebaseUID").(string)
	user, err := services.GetUserByFirebaseUID(uid)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	serverID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid server ID"})
	}

	if !services.HasPermission(user.ID, serverID, services.PermBanMembers) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	bans, err := services.GetServerBans(serverID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch bans"})
	}

	return c.JSON(bans)
}

func UnbanUser(c *fiber.Ctx) error {
	uid := c.Locals("firebaseUID").(string)
	user, err := services.GetUserByFirebaseUID(uid)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	serverID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid server ID"})
	}

	targetID, err := parseSnowflakeParam(c, "userId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	if !services.HasPermission(user.ID, serverID, services.PermBanMembers) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	if err := services.UnbanUser(serverID, targetID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to unban user"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func TransferOwnership(c *fiber.Ctx) error {
	uid := c.Locals("firebaseUID").(string)
	user, err := services.GetUserByFirebaseUID(uid)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	serverID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid server ID"})
	}

	var body struct {
		UserID models.Snowflake `json:"user_id"`
	}
	if err := c.BodyParser(&body); err != nil || body.UserID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user_id is required"})
	}

	if err := services.TransferOwnership(user.ID, body.UserID, serverID); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}

	services.RecordAuditLog(serverID, user.ID, "ownership_transfer", &body.UserID, nil)

	server, _ := services.GetServerByID(serverID)
	broadcastEvent(serverID, "server_update", server)
	return c.JSON(server)
}

func DeleteServer(c *fiber.Ctx) error {
	uid := c.Locals("firebaseUID").(string)
	user, err := services.GetUserByFirebaseUID(uid)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	serverID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid server ID"})
	}

	if err := services.DeleteServer(serverID, user.ID); err != nil {
		if err.Error() == "only the owner can delete the server" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete server"})
	}

	services.RecordAuditLog(serverID, user.ID, "server_delete", nil, nil)

	broadcastEvent(serverID, "server_delete", fiber.Map{"id": serverID})

	return c.SendStatus(fiber.StatusNoContent)
}

func GetPublicServers(c *fiber.Ctx) error {
	uid := c.Locals("firebaseUID").(string)
	user, err := services.GetUserByFirebaseUID(uid)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	servers, hasMore, err := services.GetPublicServers(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch public servers"})
	}

	return c.JSON(fiber.Map{"servers": servers, "has_more": hasMore})
}

func JoinPublicServer(c *fiber.Ctx) error {
	uid := c.Locals("firebaseUID").(string)
	user, err := services.GetUserByFirebaseUID(uid)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	serverID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid server ID"})
	}

	if services.IsServerMember(user.ID, serverID) {
		server, _ := services.GetServerByID(serverID)
		return c.JSON(server)
	}

	if user.IsRestricted() {
		count, _ := services.CountUserServers(user.ID)
		if count >= 2 {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Verify your email to join more servers"})
		}
	}

	server, err := services.JoinPublicServer(user.ID, serverID)
	if err != nil {
		if err.Error() == "you are banned from this server" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
		if err.Error() == "server is not public" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Server not found"})
	}

	if newMember, err := services.GetServerMember(user.ID, server.ID); err == nil {
		broadcastEvent(server.ID, "member_join", newMember)
	}

	if msg := services.PostSystemAnnouncement(server, "join", "joined the server", user.ID); msg != nil {
		broadcastEvent(server.ID, "message", msg)
	}

	return c.JSON(server)
}
