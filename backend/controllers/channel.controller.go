package controllers

import (
	"chatcoal/services"
	"chatcoal/ws"

	"github.com/gofiber/fiber/v2"
)

func GetChannels(c *fiber.Ctx) error {
	serverID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid server ID"})
	}

	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil || !services.IsServerMember(user.ID, serverID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	channels, err := services.GetChannelsByServerID(serverID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch channels"})
	}

	return c.JSON(channels)
}

func CreateChannel(c *fiber.Ctx) error {
	serverID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid server ID"})
	}

	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil || !services.IsServerMember(user.ID, serverID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	if !services.HasPermission(user.ID, serverID, services.PermManageChannels) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	var body struct {
		Name  string `json:"name" validate:"required,min=1,max=100"`
		Type  string `json:"type"`
		Topic string `json:"topic" validate:"max=1024"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}
	if msg := validateBody(&body); msg != "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": msg})
	}

	if body.Type == "" {
		body.Type = "text"
	}
	if body.Type != "text" && body.Type != "audio" && body.Type != "forum" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Type must be 'text', 'audio', or 'forum'"})
	}

	channel, err := services.CreateChannel(body.Name, serverID, body.Type, body.Topic)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create channel"})
	}

	broadcastEvent(serverID, "channel_create", channel)
	return c.Status(fiber.StatusCreated).JSON(channel)
}

func UpdateChannel(c *fiber.Ctx) error {
	channelID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid channel ID"})
	}

	ch, err := services.GetChannelByID(channelID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Channel not found"})
	}

	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil || !services.IsServerMember(user.ID, ch.ServerID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	if !services.HasPermission(user.ID, ch.ServerID, services.PermManageChannels) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	var body struct {
		Name  string  `json:"name" validate:"max=100"`
		Topic *string `json:"topic" validate:"max=1024"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}
	if msg := validateBody(&body); msg != "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": msg})
	}

	channel, err := services.UpdateChannel(channelID, body.Name, body.Topic)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update channel"})
	}

	return c.JSON(channel)
}

func GetVoiceStates(c *fiber.Ctx) error {
	serverID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid server ID"})
	}

	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil || !services.IsServerMember(user.ID, serverID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	states := ws.GetHub().GetVoiceStates(serverID)
	return c.JSON(states)
}

func DeleteChannel(c *fiber.Ctx) error {
	channelID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid channel ID"})
	}

	ch, err := services.GetChannelByID(channelID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Channel not found"})
	}

	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil || !services.IsServerMember(user.ID, ch.ServerID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	if !services.HasPermission(user.ID, ch.ServerID, services.PermManageChannels) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	if err := services.DeleteChannel(channelID, ch.ServerID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete channel"})
	}

	broadcastEvent(ch.ServerID, "channel_delete", fiber.Map{"id": channelID})
	return c.SendStatus(fiber.StatusNoContent)
}
