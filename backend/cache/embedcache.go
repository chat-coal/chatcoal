package cache

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"
)

const (
	ogCacheTTL       = 1 * time.Hour
	ogDomainRLWindow = 1 * time.Minute
	ogDomainRLMax    = 10
)

func ogCacheKey(url string) string {
	h := sha256.Sum256([]byte(url))
	return fmt.Sprintf("og:%x", h)
}

func GetOGCache(url string) (string, bool) {
	if Redis == nil {
		return "", false
	}
	val, err := Redis.Get(context.Background(), ogCacheKey(url)).Result()
	if err != nil {
		return "", false
	}
	return val, true
}

func SetOGCache(url string, jsonData string) {
	if Redis == nil {
		return
	}
	logRedisErr(Redis.Set(context.Background(), ogCacheKey(url), jsonData, ogCacheTTL).Err(), "SetOGCache")
}

// OGDomainRateLimitOK returns true if the domain has not exceeded 10 fetches per minute.
func OGDomainRateLimitOK(domain string) bool {
	if Redis == nil {
		return true
	}
	key := fmt.Sprintf("ogrl:%s", domain)
	count, err := Redis.Incr(context.Background(), key).Result()
	if err != nil {
		return true
	}
	if count == 1 {
		Redis.Expire(context.Background(), key, ogDomainRLWindow)
	}
	return count <= int64(ogDomainRLMax)
}
