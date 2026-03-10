import { ref } from 'vue'
import { defineStore } from 'pinia'
import api from '@/services/api.service'
import { useAuthStore } from '@/stores/auth'

export const useMessagesStore = defineStore('messages', () => {
  const messages = ref([])
  const loading = ref(false)
  const hasMore = ref(true)
  const replyingTo = ref(null)
  let tempIdCounter = 0

  async function fetchMessages(channelId, before = null) {
    loading.value = true
    try {
      const fetched = await api.getMessages(channelId, before)
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

  async function sendMessage(channelId, content, file = null, replyToId = null, gifDims = {}) {
    const authStore = useAuthStore()
    const tempId = `_temp_${Date.now()}_${++tempIdCounter}`
    const optimistic = {
      id: tempId,
      content: content || '',
      author_id: authStore.dbUser?.id,
      author: authStore.dbUser ? { ...authStore.dbUser } : null,
      channel_id: channelId,
      created_at: new Date().toISOString(),
      reply_to_id: replyToId || null,
      reply_to: replyToId && replyingTo.value ? {
        id: replyingTo.value.id,
        content: replyingTo.value.content,
        author: replyingTo.value.author,
      } : null,
      image_width: gifDims.imageWidth || 0,
      image_height: gifDims.imageHeight || 0,
      reactions: [],
      _sending: true,
    }
    messages.value.push(optimistic)

    try {
      const message = await api.sendMessage(channelId, content, file, replyToId, gifDims)
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

  function setReplyingTo(message) {
    replyingTo.value = message
  }

  function clearReplyingTo() {
    replyingTo.value = null
  }

  async function toggleReaction(messageId, emoji) {
    const res = await api.toggleReaction(messageId, emoji)
    const idx = messages.value.findIndex((m) => m.id === messageId)
    if (idx !== -1) {
      messages.value[idx] = { ...messages.value[idx], reactions: res.reactions }
    }
  }

  function addMessage(message) {
    if (!messages.value.find((m) => m.id === message.id)) {
      messages.value.push(message)
    }
  }

  function updateMessage(updated) {
    const idx = messages.value.findIndex((m) => m.id === updated.id)
    if (idx !== -1) {
      messages.value[idx] = updated
    }
  }

  function removeMessage(id) {
    messages.value = messages.value.filter((m) => m.id !== id)
  }

  function clear() {
    messages.value = []
    hasMore.value = true
    replyingTo.value = null
  }

  return { messages, loading, hasMore, replyingTo, fetchMessages, sendMessage, addMessage, updateMessage, removeMessage, toggleReaction, setReplyingTo, clearReplyingTo, clear }
})
