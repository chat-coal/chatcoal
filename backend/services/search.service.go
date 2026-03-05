package services

import (
	"chatcoal/database"
	"chatcoal/models"
	"strings"

	"gorm.io/gorm"
)

const SearchPageLimit = 25

// sanitizeFulltextQuery strips MySQL boolean mode operators to prevent syntax errors.
func sanitizeFulltextQuery(q string) string {
	replacer := strings.NewReplacer(
		"+", " ", "-", " ", ">", " ", "<", " ",
		"(", " ", ")", " ", "~", " ", "*", " ",
		"\"", " ", "@", " ", "!", " ",
	)
	return strings.Join(strings.Fields(replacer.Replace(q)), " ")
}

type SearchResult struct {
	models.Message
	ChannelName string `json:"channel_name"`
}

func SearchMessages(serverID models.Snowflake, query string, before models.Snowflake) ([]SearchResult, error) {
	// Step 1: Find matching message IDs with channel filter
	var ids []models.Snowflake
	q := database.Database.
		Table("messages").
		Select("messages.id").
		Joins("JOIN channels ON channels.id = messages.channel_id").
		Where("channels.server_id = ?", serverID).
		Where("MATCH(messages.content) AGAINST(? IN BOOLEAN MODE)", sanitizeFulltextQuery(query))

	if before > 0 {
		q = q.Where("messages.id < ?", before)
	}

	if err := q.Order("messages.id DESC").Limit(SearchPageLimit).Pluck("messages.id", &ids).Error; err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return []SearchResult{}, nil
	}

	// Step 2: Load full messages with preloads
	var messages []models.Message
	err := database.Database.
		Preload("Author").
		Preload("Reactions").
		Preload("ReplyTo", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, content, author_id").Preload("Author", func(db *gorm.DB) *gorm.DB {
				return db.Select("id, display_name, avatar_url")
			})
		}).
		Where("id IN ?", ids).
		Order("id DESC").
		Find(&messages).Error
	if err != nil {
		return nil, err
	}

	// Step 3: Get channel names
	channelNames := map[models.Snowflake]string{}
	for _, msg := range messages {
		channelNames[msg.ChannelID] = ""
	}
	var channelIDs []models.Snowflake
	for id := range channelNames {
		channelIDs = append(channelIDs, id)
	}
	var channels []models.Channel
	if err := database.Database.Where("id IN ?", channelIDs).Find(&channels).Error; err != nil {
		return nil, err
	}
	for _, ch := range channels {
		channelNames[ch.ID] = ch.Name
	}

	// Build results
	results := make([]SearchResult, len(messages))
	for i, msg := range messages {
		results[i] = SearchResult{
			Message:     msg,
			ChannelName: channelNames[msg.ChannelID],
		}
	}

	return results, nil
}
