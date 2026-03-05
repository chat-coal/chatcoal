package services

import (
	"encoding/json"

	"chatcoal/database"
	"chatcoal/models"

	"github.com/gofiber/fiber/v2/log"
)

func RecordAuditLog(serverID, actorID models.Snowflake, action string, targetID *models.Snowflake, meta map[string]interface{}) {
	entry := models.AuditLog{
		ServerID: serverID,
		ActorID:  actorID,
		Action:   action,
		TargetID: targetID,
	}

	if meta != nil {
		if raw, err := json.Marshal(meta); err == nil {
			s := string(raw)
			entry.Metadata = &s
		}
	}

	if err := database.Database.Create(&entry).Error; err != nil {
		log.Warnf("audit: failed to record %s: %v", action, err)
	}
}
