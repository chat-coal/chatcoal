package services

import (
	"crypto/rand"
	"chatcoal/cache"
	"chatcoal/database"
	"chatcoal/models"
	"encoding/hex"
	"errors"
	"time"
)

func generateInviteCode() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func CreateServer(name string, ownerID models.Snowflake, isPublic bool) (*models.Server, error) {
	server := models.Server{
		Name:       name,
		OwnerID:    ownerID,
		InviteCode: generateInviteCode(),
		IsPublic:   isPublic,
	}

	tx := database.Database.Begin()

	if err := tx.Create(&server).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Create default #general channel
	channel := models.Channel{
		Name:     "general",
		ServerID: server.ID,
		Type:     "text",
		Position: 0,
	}
	if err := tx.Create(&channel).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Add owner as member
	member := models.ServerMember{
		UserID:   ownerID,
		ServerID: server.ID,
		Role:     "owner",
		JoinedAt: time.Now(),
	}
	if err := tx.Create(&member).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Create an invite row matching the legacy invite code
	if err := CreateInviteInTx(tx, server.ID, ownerID, server.InviteCode); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	cache.SetMemberWithRole(ownerID, server.ID, "owner")
	cache.InvalidateUserServers(ownerID)
	return &server, nil
}

func CountUserServers(userID models.Snowflake) (int64, error) {
	var count int64
	err := database.Database.Model(&models.ServerMember{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

func GetServersByUserID(userID models.Snowflake) ([]models.Server, error) {
	if cached := cache.GetUserServers(userID); cached != nil {
		return cached, nil
	}
	servers := make([]models.Server, 0)
	err := database.Database.
		Joins("JOIN server_members ON server_members.server_id = servers.id").
		Where("server_members.user_id = ?", userID).
		Find(&servers).Error
	if err != nil {
		return nil, err
	}
	cache.SetUserServers(userID, servers)
	return servers, nil
}

func GetServerByID(id models.Snowflake) (*models.Server, error) {
	var server models.Server
	if err := database.Database.First(&server, id).Error; err != nil {
		return nil, err
	}
	return &server, nil
}

func GetServerByInviteCode(code string) (*models.Server, error) {
	var server models.Server
	if err := database.Database.Where("invite_code = ?", code).First(&server).Error; err != nil {
		return nil, err
	}
	return &server, nil
}

func JoinServer(userID models.Snowflake, inviteCode string) (*models.Server, error) {
	// Try new invite table first
	invite, err := GetInviteByCode(inviteCode)
	var server *models.Server
	if err == nil {
		server = &invite.Server
		UseInvite(invite.ID)
	} else {
		// Fall back to legacy Server.InviteCode
		server, err = GetServerByInviteCode(inviteCode)
		if err != nil {
			return nil, err
		}
	}

	if banned, _ := IsUserBanned(userID, server.ID); banned {
		return nil, errors.New("you are banned from this server")
	}

	member := models.ServerMember{
		UserID:   userID,
		ServerID: server.ID,
		Role:     "member",
		JoinedAt: time.Now(),
	}
	if err := database.Database.Create(&member).Error; err != nil {
		return nil, err
	}

	cache.SetMember(userID, server.ID)
	cache.InvalidateUserServers(userID)
	return server, nil
}

const MemberPageLimit = 100

func GetServerMembers(serverID models.Snowflake, after models.Snowflake) ([]models.ServerMember, error) {
	var members []models.ServerMember
	query := database.Database.Preload("User").
		Joins("JOIN users ON users.id = server_members.user_id").
		Where("server_members.server_id = ?", serverID).
		Where("users.firebase_uid NOT LIKE ?", "deleted:%")

	if after > 0 {
		query = query.Where("server_members.id > ?", after)
	}

	err := query.Order("server_members.id ASC").Limit(MemberPageLimit).Find(&members).Error
	return members, err
}

func GetServerMember(userID models.Snowflake, serverID models.Snowflake) (*models.ServerMember, error) {
	var member models.ServerMember
	err := database.Database.Preload("User").
		Where("user_id = ? AND server_id = ?", userID, serverID).
		First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func LeaveServer(userID models.Snowflake, serverID models.Snowflake) (*models.ServerMember, error) {
	member, err := GetServerMember(userID, serverID)
	if err != nil {
		return nil, err
	}
	if member.Role == "owner" {
		return nil, errors.New("owner cannot leave server")
	}
	if err := database.Database.Delete(member).Error; err != nil {
		return nil, err
	}
	cache.InvalidateMember(userID, serverID)
	cache.InvalidateUserServers(userID)
	return member, nil
}

func IsServerMember(userID models.Snowflake, serverID models.Snowflake) bool {
	return cache.IsMember(userID, serverID)
}

// GetSystemChannel returns the channel where system announcements should be
// posted, or nil if the feature is disabled.
func GetSystemChannel(server *models.Server) (*models.Channel, error) {
	if !server.ShowJoinLeave {
		return nil, nil
	}
	if server.SystemChannelID != nil && *server.SystemChannelID != 0 {
		ch, err := GetChannelByID(*server.SystemChannelID)
		if err == nil {
			return ch, nil
		}
		// Channel was deleted — fall through to default
	}
	// Fall back to first text channel
	var ch models.Channel
	err := database.Database.
		Where("server_id = ? AND type = ?", server.ID, "text").
		Order("position ASC, id ASC").
		First(&ch).Error
	if err != nil {
		return nil, err
	}
	return &ch, nil
}

// PostSystemAnnouncement creates a system message in the designated channel.
func PostSystemAnnouncement(server *models.Server, msgType, content string, userID models.Snowflake) *models.Message {
	ch, err := GetSystemChannel(server)
	if ch == nil || err != nil {
		return nil
	}
	msg, err := CreateSystemMessage(msgType, content, ch.ID, server.ID, userID)
	if err != nil {
		return nil
	}
	return msg
}

func UpdateServer(serverID models.Snowflake, ownerID models.Snowflake, name string, iconURL string, clearIcon bool, isPublic *bool, showJoinLeave *bool, systemChannelID *models.Snowflake) (*models.Server, error) {
	server, err := GetServerByID(serverID)
	if err != nil {
		return nil, err
	}
	if server.OwnerID != ownerID {
		return nil, errors.New("only the owner can update the server")
	}

	updates := map[string]interface{}{}
	if name != "" {
		updates["name"] = name
	}
	if clearIcon {
		updates["icon_url"] = ""
	} else if iconURL != "" {
		updates["icon_url"] = iconURL
	}
	if isPublic != nil {
		updates["is_public"] = *isPublic
	}
	if showJoinLeave != nil {
		updates["show_join_leave"] = *showJoinLeave
	}
	if systemChannelID != nil {
		if *systemChannelID == 0 {
			updates["system_channel_id"] = nil
		} else {
			updates["system_channel_id"] = *systemChannelID
		}
	}

	if len(updates) > 0 {
		if err := database.Database.Model(server).Updates(updates).Error; err != nil {
			return nil, err
		}
		// Re-fetch so the returned struct reflects the new values (GORM's Updates(map)
		// executes the SQL but does not update the Go struct fields in place).
		if err := database.Database.First(server, serverID).Error; err != nil {
			return nil, err
		}
		// Invalidate the server list cache for all members so they see the updated
		// server immediately (e.g. new icon URL) instead of the stale cached version.
		var memberUserIDs []models.Snowflake
		database.Database.Model(&models.ServerMember{}).
			Where("server_id = ?", serverID).
			Pluck("user_id", &memberUserIDs)
		for _, uid := range memberUserIDs {
			cache.InvalidateUserServers(uid)
		}
	}

	return server, nil
}

// GetPublicServers returns public servers the given user has not yet joined,
// ordered by member count descending. hasMore is true when public servers exist
// that were excluded because the user is already a member.
func GetPublicServers(userID models.Snowflake) ([]models.PublicServer, bool, error) {
	subquery := database.Database.Table("server_members").
		Select("server_id").
		Where("user_id = ?", userID)

	results := make([]models.PublicServer, 0)
	err := database.Database.
		Table("servers").
		Select("servers.id, servers.name, servers.icon_url, COUNT(server_members.id) AS member_count").
		Joins("LEFT JOIN server_members ON server_members.server_id = servers.id").
		Where("servers.is_public = TRUE").
		Where("servers.id NOT IN (?)", subquery).
		Group("servers.id").
		Order("member_count DESC").
		Limit(100).
		Scan(&results).Error
	if err != nil {
		return results, false, err
	}

	// Check if there are any public servers the user is already a member of
	var excluded int64
	database.Database.Table("servers").
		Joins("JOIN server_members ON server_members.server_id = servers.id AND server_members.user_id = ?", userID).
		Where("servers.is_public = TRUE").
		Count(&excluded)

	return results, excluded > 0, nil
}

// JoinPublicServer adds userID to a public server without an invite code.
func JoinPublicServer(userID models.Snowflake, serverID models.Snowflake) (*models.Server, error) {
	server, err := GetServerByID(serverID)
	if err != nil {
		return nil, err
	}
	if !server.IsPublic {
		return nil, errors.New("server is not public")
	}

	if banned, _ := IsUserBanned(userID, server.ID); banned {
		return nil, errors.New("you are banned from this server")
	}

	member := models.ServerMember{
		UserID:   userID,
		ServerID: server.ID,
		Role:     "member",
		JoinedAt: time.Now(),
	}
	if err := database.Database.Create(&member).Error; err != nil {
		return nil, err
	}

	cache.SetMember(userID, server.ID)
	cache.InvalidateUserServers(userID)
	return server, nil
}

func UpdateMemberRole(actorID, targetID, serverID models.Snowflake, newRole string) (*models.ServerMember, error) {
	if newRole != "admin" && newRole != "member" {
		return nil, errors.New("invalid role")
	}
	if actorID == targetID {
		return nil, errors.New("cannot change your own role")
	}
	if !HasPermission(actorID, serverID, PermManageRoles) {
		return nil, errors.New("insufficient permissions")
	}

	target, err := GetServerMember(targetID, serverID)
	if err != nil {
		return nil, errors.New("member not found")
	}
	if target.Role == "owner" {
		return nil, errors.New("cannot change the owner's role")
	}

	target.Role = newRole
	if err := database.Database.Save(target).Error; err != nil {
		return nil, err
	}
	cache.SetMemberWithRole(targetID, serverID, newRole)
	return target, nil
}

func KickMember(actorID, targetID, serverID models.Snowflake) (*models.ServerMember, error) {
	if !CanKick(actorID, targetID, serverID) {
		return nil, errors.New("insufficient permissions")
	}

	member, err := GetServerMember(targetID, serverID)
	if err != nil {
		return nil, errors.New("member not found")
	}

	if err := database.Database.Delete(member).Error; err != nil {
		return nil, err
	}
	cache.InvalidateMember(targetID, serverID)
	cache.InvalidateUserServers(targetID)
	return member, nil
}

func TransferOwnership(currentOwnerID, newOwnerID, serverID models.Snowflake) error {
	server, err := GetServerByID(serverID)
	if err != nil {
		return err
	}
	if server.OwnerID != currentOwnerID {
		return errors.New("only the owner can transfer ownership")
	}
	if currentOwnerID == newOwnerID {
		return errors.New("already the owner")
	}

	// Verify new owner is a member
	if !IsServerMember(newOwnerID, serverID) {
		return errors.New("target is not a member")
	}

	tx := database.Database.Begin()

	// Update server owner
	if err := tx.Model(server).Update("owner_id", newOwnerID).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Set new owner's role to "owner"
	if err := tx.Model(&models.ServerMember{}).
		Where("user_id = ? AND server_id = ?", newOwnerID, serverID).
		Update("role", "owner").Error; err != nil {
		tx.Rollback()
		return err
	}

	// Set old owner's role to "admin"
	if err := tx.Model(&models.ServerMember{}).
		Where("user_id = ? AND server_id = ?", currentOwnerID, serverID).
		Update("role", "admin").Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	cache.SetMemberWithRole(newOwnerID, serverID, "owner")
	cache.SetMemberWithRole(currentOwnerID, serverID, "admin")

	return nil
}

// CleanupUserServers deletes any server where userID is the sole remaining
// admin or owner. Called during account deletion so orphaned servers are
// removed along with their channels, members, and invites.
func CleanupUserServers(userID models.Snowflake) error {
	var entries []models.ServerMember
	if err := database.Database.
		Where("user_id = ? AND role IN ?", userID, []string{"owner", "admin"}).
		Find(&entries).Error; err != nil {
		return err
	}

	for _, entry := range entries {
		var otherAdmins int64
		database.Database.Model(&models.ServerMember{}).
			Where("server_id = ? AND user_id != ? AND role IN ?", entry.ServerID, userID, []string{"owner", "admin"}).
			Count(&otherAdmins)
		if otherAdmins > 0 {
			continue
		}

		server, err := GetServerByID(entry.ServerID)
		if err != nil {
			continue
		}

		var memberUserIDs []models.Snowflake
		database.Database.Model(&models.ServerMember{}).
			Where("server_id = ?", entry.ServerID).
			Pluck("user_id", &memberUserIDs)

		tx := database.Database.Begin()
		if err := tx.Exec("DELETE FROM messages WHERE channel_id IN (SELECT id FROM channels WHERE server_id = ?)", entry.ServerID).Error; err != nil {
			tx.Rollback()
			continue
		}
		if err := tx.Where("server_id = ?", entry.ServerID).Delete(&models.Channel{}).Error; err != nil {
			tx.Rollback()
			continue
		}
		if err := tx.Where("server_id = ?", entry.ServerID).Delete(&models.ServerMember{}).Error; err != nil {
			tx.Rollback()
			continue
		}
		if err := tx.Where("server_id = ?", entry.ServerID).Delete(&models.Invite{}).Error; err != nil {
			tx.Rollback()
			continue
		}
		if err := tx.Delete(server).Error; err != nil {
			tx.Rollback()
			continue
		}
		if err := tx.Commit().Error; err != nil {
			continue
		}

		for _, uid := range memberUserIDs {
			cache.InvalidateMember(uid, entry.ServerID)
			cache.InvalidateUserServers(uid)
		}
		cache.InvalidateServerChannels(entry.ServerID)
	}
	return nil
}

// IsUserBanned checks if a user is banned from a server.
func IsUserBanned(userID, serverID models.Snowflake) (bool, error) {
	var count int64
	err := database.Database.Model(&models.ServerBan{}).
		Where("user_id = ? AND server_id = ?", userID, serverID).
		Count(&count).Error
	return count > 0, err
}

// BanMember bans a user from a server: creates ban record + removes membership.
func BanMember(actorID, targetID, serverID models.Snowflake, reason *string) (*models.ServerMember, error) {
	if !CanBan(actorID, targetID, serverID) {
		return nil, errors.New("insufficient permissions")
	}

	member, err := GetServerMember(targetID, serverID)
	if err != nil {
		return nil, errors.New("member not found")
	}

	tx := database.Database.Begin()

	ban := models.ServerBan{
		ServerID: serverID,
		UserID:   targetID,
		BannedBy: actorID,
		Reason:   reason,
	}
	if err := tx.Create(&ban).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Delete(member).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	cache.InvalidateMember(targetID, serverID)
	cache.InvalidateUserServers(targetID)
	return member, nil
}

// GetServerBans returns all bans for a server with user info preloaded.
func GetServerBans(serverID models.Snowflake) ([]models.ServerBan, error) {
	var bans []models.ServerBan
	err := database.Database.Preload("User").
		Where("server_id = ?", serverID).
		Order("created_at DESC").
		Find(&bans).Error
	return bans, err
}

// UnbanUser removes a ban record for a user in a server.
func UnbanUser(serverID, userID models.Snowflake) error {
	return database.Database.
		Where("server_id = ? AND user_id = ?", serverID, userID).
		Delete(&models.ServerBan{}).Error
}

func DeleteServer(serverID models.Snowflake, ownerID models.Snowflake) error {
	server, err := GetServerByID(serverID)
	if err != nil {
		return err
	}
	if server.OwnerID != ownerID {
		return errors.New("only the owner can delete the server")
	}

	// Fetch member IDs before deleting so we can invalidate cache
	var memberUserIDs []models.Snowflake
	database.Database.Model(&models.ServerMember{}).
		Where("server_id = ?", serverID).
		Pluck("user_id", &memberUserIDs)

	tx := database.Database.Begin()

	// Delete messages in all channels of this server
	if err := tx.Exec("DELETE FROM messages WHERE channel_id IN (SELECT id FROM channels WHERE server_id = ?)", serverID).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("server_id = ?", serverID).Delete(&models.Channel{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("server_id = ?", serverID).Delete(&models.ServerMember{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("server_id = ?", serverID).Delete(&models.Invite{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Delete(server).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	for _, uid := range memberUserIDs {
		cache.InvalidateMember(uid, serverID)
		cache.InvalidateUserServers(uid)
	}
	cache.InvalidateServerChannels(serverID)

	return nil
}
