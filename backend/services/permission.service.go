package services

import (
	"chatcoal/cache"
	"chatcoal/models"
)

// Permission constants
const (
	PermManageServer   = "manage_server"
	PermManageRoles    = "manage_roles"
	PermManageChannels = "manage_channels"
	PermManageMessages = "manage_messages"
	PermManageInvites  = "manage_invites"
	PermKickMembers    = "kick_members"
	PermBanMembers     = "ban_members"
)

var roleLevel = map[string]int{
	"owner":  3,
	"admin":  2,
	"member": 1,
}

var permissionMinRole = map[string]string{
	PermManageServer:   "owner",
	PermManageRoles:    "owner",
	PermManageChannels: "admin",
	PermManageMessages: "admin",
	PermManageInvites:  "admin",
	PermKickMembers:    "admin",
	PermBanMembers:     "admin",
}

// GetMemberRole returns the role of a user in a server (cache-first, DB fallback).
func GetMemberRole(userID, serverID models.Snowflake) string {
	return cache.GetMemberRole(userID, serverID)
}

// HasPermission checks if a user has the given permission in a server.
func HasPermission(userID, serverID models.Snowflake, permission string) bool {
	role := GetMemberRole(userID, serverID)
	if role == "" {
		return false
	}
	minRole, ok := permissionMinRole[permission]
	if !ok {
		return false
	}
	return roleLevel[role] >= roleLevel[minRole]
}

// CanKick checks if actor can kick target: actor must be strictly higher role and at least admin.
func CanKick(actorID, targetID, serverID models.Snowflake) bool {
	if actorID == targetID {
		return false
	}
	actorRole := GetMemberRole(actorID, serverID)
	targetRole := GetMemberRole(targetID, serverID)
	if actorRole == "" || targetRole == "" {
		return false
	}
	return roleLevel[actorRole] >= roleLevel["admin"] && roleLevel[actorRole] > roleLevel[targetRole]
}

// CanBan checks if actor can ban target: same rules as CanKick.
func CanBan(actorID, targetID, serverID models.Snowflake) bool {
	if actorID == targetID {
		return false
	}
	actorRole := GetMemberRole(actorID, serverID)
	targetRole := GetMemberRole(targetID, serverID)
	if actorRole == "" || targetRole == "" {
		return false
	}
	return roleLevel[actorRole] >= roleLevel["admin"] && roleLevel[actorRole] > roleLevel[targetRole]
}
