package services

import (
	"encoding/json"
	"regexp"
	"strconv"

	"chatcoal/models"
)

var mentionRegex = regexp.MustCompile(`<@(\d{1,20})>`)

// ParseMentions extracts deduplicated user IDs from <@id> tokens in content.
func ParseMentions(content string) []models.Snowflake {
	matches := mentionRegex.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}

	seen := make(map[models.Snowflake]bool, len(matches))
	var ids []models.Snowflake
	for _, m := range matches {
		n, err := strconv.ParseInt(m[1], 10, 64)
		if err != nil {
			continue
		}
		id := models.Snowflake(n)
		if !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	return ids
}

// MentionsToJSON serializes mention IDs to a JSON array of strings.
func MentionsToJSON(ids []models.Snowflake) json.RawMessage {
	if len(ids) == 0 {
		return nil
	}
	strs := make([]string, len(ids))
	for i, id := range ids {
		strs[i] = id.String()
	}
	data, _ := json.Marshal(strs)
	return data
}

// ResolveMentionsForDisplay replaces <@id> tokens with @DisplayName for push body text.
func ResolveMentionsForDisplay(content string, mentionIDs []models.Snowflake) string {
	if len(mentionIDs) == 0 {
		return content
	}

	// Build lookup map
	nameMap := make(map[string]string, len(mentionIDs))
	for _, id := range mentionIDs {
		user, err := GetUserByID(id)
		if err != nil || user == nil {
			continue
		}
		display := user.DisplayName
		if display == "" && user.Username != nil {
			display = *user.Username
		}
		if display == "" {
			display = "Unknown"
		}
		nameMap[id.String()] = display
	}

	return mentionRegex.ReplaceAllStringFunc(content, func(match string) string {
		sub := mentionRegex.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		if name, ok := nameMap[sub[1]]; ok {
			return "@" + name
		}
		return match
	})
}
