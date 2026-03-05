package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"chatcoal/cache"

	"github.com/gofiber/fiber/v2/log"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	client   *minio.Client
	bucket   string
	endpoint string
)

// Connect initialises the S3-compatible client using Railway env vars.
// Non-fatal: if any env var is missing or connection fails the app
// falls back to local ./uploads/ storage.
func Connect() error {
	rawEndpoint := os.Getenv("S3_ENDPOINT")
	accessKey := os.Getenv("S3_ACCESS_KEY")
	secretKey := os.Getenv("S3_SECRET_KEY")
	bucket = os.Getenv("S3_BUCKET")

	// Strip protocol prefix — minio-go adds it based on the Secure flag
	endpoint = strings.TrimPrefix(strings.TrimPrefix(rawEndpoint, "https://"), "http://")

	log.Debugf("[S3] endpoint=%q (raw=%q) bucket=%q accessKey=%q (secretKey len=%d)", endpoint, rawEndpoint, bucket, accessKey, len(secretKey))

	if endpoint == "" || accessKey == "" || secretKey == "" || bucket == "" {
		log.Warn("S3 env vars not set — using local file storage")
		return nil
	}

	useSSL := true
	if os.Getenv("S3_USE_SSL") == "false" {
		useSSL = false
	}
	log.Debugf("[S3] useSSL=%v", useSSL)

	c, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Errorf("[S3] client init failed: %v", err)
		return fmt.Errorf("s3 client init: %w", err)
	}

	log.Debug("[S3] client created, checking bucket exists...")

	// Verify bucket exists
	exists, err := c.BucketExists(context.Background(), bucket)
	if err != nil {
		log.Errorf("[S3] bucket check failed: %v", err)
		return fmt.Errorf("s3 bucket check: %w", err)
	}
	if !exists {
		log.Errorf("[S3] bucket %q does not exist", bucket)
		return fmt.Errorf("s3 bucket %q does not exist", bucket)
	}

	client = c
	log.Infof("[S3] connected: %s/%s", endpoint, bucket)
	return nil
}

// Available returns true when S3 is configured and connected.
func Available() bool {
	return client != nil
}

// Upload stores a multipart file in S3 and returns its public URL.
// key is the object key (e.g. "avatars/1_17000000.png").
func Upload(ctx context.Context, file *multipart.FileHeader, key string, contentType string) (string, error) {
	log.Debugf("[S3] Upload: key=%q contentType=%q size=%d", key, contentType, file.Size)

	src, err := file.Open()
	if err != nil {
		log.Errorf("[S3] Upload: failed to open file: %v", err)
		return "", err
	}
	defer src.Close()

	return UploadReader(ctx, src, file.Size, key, contentType)
}

// UploadReader stores data from an io.Reader in S3 and returns a relative
// serving path (/api/files/<key>) that the file-serving endpoint will resolve
// to a presigned URL.
func UploadReader(ctx context.Context, r io.Reader, size int64, key string, contentType string) (string, error) {
	log.Debugf("[S3] PutObject: bucket=%q key=%q size=%d contentType=%q", bucket, key, size, contentType)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	info, err := client.PutObject(ctx, bucket, key, r, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		log.Errorf("[S3] PutObject failed: %v", err)
		return "", fmt.Errorf("s3 upload: %w", err)
	}

	servePath := "/api/files/" + key
	log.Debugf("[S3] PutObject success: etag=%q size=%d servePath=%q", info.ETag, info.Size, servePath)
	return servePath, nil
}

// Delete removes an object from S3 by key and invalidates its presigned URL cache entry.
func Delete(ctx context.Context, key string) error {
	log.Debugf("[S3] Delete: key=%q", key)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err := client.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		log.Errorf("[S3] Delete failed: key=%q err=%v", key, err)
		return err
	}
	cache.InvalidatePresignedURL(key)
	return nil
}

// DeleteFileByURL removes the file associated with a URL (S3 or local).
// Errors are logged but not returned — this is best-effort cleanup.
func DeleteFileByURL(fileURL string) {
	if fileURL == "" {
		return
	}
	if Available() {
		key := KeyFromURL(fileURL)
		if key == "" {
			log.Warnf("[S3] DeleteFileByURL: could not extract key from URL %q — file not deleted", fileURL)
			return
		}
		if err := Delete(context.Background(), key); err != nil {
			log.Errorf("[S3] DeleteFileByURL: failed to delete key=%q: %v", key, err)
		}
	} else if strings.HasPrefix(fileURL, "/uploads/") {
		// Resolve the path and verify it stays within ./uploads/ to prevent path traversal.
		resolved := filepath.Clean("." + fileURL)
		absUploads, err := filepath.Abs("./uploads")
		if err != nil {
			log.Errorf("[local] DeleteFileByURL: could not resolve uploads dir: %v", err)
			return
		}
		absResolved, err := filepath.Abs(resolved)
		if err != nil || !strings.HasPrefix(absResolved, absUploads+string(os.PathSeparator)) {
			log.Warnf("[local] DeleteFileByURL: path traversal attempt rejected: %q", fileURL)
			return
		}
		if err := os.Remove(absResolved); err != nil {
			log.Errorf("[local] DeleteFileByURL: failed to remove %q: %v", fileURL, err)
		}
	}
}

// DeleteFileByURLErr removes the file associated with a URL and returns any error.
// Use this when the caller needs to abort a DB delete on storage failure.
func DeleteFileByURLErr(fileURL string) error {
	if fileURL == "" {
		return nil
	}
	if Available() {
		key := KeyFromURL(fileURL)
		if key == "" {
			return fmt.Errorf("could not extract key from URL %q", fileURL)
		}
		return Delete(context.Background(), key)
	} else if strings.HasPrefix(fileURL, "/uploads/") {
		resolved := filepath.Clean("." + fileURL)
		absUploads, err := filepath.Abs("./uploads")
		if err != nil {
			return fmt.Errorf("could not resolve uploads dir: %w", err)
		}
		absResolved, err := filepath.Abs(resolved)
		if err != nil || !strings.HasPrefix(absResolved, absUploads+string(os.PathSeparator)) {
			return fmt.Errorf("path traversal attempt rejected: %q", fileURL)
		}
		return os.Remove(absResolved)
	}
	return nil
}

// KeyFromURL extracts the S3 object key from a stored URL.
// Handles both the new format (/api/files/<key>) and legacy direct S3 URLs.
func KeyFromURL(fileURL string) string {
	const apiPrefix = "/api/files/"
	if strings.HasPrefix(fileURL, apiPrefix) {
		return strings.TrimPrefix(fileURL, apiPrefix)
	}
	// Legacy: direct S3 URL
	prefix := PublicURL("")
	key := strings.TrimPrefix(fileURL, prefix)
	if key != fileURL && key != "" {
		return key
	}
	return ""
}

// PresignedGetURL generates a temporary presigned URL for downloading an S3 object.
func PresignedGetURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	reqParams := make(url.Values)
	presigned, err := client.PresignedGetObject(ctx, bucket, key, expiry, reqParams)
	if err != nil {
		return "", fmt.Errorf("s3 presign: %w", err)
	}
	return presigned.String(), nil
}

// PublicURL returns the public URL for an S3 object (legacy, used for key extraction).
func PublicURL(key string) string {
	proto := "https"
	if os.Getenv("S3_USE_SSL") == "false" {
		proto = "http"
	}
	return fmt.Sprintf("%s://%s/%s/%s", proto, endpoint, bucket, key)
}
