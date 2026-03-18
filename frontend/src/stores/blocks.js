import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '@/services/api.service'

export const useBlocksStore = defineStore('blocks', () => {
  const blockedUsers = ref([])
  const blockedUserIds = computed(() => new Set(blockedUsers.value.map(b => String(b.blocked_id))))

  async function fetchBlockedUsers() {
    blockedUsers.value = await api.getBlockedUsers()
  }

  async function blockUser(userId) {
    await api.blockUser(userId)
    await fetchBlockedUsers()
  }

  async function unblockUser(userId) {
    await api.unblockUser(userId)
    blockedUsers.value = blockedUsers.value.filter(b => String(b.blocked_id) !== String(userId))
  }

  function isBlocked(userId) {
    return blockedUserIds.value.has(String(userId))
  }

  return { blockedUsers, blockedUserIds, fetchBlockedUsers, blockUser, unblockUser, isBlocked }
})
