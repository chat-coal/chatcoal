<script setup>
import { ref, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useSearchStore } from '@/stores/search'
import { useServersStore } from '@/stores/servers'
import { useChannelsStore } from '@/stores/channels'
import { getAvatarColor, getDefaultAvatarStyle, resolveFileUrl, cssBackgroundUrl } from '@/utils/avatar'

const router = useRouter()
const searchStore = useSearchStore()
const serversStore = useServersStore()
const channelsStore = useChannelsStore()
const inputRef = ref(null)
const localQuery = ref('')
let debounceTimer = null

onMounted(() => {
  inputRef.value?.focus()
})

watch(() => searchStore.isOpen, (open) => {
  if (open) {
    setTimeout(() => inputRef.value?.focus(), 50)
  }
})

function handleInput() {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    if (serversStore.currentServer && localQuery.value.trim()) {
      searchStore.search(serversStore.currentServer.id, localQuery.value.trim())
    } else {
      searchStore.results = []
    }
  }, 300)
}

function navigateToResult(result) {
  const channel = channelsStore.channels.find((c) => c.id === result.channel_id)
  if (channel) {
    channelsStore.selectChannel(channel)
    router.push(`/channels/${serversStore.currentServer.id}/${channel.id}`)
  }
  searchStore.close()
}

function formatTime(dateStr) {
  const d = new Date(dateStr)
  const today = new Date()
  if (d.toDateString() === today.toDateString()) {
    return 'Today at ' + d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }
  return d.toLocaleDateString() + ' ' + d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

function truncate(text, len = 120) {
  if (!text || text.length <= len) return text
  return text.slice(0, len) + '...'
}

</script>

<template>
  <div class="w-[360px] bg-[var(--surface)] border-l border-[var(--surface-border)] flex flex-col shrink-0">
    <!-- Header -->
    <div class="h-13 px-4 flex items-center gap-2 border-b border-[var(--surface-border)]">
      <svg class="w-4 h-4 text-[var(--text-4)] shrink-0" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" />
      </svg>
      <input
        ref="inputRef"
        v-model="localQuery"
        @input="handleInput"
        @keyup.escape="searchStore.close()"
        placeholder="Search messages..."
        class="flex-1 bg-transparent text-[var(--text-1)] text-sm outline-none placeholder-[var(--text-4)] border-none"
      />
      <button
        @click="searchStore.close()"
        class="text-[var(--text-4)] hover:text-[var(--text-1)] cursor-pointer p-1 rounded-lg hover:bg-[var(--surface-2)] transition-colors duration-100"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
        </svg>
      </button>
    </div>

    <!-- Results -->
    <div class="flex-1 overflow-y-auto px-3 py-3 scrollbar-light">
      <div v-if="searchStore.loading && searchStore.results.length === 0" class="text-center py-8">
        <p class="text-[var(--text-4)] text-sm">Searching...</p>
      </div>

      <div v-else-if="searchStore.results.length > 0" class="space-y-1">
        <button
          v-for="result in searchStore.results"
          :key="result.id"
          @click="navigateToResult(result)"
          class="w-full text-left p-3 rounded-xl hover:bg-[var(--surface-2)] cursor-pointer transition-colors duration-100"
        >
          <div class="flex items-center gap-1.5 mb-1">
            <span class="text-[#E8521A] text-xs font-medium">#{{ result.channel_name }}</span>
            <span class="text-[var(--text-4)] text-[10px]">&middot;</span>
            <span class="text-[var(--text-4)] text-[10px]">{{ formatTime(result.created_at) }}</span>
          </div>
          <div class="flex items-center gap-2 mb-1">
            <div
              v-if="result.author?.avatar_url"
              class="w-5 h-5 rounded-full bg-cover bg-center shrink-0"
              :style="{ backgroundImage: cssBackgroundUrl(resolveFileUrl(result.author.avatar_url)) }"
            ></div>
            <div
              v-else
              class="w-5 h-5 rounded-full flex items-center justify-center text-white text-[9px] font-bold shrink-0"
              :style="getDefaultAvatarStyle(result.author_id)"
            >
              {{ (result.author?.display_name || '?')[0].toUpperCase() }}
            </div>
            <span class="text-[var(--text-1)] text-xs font-medium">{{ result.author?.display_name || 'Unknown' }}</span>
          </div>
          <p class="text-[var(--text-3)] text-xs leading-relaxed">{{ truncate(result.content) }}</p>
        </button>

        <button
          v-if="searchStore.hasMore"
          @click="searchStore.loadMore(serversStore.currentServer?.id)"
          class="w-full text-center text-[#E8521A] text-xs py-2 hover:text-[#D44818] cursor-pointer font-medium"
        >
          Load more results
        </button>
      </div>

      <div v-else-if="searchStore.query && !searchStore.loading" class="text-center py-8">
        <p class="text-[var(--text-4)] text-sm">No results found</p>
      </div>

      <div v-else class="text-center py-8">
        <svg class="w-10 h-10 text-[var(--text-4)] mx-auto mb-3 opacity-40" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" />
        </svg>
        <p class="text-[var(--text-4)] text-sm">Search messages in this server</p>
      </div>
    </div>
  </div>
</template>
