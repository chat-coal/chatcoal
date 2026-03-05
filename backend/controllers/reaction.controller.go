package controllers

import (
	"net/url"

	"chatcoal/models"
	"chatcoal/services"

	"github.com/gofiber/fiber/v2"
)

func ToggleMessageReaction(c *fiber.Ctx) error {
	messageID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid message ID"})
	}
	rawEmoji := c.Params("emoji")
	emoji, _ := url.PathUnescape(rawEmoji)
	if emoji == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Emoji is required"})
	}
	if len([]rune(emoji)) > 50 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Emoji is too long"})
	}
	if !isValidEmoji(emoji) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid emoji"})
	}

	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Verify the message exists and user is a server member
	var msgInfo struct {
		ID        models.Snowflake
		ChannelID models.Snowflake `gorm:"column:channel_id"`
	}
	services.GetMessageForDelete(messageID, &msgInfo)
	if msgInfo.ChannelID == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Message not found"})
	}

	ch, err := services.GetChannelByID(msgInfo.ChannelID)
	if err != nil || !services.IsServerMember(user.ID, ch.ServerID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	_, err = services.ToggleReaction(messageID, user.ID, emoji)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to toggle reaction"})
	}

	reactions := services.GetReactions(messageID)

	// Broadcast to server
	broadcastEvent(ch.ServerID, "reaction_update", fiber.Map{
		"message_id": messageID,
		"reactions":  reactions,
	})

	return c.JSON(fiber.Map{"reactions": reactions})
}

func ToggleDMMessageReaction(c *fiber.Ctx) error {
	messageID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid message ID"})
	}
	rawEmoji := c.Params("emoji")
	emoji, _ := url.PathUnescape(rawEmoji)
	if emoji == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Emoji is required"})
	}
	if len([]rune(emoji)) > 50 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Emoji is too long"})
	}
	if !isValidEmoji(emoji) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid emoji"})
	}

	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Verify the DM message exists and user is a participant
	var msgInfo struct {
		ID          models.Snowflake
		DMChannelID models.Snowflake `gorm:"column:dm_channel_id"`
	}
	services.GetDMMessageForDelete(messageID, &msgInfo)
	if msgInfo.DMChannelID == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Message not found"})
	}

	if !services.IsDMParticipant(user.ID, msgInfo.DMChannelID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	_, err = services.ToggleDMReaction(messageID, user.ID, emoji)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to toggle reaction"})
	}

	reactions := services.GetDMReactions(messageID)

	// Broadcast to both DM participants
	channel, _ := services.GetDMChannelByID(msgInfo.DMChannelID)
	if channel != nil {
		broadcastDMEvent("dm_reaction_update", fiber.Map{
			"message_id": messageID,
			"reactions":  reactions,
		}, channel.User1ID, channel.User2ID)
	}

	return c.JSON(fiber.Map{"reactions": reactions})
}
