package controllers

import (
	"chatcoal/models"
	"chatcoal/services"

	"github.com/gofiber/fiber/v2"
)

func GetVoiceToken(c *fiber.Ctx) error {
	serverID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid server ID"})
	}

	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil || !services.IsServerMember(user.ID, serverID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	var body struct {
		ChannelID models.Snowflake `json:"channel_id"`
	}
	if err := c.BodyParser(&body); err != nil || body.ChannelID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "channel_id is required"})
	}

	ch, err := services.GetChannelByID(body.ChannelID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Channel not found"})
	}
	if ch.ServerID != serverID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Channel does not belong to this server"})
	}
	if ch.Type != "audio" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Channel is not an audio channel"})
	}

	token, err := services.GenerateLiveKitToken(user.ID, body.ChannelID, user.DisplayName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	return c.JSON(fiber.Map{"token": token})
}
