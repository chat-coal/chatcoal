<script setup>
import { ref, computed, onMounted } from 'vue'
import { useServersStore } from '@/stores/servers'
import { useToastStore } from '@/stores/toast'
import api, { API_URL } from '@/services/api.service'
import { getAvatarColor, getDefaultAvatarStyle, resolveFileUrl, cssBackgroundUrl } from '@/utils/avatar'
import { linkify } from '@/utils/linkify'
import LinkEmbedCard from './LinkEmbedCard.vue'

const props = defineProps({
  channelId: { type: String, required: true },
})

const emit = defineEmits(['close', 'scroll-to-message'])

const serversStore = useServersStore()
const toastStore = useToastStore()
const pins = ref([])
const loading = ref(true)

onMounted(async () => {
  try {
    pins.value = await api.getPinnedMessages(props.channelId)
  } finally {
    loading.value = false
  }
})

function formatTime(dateStr) {
  const d = new Date(dateStr)
  return d.toLocaleDateString() + ' ' + d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

function fileFullUrl(url) {
  if (!url) return ''
  if (url.startsWith('http')) return url
  return API_URL + url
}

const imageExts = ['.jpg', '.jpeg', '.png', '.gif', '.webp']
function isImage(url) {
  if (!url) return false
  return imageExts.some((ext) => url.toLowerCase().endsWith(ext))
}

const giphyUrlPattern = /^https?:\/\/media[0-9]*\.giphy\.com\/media\/[^\s]+\.gif(\?[^\s]*)?$/

function isGifMessage(msg) {
  const c = msg?.content?.trim()
  if (!c) return false
  return giphyUrlPattern.test(c)
}

function gifContainerStyle(msg) {
  const w = msg.image_width
  const h = msg.image_height
  if (!w || !h) return {}
  return {
    width: Math.min(w, 280) + 'px',
    maxWidth: '100%',
    aspectRatio: w + ' / ' + h,
    maxHeight: '150px',
  }
}

function getEmbeds(msg) {
  if (!msg?.embeds) return []
  if (typeof msg.embeds === 'string') {
    try { return JSON.parse(msg.embeds) } catch { return [] }
  }
  return msg.embeds
}

async function unpinMessage(pin) {
  try {
    await api.unpinMessage(pin.message_id)
    pins.value = pins.value.filter((p) => p.id !== pin.id)
  } catch {
    toastStore.add('Failed to unpin message')
  }
}

function addPin(pin) {
  if (!pins.value.find((p) => p.message_id === pin.message_id)) {
    pins.value.unshift(pin)
  }
}

function removePin(messageId) {
  pins.value = pins.value.filter((p) => p.message_id !== messageId)
}

defineExpose({ addPin, removePin })
</script>

<template>
  <div class="w-[320px] bg-[var(--surface)] border-l border-[var(--surface-border)] flex flex-col shrink-0 overflow-hidden">
    <!-- Header -->
    <div class="h-13 px-4 flex items-center justify-between border-b border-[var(--surface-border)] shrink-0">
      <div class="flex items-center gap-2">
        <svg class="w-4 h-4 text-[var(--text-3)]" fill="currentColor" viewBox="0 0 24 24">
          <path d="M16 3a1 1 0 0 1 .117 1.993L16 5h-.08l-.349 2.792A4.5 4.5 0 0 1 17 11v2.5h.5a1 1 0 0 1 .117 1.993L17.5 15.5h-4.5V21a1 1 0 0 1-1.993.117L11 21v-5.5H6.5a1 1 0 0 1-.117-1.993L6.5 13.5H7V11a4.5 4.5 0 0 1 1.429-3.208L8.08 5H8a1 1 0 0 1-.117-1.993L8 3h8Z" />
        </svg>
        <h3 class="text-[var(--text-1)] font-semibold text-sm">Pinned Messages</h3>
      </div>
      <button
        @click="emit('close')"
        class="text-[var(--text-4)] hover:text-[var(--text-1)] p-1.5 rounded-lg hover:bg-[var(--surface-3)] transition-colors duration-150 cursor-pointer"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="M18 6L6 18M6 6l12 12" />
        </svg>
      </button>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-y-auto">
      <div v-if="loading" class="flex items-center justify-center py-16">
        <div class="w-5 h-5 border-2 border-[var(--text-4)] border-t-[#E8521A] rounded-full animate-spin"></div>
      </div>

      <div v-else-if="pins.length === 0" class="flex flex-col items-center justify-center py-16 px-6 text-center">
        <div class="w-12 h-12 rounded-2xl bg-[var(--surface-2)] flex items-center justify-center mb-3">
          <svg class="w-6 h-6 text-[var(--text-4)]" fill="currentColor" viewBox="0 0 24 24">
            <path d="M16 3a1 1 0 0 1 .117 1.993L16 5h-.08l-.349 2.792A4.5 4.5 0 0 1 17 11v2.5h.5a1 1 0 0 1 .117 1.993L17.5 15.5h-4.5V21a1 1 0 0 1-1.993.117L11 21v-5.5H6.5a1 1 0 0 1-.117-1.993L6.5 13.5H7V11a4.5 4.5 0 0 1 1.429-3.208L8.08 5H8a1 1 0 0 1-.117-1.993L8 3h8Z" />
          </svg>
        </div>
        <p class="text-[var(--text-3)] text-sm">No pinned messages yet</p>
        <p class="text-[var(--text-4)] text-xs mt-1">Admins can pin important messages</p>
      </div>

      <div v-else class="p-3 space-y-2">
        <div
          v-for="pin in pins"
          :key="pin.id"
          class="bg-[var(--surface-2)] rounded-xl p-3 border border-[var(--surface-border)] group"
        >
          <!-- Author + time -->
          <div class="flex items-center gap-2 mb-2">
            <div class="shrink-0">
              <div
                v-if="pin.message?.author?.avatar_url"
                class="w-6 h-6 rounded-full bg-cover bg-center"
                :style="{ backgroundImage: cssBackgroundUrl(resolveFileUrl(pin.message.author.avatar_url)) }"
              ></div>
              <div
                v-else
                class="w-6 h-6 rounded-full flex items-center justify-center text-white text-[9px] font-bold"
                :style="getDefaultAvatarStyle(pin.message?.author_id || 0)"
              >
                {{ (pin.message?.author?.display_name || '?')[0].toUpperCase() }}
              </div>
            </div>
            <span class="text-[var(--text-1)] text-xs font-semibold truncate flex-1">
              {{ pin.message?.author?.display_name || 'Unknown' }}
            </span>
            <span class="text-[var(--text-4)] text-[10px] shrink-0">{{ formatTime(pin.message?.created_at) }}</span>
          </div>

          <!-- Content: inline GIF -->
          <div v-if="isGifMessage(pin.message)" class="mb-2">
            <div
              v-if="pin.message.image_width && pin.message.image_height"
              :style="gifContainerStyle(pin.message)"
              class="rounded-lg overflow-hidden bg-[var(--surface-3)]"
            >
              <img
                :src="pin.message.content.trim()"
                alt="GIF"
                class="w-full h-full object-cover"
              />
            </div>
            <img
              v-else
              :src="pin.message.content.trim()"
              alt="GIF"
              class="max-w-full max-h-[150px] rounded-lg object-cover"
            />
          </div>

          <!-- Content: linkified text -->
          <p v-else-if="pin.message?.content" class="text-[var(--text-2)] text-xs leading-relaxed break-words mb-2">
            <template v-for="(part, i) in linkify(pin.message.content)" :key="i">
              <a
                v-if="part.type === 'link'"
                :href="part.value"
                target="_blank"
                rel="noopener noreferrer"
                class="text-[#E8521A] hover:underline break-all"
              >{{ part.value }}</a>
              <template v-else>{{ part.value }}</template>
            </template>
          </p>

          <!-- Image attachment -->
          <div v-if="pin.message?.file_url && isImage(pin.message.file_url)" class="mb-2">
            <div
              v-if="pin.message.image_width && pin.message.image_height"
              :style="{
                width: Math.min(pin.message.image_width, 280) + 'px',
                maxWidth: '100%',
                aspectRatio: pin.message.image_width + ' / ' + pin.message.image_height,
                maxHeight: '150px',
              }"
              class="rounded-lg overflow-hidden bg-[var(--surface-3)]"
            >
              <img
                :src="fileFullUrl(pin.message.file_url)"
                :alt="pin.message.file_name"
                class="w-full h-full object-cover"
              />
            </div>
            <img
              v-else
              :src="fileFullUrl(pin.message.file_url)"
              :alt="pin.message.file_name"
              class="max-w-full max-h-[150px] rounded-lg object-cover"
            />
          </div>

          <!-- Link embed cards -->
          <LinkEmbedCard
            v-for="(embed, i) in getEmbeds(pin.message)"
            :key="i"
            :embed="embed"
            class="!max-w-full !mt-1"
          />

          <!-- Actions -->
          <div class="flex items-center gap-1 mt-2">
            <button
              @click="emit('scroll-to-message', pin.message_id)"
              class="flex-1 text-[10px] text-[var(--text-4)] hover:text-[var(--text-1)] py-1 px-2 rounded-lg hover:bg-[var(--surface-3)] transition-colors duration-100 cursor-pointer text-left"
            >
              Jump to message
            </button>
            <button
              v-if="serversStore.canManageMessages"
              @click="unpinMessage(pin)"
              class="text-[10px] text-[var(--text-4)] hover:text-[#E8521A] py-1 px-2 rounded-lg hover:bg-[#E8521A]/10 transition-colors duration-100 cursor-pointer opacity-0 group-hover:opacity-100"
              title="Unpin"
            >
              Unpin
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
