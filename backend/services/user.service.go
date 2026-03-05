package services

import (
	"chatcoal/cache"
	"chatcoal/database"
	"chatcoal/models"
	"fmt"
	"strings"
)

func GetOrCreateUser(firebaseUID string, displayName string, avatarURL string, isAnonymous bool, emailVerified bool) (*models.User, error) {
	var user models.User
	if err := database.Database.Where("firebase_uid = ?", firebaseUID).First(&user).Error; err == nil {
		dirty := false
		// Clear is_anonymous when a guest upgrades to a permanent account.
		if !isAnonymous && user.IsAnonymous {
			database.Database.Model(&user).Update("is_anonymous", false)
			user.IsAnonymous = false
			dirty = true
		}
		// Sync email_verified with the current token claim.
		if emailVerified && !user.EmailVerified {
			database.Database.Model(&user).Update("email_verified", true)
			user.EmailVerified = true
			dirty = true
		} else if !emailVerified && user.EmailVerified && !isAnonymous {
			// Downgrade when a formerly-anonymous user links an unverified email.
			database.Database.Model(&user).Update("email_verified", false)
			user.EmailVerified = false
			dirty = true
		}
		// Auto-assign a display name for existing anon users that don't have one
		// (e.g. accounts created before this logic was added).
		if isAnonymous && user.DisplayName == "" {
			name := randomAnonName()
			database.Database.Model(&user).Update("display_name", name)
			user.DisplayName = name
			dirty = true
		}
		if dirty {
			cache.InvalidateUserByFirebaseUID(firebaseUID)
		}
		return &user, nil
	}

	if isAnonymous && displayName == "" {
		displayName = randomAnonName()
	}

	user = models.User{
		FirebaseUID:   firebaseUID,
		DisplayName:   displayName,
		AvatarURL:     avatarURL,
		Status:        "online",
		IsAnonymous:   isAnonymous,
		EmailVerified: emailVerified,
	}
	if err := database.Database.Create(&user).Error; err != nil {
		// Race condition: another request inserted the same user concurrently.
		// Use a fresh struct so GORM v2 doesn't append "AND id = <new-snowflake>"
		// to the WHERE clause (it does that when the struct has a non-zero PK).
		if strings.Contains(err.Error(), "1062") || strings.Contains(err.Error(), "Duplicate entry") {
			var existing models.User
			if err2 := database.Database.Where("firebase_uid = ?", firebaseUID).First(&existing).Error; err2 == nil {
				return &existing, nil
			}
		}
		return nil, err
	}
	cache.InvalidateUserByFirebaseUID(firebaseUID)
	return &user, nil
}

func GetUserByFirebaseUID(firebaseUID string) (*models.User, error) {
	return cache.GetUser(firebaseUID)
}

func GetUserByID(id models.Snowflake) (*models.User, error) {
	var user models.User
	if err := database.Database.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func CheckUsernameAvailable(username string, excludeUserID models.Snowflake) (bool, error) {
	var count int64
	query := database.Database.Model(&models.User{}).Where("username = ?", username)
	if excludeUserID != 0 {
		query = query.Where("id != ?", excludeUserID)
	}
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count == 0, nil
}

func UpdateUser(userID models.Snowflake, displayName string, username string, avatarURL string, clearAvatar bool, status string) (*models.User, error) {
	updates := map[string]interface{}{}
	if displayName != "" {
		updates["display_name"] = displayName
	}
	if username != "" {
		updates["username"] = username
	}
	if clearAvatar {
		updates["avatar_url"] = ""
	} else if avatarURL != "" {
		updates["avatar_url"] = avatarURL
	}
	if status != "" {
		updates["status"] = status
	}

	if len(updates) > 0 {
		if err := database.Database.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	cache.InvalidateUser(userID)
	user, err := GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	cache.SetUser(user)
	return user, nil
}

func DeleteUser(userID models.Snowflake, firebaseUID string) error {
	// Delete any server where this user is the only remaining admin/owner.
	CleanupUserServers(userID)

	updates := map[string]interface{}{
		// Tombstone the firebase_uid so this record is never matched again if
		// the same UID is reused or the client session lingers after deletion.
		"firebase_uid": fmt.Sprintf("deleted:%d", userID),
		"display_name": "Deleted User",
		"username":     nil,
		"avatar_url":   "",
		"status":       "offline",
	}
	if err := database.Database.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		return err
	}
	cache.InvalidateUser(userID)
	cache.InvalidateUserByFirebaseUID(firebaseUID)
	return nil
}

// GetOrCreateFederatedUser creates or updates a remote (federated) user record.
// federatedUID is the synthetic firebase_uid, e.g. "fed:alice@instance-a.com".
func GetOrCreateFederatedUser(federatedUID, homeInstance, displayName, avatarURL string) (*models.User, error) {
	var user models.User
	if err := database.Database.Where("firebase_uid = ?", federatedUID).First(&user).Error; err == nil {
		// Sync profile from home instance on every login.
		database.Database.Model(&user).Updates(map[string]interface{}{
			"display_name": displayName,
			"avatar_url":   avatarURL,
		})
		cache.InvalidateUser(user.ID)
		return &user, nil
	}

	user = models.User{
		FirebaseUID:  federatedUID,
		HomeInstance: &homeInstance,
		DisplayName:  displayName,
		AvatarURL:    avatarURL,
		Status:       "online",
		IsAnonymous:  false,
	}
	if err := database.Database.Create(&user).Error; err != nil {
		if strings.Contains(err.Error(), "1062") || strings.Contains(err.Error(), "Duplicate entry") {
			var existing models.User
			if err2 := database.Database.Where("firebase_uid = ?", federatedUID).First(&existing).Error; err2 == nil {
				return &existing, nil
			}
		}
		return nil, err
	}
	cache.InvalidateUserByFirebaseUID(federatedUID)
	return &user, nil
}

func IsUserDeleted(userID models.Snowflake) bool {
	user, err := GetUserByID(userID)
	if err != nil {
		return false
	}
	return strings.HasPrefix(user.FirebaseUID, "deleted:")
}

func UpdateUserStatus(userID models.Snowflake, status string) error {
	err := database.Database.Model(&models.User{}).Where("id = ?", userID).Update("status", status).Error
	if err == nil {
		cache.InvalidateUser(userID)
	}
	return err
}
