package controllers

import (
	"chatcoal/services"

	"github.com/gofiber/fiber/v2"
)

// GetFederationPolicy returns the default policy and all instance-specific policies.
func GetFederationPolicy(c *fiber.Ctx) error {
	policies, err := services.GetInstancePolicies()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch policies"})
	}
	return c.JSON(fiber.Map{
		"default_policy": services.GetDefaultPolicy(),
		"instances":      policies,
	})
}

// UpdateFederationPolicy updates the default federation policy (open/closed).
func UpdateFederationPolicy(c *fiber.Ctx) error {
	var body struct {
		Policy string `json:"policy"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := services.SetDefaultPolicy(body.Policy); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"default_policy": body.Policy})
}

// AddInstancePolicy adds or updates an allow/block policy for a specific domain.
func AddInstancePolicy(c *fiber.Ctx) error {
	uid := c.Locals("firebaseUID").(string)
	user, err := services.GetUserByFirebaseUID(uid)
	if err != nil || user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "user not found"})
	}

	var body struct {
		Domain string `json:"domain"`
		Policy string `json:"policy"`
		Note   string `json:"note"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	policy, err := services.AddInstancePolicy(body.Domain, body.Policy, body.Note, user.ID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(policy)
}

// RemoveInstancePolicy removes the policy for a specific domain.
func RemoveInstancePolicy(c *fiber.Ctx) error {
	domain := c.Params("domain")
	if domain == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "domain is required"})
	}
	if err := services.RemoveInstancePolicy(domain); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
