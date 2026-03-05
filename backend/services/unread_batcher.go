package services

import (
	"strings"
	"sync"
	"time"

	"chatcoal/database"
	"chatcoal/models"

	"github.com/gofiber/fiber/v2/log"
)

const unreadFlushInterval = 2 * time.Second

type channelKey struct {
	ChannelID models.Snowflake
	ServerID  models.Snowflake
}

type pendingEntry struct {
	Total     int
	PerAuthor map[models.Snowflake]int
}

var (
	unreadMu      sync.Mutex
	unreadPending map[channelKey]*pendingEntry

	unreadStop chan struct{}
	unreadDone chan struct{}
)

// EnqueueUnread records an unread increment for later batch flushing.
// It is non-blocking and does not hit the database.
func EnqueueUnread(channelID, serverID, authorID models.Snowflake) {
	unreadMu.Lock()
	key := channelKey{ChannelID: channelID, ServerID: serverID}
	entry := unreadPending[key]
	if entry == nil {
		entry = &pendingEntry{PerAuthor: make(map[models.Snowflake]int)}
		unreadPending[key] = entry
	}
	entry.Total++
	entry.PerAuthor[authorID]++
	unreadMu.Unlock()
}

// StartUnreadBatcher launches the background goroutine that periodically
// flushes accumulated unread increments to the database.
func StartUnreadBatcher() {
	unreadPending = make(map[channelKey]*pendingEntry)
	unreadStop = make(chan struct{})
	unreadDone = make(chan struct{})

	go func() {
		defer close(unreadDone)
		ticker := time.NewTicker(unreadFlushInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				flushUnreads()
			case <-unreadStop:
				flushUnreads()
				return
			}
		}
	}()
}

// StopUnreadBatcher signals the flush goroutine to stop and waits for a
// final flush to complete.
func StopUnreadBatcher() {
	close(unreadStop)
	<-unreadDone
}

// flushUnreads swaps the pending map, then processes each channel entry.
func flushUnreads() {
	unreadMu.Lock()
	batch := unreadPending
	unreadPending = make(map[channelKey]*pendingEntry)
	unreadMu.Unlock()

	if len(batch) == 0 {
		return
	}

	for key, entry := range batch {
		flushChannel(key, entry)
	}
}

// flushChannel writes a single batched INSERT for one channel.
func flushChannel(key channelKey, entry *pendingEntry) {
	// Fetch all server members once.
	var memberIDs []models.Snowflake
	if err := database.Database.
		Table("server_members").
		Where("server_id = ?", key.ServerID).
		Pluck("user_id", &memberIDs).Error; err != nil || len(memberIDs) == 0 {
		return
	}

	// Get muted user IDs for this channel/server so we can skip them.
	mutedUsers := GetMutedUserIDsForChannel(key.ChannelID, key.ServerID)

	// Build per-member increment: total minus the messages they authored.
	placeholders := make([]string, 0, len(memberIDs))
	args := make([]any, 0, len(memberIDs)*5)
	for _, uid := range memberIDs {
		if mutedUsers[uid] {
			continue
		}
		inc := entry.Total - entry.PerAuthor[uid]
		if inc <= 0 {
			continue
		}
		placeholders = append(placeholders, "(?, ?, 'server', ?, ?, 0, ?, NOW())")
		args = append(args, models.GenerateID(), uid, key.ChannelID, key.ServerID, inc)
	}

	if len(placeholders) == 0 {
		return
	}

	sql := "INSERT INTO read_states (id, user_id, channel_type, channel_ref_id, server_id, last_read_message_id, unread_count, updated_at) VALUES " +
		strings.Join(placeholders, ", ") +
		" ON DUPLICATE KEY UPDATE unread_count = unread_count + VALUES(unread_count), server_id = VALUES(server_id), updated_at = NOW()"

	if err := database.Database.Exec(sql, args...).Error; err != nil {
		log.Errorf("[unreadBatcher] flush failed for channel %d: %v", key.ChannelID, err)
	}
}
