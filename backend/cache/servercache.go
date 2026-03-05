package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"chatcoal/metrics"
	"chatcoal/models"
)

const userServersTTL = 5 * time.Minute

func userServersKey(userID models.Snowflake) string {
	return fmt.Sprintf("servers:user:%d", userID)
}

// GetUserServers returns the cached server list for a user, or nil on cache miss.
func GetUserServers(userID models.Snowflake) []models.Server {
	if Redis == nil {
		return nil
	}
	data, err := Redis.Get(context.Background(), userServersKey(userID)).Bytes()
	if err != nil {
		metrics.CacheMisses.UserServers.Add(1)
		return nil
	}
	var servers []models.Server
	if json.Unmarshal(data, &servers) != nil {
		return nil
	}
	metrics.CacheHits.UserServers.Add(1)
	return servers
}

// SetUserServers writes the server list for a user to cache.
func SetUserServers(userID models.Snowflake, servers []models.Server) {
	if Redis == nil {
		return
	}
	data, err := json.Marshal(servers)
	if err != nil {
		return
	}
	Redis.Set(context.Background(), userServersKey(userID), data, userServersTTL)
}

// InvalidateUserServers removes a user's server list from cache.
func InvalidateUserServers(userID models.Snowflake) {
	if Redis == nil {
		return
	}
	Redis.Del(context.Background(), userServersKey(userID))
}
