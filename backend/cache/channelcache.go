package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"chatcoal/metrics"
	"chatcoal/models"
)

const serverChannelsTTL = 5 * time.Minute

func serverChannelsKey(serverID models.Snowflake) string {
	return fmt.Sprintf("channels:server:%d", serverID)
}

// GetServerChannels returns the cached channel list for a server, or nil on cache miss.
func GetServerChannels(serverID models.Snowflake) []models.Channel {
	if Redis == nil {
		return nil
	}
	data, err := Redis.Get(context.Background(), serverChannelsKey(serverID)).Bytes()
	if err != nil {
		metrics.CacheMisses.ServerChannels.Add(1)
		return nil
	}
	var channels []models.Channel
	if json.Unmarshal(data, &channels) != nil {
		return nil
	}
	metrics.CacheHits.ServerChannels.Add(1)
	return channels
}

// SetServerChannels writes the channel list for a server to cache.
func SetServerChannels(serverID models.Snowflake, channels []models.Channel) {
	if Redis == nil {
		return
	}
	data, err := json.Marshal(channels)
	if err != nil {
		return
	}
	Redis.Set(context.Background(), serverChannelsKey(serverID), data, serverChannelsTTL)
}

// InvalidateServerChannels removes a server's channel list from cache.
func InvalidateServerChannels(serverID models.Snowflake) {
	if Redis == nil {
		return
	}
	Redis.Del(context.Background(), serverChannelsKey(serverID))
}
