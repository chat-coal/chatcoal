<script setup>
import { ref, computed, nextTick } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useServersStore } from '@/stores/servers'
import { useMessagesStore } from '@/stores/messages'
import { useDMsStore } from '@/stores/dms'
import { useForumStore } from '@/stores/forum'
import { useToastStore } from '@/stores/toast'
import api, { API_URL } from '@/services/api.service'
import { getAvatarColor, getDefaultAvatarStyle, resolveFileUrl, cssBackgroundUrl } from '@/utils/avatar'
import { linkify } from '@/utils/linkify'
import UserProfilePopover from './UserProfilePopover.vue'
import LinkEmbedCard from './LinkEmbedCard.vue'

const props = defineProps({
  message: { type: Object, required: true },
  mode: { type: String, default: 'server' }, // 'server' or 'dm' or 'forum'
  pinnedMessageIds: { type: Set, default: () => new Set() },
  serverMember: { type: Object, default: null },
  selecting: { type: Boolean, default: false },
  selected: { type: Boolean, default: false },
})

const emit = defineEmits(['reply', 'scroll-to-message', 'pin', 'unpin', 'toggle-select'])

const authStore = useAuthStore()
const serversStore = useServersStore()
const messagesStore = useMessagesStore()
const dmsStore = useDMsStore()
const forumStore = useForumStore()
const toastStore = useToastStore()
const editing = ref(false)
const editContent = ref('')
const editInput = ref(null)
const hovering = ref(false)
const showEmojiPicker = ref(false)
const showDeleteConfirm = ref(false)
const profileAnchor = ref(null)
const showProfile = ref(false)

function openProfile(event) {
  if (props.mode === 'dm') return
  profileAnchor.value = event.currentTarget
  showProfile.value = true
}

const isPinned = computed(() => props.pinnedMessageIds.has(props.message.id))

async function pinMsg() {
  try {
    const pin = await api.pinMessage(props.message.id)
    emit('pin', pin)
  } catch {
    toastStore.add('Failed to pin message')
  }
}

async function unpinMsg() {
  try {
    await api.unpinMessage(props.message.id)
    emit('unpin', props.message.id)
  } catch {
    toastStore.add('Failed to unpin message')
  }
}

const quickEmojis = ['👍', '❤️', '😂', '😮', '😢', '😡', '🔥', '👀']

const imageExts = ['.jpg', '.jpeg', '.png', '.gif', '.webp']

const isImage = computed(() => {
  if (!props.message.file_url) return false
  const url = props.message.file_url.toLowerCase()
  return imageExts.some((ext) => url.endsWith(ext))
})

const imageContainerStyle = computed(() => {
  const w = props.message.image_width
  const h = props.message.image_height
  if (!w || !h) return {}
  const maxW = 400
  const maxH = 300
  const displayW = Math.min(w, maxW)
  const displayH = Math.min((displayW / w) * h, maxH)
  return {
    width: displayW + 'px',
    maxWidth: 'calc(100vw - 6rem)',
    aspectRatio: w + ' / ' + h,
    maxHeight: maxH + 'px',
  }
})

const fileFullUrl = computed(() => {
  if (!props.message.file_url) return ''
  if (props.message.file_url.startsWith('http')) return props.message.file_url
  return API_URL + props.message.file_url
})

function formatFileSize(bytes) {
  if (!bytes) return ''
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

function formatTime(dateStr) {
  const d = new Date(dateStr)
  const today = new Date()
  if (d.toDateString() === today.toDateString()) {
    return 'Today at ' + d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }
  return d.toLocaleDateString() + ' ' + d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

function startEdit() {
  editContent.value = props.message.content
  editing.value = true
  nextTick(() => editInput.value?.focus())
}

async function saveEdit() {
  if (!editContent.value.trim()) return
  try {
    if (props.mode === 'dm') {
      const updated = await api.editDMMessage(props.message.id, editContent.value)
      dmsStore.updateMessage(updated)
    } else {
      const updated = await api.editMessage(props.message.id, editContent.value)
      messagesStore.updateMessage(updated)
    }
    editing.value = false
  } catch {
    toastStore.add('Failed to edit message')
  }
}

function confirmDelete() {
  showDeleteConfirm.value = true
}

async function deleteMsg() {
  showDeleteConfirm.value = false
  try {
    if (props.mode === 'dm') {
      await api.deleteDMMessage(props.message.id)
    } else {
      await api.deleteMessage(props.message.id)
    }
  } catch {
    toastStore.add('Failed to delete message')
  }
}

async function toggleReaction(emoji) {
  showEmojiPicker.value = false
  try {
    if (props.mode === 'dm') {
      await dmsStore.toggleReaction(props.message.id, emoji)
    } else {
      await messagesStore.toggleReaction(props.message.id, emoji)
    }
  } catch {
    toastStore.add('Failed to add reaction')
  }
}

function hasUserReacted(reaction) {
  return reaction.user_ids?.includes(authStore.dbUser?.id)
}

const isSystemMessage = computed(() => props.message.type && props.message.type !== 'user')

const isOptimistic = computed(() => props.message._sending || props.message._failed)

function dismissFailed() {
  if (props.mode === 'dm') {
    dmsStore.removeMessage(props.message.id)
  } else if (props.mode === 'forum') {
    forumStore.removeMessage(props.message.id)
  } else {
    messagesStore.removeMessage(props.message.id)
  }
}

const giphyUrlPattern = /^https?:\/\/media[0-9]*\.giphy\.com\/media\/[^\s]+\.gif(\?[^\s]*)?$/

const isGifMessage = computed(() => {
  const c = props.message.content?.trim()
  if (!c) return false
  return giphyUrlPattern.test(c)
})

const gifLoaded = ref(false)

const gifContainerStyle = computed(() => {
  const w = props.message.image_width
  const h = props.message.image_height
  if (!w || !h) return {}
  const maxW = 400
  const maxH = 300
  const displayW = Math.min(w, maxW)
  const displayH = Math.min((displayW / w) * h, maxH)
  return {
    width: displayW + 'px',
    maxWidth: 'calc(100vw - 6rem)',
    aspectRatio: w + ' / ' + h,
    maxHeight: maxH + 'px',
  }
})

const contentParts = computed(() => linkify(props.message.content))
const embeds = computed(() => {
  if (!props.message.embeds) return []
  if (typeof props.message.embeds === 'string') {
    try { return JSON.parse(props.message.embeds) } catch { return [] }
  }
  return props.message.embeds
})

const isAuthor = () => authStore.dbUser?.id === props.message.author_id
const canDelete = () => isAuthor() || (props.mode === 'server' && serversStore.canManageMessages)
</script>

<template>
  <!-- System message (join/leave) -->
  <div v-if="isSystemMessage" class="flex items-center gap-2 py-1 px-3 -mx-3">
    <!-- Join icon -->
    <svg v-if="message.type === 'join'" class="w-4 h-4 text-emerald-500 shrink-0" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
    </svg>
    <!-- Leave icon -->
    <svg v-else class="w-4 h-4 text-[var(--text-4)] shrink-0" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" d="M15.75 9V5.25A2.25 2.25 0 0 0 13.5 3h-6a2.25 2.25 0 0 0-2.25 2.25v13.5A2.25 2.25 0 0 0 7.5 21h6a2.25 2.25 0 0 0 2.25-2.25V15m3 0 3-3m0 0-3-3m3 3H9" />
    </svg>
    <p class="text-sm text-[var(--text-3)]">
      <span class="font-semibold text-[var(--text-1)]">{{ message.author?.display_name || 'Unknown' }}</span>
      {{ message.content }}
    </p>
    <span class="text-[var(--text-4)] text-[11px] shrink-0">{{ formatTime(message.created_at) }}</span>
  </div>

  <!-- Regular user message -->
  <div
    v-else
    class="flex gap-3.5 py-1.5 px-3 -mx-3 rounded-xl group transition-colors duration-100"
    :class="[
      selecting && selected ? 'bg-[#E8521A]/8' : hovering ? 'bg-[var(--surface-2)]' : '',
      message._sending ? 'opacity-50' : '',
      selecting ? 'cursor-pointer' : '',
    ]"
    @mouseenter="hovering = true"
    @mouseleave="hovering = false; showEmojiPicker = false"
    @click="selecting && !isSystemMessage ? emit('toggle-select', message.id) : null"
  >
    <!-- Selection checkbox -->
    <div v-if="selecting" class="shrink-0 flex items-center self-center">
      <div
        class="w-5 h-5 rounded-md border-2 flex items-center justify-center transition-all duration-100"
        :class="selected
          ? 'bg-[#E8521A] border-[#E8521A]'
          : 'border-[var(--surface-border)] hover:border-[var(--text-4)]'"
      >
        <svg v-if="selected" class="w-3 h-3 text-white" fill="none" stroke="currentColor" stroke-width="3" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" />
        </svg>
      </div>
    </div>

    <!-- Avatar -->
    <div class="shrink-0 mt-0.5">
      <div
        v-if="message.author?.avatar_url"
        class="w-9 h-9 rounded-full bg-cover bg-center cursor-pointer"
        :style="{ backgroundImage: cssBackgroundUrl(resolveFileUrl(message.author.avatar_url)) }"
        @click="openProfile"
      ></div>
      <div
        v-else
        class="w-9 h-9 rounded-full flex items-center justify-center text-white font-bold text-xs cursor-pointer"
        :style="getDefaultAvatarStyle(message.author_id)"
        @click="openProfile"
      >
        {{ (message.author?.display_name || '?')[0].toUpperCase() }}
      </div>
    </div>

    <!-- Content -->
    <div class="flex-1 min-w-0">
      <!-- Reply reference -->
      <button
        v-if="message.reply_to"
        @click="emit('scroll-to-message', message.reply_to.id)"
        class="flex items-center gap-1.5 mb-1 text-xs text-[var(--text-4)] hover:text-[var(--text-2)] cursor-pointer transition-colors duration-100"
      >
        <svg class="w-3 h-3 shrink-0 opacity-60" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="M9 15 3 9m0 0 6-6M3 9h12a6 6 0 0 1 0 12h-3" />
        </svg>
        <span class="font-medium text-[var(--text-2)]">{{ message.reply_to.author?.display_name || 'Unknown' }}</span>
        <span class="truncate max-w-[300px] opacity-75">{{ message.reply_to.content || 'Message deleted' }}</span>
      </button>

      <div class="flex items-baseline gap-2 flex-wrap">
        <button
          class="font-semibold text-sm cursor-pointer hover:underline"
          :class="message.author?.is_anonymous ? 'text-[var(--text-3)]' : 'text-[var(--text-1)]'"
          @click="openProfile"
        >{{ message.author?.display_name || 'Unknown' }}</button>
        <span
          v-if="message.author?.home_instance"
          class="text-[10px] font-medium text-[var(--text-4)] bg-[var(--surface-2)] px-1.5 py-0.5 rounded-full leading-none"
          :title="`From ${message.author.home_instance}`"
        >@{{ message.author.home_instance }}</span>
        <span v-if="message._sending" class="text-[var(--text-4)] text-[11px] italic">Sending...</span>
        <span v-else-if="message._failed" class="text-red-400 text-[11px]">Failed to send</span>
        <span v-else class="text-[var(--text-4)] text-[11px]">{{ formatTime(message.created_at) }}</span>
        <span v-if="message.edited" class="text-[var(--text-4)] text-[11px] italic">(edited)</span>
        <span v-if="isPinned" class="text-[#E8521A] text-[10px]" title="Pinned">📌</span>
      </div>

      <div v-if="editing" class="mt-1.5">
        <input
          ref="editInput"
          v-model="editContent"
          @keyup.enter="saveEdit"
          @keyup.escape="editing = false"
          class="w-full bg-[var(--card)] text-[var(--text-1)] px-3 py-2 rounded-xl border border-[var(--surface-border)] text-sm"
        />
        <div class="text-[11px] text-[var(--text-4)] mt-1.5 flex gap-1">
          escape to <button @click="editing = false" class="text-[#E8521A] hover:underline cursor-pointer font-medium">cancel</button>
          <span class="mx-0.5">&middot;</span>
          enter to <button @click="saveEdit" class="text-[#E8521A] hover:underline cursor-pointer font-medium">save</button>
        </div>
      </div>
      <!-- Inline GIF (Giphy URL as entire message) -->
      <div v-else-if="isGifMessage" class="mt-1">
        <div
          v-if="message.image_width && message.image_height"
          :style="gifContainerStyle"
          class="rounded-xl overflow-hidden bg-[var(--surface-3)] relative"
        >
          <div v-if="!gifLoaded" class="absolute inset-0 animate-pulse bg-[var(--surface-3)]" />
          <img
            :src="message.content.trim()"
            alt="GIF"
            class="w-full h-full object-cover"
            @load="gifLoaded = true"
          />
        </div>
        <img
          v-else
          :src="message.content.trim()"
          alt="GIF"
          class="max-w-full sm:max-w-[400px] max-h-[300px] rounded-xl object-cover"
        />
      </div>

      <p v-else-if="message.content" class="text-[var(--text-2)] text-sm leading-relaxed break-words">
        <template v-for="(part, i) in contentParts" :key="i">
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

      <!-- File attachment: inline image -->
      <div v-if="message.file_url && isImage" class="mt-2">
        <div
          v-if="message.image_width && message.image_height"
          :style="imageContainerStyle"
          class="rounded-xl overflow-hidden bg-[var(--surface-3)]"
        >
          <img
            :src="fileFullUrl"
            :alt="message.file_name"
            class="w-full h-full object-cover cursor-pointer"
            @click="() => globalThis.open(fileFullUrl, '_blank')"
          />
        </div>
        <img
          v-else
          :src="fileFullUrl"
          :alt="message.file_name"
          class="max-w-full sm:max-w-[400px] max-h-[300px] rounded-xl object-cover cursor-pointer"
          @click="() => globalThis.open(fileFullUrl, '_blank')"
        />
      </div>

      <!-- File attachment: download card -->
      <div v-else-if="message.file_url" class="mt-2">
        <a
          :href="fileFullUrl"
          target="_blank"
          class="inline-flex items-center gap-2.5 bg-[var(--surface-2)] hover:bg-[var(--surface-3)] px-3.5 py-2.5 rounded-xl transition-colors duration-100 border border-[var(--surface-border)]"
        >
          <svg class="w-5 h-5 text-[var(--text-3)] shrink-0" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z" />
          </svg>
          <div class="min-w-0">
            <div class="text-sm text-[#E8521A] font-medium truncate">{{ message.file_name }}</div>
            <div class="text-[11px] text-[var(--text-4)]">{{ formatFileSize(message.file_size) }}</div>
          </div>
        </a>
      </div>

      <!-- Link embed cards -->
      <LinkEmbedCard v-for="(embed, i) in embeds" :key="i" :embed="embed" />

      <!-- Reaction pills -->
      <div v-if="message.reactions?.length" class="flex flex-wrap gap-1 mt-1.5">
        <button
          v-for="reaction in message.reactions"
          :key="reaction.emoji"
          @click="toggleReaction(reaction.emoji)"
          class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs cursor-pointer transition-colors duration-100 border"
          :class="hasUserReacted(reaction)
            ? 'bg-[#E8521A]/15 border-[#E8521A]/40 text-[#E8521A]'
            : 'bg-[var(--surface-2)] border-[var(--surface-border)] text-[var(--text-3)] hover:bg-[var(--surface-3)]'"
        >
          <span>{{ reaction.emoji }}</span>
          <span class="font-medium">{{ reaction.count }}</span>
        </button>
      </div>
    </div>

    <!-- Dismiss failed message -->
    <div v-if="message._failed" class="flex items-start shrink-0 -mt-0.5">
      <button
        @click="dismissFailed"
        class="p-1.5 text-[var(--text-4)] hover:text-[var(--text-1)] hover:bg-[var(--surface-3)] rounded-lg cursor-pointer transition-colors duration-100"
        title="Dismiss"
      >
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
        </svg>
      </button>
    </div>

    <!-- Action buttons -->
    <div v-if="hovering && !editing && !isOptimistic && !selecting" class="flex items-start gap-0.5 shrink-0 -mt-0.5 relative">
      <!-- Emoji reaction button -->
      <button
        @click="showEmojiPicker = !showEmojiPicker"
        class="p-1.5 text-[var(--text-4)] hover:text-[var(--text-1)] hover:bg-[var(--surface-3)] rounded-lg cursor-pointer transition-colors duration-100"
        title="Add reaction"
      >
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="M15.182 15.182a4.5 4.5 0 0 1-6.364 0M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0ZM9.75 9.75c0 .414-.168.75-.375.75S9 10.164 9 9.75 9.168 9 9.375 9s.375.336.375.75Zm-.375 0h.008v.015h-.008V9.75Zm5.625 0c0 .414-.168.75-.375.75s-.375-.336-.375-.75.168-.75.375-.75.375.336.375.75Zm-.375 0h.008v.015h-.008V9.75Z" />
        </svg>
      </button>

      <!-- Emoji quick picker -->
      <div
        v-if="showEmojiPicker"
        class="absolute right-0 top-8 z-50 bg-[var(--card)] border border-[var(--surface-border)] rounded-xl shadow-xl p-1.5 flex gap-0.5"
      >
        <button
          v-for="emoji in quickEmojis"
          :key="emoji"
          @click="toggleReaction(emoji)"
          class="w-8 h-8 flex items-center justify-center rounded-lg hover:bg-[var(--surface-2)] cursor-pointer transition-colors duration-100 text-base"
        >
          {{ emoji }}
        </button>
      </div>

      <!-- Pin / Unpin button (admins, server mode only) -->
      <button
        v-if="mode === 'server' && serversStore.canManageMessages && !isPinned"
        @click="pinMsg"
        class="p-1.5 text-[var(--text-4)] hover:text-[var(--text-1)] hover:bg-[var(--surface-3)] rounded-lg cursor-pointer transition-colors duration-100"
        title="Pin message"
      >
        <svg class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 24 24">
          <path d="M16 3a1 1 0 0 1 .117 1.993L16 5h-.08l-.349 2.792A4.5 4.5 0 0 1 17 11v2.5h.5a1 1 0 0 1 .117 1.993L17.5 15.5h-4.5V21a1 1 0 0 1-1.993.117L11 21v-5.5H6.5a1 1 0 0 1-.117-1.993L6.5 13.5H7V11a4.5 4.5 0 0 1 1.429-3.208L8.08 5H8a1 1 0 0 1-.117-1.993L8 3h8Z" />
        </svg>
      </button>
      <button
        v-if="mode === 'server' && serversStore.canManageMessages && isPinned"
        @click="unpinMsg"
        class="p-1.5 text-[#E8521A] hover:bg-[#E8521A]/10 rounded-lg cursor-pointer transition-colors duration-100"
        title="Unpin message"
      >
        <svg class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 24 24">
          <path d="M16 3a1 1 0 0 1 .117 1.993L16 5h-.08l-.349 2.792A4.5 4.5 0 0 1 17 11v2.5h.5a1 1 0 0 1 .117 1.993L17.5 15.5h-4.5V21a1 1 0 0 1-1.993.117L11 21v-5.5H6.5a1 1 0 0 1-.117-1.993L6.5 13.5H7V11a4.5 4.5 0 0 1 1.429-3.208L8.08 5H8a1 1 0 0 1-.117-1.993L8 3h8Z" />
        </svg>
      </button>

      <button
        v-if="mode !== 'dm'"
        @click="emit('reply', message)"
        class="p-1.5 text-[var(--text-4)] hover:text-[var(--text-1)] hover:bg-[var(--surface-3)] rounded-lg cursor-pointer transition-colors duration-100"
        title="Reply"
      >
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="M9 15 3 9m0 0 6-6M3 9h12a6 6 0 0 1 0 12h-3" />
        </svg>
      </button>
      <button
        v-if="isAuthor()"
        @click="startEdit"
        class="p-1.5 text-[var(--text-4)] hover:text-[var(--text-1)] hover:bg-[var(--surface-3)] rounded-lg cursor-pointer transition-colors duration-100"
        title="Edit"
      >
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L10.582 16.07a4.5 4.5 0 0 1-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 0 1 1.13-1.897l8.932-8.931Z" />
        </svg>
      </button>
      <button
        v-if="canDelete()"
        @click="confirmDelete"
        class="p-1.5 text-[var(--text-4)] hover:text-[#E8521A] hover:bg-[#E8521A]/10 rounded-lg cursor-pointer transition-colors duration-100"
        title="Delete"
      >
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0" />
        </svg>
      </button>
    </div>
  </div>

  <!-- Delete confirmation modal -->
  <Teleport to="body">
    <div v-if="showDeleteConfirm" class="fixed inset-0 bg-[var(--backdrop)] backdrop-blur-sm flex items-center justify-center z-[60]" @click.self="showDeleteConfirm = false">
      <div class="bg-[var(--modal-bg)] rounded-2xl p-7 w-full max-w-sm shadow-2xl shadow-black/10 animate-scale-in border border-[var(--modal-border)]">
        <h2 class="font-display text-xl font-bold text-[var(--text-1)] mb-2">Delete Message</h2>
        <p class="text-[var(--text-3)] text-sm mb-5">
          Are you sure you want to delete this message? This cannot be undone.
        </p>
        <div class="flex justify-end gap-3">
          <button @click="showDeleteConfirm = false" class="text-[var(--text-3)] hover:text-[var(--text-1)] px-4 py-2.5 rounded-xl cursor-pointer font-medium transition-colors duration-150">
            Cancel
          </button>
          <button
            @click="deleteMsg"
            class="bg-[#E8521A] text-white px-5 py-2.5 rounded-xl hover:bg-[#D44818] cursor-pointer font-semibold shadow-lg shadow-[#E8521A]/15 transition-all duration-200"
          >
            Delete
          </button>
        </div>
      </div>
    </div>
  </Teleport>

  <!-- User profile popover -->
  <UserProfilePopover
    v-if="showProfile && mode === 'server'"
    :user-id="message.author_id"
    :anchor-el="profileAnchor"
    :server-member="serverMember"
    @close="showProfile = false"
  />
</template>
