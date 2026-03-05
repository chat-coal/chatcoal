import { ref } from 'vue'
import { defineStore } from 'pinia'
import api from '@/services/api.service'

export const useChannelsStore = defineStore('channels', () => {
  const channels = ref([])
  const currentChannel = ref(null)
  const loading = ref(false)

  async function fetchChannels(serverId) {
    loading.value = true
    try {
      channels.value = await api.getChannels(serverId)
    } finally {
      loading.value = false
    }
  }

  async function createChannel(serverId, name, type = 'text') {
    const channel = await api.createChannel(serverId, name, type)
    addChannel(channel)
    return channel
  }

  async function updateChannel(channelId, data) {
    const updated = await api.updateChannel(channelId, data)
    const idx = channels.value.findIndex((c) => c.id === updated.id)
    if (idx !== -1) channels.value[idx] = updated
    if (currentChannel.value?.id === updated.id) currentChannel.value = updated
    return updated
  }

  function addChannel(channel) {
    if (!channels.value.find((c) => String(c.id) === String(channel.id))) {
      channels.value.push(channel)
    }
  }

  function removeChannel(channelId) {
    channels.value = channels.value.filter((c) => c.id !== channelId)
    if (currentChannel.value?.id === channelId) {
      currentChannel.value = null
    }
  }

  function selectChannel(channel) {
    currentChannel.value = channel
  }

  function updateChannelLocal(channelId, data) {
    const idx = channels.value.findIndex((c) => String(c.id) === String(channelId))
    if (idx !== -1) channels.value[idx] = { ...channels.value[idx], ...data }
    if (currentChannel.value && String(currentChannel.value.id) === String(channelId)) {
      currentChannel.value = { ...currentChannel.value, ...data }
    }
  }

  async function reorderChannels(serverId, channelIds) {
    await api.reorderChannels(serverId, channelIds)
  }

  function applyReorder(updatedChannels) {
    // Build a position map from the server-provided channel list
    const posMap = new Map()
    for (const ch of updatedChannels) {
      posMap.set(String(ch.id), ch.position)
    }
    // Update positions on existing channel objects
    for (const ch of channels.value) {
      const pos = posMap.get(String(ch.id))
      if (pos !== undefined) ch.position = pos
    }
    // Re-sort by position, then id
    channels.value.sort((a, b) => a.position - b.position || String(a.id).localeCompare(String(b.id)))
  }

  function clear() {
    channels.value = []
    currentChannel.value = null
  }

  return { channels, currentChannel, loading, fetchChannels, createChannel, updateChannel, updateChannelLocal, addChannel, removeChannel, selectChannel, reorderChannels, applyReorder, clear }
})
