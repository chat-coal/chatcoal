import { ref } from 'vue'
import { defineStore } from 'pinia'
import api from '@/services/api.service'

export const useSearchStore = defineStore('search', () => {
  const results = ref([])
  const query = ref('')
  const isOpen = ref(false)
  const loading = ref(false)
  const hasMore = ref(false)

  async function search(serverId, q) {
    query.value = q
    if (!q.trim()) {
      results.value = []
      hasMore.value = false
      return
    }
    loading.value = true
    try {
      const data = await api.searchMessages(serverId, q)
      results.value = data
      hasMore.value = data.length === 25
    } catch {
      results.value = []
    }
    loading.value = false
  }

  async function loadMore(serverId) {
    if (!hasMore.value || loading.value || results.value.length === 0) return
    const before = results.value[results.value.length - 1].id
    loading.value = true
    try {
      const data = await api.searchMessages(serverId, query.value, before)
      results.value = [...results.value, ...data]
      hasMore.value = data.length === 25
    } catch {
      // ignore
    }
    loading.value = false
  }

  function open() {
    isOpen.value = true
  }

  function close() {
    isOpen.value = false
    results.value = []
    query.value = ''
    hasMore.value = false
  }

  return { results, query, isOpen, loading, hasMore, search, loadMore, open, close }
})
