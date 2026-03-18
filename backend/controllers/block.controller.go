package controllers

import (
	"chatcoal/models"
	"chatcoal/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func BlockUser(c *fiber.Ctx) error {
	uid := c.Locals("firebaseUID").(string)
	user, err := services.GetUserByFirebaseUID(uid)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	targetID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	block, err := services.BlockUser(user.ID, models.Snowflake(targetID))
	if err != nil {
		if err.Error() == "cannot block yourself" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Already blocked"})
	}

	return c.Status(fiber.StatusCreated).JSON(block)
}

func UnblockUser(c *fiber.Ctx) error {
	uid := c.Locals("firebaseUID").(string)
	user, err := services.GetUserByFirebaseUID(uid)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	targetID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	if err := services.UnblockUser(user.ID, models.Snowflake(targetID)); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Not blocked"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func GetBlockedUsers(c *fiber.Ctx) error {
	uid := c.Locals("firebaseUID").(string)
	user, err := services.GetUserByFirebaseUID(uid)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	blocks, err := services.GetBlockedUsers(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch blocked users"})
	}

	return c.JSON(blocks)
}
