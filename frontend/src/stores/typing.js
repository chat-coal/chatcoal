import { ref } from 'vue'
import { defineStore } from 'pinia'

export const useTypingStore = defineStore('typing', () => {
  // { "channel:5": { userId: timeoutId }, "dm:3": { userId: timeoutId } }
  const typers = ref({})

  function _key(type, channelId) {
    return `${type}:${channelId}`
  }

  function addTyper(type, channelId, userId) {
    const key = _key(type, channelId)
    const current = { ...typers.value[key] }

    // Clear existing timeout for this user
    if (current[userId]) clearTimeout(current[userId])

    // Auto-expire after 3 seconds
    current[userId] = setTimeout(() => {
      removeTyper(type, channelId, userId)
    }, 3000)

    typers.value = { ...typers.value, [key]: current }
  }

  function removeTyper(type, channelId, userId) {
    const key = _key(type, channelId)
    const current = typers.value[key]
    if (!current || !current[userId]) return

    clearTimeout(current[userId])
    const updated = { ...current }
    delete updated[userId]

    if (Object.keys(updated).length === 0) {
      const newTypers = { ...typers.value }
      delete newTypers[key]
      typers.value = newTypers
    } else {
      typers.value = { ...typers.value, [key]: updated }
    }
  }

  function getTyperIds(type, channelId) {
    const key = _key(type, channelId)
    const current = typers.value[key]
    return current ? Object.keys(current) : []
  }

  return { typers, addTyper, removeTyper, getTyperIds }
})
