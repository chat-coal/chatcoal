package controllers

import (
	"chatcoal/cache"
	"chatcoal/models"
	"chatcoal/services"

	"github.com/gofiber/fiber/v2"
)

func GetForumPosts(c *fiber.Ctx) error {
	channelID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid channel ID"})
	}

	ch, err := services.GetChannelByID(channelID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Channel not found"})
	}
	if ch.Type != "forum" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Not a forum channel"})
	}

	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil || !services.IsServerMember(user.ID, ch.ServerID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	before := parseSnowflakeQuery(c, "before")

	posts, err := services.GetForumPosts(channelID, before)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch posts"})
	}

	return c.JSON(posts)
}

func CreateForumPost(c *fiber.Ctx) error {
	channelID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid channel ID"})
	}

	ch, err := services.GetChannelByID(channelID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Channel not found"})
	}
	if ch.Type != "forum" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Not a forum channel"})
	}

	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil || !services.IsServerMember(user.ID, ch.ServerID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	var body struct {
		Title   string `json:"title" validate:"required,min=1,max=200"`
		Content string `json:"content" validate:"required,min=1,max=4000"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}
	if msg := validateBody(&body); msg != "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": msg})
	}

	post, err := services.CreateForumPost(body.Title, body.Content, channelID, user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create post"})
	}

	broadcastEvent(ch.ServerID, "forum_post", post)
	return c.Status(fiber.StatusCreated).JSON(post)
}

func GetForumPostByIDHandler(c *fiber.Ctx) error {
	postID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid post ID"})
	}

	post, err := services.GetForumPostByID(postID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Post not found"})
	}

	ch, err := services.GetChannelByID(post.ChannelID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Channel not found"})
	}

	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil || !services.IsServerMember(user.ID, ch.ServerID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	return c.JSON(post)
}

func DeleteForumPostHandler(c *fiber.Ctx) error {
	postID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid post ID"})
	}

	post, err := services.GetForumPostByID(postID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Post not found"})
	}

	ch, err := services.GetChannelByID(post.ChannelID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Channel not found"})
	}

	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil || !services.IsServerMember(user.ID, ch.ServerID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	hasPerm := services.HasPermission(user.ID, ch.ServerID, services.PermManageMessages)
	if err := services.DeleteForumPost(postID, user.ID, hasPerm); err != nil {
		if services.IsForbiddenError(err) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Not allowed"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete post"})
	}

	broadcastEvent(ch.ServerID, "forum_post_delete", fiber.Map{"id": postID, "channel_id": post.ChannelID})
	return c.SendStatus(fiber.StatusNoContent)
}

func EditForumPostHandler(c *fiber.Ctx) error {
	postID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid post ID"})
	}

	post, err := services.GetForumPostByID(postID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Post not found"})
	}

	ch, err := services.GetChannelByID(post.ChannelID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Channel not found"})
	}

	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil || !services.IsServerMember(user.ID, ch.ServerID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	var body struct {
		Title   string `json:"title" validate:"max=200"`
		Content string `json:"content" validate:"max=4000"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}
	if body.Title == "" && body.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Title or content required"})
	}
	if msg := validateBody(&body); msg != "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": msg})
	}

	updated, err := services.UpdateForumPost(postID, user.ID, body.Title, body.Content)
	if err != nil {
		if services.IsForbiddenError(err) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Not allowed"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update post"})
	}

	broadcastEvent(ch.ServerID, "forum_post_edit", updated)
	return c.JSON(updated)
}

func GetForumPostMessages(c *fiber.Ctx) error {
	postID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid post ID"})
	}

	post, err := services.GetForumPostByID(postID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Post not found"})
	}

	ch, err := services.GetChannelByID(post.ChannelID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Channel not found"})
	}

	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil || !services.IsServerMember(user.ID, ch.ServerID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	before := parseSnowflakeQuery(c, "before")

	fpID := postID
	messages, err := services.GetMessages(post.ChannelID, before, &fpID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch messages"})
	}

	return c.JSON(messages)
}

func SendForumPostMessage(c *fiber.Ctx) error {
	postID, err := parseSnowflakeParam(c, "id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid post ID"})
	}

	post, err := services.GetForumPostByID(postID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Post not found"})
	}

	ch, err := services.GetChannelByID(post.ChannelID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Channel not found"})
	}

	uid := c.Locals("firebaseUID").(string)
	user, _ := services.GetUserByFirebaseUID(uid)
	if user == nil || !services.IsServerMember(user.ID, ch.ServerID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}

	if user.IsRestricted() && !cache.AnonMessageRateLimitOK(int64(user.ID)) {
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "Sending too fast. Verify your email to remove this limit"})
	}

	var body struct {
		Content     string            `json:"content" validate:"required,max=4000"`
		ReplyToID   *models.Snowflake `json:"reply_to_id"`
		ImageWidth  int               `json:"image_width"`
		ImageHeight int               `json:"image_height"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}
	if msg := validateBody(&body); msg != "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": msg})
	}

	fpID := postID
	message, err := services.CreateMessage(body.Content, post.ChannelID, ch.ServerID, user.ID, "", "", 0, body.ImageWidth, body.ImageHeight, body.ReplyToID, &fpID)
	if err != nil {
		if services.IsValidationError(err) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": services.ValidationErrorMessage(err)})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to send message"})
	}

	broadcastEvent(ch.ServerID, "forum_message", message)
	return c.Status(fiber.StatusCreated).JSON(message)
}
