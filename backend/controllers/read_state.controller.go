package controllers

import (
	"chatcoal/models"
	"chatcoal/services"

	"github.com/gofiber/fiber/v2"
)

// MarkChannelAsRead marks a server channel as read for the current user.
func MarkChannelAsRead(c *fiber.Ctx) error {
	channelID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid channel ID"})
	}

	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var body struct {
		MessageID models.Snowflake `json:"message_id"`
	}
	if err := c.BodyParser(&body); err != nil || body.MessageID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "message_id is required"})
	}

	if err := services.UpdateReadState(user.ID, "server", channelID, body.MessageID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update read state"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// MarkDMAsRead marks a DM channel as read for the current user.
func MarkDMAsRead(c *fiber.Ctx) error {
	dmChannelID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid DM channel ID"})
	}

	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	if !services.IsDMParticipant(user.ID, dmChannelID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	var body struct {
		MessageID models.Snowflake `json:"message_id"`
	}
	if err := c.BodyParser(&body); err != nil || body.MessageID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "message_id is required"})
	}

	if err := services.UpdateReadState(user.ID, "dm", dmChannelID, body.MessageID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update read state"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetUnreadCounts returns all unread counts for the current user.
func GetUnreadCounts(c *fiber.Ctx) error {
	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	counts, err := services.GetAllUnreadCounts(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch unread counts"})
	}

	if counts == nil {
		counts = []services.UnreadCount{}
	}

	return c.JSON(counts)
}
