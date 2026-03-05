package cache

import (
	"context"
	"time"

	"chatcoal/metrics"
)

// presignedURLTTL is shorter than the 1-hour presigned URL expiry to avoid
// serving expired URLs cached right at the boundary.
const presignedURLTTL = 55 * time.Minute

func presignedURLKey(key string) string {
	return "presigned:" + key
}

// GetPresignedURL returns a cached presigned URL for the given S3 key, or "" on miss.
func GetPresignedURL(key string) string {
	if Redis == nil {
		return ""
	}
	url, err := Redis.Get(context.Background(), presignedURLKey(key)).Result()
	if err != nil {
		metrics.CacheMisses.Presigned.Add(1)
		return ""
	}
	metrics.CacheHits.Presigned.Add(1)
	return url
}

// SetPresignedURL caches a presigned URL for the given S3 key.
func SetPresignedURL(key string, url string) {
	if Redis == nil {
		return
	}
	logRedisErr(Redis.Set(context.Background(), presignedURLKey(key), url, presignedURLTTL).Err(), "SetPresignedURL")
}

// InvalidatePresignedURL removes a cached presigned URL for the given S3 key.
// Call this whenever an S3 object is deleted so stale presigned URLs are not served.
func InvalidatePresignedURL(key string) {
	if Redis == nil {
		return
	}
	logRedisErr(Redis.Del(context.Background(), presignedURLKey(key)).Err(), "InvalidatePresignedURL")
}
