import { ref } from 'vue'
import { defineStore } from 'pinia'
import api from '@/services/api.service'

export const useInvitesStore = defineStore('invites', () => {
  const invites = ref([])

  async function fetchInvites(serverId) {
    invites.value = await api.getInvites(serverId)
  }

  async function createInvite(serverId, options = {}) {
    const invite = await api.createInvite(serverId, options)
    invites.value.unshift(invite)
    return invite
  }

  async function deleteInvite(serverId, inviteId) {
    await api.deleteInvite(serverId, inviteId)
    invites.value = invites.value.filter((i) => i.id !== inviteId)
  }

  function clear() {
    invites.value = []
  }

  return { invites, fetchInvites, createInvite, deleteInvite, clear }
})
