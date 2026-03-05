import { ref } from 'vue'
import { defineStore } from 'pinia'
import api from '@/services/api.service'
import { useAuthStore } from '@/stores/auth'

export const useDMsStore = defineStore('dms', () => {
  const dmChannels = ref([])
  const currentDMChannel = ref(null)
  const messages = ref([])
  const loading = ref(false)
  const hasMore = ref(true)
  let tempIdCounter = 0

  async function fetchDMChannels() {
    const fetched = await api.getDMChannels()
    dmChannels.value = fetched
  }

  async function openDM(userId) {
    const channel = await api.createOrGetDMChannel(userId)
    // Add to list if not already there
    if (!dmChannels.value.find((c) => c.id === channel.id)) {
      dmChannels.value.unshift(channel)
    }
    return channel
  }

  function selectDM(channel) {
    currentDMChannel.value = channel
  }

  async function fetchMessages(dmChannelId, before = null) {
    loading.value = true
    try {
      const fetched = await api.getDMMessages(dmChannelId, before)
      if (before) {
        messages.value = [...fetched.reverse(), ...messages.value]
      } else {
        messages.value = fetched.reverse()
      }
      hasMore.value = fetched.length === 50
    } finally {
      loading.value = false
    }
  }

  async function sendMessage(dmChannelId, content, file = null) {
    const authStore = useAuthStore()
    const tempId = `_temp_${Date.now()}_${++tempIdCounter}`
    const optimistic = {
      id: tempId,
      content: content || '',
      author_id: authStore.dbUser?.id,
      author: authStore.dbUser ? { ...authStore.dbUser } : null,
      dm_channel_id: dmChannelId,
      created_at: new Date().toISOString(),
      reactions: [],
      _sending: true,
    }
    messages.value.push(optimistic)

    try {
      const message = await api.sendDMMessage(dmChannelId, content, file)
      if (messages.value.find(m => m.id === tempId)) {
        messages.value = messages.value.filter(m => m.id !== tempId)
        addMessage(message)
      }
    } catch (err) {
      const idx = messages.value.findIndex(m => m.id === tempId)
      if (idx !== -1) {
        messages.value[idx] = { ...messages.value[idx], _sending: false, _failed: true, _error: err?.response?.data?.error || 'Failed to send' }
      }
    }
  }

  async function toggleReaction(messageId, emoji) {
    const res = await api.toggleDMReaction(messageId, emoji)
    const idx = messages.value.findIndex((m) => m.id === messageId)
    if (idx !== -1) {
      messages.value[idx] = { ...messages.value[idx], reactions: res.reactions }
    }
  }

  function addMessage(message) {
    if (!messages.value.find((m) => m.id === message.id)) {
      messages.value.push(message)
    }
    // Update last_message on the channel in our list
    const ch = dmChannels.value.find((c) => c.id === message.dm_channel_id)
    if (ch) {
      ch.last_message = {
        id: message.id,
        content: message.content,
        author_id: message.author_id,
        created_at: message.created_at,
      }
      // Move to top
      dmChannels.value = [ch, ...dmChannels.value.filter((c) => c.id !== ch.id)]
    }
  }

  function updateMessage(updated) {
    const idx = messages.value.findIndex((m) => m.id === updated.id)
    if (idx !== -1) {
      messages.value[idx] = updated
    }
    // Update sidebar last_message if this is the latest message
    const ch = dmChannels.value.find((c) => c.id === updated.dm_channel_id)
    if (ch && ch.last_message && ch.last_message.id === updated.id) {
      ch.last_message.content = updated.content
    }
  }

  function removeMessage(id) {
    messages.value = messages.value.filter((m) => m.id !== id)
  }

  function clear() {
    messages.value = []
    hasMore.value = true
  }

  function clearAll() {
    messages.value = []
    hasMore.value = true
    dmChannels.value = []
    currentDMChannel.value = null
  }

  return {
    dmChannels,
    currentDMChannel,
    messages,
    loading,
    hasMore,
    fetchDMChannels,
    openDM,
    selectDM,
    fetchMessages,
    sendMessage,
    addMessage,
    updateMessage,
    removeMessage,
    toggleReaction,
    clear,
    clearAll,
  }
})
