import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import api from '@/services/api.service'
import { subscribe, unsubscribe } from '@/services/websocket.service'

export const useServersStore = defineStore('servers', () => {
  const servers = ref([])
  const currentServer = ref(null)
  const currentMemberRole = ref(null)

  const isOwner = computed(() => currentMemberRole.value === 'owner')
  const isAdmin = computed(() => currentMemberRole.value === 'owner' || currentMemberRole.value === 'admin')
  const canManageChannels = computed(() => isAdmin.value)
  const canManageMessages = computed(() => isAdmin.value)
  const canManageInvites = computed(() => isAdmin.value)

  function canKick(targetRole) {
    const levels = { owner: 3, admin: 2, member: 1 }
    return isAdmin.value && levels[currentMemberRole.value] > (levels[targetRole] || 0)
  }

  function canBan(targetRole) {
    const levels = { owner: 3, admin: 2, member: 1 }
    return isAdmin.value && levels[currentMemberRole.value] > (levels[targetRole] || 0)
  }

  async function fetchServers() {
    servers.value = await api.getServers()
  }

  async function createServer(name, isPublic = false) {
    const server = await api.createServer(name, isPublic)
    servers.value.push(server)
    subscribe(server.id)
    return server
  }

  async function joinServer(inviteCode) {
    const server = await api.joinServer(inviteCode)
    if (!servers.value.find((s) => s.id === server.id)) {
      servers.value.push(server)
      subscribe(server.id)
    }
    return server
  }

  async function joinPublicServer(serverId) {
    const server = await api.joinPublicServer(serverId)
    if (!servers.value.find((s) => s.id === server.id)) {
      servers.value.push(server)
      subscribe(server.id)
    }
    return server
  }

  async function updateServer(serverId, { name, iconFile, clearIcon, isPublic, showJoinLeave, systemChannelId }) {
    const updated = await api.updateServer(serverId, { name, iconFile, clearIcon, isPublic, showJoinLeave, systemChannelId })
    const idx = servers.value.findIndex((s) => s.id === serverId)
    if (idx !== -1) servers.value[idx] = updated
    if (currentServer.value?.id === serverId) currentServer.value = updated
    return updated
  }

  async function deleteServer(serverId) {
    await api.deleteServer(serverId)
    servers.value = servers.value.filter((s) => s.id !== serverId)
    unsubscribe(serverId)
    if (currentServer.value?.id === serverId) currentServer.value = null
  }

  async function leaveServer(serverId) {
    await api.leaveServer(serverId)
    servers.value = servers.value.filter((s) => s.id !== serverId)
    unsubscribe(serverId)
    if (currentServer.value?.id === serverId) {
      currentServer.value = null
    }
  }

  function removeServer(serverId) {
    servers.value = servers.value.filter((s) => s.id !== serverId)
    unsubscribe(serverId)
    if (currentServer.value?.id === serverId) currentServer.value = null
  }

  function selectServer(server) {
    currentServer.value = server
    if (!server) currentMemberRole.value = null
  }

  function setMemberRole(role) {
    currentMemberRole.value = role
  }

  function patchCurrentServer(fields) {
    if (currentServer.value) {
      currentServer.value = { ...currentServer.value, ...fields }
      const idx = servers.value.findIndex((s) => s.id === currentServer.value.id)
      if (idx !== -1) servers.value[idx] = currentServer.value
    }
  }

  function clear() {
    servers.value = []
    currentServer.value = null
    currentMemberRole.value = null
  }

  return {
    servers, currentServer, currentMemberRole,
    isOwner, isAdmin, canManageChannels, canManageMessages, canManageInvites, canKick, canBan,
    fetchServers, createServer, joinServer, joinPublicServer, updateServer, deleteServer, leaveServer, removeServer, selectServer, setMemberRole, patchCurrentServer, clear,
  }
})
