package controllers

import (
	"chatcoal/services"

	"github.com/gofiber/fiber/v2"
)

func CreateReport(c *fiber.Ctx) error {
	uid := c.Locals("firebaseUID").(string)
	user, err := services.GetUserByFirebaseUID(uid)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	var body struct {
		TargetType string `json:"target_type" validate:"required,oneof=user message dm_message forum_post"`
		TargetID   string `json:"target_id" validate:"required"`
		Reason     string `json:"reason" validate:"required,min=1,max=1000"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}
	if msg := validateBody(&body); msg != "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": msg})
	}

	targetID, err := parseSnowflakeString(body.TargetID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid target ID"})
	}

	report, err := services.CreateReport(user.ID, body.TargetType, targetID, body.Reason)
	if err != nil {
		switch err.Error() {
		case "already reported":
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "You have already reported this"})
		case "invalid target type":
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		default:
			if err.Error() == "user not found" || err.Error() == "message not found" || err.Error() == "forum post not found" {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create report"})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(report)
}

func GetReports(c *fiber.Ctx) error {
	status := c.Query("status", "")
	before := parseSnowflakeQuery(c, "before")

	reports, err := services.GetReports(status, before)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch reports"})
	}

	return c.JSON(reports)
}

func UpdateReportStatus(c *fiber.Ctx) error {
	reportID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid report ID"})
	}

	var body struct {
		Status string `json:"status" validate:"required,oneof=pending reviewed dismissed"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}
	if msg := validateBody(&body); msg != "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": msg})
	}

	report, err := services.UpdateReportStatus(reportID, body.Status)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Report not found"})
	}

	return c.JSON(report)
}
