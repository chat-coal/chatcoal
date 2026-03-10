import axios from 'axios'
import { auth } from './firebase'
import { useToastStore } from '../stores/toast'

export const API_URL = import.meta.env.VITE_API_URL

const api = axios.create({
  baseURL: API_URL,
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 429) {
      const toast = useToastStore()
      toast.add('Too many requests — please slow down and try again.', 'error')
    }
    return Promise.reject(error)
  },
)

api.interceptors.request.use(async (config) => {
  // Federation session tokens take priority over Firebase tokens.
  const fedToken = localStorage.getItem('fed_token')
  if (fedToken) {
    config.headers.Authorization = `Bearer ${fedToken}`
    return config
  }
  const user = auth.currentUser
  if (user) {
    const token = await user.getIdToken()
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// The backend returns raw reaction rows (one per user) on message load,
// but grouped reactions (emoji/count/user_ids) from toggle endpoints.
// Normalize raw rows into the grouped format the UI expects.
function normalizeReactions(reactions) {
  if (!reactions?.length) return reactions
  // Already grouped format
  if (reactions[0].count !== undefined || reactions[0].user_ids !== undefined) return reactions
  // Raw format — group by emoji
  const map = new Map()
  for (const r of reactions) {
    if (!r.emoji) continue
    if (!map.has(r.emoji)) map.set(r.emoji, [])
    if (r.user_id) map.get(r.emoji).push(r.user_id)
  }
  return Array.from(map, ([emoji, ids]) => ({ emoji, count: ids.length, user_ids: ids }))
}

function normalizeMessage(msg) {
  if (msg && msg.reactions) msg.reactions = normalizeReactions(msg.reactions)
  return msg
}

export { normalizeMessage }

export default {
  // Auth
  async login() {
    const res = await api.post('/api/auth/login')
    return res.data
  },
  async getMe() {
    const res = await api.get('/api/auth/me')
    return res.data
  },
  async updateProfile({ displayName, username, avatarFile, clearAvatar, status }) {
    const formData = new FormData()
    if (displayName) formData.append('display_name', displayName)
    if (username) formData.append('username', username)
    if (avatarFile) formData.append('avatar', avatarFile)
    if (clearAvatar) formData.append('clear_avatar', 'true')
    if (status) formData.append('status', status)
    const res = await api.put('/api/auth/profile', formData)
    return res.data
  },
  async checkUsername(username) {
    const res = await api.get('/api/auth/check-username', { params: { username } })
    return res.data
  },
  async deleteAccount() {
    await api.delete('/api/auth/account')
  },

  // Servers
  async getServers() {
    const res = await api.get('/api/servers')
    return res.data
  },
  async getPublicServers() {
    const res = await api.get('/api/servers/public')
    return res.data
  },
  async createServer(name, isPublic = false) {
    const res = await api.post('/api/servers', { name, is_public: isPublic })
    return res.data
  },
  async joinServer(inviteCode) {
    const res = await api.post('/api/servers/join', { invite_code: inviteCode })
    return res.data
  },
  async joinPublicServer(serverId) {
    const res = await api.post(`/api/servers/${serverId}/join`)
    return res.data
  },
  async updateServer(serverId, { name, iconFile, clearIcon, isPublic, showJoinLeave, systemChannelId }) {
    const formData = new FormData()
    if (name) formData.append('name', name)
    if (iconFile) formData.append('icon', iconFile)
    if (clearIcon) formData.append('clear_icon', 'true')
    if (isPublic !== undefined) formData.append('is_public', isPublic ? 'true' : 'false')
    if (showJoinLeave !== undefined) formData.append('show_join_leave', showJoinLeave ? 'true' : 'false')
    if (systemChannelId !== undefined) formData.append('system_channel_id', String(systemChannelId))
    const res = await api.put(`/api/servers/${serverId}`, formData)
    return res.data
  },
  async deleteServer(serverId) {
    await api.delete(`/api/servers/${serverId}`)
  },
  async leaveServer(serverId) {
    await api.delete(`/api/servers/${serverId}/leave`)
  },
  async getServerMembers(serverId) {
    const res = await api.get(`/api/servers/${serverId}/members`)
    return res.data
  },

  // Role management
  async updateMemberRole(serverId, userId, role) {
    const res = await api.patch(`/api/servers/${serverId}/members/${userId}/role`, { role })
    return res.data
  },
  async kickMember(serverId, userId) {
    await api.delete(`/api/servers/${serverId}/members/${userId}`)
  },
  async banMember(serverId, userId, reason) {
    await api.post(`/api/servers/${serverId}/bans/${userId}`, { reason: reason || null })
  },
  async getServerBans(serverId) {
    const res = await api.get(`/api/servers/${serverId}/bans`)
    return res.data
  },
  async unbanUser(serverId, userId) {
    await api.delete(`/api/servers/${serverId}/bans/${userId}`)
  },
  async transferOwnership(serverId, userId) {
    const res = await api.post(`/api/servers/${serverId}/transfer`, { user_id: userId })
    return res.data
  },

  // Invites
  async getInvites(serverId) {
    const res = await api.get(`/api/servers/${serverId}/invites`)
    return res.data
  },
  async createInvite(serverId, options = {}) {
    const res = await api.post(`/api/servers/${serverId}/invites`, options)
    return res.data
  },
  async deleteInvite(serverId, inviteId) {
    await api.delete(`/api/servers/${serverId}/invites/${inviteId}`)
  },
  async resolveInvite(code) {
    const res = await api.get(`/api/invites/${code}`)
    return res.data
  },

  // Voice
  async getVoiceStates(serverId) {
    const res = await api.get(`/api/servers/${serverId}/voice-states`)
    return res.data
  },
  async getVoiceToken(serverId, channelId) {
    const res = await api.post(`/api/servers/${serverId}/voice-token`, { channel_id: channelId })
    return res.data
  },

  // Channels
  async getChannels(serverId) {
    const res = await api.get(`/api/servers/${serverId}/channels`)
    return res.data
  },
  async createChannel(serverId, name, type = 'text') {
    const res = await api.post(`/api/servers/${serverId}/channels`, { name, type })
    return res.data
  },
  async updateChannel(channelId, data) {
    const res = await api.put(`/api/channels/${channelId}`, data)
    return res.data
  },
  async deleteChannel(channelId) {
    await api.delete(`/api/channels/${channelId}`)
  },
  async reorderChannels(serverId, channelIds) {
    const res = await api.put(`/api/servers/${serverId}/channels/reorder`, { channel_ids: channelIds })
    return res.data
  },

  // Messages
  async getMessages(channelId, before = null) {
    const params = before ? { before } : {}
    const res = await api.get(`/api/channels/${channelId}/messages`, { params })
    return res.data.map(normalizeMessage)
  },
  async sendMessage(channelId, content, file = null, replyToId = null, { imageWidth, imageHeight } = {}) {
    if (file) {
      const formData = new FormData()
      formData.append('file', file)
      if (content) formData.append('content', content)
      if (replyToId) formData.append('reply_to_id', replyToId)
      const res = await api.post(`/api/channels/${channelId}/messages`, formData)
      return normalizeMessage(res.data)
    }
    const body = { content }
    if (replyToId) body.reply_to_id = replyToId
    if (imageWidth) body.image_width = imageWidth
    if (imageHeight) body.image_height = imageHeight
    const res = await api.post(`/api/channels/${channelId}/messages`, body)
    return normalizeMessage(res.data)
  },
  async editMessage(messageId, content) {
    const res = await api.put(`/api/messages/${messageId}`, { content })
    return res.data
  },
  async deleteMessage(messageId) {
    await api.delete(`/api/messages/${messageId}`)
  },
  async bulkDeleteMessages(ids) {
    const results = await Promise.allSettled(ids.map((id) => api.delete(`/api/messages/${id}`)))
    const failed = results.filter((r) => r.status === 'rejected')
    return { total: ids.length, failed: failed.length }
  },
  async toggleReaction(messageId, emoji) {
    const res = await api.put(`/api/messages/${messageId}/reactions/${emoji}`)
    return res.data
  },

  // DMs
  async getDMChannels() {
    const res = await api.get('/api/dms')
    return res.data
  },
  async createOrGetDMChannel(userId) {
    const res = await api.post('/api/dms', { user_id: userId })
    return res.data
  },
  async getDMMessages(dmChannelId, before = null) {
    const params = before ? { before } : {}
    const res = await api.get(`/api/dms/${dmChannelId}/messages`, { params })
    return res.data.map(normalizeMessage)
  },
  async sendDMMessage(dmChannelId, content, file = null, { imageWidth, imageHeight } = {}) {
    if (file) {
      const formData = new FormData()
      formData.append('file', file)
      if (content) formData.append('content', content)
      const res = await api.post(`/api/dms/${dmChannelId}/messages`, formData)
      return normalizeMessage(res.data)
    }
    const body = { content }
    if (imageWidth) body.image_width = imageWidth
    if (imageHeight) body.image_height = imageHeight
    const res = await api.post(`/api/dms/${dmChannelId}/messages`, body)
    return normalizeMessage(res.data)
  },
  async toggleDMReaction(messageId, emoji) {
    const res = await api.put(`/api/dm-messages/${messageId}/reactions/${emoji}`)
    return res.data
  },
  async editDMMessage(messageId, content) {
    const res = await api.put(`/api/dm-messages/${messageId}`, { content })
    return res.data
  },
  async deleteDMMessage(messageId) {
    await api.delete(`/api/dm-messages/${messageId}`)
  },

  // Read State
  async markChannelAsRead(channelId, messageId) {
    await api.put(`/api/channels/${channelId}/read`, { message_id: messageId })
  },
  async markDMAsRead(dmChannelId, messageId) {
    await api.put(`/api/dms/${dmChannelId}/read`, { message_id: messageId })
  },
  async getUnreadCounts() {
    const res = await api.get('/api/unread')
    return res.data
  },

  // Notification Settings
  async getNotificationSettings() {
    const res = await api.get('/api/notification-settings')
    return res.data
  },
  async updateNotificationSetting(targetType, targetId, muted) {
    const res = await api.put('/api/notification-settings', {
      target_type: targetType,
      target_id: targetId,
      muted,
    })
    return res.data
  },

  // Pins
  async getPinnedMessages(channelId) {
    const res = await api.get(`/api/channels/${channelId}/pins`)
    return res.data
  },
  async pinMessage(messageId) {
    const res = await api.put(`/api/messages/${messageId}/pin`)
    return res.data
  },
  async unpinMessage(messageId) {
    await api.delete(`/api/messages/${messageId}/pin`)
  },

  // User profile
  async getUserProfile(userId) {
    const res = await api.get(`/api/users/${userId}`)
    return res.data
  },

  // Search
  async searchMessages(serverId, query, before = null) {
    const params = { q: query }
    if (before) params.before = before
    const res = await api.get(`/api/servers/${serverId}/search`, { params })
    return res.data.map(normalizeMessage)
  },

  // Federation
  async beginFederation(federatedId) {
    return api.post('/api/federation/begin', { federated_id: federatedId })
  },
  async federationAuthorize(visiting, nonce, callback) {
    return api.post('/api/federation/authorize', { visiting, nonce, callback })
  },
  async verifyFederation(token) {
    return api.post('/api/federation/verify', { token })
  },

  // Admin — Federation Policy
  async getFederationPolicy() {
    const res = await api.get('/api/admin/federation/policy')
    return res.data
  },
  async updateFederationPolicy(policy) {
    const res = await api.put('/api/admin/federation/policy', { policy })
    return res.data
  },
  async addInstancePolicy(domain, policy, note = '') {
    const res = await api.post('/api/admin/federation/instances', { domain, policy, note })
    return res.data
  },
  async removeInstancePolicy(domain) {
    await api.delete(`/api/admin/federation/instances/${domain}`)
  },

  // Channel Federation
  async enableChannelFederation(serverId, channelId) {
    const res = await api.post(`/api/servers/${serverId}/channels/${channelId}/federation`)
    return res.data
  },
  async disableChannelFederation(serverId, channelId) {
    await api.delete(`/api/servers/${serverId}/channels/${channelId}/federation`)
  },
  async linkRemoteChannel(serverId, channelId, remoteAddress) {
    const res = await api.post(`/api/servers/${serverId}/channels/${channelId}/federation/link`, { remote_address: remoteAddress })
    return res.data
  },
  async unlinkRemoteChannel(serverId, channelId, linkId) {
    await api.delete(`/api/servers/${serverId}/channels/${channelId}/federation/link/${linkId}`)
  },

  // Giphy
  async searchGifs(query, offset = 0) {
    const res = await api.get('/api/giphy/search', { params: { q: query, limit: 20, offset } })
    return res.data
  },
  async trendingGifs(offset = 0) {
    const res = await api.get('/api/giphy/trending', { params: { limit: 20, offset } })
    return res.data
  },

  // Forum posts
  async getForumPosts(channelId, before = null) {
    const params = before ? { before } : {}
    const res = await api.get(`/api/channels/${channelId}/posts`, { params })
    return res.data
  },
  async createForumPost(channelId, title, content) {
    const res = await api.post(`/api/channels/${channelId}/posts`, { title, content })
    return res.data
  },
  async getForumPost(postId) {
    const res = await api.get(`/api/forum-posts/${postId}`)
    return res.data
  },
  async editForumPost(postId, title, content) {
    const res = await api.put(`/api/forum-posts/${postId}`, { title, content })
    return res.data
  },
  async deleteForumPost(postId) {
    await api.delete(`/api/forum-posts/${postId}`)
  },
  async getForumPostMessages(postId, before = null) {
    const params = before ? { before } : {}
    const res = await api.get(`/api/forum-posts/${postId}/messages`, { params })
    return res.data.map(normalizeMessage)
  },
  async sendForumPostMessage(postId, content, replyToId = null, { imageWidth, imageHeight } = {}) {
    const body = { content }
    if (replyToId) body.reply_to_id = replyToId
    if (imageWidth) body.image_width = imageWidth
    if (imageHeight) body.image_height = imageHeight
    const res = await api.post(`/api/forum-posts/${postId}/messages`, body)
    return res.data
  },
}
