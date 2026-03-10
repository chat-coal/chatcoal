<script setup>
import { ref, onMounted, onBeforeUnmount, nextTick, watch } from 'vue'
import api from '@/services/api.service'

const emit = defineEmits(['select', 'close'])

const query = ref('')
const gifs = ref([])
const loading = ref(false)
const loadingMore = ref(false)
const offset = ref(0)
const hasMore = ref(true)
const searchInput = ref(null)
const scrollContainer = ref(null)
const pickerRef = ref(null)
let debounceTimer = null

onMounted(() => {
  loadTrending()
  nextTick(() => searchInput.value?.focus())
  document.addEventListener('mousedown', onClickOutside)
})

onBeforeUnmount(() => {
  document.removeEventListener('mousedown', onClickOutside)
  clearTimeout(debounceTimer)
})

function onClickOutside(e) {
  if (pickerRef.value && !pickerRef.value.contains(e.target)) {
    emit('close')
  }
}

async function loadTrending() {
  loading.value = true
  try {
    const res = await api.trendingGifs(0)
    gifs.value = res.data || []
    offset.value = gifs.value.length
    hasMore.value = (res.pagination?.total_count || 0) > offset.value
  } catch {
    gifs.value = []
  } finally {
    loading.value = false
  }
}

watch(query, (val) => {
  clearTimeout(debounceTimer)
  if (!val.trim()) {
    debounceTimer = setTimeout(() => {
      offset.value = 0
      loadTrending()
    }, 300)
    return
  }
  debounceTimer = setTimeout(() => {
    offset.value = 0
    search()
  }, 300)
})

async function search() {
  const q = query.value.trim()
  if (!q) return
  loading.value = true
  try {
    const res = await api.searchGifs(q, 0)
    gifs.value = res.data || []
    offset.value = gifs.value.length
    hasMore.value = (res.pagination?.total_count || 0) > offset.value
  } catch {
    gifs.value = []
  } finally {
    loading.value = false
  }
}

async function loadMore() {
  if (loadingMore.value || !hasMore.value) return
  loadingMore.value = true
  try {
    const q = query.value.trim()
    const res = q ? await api.searchGifs(q, offset.value) : await api.trendingGifs(offset.value)
    const newGifs = res.data || []
    gifs.value.push(...newGifs)
    offset.value += newGifs.length
    hasMore.value = (res.pagination?.total_count || 0) > offset.value
  } catch {
    // ignore
  } finally {
    loadingMore.value = false
  }
}

function onScroll() {
  const el = scrollContainer.value
  if (!el) return
  if (el.scrollTop + el.clientHeight >= el.scrollHeight - 100) {
    loadMore()
  }
}

function selectGif(gif) {
  const img = gif.images?.original || gif.images?.downsized
  if (!img?.url) return
  emit('select', {
    url: img.url,
    width: parseInt(img.width) || 0,
    height: parseInt(img.height) || 0,
  })
}

function getPreviewUrl(gif) {
  return gif.images?.fixed_width?.url || gif.images?.preview_gif?.url || gif.images?.original?.url
}

function getPreviewAspect(gif) {
  const img = gif.images?.fixed_width
  const w = parseInt(img?.width) || 200
  const h = parseInt(img?.height) || 150
  return `${w} / ${h}`
}
</script>

<template>
  <div
    ref="pickerRef"
    class="absolute bottom-full mb-2 left-0 right-0 bg-[var(--card)] border border-[var(--surface-border)] rounded-2xl shadow-2xl shadow-black/15 overflow-hidden z-50 flex flex-col"
    style="height: 400px; max-height: 60vh"
  >
    <!-- Search bar -->
    <div class="p-3 border-b border-[var(--surface-border)]">
      <input
        ref="searchInput"
        v-model="query"
        placeholder="Search GIFs..."
        class="w-full bg-[var(--surface-2)] text-[var(--text-1)] placeholder-[var(--text-4)] px-3 py-2 rounded-xl text-sm outline-none border border-[var(--surface-border)] focus:border-[#E8521A]/40 transition-colors duration-150"
      />
    </div>

    <!-- GIF grid -->
    <div ref="scrollContainer" @scroll="onScroll" class="flex-1 overflow-y-auto p-2">
      <div v-if="loading" class="flex items-center justify-center h-full">
        <div class="text-[var(--text-4)] text-sm">Loading...</div>
      </div>

      <div v-else-if="!gifs.length" class="flex items-center justify-center h-full">
        <div class="text-[var(--text-4)] text-sm">No GIFs found</div>
      </div>

      <div v-else class="columns-2 gap-1.5">
        <button
          v-for="gif in gifs"
          :key="gif.id"
          @click="selectGif(gif)"
          class="w-full mb-1.5 rounded-lg overflow-hidden cursor-pointer hover:opacity-80 transition-opacity duration-100 break-inside-avoid block"
        >
          <img
            :src="getPreviewUrl(gif)"
            :alt="gif.title"
            :style="{ aspectRatio: getPreviewAspect(gif) }"
            class="w-full object-cover bg-[var(--surface-3)]"
            loading="lazy"
          />
        </button>
      </div>

      <div v-if="loadingMore" class="py-3 text-center">
        <span class="text-[var(--text-4)] text-xs">Loading more...</span>
      </div>
    </div>

    <!-- Powered by Giphy attribution -->
    <div class="px-3 py-2 border-t border-[var(--surface-border)] flex items-center justify-end">
      <span class="text-[10px] text-[var(--text-4)] mr-1.5">Powered by</span>
      <svg viewBox="0 0 163 35" class="h-3.5 opacity-50" fill="currentColor" xmlns="http://www.w3.org/2000/svg">
        <text x="0" y="28" font-size="28" font-weight="bold" fill="var(--text-4)">GIPHY</text>
      </svg>
    </div>
  </div>
</template>
