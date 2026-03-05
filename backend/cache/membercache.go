package cache

import (
	"context"
	"fmt"
	"time"

	"chatcoal/database"
	"chatcoal/metrics"
	"chatcoal/models"
)

const memberTTL = 10 * time.Minute

func memberKey(userID, serverID models.Snowflake) string {
	return fmt.Sprintf("member:%d:%d", userID, serverID)
}

// IsMember checks membership, using Redis cache when available.
// Uses singleflight to coalesce concurrent cache misses for the same key.
func IsMember(userID, serverID models.Snowflake) bool {
	ctx := context.Background()

	if Redis != nil {
		val, err := Redis.Get(ctx, memberKey(userID, serverID)).Result()
		if err == nil {
			metrics.CacheHits.Member.Add(1)
			return val != ""
		}
		metrics.CacheMisses.Member.Add(1)
	}

	// Cache miss — coalesce concurrent lookups via singleflight
	v, err, _ := sfGroup.Do(memberKey(userID, serverID), func() (interface{}, error) {
		var member models.ServerMember
		err := database.Database.
			Where("user_id = ? AND server_id = ?", userID, serverID).
			First(&member).Error

		if err != nil {
			return false, nil
		}

		// Write-through with role
		if Redis != nil {
			role := member.Role
			if role == "" {
				role = "member"
			}
			logRedisErr(Redis.Set(ctx, memberKey(userID, serverID), role, memberTTL).Err(), "IsMember")
		}

		return true, nil
	})
	if err != nil {
		return false
	}
	return v.(bool)
}

// SetMember explicitly marks a user as a server member in cache (backward compat).
func SetMember(userID, serverID models.Snowflake) {
	SetMemberWithRole(userID, serverID, "member")
}

// SetMemberWithRole marks a user as a server member with a specific role in cache.
func SetMemberWithRole(userID, serverID models.Snowflake, role string) {
	if Redis == nil {
		return
	}
	if role == "" {
		role = "member"
	}
	logRedisErr(Redis.Set(context.Background(), memberKey(userID, serverID), role, memberTTL).Err(), "SetMemberWithRole")
}

// GetMemberRole returns the cached role for a user in a server.
// Falls back to DB on cache miss. Returns "" if not a member.
// Uses singleflight to coalesce concurrent cache misses for the same key.
func GetMemberRole(userID, serverID models.Snowflake) string {
	ctx := context.Background()

	if Redis != nil {
		val, err := Redis.Get(ctx, memberKey(userID, serverID)).Result()
		if err == nil {
			metrics.CacheHits.Member.Add(1)
			// Backward compat: "1" means "member"
			if val == "1" {
				return "member"
			}
			return val
		}
		metrics.CacheMisses.Member.Add(1)
	}

	// Cache miss — coalesce concurrent lookups via singleflight
	v, err, _ := sfGroup.Do("role:"+memberKey(userID, serverID), func() (interface{}, error) {
		var member models.ServerMember
		err := database.Database.
			Where("user_id = ? AND server_id = ?", userID, serverID).
			First(&member).Error

		if err != nil {
			return "", nil
		}

		role := member.Role
		if role == "" {
			role = "member"
		}

		// Write-through
		if Redis != nil {
			logRedisErr(Redis.Set(ctx, memberKey(userID, serverID), role, memberTTL).Err(), "GetMemberRole")
		}

		return role, nil
	})
	if err != nil {
		return ""
	}
	return v.(string)
}

// InvalidateMember removes a single membership cache entry.
func InvalidateMember(userID, serverID models.Snowflake) {
	if Redis == nil {
		return
	}
	Redis.Del(context.Background(), memberKey(userID, serverID))
}

// InvalidateServerMembers removes all membership cache entries for a server.
func InvalidateServerMembers(serverID models.Snowflake) {
	if Redis == nil {
		return
	}
	ctx := context.Background()
	pattern := fmt.Sprintf("member:*:%d", serverID)

	var cursor uint64
	for {
		keys, next, err := Redis.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			break
		}
		if len(keys) > 0 {
			Redis.Del(ctx, keys...)
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
}
