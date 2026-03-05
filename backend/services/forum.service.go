package services

import (
	"chatcoal/database"
	"chatcoal/models"
)

const ForumPostPageLimit = 25

func GetForumPosts(channelID models.Snowflake, before models.Snowflake) ([]models.ForumPost, error) {
	var posts []models.ForumPost
	query := database.Database.Preload("Author").Where("channel_id = ?", channelID)

	if before > 0 {
		query = query.Where("id < ?", before)
	}

	err := query.Order("COALESCE(last_reply_at, created_at) DESC").Limit(ForumPostPageLimit).Find(&posts).Error
	return posts, err
}

func CreateForumPost(title, content string, channelID, authorID models.Snowflake) (*models.ForumPost, error) {
	post := models.ForumPost{
		Title:     title,
		Content:   content,
		ChannelID: channelID,
		AuthorID:  authorID,
	}
	if err := database.Database.Create(&post).Error; err != nil {
		return nil, err
	}
	database.Database.Preload("Author").First(&post, post.ID)
	return &post, nil
}

func GetForumPostByID(id models.Snowflake) (*models.ForumPost, error) {
	var post models.ForumPost
	if err := database.Database.Preload("Author").First(&post, id).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

func UpdateForumPost(id models.Snowflake, authorID models.Snowflake, title, content string) (*models.ForumPost, error) {
	var post models.ForumPost
	if err := database.Database.First(&post, id).Error; err != nil {
		return nil, err
	}
	if post.AuthorID != authorID {
		return nil, fiber_forbidden()
	}
	if title != "" {
		post.Title = title
	}
	if content != "" {
		post.Content = content
	}
	if err := database.Database.Save(&post).Error; err != nil {
		return nil, err
	}
	database.Database.Preload("Author").First(&post, post.ID)
	return &post, nil
}

func DeleteForumPost(id models.Snowflake, authorID models.Snowflake, hasManagePerm bool) error {
	var post models.ForumPost
	if err := database.Database.First(&post, id).Error; err != nil {
		return err
	}
	if post.AuthorID != authorID && !hasManagePerm {
		return fiber_forbidden()
	}
	return database.Database.Delete(&post).Error
}
