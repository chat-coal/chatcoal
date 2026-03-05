package controllers

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"chatcoal/middleware"
	"chatcoal/models"
	"chatcoal/services"
	"chatcoal/storage"
	"chatcoal/ws"

	"firebase.google.com/go/v4/auth"
	"github.com/gofiber/fiber/v2"
)

var usernameRe = regexp.MustCompile(`^[a-zA-Z0-9_]{2,32}$`)

func Login(c *fiber.Ctx) error {
	uid := c.Locals("firebaseUID").(string)

	// Extract anonymous status and email verification from Firebase claims.
	// Display name is intentionally left empty for new users so they
	// are directed through onboarding to choose their own username.
	// Avatar is not imported from social providers — all users start
	// without an avatar and may upload one through profile settings.
	isAnonymous := false
	emailVerified := true // social/anon default to true
	if token, ok := c.Locals("firebaseToken").(*auth.Token); ok && token != nil {
		isAnonymous = token.Firebase.SignInProvider == "anonymous"
		if !isAnonymous {
			if ev, ok := token.Claims["email_verified"].(bool); ok {
				emailVerified = ev
			}
		}
	} else if anon, ok := c.Locals("firebaseIsAnonymous").(bool); ok {
		isAnonymous = anon
		if !isAnonymous {
			if ev, ok := c.Locals("firebaseEmailVerified").(bool); ok {
				emailVerified = ev
			}
		}
	}

	user, err := services.GetOrCreateUser(uid, "", "", isAnonymous, emailVerified)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create user"})
	}

	resp := userToMap(user)
	resp["is_site_admin"] = middleware.IsSiteAdmin(uid)
	return c.JSON(resp)
}

func GetMe(c *fiber.Ctx) error {
	uid := c.Locals("firebaseUID").(string)
	user, err := services.GetUserByFirebaseUID(uid)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}
	resp := userToMap(user)
	resp["is_site_admin"] = middleware.IsSiteAdmin(uid)
	return c.JSON(resp)
}

func UpdateProfile(c *fiber.Ctx) error {
	uid := c.Locals("firebaseUID").(string)
	user, err := services.GetUserByFirebaseUID(uid)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	displayName := c.FormValue("display_name")
	username := c.FormValue("username")
	status := c.FormValue("status")
	avatarURL := ""

	// Anonymous users cannot change their display name
	if displayName != "" && user.IsAnonymous {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Verify your email to unlock this feature"})
	}

	// Validate display name length
	if len([]rune(displayName)) > 64 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Display name is too long (max 64 characters)"})
	}

	// Validate status
	if status != "" && status != "online" && status != "invisible" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Status must be 'online' or 'invisible'"})
	}

	// Validate and check username availability
	if username != "" {
		if user.IsAnonymous {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Verify your email to unlock this feature"})
		}
		if !usernameRe.MatchString(username) || strings.HasPrefix(username, "_") || strings.HasSuffix(username, "_") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Username must be 2–32 characters: letters, numbers, underscores, not starting or ending with underscore"})
		}
		available, checkErr := services.CheckUsernameAvailable(username, user.ID)
		if checkErr != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not verify username"})
		}
		if !available {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Username is already taken"})
		}
	}

	// Handle avatar upload or clear
	oldAvatarURL := user.AvatarURL
	clearAvatar := c.FormValue("clear_avatar") == "true"
	file, err := c.FormFile("avatar")
	if err == nil && file != nil && user.IsRestricted() {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Verify your email to unlock this feature"})
	}
	if err == nil && file != nil {
		ext := strings.ToLower(filepath.Ext(file.Filename))
		allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
		if !allowed[ext] {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid file type. Allowed: jpg, jpeg, png, gif, webp"})
		}
		if err := checkMagicBytes(file); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid file type. Allowed: jpg, jpeg, png, gif, webp"})
		}

		filename := fmt.Sprintf("%d_%s%s", user.ID, stamp(), ext)
		url, uploadErr := uploadFile(c, file, "avatars", filename)
		if uploadErr != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save avatar"})
		}
		avatarURL = url
		clearAvatar = false // uploading a new file takes precedence
	}

	oldStatus := user.Status
	updated, err := services.UpdateUser(user.ID, displayName, username, avatarURL, clearAvatar, status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update profile"})
	}

	// Delete old avatar from storage after successful update
	if (avatarURL != "" || clearAvatar) && oldAvatarURL != "" {
		storage.DeleteFileByURL(oldAvatarURL)
	}

	// Broadcast presence change if status changed
	if status != "" && status != oldStatus {
		effectiveStatus := status
		if status == "invisible" {
			effectiveStatus = "offline"
		}
		ws.GetHub().BroadcastPresenceChange(updated.ID, effectiveStatus)
	}

	// Broadcast profile change if display_name or avatar changed
	if updated.DisplayName != user.DisplayName || updated.AvatarURL != user.AvatarURL {
		ws.GetHub().BroadcastUserProfileChange(updated.ID, updated.DisplayName, updated.AvatarURL)
	}

	return c.JSON(updated)
}

func DeleteAccount(c *fiber.Ctx) error {
	uid := c.Locals("firebaseUID").(string)
	isFederated := strings.HasPrefix(uid, "fed:")

	if !isFederated {
		// Require a freshly issued token to guard against session-hijacking.
		// Bypass the Redis cache so we always get the real issuance time.
		rawToken := strings.TrimPrefix(c.Get("Authorization"), "Bearer ")
		decodedToken, err := middleware.VerifyTokenFull(rawToken)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
		}
		if time.Since(time.Unix(decodedToken.IssuedAt, 0)) > 5*time.Minute {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Recent authentication required. Please sign in again to delete your account"})
		}
	}

	user, err := services.GetUserByFirebaseUID(uid)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	// Delete avatar from storage if present
	if user.AvatarURL != "" {
		storage.DeleteFileByURL(user.AvatarURL)
	}

	// Anonymize the DB record so FK constraints are satisfied and
	// messages remain visible as "Deleted User"
	if err := services.DeleteUser(user.ID, uid); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete account"})
	}

	// Delete the Firebase user so they cannot log back in (skip for federated users)
	if !isFederated {
		if err := middleware.DeleteFirebaseUser(uid); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete account"})
		}
	} else {
		// Invalidate the federation session
		rawToken := strings.TrimPrefix(c.Get("Authorization"), "Bearer ")
		services.DeleteFederationSession(rawToken)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func CheckUsername(c *fiber.Ctx) error {
	username := c.Query("username")
	if username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Username is required"})
	}
	if !usernameRe.MatchString(username) || strings.HasPrefix(username, "_") || strings.HasSuffix(username, "_") {
		return c.JSON(fiber.Map{"available": false})
	}

	uid := c.Locals("firebaseUID").(string)
	user, err := services.GetUserByFirebaseUID(uid)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	available, err := services.CheckUsernameAvailable(username, user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Server error"})
	}
	return c.JSON(fiber.Map{"available": available})
}

// userToMap converts a User to a map so extra fields (like is_site_admin) can be added.
func userToMap(u *models.User) fiber.Map {
	m := fiber.Map{
		"id":             u.ID,
		"firebase_uid":   u.FirebaseUID,
		"display_name":   u.DisplayName,
		"avatar_url":     u.AvatarURL,
		"status":         u.Status,
		"is_anonymous":   u.IsAnonymous,
		"email_verified": u.EmailVerified,
		"created_at":     u.CreatedAt,
		"updated_at":     u.UpdatedAt,
	}
	if u.Username != nil {
		m["username"] = *u.Username
	} else {
		m["username"] = nil
	}
	if u.HomeInstance != nil {
		m["home_instance"] = *u.HomeInstance
	}
	return m
}
