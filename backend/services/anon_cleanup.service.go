package services

import (
	"chatcoal/cache"
	"chatcoal/database"
	"chatcoal/models"
	"chatcoal/storage"
	"time"

	"github.com/gofiber/fiber/v2/log"
)

// DeleteFirebaseUserFunc is set by server.go to break the import cycle
// between services and middleware.
var DeleteFirebaseUserFunc func(uid string) error

// StartAnonCleanup launches a background goroutine that periodically deletes
// anonymous accounts that have been inactive for more than 7 days.
func StartAnonCleanup() {
	go func() {
		// Run once on startup after a short delay, then every hour.
		time.Sleep(30 * time.Second)
		cleanupInactiveAnons()

		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			cleanupInactiveAnons()
		}
	}()
}

func cleanupInactiveAnons() {
	cutoff := time.Now().Add(-7 * 24 * time.Hour)

	var users []models.User
	if err := database.Database.
		Where("is_anonymous = ? AND updated_at < ?", true, cutoff).
		Find(&users).Error; err != nil {
		log.Warnf("anon cleanup: query failed: %v", err)
		return
	}

	if len(users) == 0 {
		return
	}

	log.Infof("anon cleanup: found %d inactive anonymous accounts to delete", len(users))

	for _, u := range users {
		// Delete avatar from storage if present
		if u.AvatarURL != "" {
			storage.DeleteFileByURL(u.AvatarURL)
		}

		// Anonymize the DB record (same as manual account deletion)
		if err := DeleteUser(u.ID, u.FirebaseUID); err != nil {
			log.Warnf("anon cleanup: failed to delete user %d: %v", u.ID, err)
			continue
		}

		// Delete the Firebase user so the UID is freed
		if DeleteFirebaseUserFunc != nil {
			if err := DeleteFirebaseUserFunc(u.FirebaseUID); err != nil {
				log.Warnf("anon cleanup: failed to delete firebase user %s: %v", u.FirebaseUID, err)
			}
		}

		cache.InvalidateUser(u.ID)
		cache.InvalidateUserByFirebaseUID(u.FirebaseUID)
	}

	log.Infof("anon cleanup: finished processing %d accounts", len(users))
}
