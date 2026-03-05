import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import api from '@/services/api.service'

export const useUnreadStore = defineStore('unread', () => {
  // channelUnread: { [channelId]: count }
  const channelUnread = ref({})
  // dmUnread: { [dmChannelId]: count }
  const dmUnread = ref({})
  // serverUnread: { [serverId]: count } — aggregate of channel unreads per server
  const serverUnread = ref({})

  const totalDMUnread = computed(() => {
    return Object.values(dmUnread.value).reduce((sum, c) => sum + c, 0)
  })

  async function fetchUnreadCounts() {
    const counts = await api.getUnreadCounts()
    const newChannel = {}
    const newDM = {}
    const newServer = {}
    for (const entry of counts) {
      if (entry.channel_type === 'server') {
        newChannel[entry.channel_ref_id] = entry.count
        newServer[entry.server_id] = (newServer[entry.server_id] || 0) + entry.count
      } else if (entry.channel_type === 'dm') {
        newDM[entry.channel_ref_id] = entry.count
      }
    }
    channelUnread.value = newChannel
    dmUnread.value = newDM
    serverUnread.value = newServer
  }

  function incrementChannel(channelId, serverId) {
    channelUnread.value = { ...channelUnread.value, [channelId]: (channelUnread.value[channelId] || 0) + 1 }
    if (serverId) {
      serverUnread.value = { ...serverUnread.value, [serverId]: (serverUnread.value[serverId] || 0) + 1 }
    }
  }

  function incrementDM(dmChannelId) {
    dmUnread.value = { ...dmUnread.value, [dmChannelId]: (dmUnread.value[dmChannelId] || 0) + 1 }
  }

  function markChannelRead(channelId, serverId) {
    const count = channelUnread.value[channelId] || 0
    const newCh = { ...channelUnread.value }
    delete newCh[channelId]
    channelUnread.value = newCh
    if (serverId && count > 0) {
      const newSrv = { ...serverUnread.value }
      newSrv[serverId] = Math.max(0, (newSrv[serverId] || 0) - count)
      if (newSrv[serverId] === 0) delete newSrv[serverId]
      serverUnread.value = newSrv
    }
  }

  function markDMRead(dmChannelId) {
    const newDM = { ...dmUnread.value }
    delete newDM[dmChannelId]
    dmUnread.value = newDM
  }

  function getServerUnread(serverId) {
    return serverUnread.value[serverId] || 0
  }

  function clear() {
    channelUnread.value = {}
    dmUnread.value = {}
    serverUnread.value = {}
  }

  return {
    channelUnread,
    dmUnread,
    serverUnread,
    totalDMUnread,
    fetchUnreadCounts,
    incrementChannel,
    incrementDM,
    markChannelRead,
    markDMRead,
    getServerUnread,
    clear,
  }
})
