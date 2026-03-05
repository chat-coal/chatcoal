import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '@/services/api.service'

export const useNotificationSettingsStore = defineStore('notificationSettings', () => {
  const settings = ref([])

  const mutedServers = computed(() => {
    const set = new Set()
    for (const s of settings.value) {
      if (s.target_type === 'server' && s.muted) set.add(s.target_id)
    }
    return set
  })

  const mutedChannels = computed(() => {
    const set = new Set()
    for (const s of settings.value) {
      if (s.target_type === 'channel' && s.muted) set.add(s.target_id)
    }
    return set
  })

  function isEffectivelyMuted(channelId, serverId) {
    return mutedChannels.value.has(channelId) || mutedServers.value.has(serverId)
  }

  async function fetchSettings() {
    try {
      settings.value = await api.getNotificationSettings()
    } catch {
      // silent fail
    }
  }

  async function toggleMute(targetType, targetId) {
    const existing = settings.value.find(
      (s) => s.target_type === targetType && s.target_id === targetId,
    )
    const newMuted = !(existing?.muted)

    try {
      const result = await api.updateNotificationSetting(targetType, targetId, newMuted)
      // Update local state
      const idx = settings.value.findIndex(
        (s) => s.target_type === targetType && s.target_id === targetId,
      )
      if (idx >= 0) {
        settings.value[idx] = result
      } else {
        settings.value.push(result)
      }
    } catch {
      // silent fail
    }
  }

  return {
    settings,
    mutedServers,
    mutedChannels,
    isEffectivelyMuted,
    fetchSettings,
    toggleMute,
  }
})
