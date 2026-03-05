package services

import (
	"chatcoal/database"
	"chatcoal/models"
	"errors"

	"gorm.io/gorm"
)

// GetDefaultPolicy returns the instance's default federation policy ("open" or "closed").
func GetDefaultPolicy() string {
	var config models.InstanceConfig
	if err := database.Database.First(&config).Error; err != nil {
		return "open"
	}
	if config.DefaultPolicy == "" {
		return "open"
	}
	return config.DefaultPolicy
}

// SetDefaultPolicy updates the instance's default federation policy.
func SetDefaultPolicy(policy string) error {
	if policy != "open" && policy != "closed" {
		return errors.New("policy must be 'open' or 'closed'")
	}
	return database.Database.Model(&models.InstanceConfig{}).Where("id = 1").Update("default_policy", policy).Error
}

// GetInstancePolicies returns all configured instance policies.
func GetInstancePolicies() ([]models.InstancePolicy, error) {
	var policies []models.InstancePolicy
	err := database.Database.Order("created_at DESC").Find(&policies).Error
	return policies, err
}

// AddInstancePolicy upserts an allow/block policy for a domain.
func AddInstancePolicy(domain, policy, note string, createdBy models.Snowflake) (*models.InstancePolicy, error) {
	if policy != "allow" && policy != "block" {
		return nil, errors.New("policy must be 'allow' or 'block'")
	}
	if domain == "" {
		return nil, errors.New("domain is required")
	}

	var existing models.InstancePolicy
	err := database.Database.Where("domain = ?", domain).First(&existing).Error
	if err == nil {
		// Update existing
		database.Database.Model(&existing).Updates(map[string]interface{}{
			"policy":     policy,
			"note":       note,
			"created_by": createdBy,
		})
		existing.Policy = policy
		existing.Note = note
		existing.CreatedBy = createdBy
		return &existing, nil
	}

	p := models.InstancePolicy{
		Domain:    domain,
		Policy:    policy,
		Note:      note,
		CreatedBy: createdBy,
	}
	if err := database.Database.Create(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

// RemoveInstancePolicy deletes the policy for the given domain.
func RemoveInstancePolicy(domain string) error {
	result := database.Database.Where("domain = ?", domain).Delete(&models.InstancePolicy{})
	if result.RowsAffected == 0 {
		return errors.New("policy not found")
	}
	return result.Error
}

// CheckInstanceAllowed returns true if the domain is allowed to federate.
// Explicit policy wins; otherwise the default policy applies.
func CheckInstanceAllowed(domain string) (bool, error) {
	var policy models.InstancePolicy
	err := database.Database.Where("domain = ?", domain).First(&policy).Error
	if err == nil {
		return policy.Policy == "allow", nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}
	// No explicit policy — fall back to default.
	def := GetDefaultPolicy()
	return def == "open", nil
}
