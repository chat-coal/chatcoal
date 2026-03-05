package controllers

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"chatcoal/storage"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	_ "golang.org/x/image/webp"
)

// uploadFile stores a file either in S3 (if available) or locally in ./uploads/.
// prefix is the sub-path (e.g. "" for messages, "avatars" for avatars, "server-icons" for icons).
// Returns the public URL of the uploaded file.
func uploadFile(c *fiber.Ctx, file *multipart.FileHeader, prefix string, savedName string) (string, error) {
	ext := strings.ToLower(filepath.Ext(file.Filename))

	if storage.Available() {
		key := savedName
		if prefix != "" {
			key = prefix + "/" + savedName
		}
		contentType := mimeFromExt(ext)
		url, err := storage.Upload(c.Context(), file, key, contentType)
		if err != nil {
			return "", err
		}
		return url, nil
	}

	// Fallback: local storage
	localDir := "./uploads"
	if prefix != "" {
		localDir = "./uploads/" + prefix
	}
	os.MkdirAll(localDir, 0755)

	savePath := filepath.Join(localDir, savedName)
	if err := c.SaveFile(file, savePath); err != nil {
		return "", err
	}

	if prefix != "" {
		return fmt.Sprintf("/uploads/%s/%s", prefix, savedName), nil
	}
	return "/uploads/" + savedName, nil
}

// generateFileName creates a unique filename with the given parts.
func generateFileName(parts ...interface{}) string {
	return fmt.Sprint(parts...)
}

func mimeFromExt(ext string) string {
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".pdf":
		return "application/pdf"
	case ".txt":
		return "text/plain"
	case ".zip":
		return "application/zip"
	default:
		return "application/octet-stream"
	}
}

// checkMagicBytes reads the first 512 bytes of a file and uses
// http.DetectContentType to verify the actual content type is allowed.
// This prevents spoofed extensions (e.g. an HTML file renamed to .jpg).
func checkMagicBytes(file *multipart.FileHeader) error {
	f, err := file.Open()
	if err != nil {
		return fmt.Errorf("could not open file")
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return fmt.Errorf("could not read file")
	}
	if n == 0 {
		return fmt.Errorf("file is empty")
	}

	detected := http.DetectContentType(buf[:n])
	// Strip charset suffix, e.g. "text/plain; charset=utf-8" → "text/plain"
	base := strings.TrimSpace(strings.SplitN(detected, ";", 2)[0])
	switch base {
	case "image/jpeg", "image/png", "image/gif", "image/webp",
		"application/pdf", "application/zip", "text/plain":
		return nil
	}
	return fmt.Errorf("file content not allowed")
}

// getImageDimensions reads just the header of a multipart file to extract width/height.
// Returns (0, 0) for non-image files or on any error.
func getImageDimensions(file *multipart.FileHeader) (int, int) {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
	default:
		return 0, 0
	}
	f, err := file.Open()
	if err != nil {
		return 0, 0
	}
	defer f.Close()
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return 0, 0
	}
	return cfg.Width, cfg.Height
}

// stamp returns a UUID v4 string for unique file names.
func stamp() string {
	return uuid.New().String()
}
