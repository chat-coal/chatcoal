package cache

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"chatcoal/models"
)

// presenceTTL is the safety-net expiry for online presence keys.
// SetOffline deletes the key immediately on clean disconnect; the TTL
// ensures stale entries are cleaned up if the server crashes.
const presenceTTL = 10 * time.Minute

func presenceUserKey(userID models.Snowflake) string {
	return fmt.Sprintf("presence:user:%d", userID)
}

// SetOnline marks a user as connected with a TTL safety net.
func SetOnline(userID models.Snowflake) {
	if Redis == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	Redis.Set(ctx, presenceUserKey(userID), 1, presenceTTL)
}

// SetOffline marks a user as disconnected by deleting their presence key.
func SetOffline(userID models.Snowflake) {
	if Redis == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	Redis.Del(ctx, presenceUserKey(userID))
}

// GetOnlineUserIDs returns the set of all online user IDs from Redis.
// Returns nil if Redis is unavailable (caller should fall back to local state).
func GetOnlineUserIDs() map[models.Snowflake]bool {
	if Redis == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	result := make(map[models.Snowflake]bool)
	var cursor uint64
	for {
		keys, next, err := Redis.Scan(ctx, cursor, "presence:user:*", 100).Result()
		if err != nil {
			if len(result) == 0 {
				return nil
			}
			break
		}
		for _, key := range keys {
			idStr := strings.TrimPrefix(key, "presence:user:")
			if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
				result[models.Snowflake(id)] = true
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return result
}
