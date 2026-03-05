import { ref } from 'vue'
import { defineStore } from 'pinia'
import api from '@/services/api.service'

export const useVoiceStore = defineStore('voice', () => {
  // channelId -> [userIds]
  const voiceStates = ref({})
  const currentVoiceChannelId = ref(null)
  const isMuted = ref(false)
  const isDeafened = ref(false)
  const isJoining = ref(false)
  const connectionMode = ref(null) // 'livekit' | 'p2p' | null

  // Push-to-talk state
  const inputMode = ref(localStorage.getItem('voiceInputMode') || 'voice_activity') // 'voice_activity' | 'push_to_talk'
  const pttKey = ref(localStorage.getItem('voicePttKey') || 'Space')
  const pttActive = ref(false)
  const globalPttAvailable = ref(false) // true when Electron global keyboard hook is active

  function handleVoiceStateUpdate({ channel_id, user_id, action }) {
    const chId = String(channel_id)
    const uid = String(user_id)
    if (action === 'join') {
      if (!voiceStates.value[chId]) {
        voiceStates.value[chId] = []
      }
      if (!voiceStates.value[chId].some((id) => String(id) === uid)) {
        voiceStates.value[chId].push(uid)
      }
    } else if (action === 'leave') {
      if (voiceStates.value[chId]) {
        voiceStates.value[chId] = voiceStates.value[chId].filter(
          (id) => String(id) !== uid,
        )
        if (voiceStates.value[chId].length === 0) {
          delete voiceStates.value[chId]
        }
      }
    }
  }

  async function fetchVoiceStates(serverId) {
    try {
      const raw = await api.getVoiceStates(serverId)
      // Normalize all IDs to strings for consistent comparisons
      const normalized = {}
      for (const [chId, userIds] of Object.entries(raw || {})) {
        normalized[String(chId)] = (userIds || []).map(String)
      }
      voiceStates.value = normalized
    } catch {
      voiceStates.value = {}
    }
  }

  function setCurrentVoiceChannel(channelId) {
    currentVoiceChannelId.value = channelId
  }

  function toggleMute() {
    isMuted.value = !isMuted.value
  }

  function toggleDeafen() {
    isDeafened.value = !isDeafened.value
    if (isDeafened.value) {
      isMuted.value = true
    }
  }

  function setInputMode(mode) {
    inputMode.value = mode
    localStorage.setItem('voiceInputMode', mode)
  }

  function setPttKey(key) {
    pttKey.value = key
    localStorage.setItem('voicePttKey', key)
  }

  function clear() {
    voiceStates.value = {}
    currentVoiceChannelId.value = null
    isMuted.value = false
    isDeafened.value = false
    isJoining.value = false
    connectionMode.value = null
    pttActive.value = false
  }

  return {
    voiceStates,
    currentVoiceChannelId,
    isMuted,
    isDeafened,
    isJoining,
    connectionMode,
    inputMode,
    pttKey,
    pttActive,
    globalPttAvailable,
    handleVoiceStateUpdate,
    fetchVoiceStates,
    setCurrentVoiceChannel,
    toggleMute,
    toggleDeafen,
    setInputMode,
    setPttKey,
    clear,
  }
})
