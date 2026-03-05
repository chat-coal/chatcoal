package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"chatcoal/database"
	"chatcoal/metrics"
	"chatcoal/models"
)

const userTTL = 5 * time.Minute

func userFBKey(firebaseUID string) string {
	return fmt.Sprintf("user:fbuid:%s", firebaseUID)
}

func userIDKey(userID models.Snowflake) string {
	return fmt.Sprintf("user:id:%d", userID)
}

// GetUser looks up a user by Firebase UID, checking Redis first.
// Uses singleflight to coalesce concurrent cache misses for the same key.
func GetUser(firebaseUID string) (*models.User, error) {
	ctx := context.Background()

	if Redis != nil {
		data, err := Redis.Get(ctx, userFBKey(firebaseUID)).Bytes()
		if err == nil {
			var user models.User
			if json.Unmarshal(data, &user) == nil {
				metrics.CacheHits.User.Add(1)
				return &user, nil
			}
		}
		metrics.CacheMisses.User.Add(1)
	}

	// Cache miss — coalesce concurrent lookups via singleflight
	v, err, _ := sfGroup.Do("user:"+firebaseUID, func() (interface{}, error) {
		var user models.User
		if err := database.Database.Where("firebase_uid = ?", firebaseUID).First(&user).Error; err != nil {
			return nil, err
		}

		// Write-through
		if Redis != nil {
			if data, err := json.Marshal(user); err == nil {
				logRedisErr(Redis.Set(ctx, userFBKey(firebaseUID), data, userTTL).Err(), "GetUser fbuid")
				logRedisErr(Redis.Set(ctx, userIDKey(user.ID), firebaseUID, userTTL).Err(), "GetUser userid")
			}
		}

		return &user, nil
	})
	if err != nil {
		return nil, err
	}
	return v.(*models.User), nil
}

// SetUser writes a user to the cache (write-through after DB fetch).
func SetUser(user *models.User) {
	if Redis == nil || user == nil {
		return
	}
	ctx := context.Background()
	if data, err := json.Marshal(user); err == nil {
		logRedisErr(Redis.Set(ctx, userFBKey(user.FirebaseUID), data, userTTL).Err(), "SetUser fbuid")
		logRedisErr(Redis.Set(ctx, userIDKey(user.ID), user.FirebaseUID, userTTL).Err(), "SetUser userid")
	}
}

// InvalidateUser removes cached user data by user ID.
func InvalidateUser(userID models.Snowflake) {
	if Redis == nil {
		return
	}
	ctx := context.Background()

	// Look up firebase UID from reverse key
	fbUID, err := Redis.Get(ctx, userIDKey(userID)).Result()
	if err == nil {
		Redis.Del(ctx, userFBKey(fbUID))
	}
	Redis.Del(ctx, userIDKey(userID))
}

// InvalidateUserByFirebaseUID removes cached user data by Firebase UID.
func InvalidateUserByFirebaseUID(firebaseUID string) {
	if Redis == nil {
		return
	}
	ctx := context.Background()
	Redis.Del(ctx, userFBKey(firebaseUID))
}
