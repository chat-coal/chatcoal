package controllers

import (
	"chatcoal/services"

	"github.com/gofiber/fiber/v2"
)

func CreateInvite(c *fiber.Ctx) error {
	serverID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid server ID"})
	}

	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil || !services.IsServerMember(user.ID, serverID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	if !services.HasPermission(user.ID, serverID, services.PermManageInvites) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	var body struct {
		MaxUses   int `json:"max_uses"`
		ExpiresIn int `json:"expires_in"`
	}
	c.BodyParser(&body)

	invite, err := services.CreateInvite(serverID, user.ID, body.MaxUses, body.ExpiresIn)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create invite"})
	}

	return c.Status(fiber.StatusCreated).JSON(invite)
}

func GetInvites(c *fiber.Ctx) error {
	serverID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid server ID"})
	}

	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil || !services.IsServerMember(user.ID, serverID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	if !services.HasPermission(user.ID, serverID, services.PermManageInvites) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	before := parseSnowflakeQuery(c, "before")
	invites, err := services.GetInvitesByServerID(serverID, before)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch invites"})
	}

	return c.JSON(invites)
}

func DeleteInvite(c *fiber.Ctx) error {
	serverID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid server ID"})
	}

	inviteID, err := parseSnowflakeParam(c, "inviteId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid invite ID"})
	}

	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil || !services.IsServerMember(user.ID, serverID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	if !services.HasPermission(user.ID, serverID, services.PermManageInvites) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
	}

	if err := services.DeleteInvite(inviteID, serverID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete invite"})
	}

	services.RecordAuditLog(serverID, user.ID, "invite_delete", &inviteID, nil)

	return c.SendStatus(fiber.StatusNoContent)
}

func ResolveInvite(c *fiber.Ctx) error {
	code := c.Params("code")
	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invite code is required"})
	}

	invite, err := services.GetInviteByCode(code)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Invalid or expired invite"})
	}

	return c.JSON(fiber.Map{
		"code":   invite.Code,
		"server": invite.Server,
	})
}
