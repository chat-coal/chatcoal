package controllers

import (
	"chatcoal/models"
	"chatcoal/services"

	"github.com/gofiber/fiber/v2"
)

func GetNotificationSettings(c *fiber.Ctx) error {
	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	settings, err := services.GetNotificationSettings(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch settings"})
	}

	return c.JSON(settings)
}

func UpdateNotificationSetting(c *fiber.Ctx) error {
	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var body struct {
		TargetType string          `json:"target_type" validate:"required,oneof=server channel"`
		TargetID   models.Snowflake `json:"target_id" validate:"required"`
		Muted      bool            `json:"muted"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}
	if body.TargetType == "" || body.TargetID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "target_type and target_id are required"})
	}
	if body.TargetType != "server" && body.TargetType != "channel" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "target_type must be 'server' or 'channel'"})
	}

	// Validate membership
	if body.TargetType == "server" {
		if !services.IsServerMember(user.ID, body.TargetID) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Not a member of this server"})
		}
	} else {
		ch, err := services.GetChannelByID(body.TargetID)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Channel not found"})
		}
		if !services.IsServerMember(user.ID, ch.ServerID) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Not a member of this server"})
		}
	}

	setting, err := services.UpsertNotificationSetting(user.ID, body.TargetType, body.TargetID, body.Muted)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update setting"})
	}

	return c.JSON(setting)
}
