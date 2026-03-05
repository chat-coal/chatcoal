package cache

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"chatcoal/models"
)

const voiceTTL = 30 * time.Minute

func voiceChannelKey(channelID models.Snowflake) string {
	return fmt.Sprintf("voice:channel:%d", channelID)
}

func voiceServerKey(serverID models.Snowflake) string {
	return fmt.Sprintf("voice:server:%d", serverID)
}

func voiceUserKey(userID models.Snowflake) string {
	return fmt.Sprintf("voice:user:%d", userID)
}

// JoinVoiceChannel records a user joining a voice channel in Redis.
func JoinVoiceChannel(serverID, channelID, userID models.Snowflake) {
	if Redis == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	pipe := Redis.Pipeline()
	chKey := voiceChannelKey(channelID)
	srvKey := voiceServerKey(serverID)
	usrKey := voiceUserKey(userID)
	pipe.SAdd(ctx, chKey, fmt.Sprintf("%d", userID))
	pipe.Expire(ctx, chKey, voiceTTL)
	pipe.SAdd(ctx, srvKey, fmt.Sprintf("%d", channelID))
	pipe.Expire(ctx, srvKey, voiceTTL)
	pipe.Set(ctx, usrKey, fmt.Sprintf("%d:%d", serverID, channelID), voiceTTL)
	pipe.Exec(ctx)
}

// LeaveVoiceChannel removes a user from their voice channel in Redis.
func LeaveVoiceChannel(userID models.Snowflake) {
	if Redis == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Look up which channel the user is in
	val, err := Redis.Get(ctx, voiceUserKey(userID)).Result()
	if err != nil {
		return
	}

	parts := strings.SplitN(val, ":", 2)
	if len(parts) != 2 {
		return
	}
	serverID, _ := strconv.ParseInt(parts[0], 10, 64)
	channelID, _ := strconv.ParseInt(parts[1], 10, 64)

	chID := models.Snowflake(channelID)
	srvID := models.Snowflake(serverID)

	pipe := Redis.Pipeline()
	pipe.SRem(ctx, voiceChannelKey(chID), fmt.Sprintf("%d", userID))
	pipe.Del(ctx, voiceUserKey(userID))
	pipe.Exec(ctx)

	// Clean up empty channel from server set
	remaining, err := Redis.SCard(ctx, voiceChannelKey(chID)).Result()
	if err == nil && remaining == 0 {
		Redis.SRem(ctx, voiceServerKey(srvID), fmt.Sprintf("%d", channelID))
		Redis.Del(ctx, voiceChannelKey(chID))
	}
}

// GetVoiceStatesFromRedis returns channelID -> []userID for a server.
// Returns nil if Redis is unavailable (caller should fall back to local state).
func GetVoiceStatesFromRedis(serverID models.Snowflake) map[models.Snowflake][]models.Snowflake {
	if Redis == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	channelStrs, err := Redis.SMembers(ctx, voiceServerKey(serverID)).Result()
	if err != nil {
		return nil
	}

	result := make(map[models.Snowflake][]models.Snowflake)
	for _, chStr := range channelStrs {
		channelID, err := strconv.ParseInt(chStr, 10, 64)
		if err != nil {
			continue
		}
		chID := models.Snowflake(channelID)

		userStrs, err := Redis.SMembers(ctx, voiceChannelKey(chID)).Result()
		if err != nil {
			continue
		}

		var userIDs []models.Snowflake
		for _, uStr := range userStrs {
			if uid, err := strconv.ParseInt(uStr, 10, 64); err == nil {
				userIDs = append(userIDs, models.Snowflake(uid))
			}
		}
		if len(userIDs) > 0 {
			result[chID] = userIDs
		}
	}
	return result
}
