package controllers

import (
	"chatcoal/services"

	"github.com/gofiber/fiber/v2"
)

func RegisterDeviceToken(c *fiber.Ctx) error {
	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var body struct {
		Token    string `json:"token"`
		Platform string `json:"platform"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}
	if body.Token == "" || body.Platform == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "token and platform are required"})
	}
	if body.Platform != "ios" && body.Platform != "android" && body.Platform != "web" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "platform must be 'ios', 'android', or 'web'"})
	}

	if err := services.RegisterDeviceToken(user.ID, body.Token, body.Platform); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to register token"})
	}

	return c.JSON(fiber.Map{"status": "ok"})
}

func UnregisterDeviceToken(c *fiber.Ctx) error {
	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var body struct {
		Token string `json:"token"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}
	if body.Token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "token is required"})
	}

	if err := services.UnregisterDeviceToken(body.Token); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to unregister token"})
	}

	return c.JSON(fiber.Map{"status": "ok"})
}
