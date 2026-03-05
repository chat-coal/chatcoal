<script setup>
import { ref, computed, watch, nextTick } from 'vue'
import { useChannelsStore } from '@/stores/channels'
import { useServersStore } from '@/stores/servers'
import { useMessagesStore } from '@/stores/messages'
import { useDMsStore } from '@/stores/dms'
import { useAuthStore } from '@/stores/auth'
import { useForumStore } from '@/stores/forum'
import { sendTyping, sendDMTyping } from '@/services/websocket.service'

const props = defineProps({
  mode: {
    type: String,
    default: 'server', // 'server' or 'dm' or 'forum'
  },
  replyingTo: {
    type: Object,
    default: null,
  },
  forumPostId: {
    type: [String, Number],
    default: null,
  },
})

const emit = defineEmits(['cancel-reply'])

const channelsStore = useChannelsStore()
const serversStore = useServersStore()
const messagesStore = useMessagesStore()
const dmsStore = useDMsStore()
const authStore = useAuthStore()
const forumStore = useForumStore()
const content = ref('')
const selectedFile = ref(null)
const fileInputRef = ref(null)
const textInputRef = ref(null)
let typingTimeout = null

watch(() => props.replyingTo, (val) => {
  if (val) nextTick(() => textInputRef.value?.focus())
})

const placeholder = computed(() => {
  if (props.mode === 'dm') {
    const dm = dmsStore.currentDMChannel
    if (!dm || !authStore.dbUser) return 'Message'
    const other = dm.user1?.id === authStore.dbUser.id ? dm.user2 : dm.user1
    return `Message @${other?.display_name || 'User'}`
  }
  return `Message #${channelsStore.currentChannel?.name || ''}`
})

const canSend = computed(() => content.value.trim() || selectedFile.value)

function formatFileSize(bytes) {
  if (!bytes) return ''
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

function openFilePicker() {
  fileInputRef.value?.click()
}

function onFileSelected(e) {
  const file = e.target.files?.[0]
  if (!file) return
  if (file.size > 25 * 1024 * 1024) {
    alert('File must be under 25MB')
    return
  }
  selectedFile.value = file
  // Reset input so same file can be re-selected
  e.target.value = ''
}

function removeFile() {
  selectedFile.value = null
}

function send() {
  if (!canSend.value) return
  const text = content.value.trim()
  const file = selectedFile.value
  const replyId = props.replyingTo?.id || null

  // Clear input immediately for snappy UX
  content.value = ''
  selectedFile.value = null
  if (replyId) emit('cancel-reply')

  // Send in background — errors shown on the message itself
  if (props.mode === 'forum') {
    forumStore.sendMessage(props.forumPostId, text, replyId)
  } else if (props.mode === 'dm') {
    dmsStore.sendMessage(dmsStore.currentDMChannel.id, text, file)
  } else {
    messagesStore.sendMessage(channelsStore.currentChannel.id, text, file, replyId)
  }
}

function handleInput() {
  if (!typingTimeout) {
    if (props.mode === 'dm') {
      const dm = dmsStore.currentDMChannel
      if (dm && authStore.dbUser) {
        const otherUserId = dm.user1_id === authStore.dbUser.id ? dm.user2_id : dm.user1_id
        sendDMTyping(dm.id, otherUserId)
      }
    } else {
      sendTyping(channelsStore.currentChannel.id, serversStore.currentServer?.id)
    }
  }
  clearTimeout(typingTimeout)
  typingTimeout = setTimeout(() => {
    typingTimeout = null
  }, 2000)
}
</script>

<template>
  <div class="px-5 pb-5 pt-1">
    <!-- Replying to bar -->
    <div v-if="replyingTo" class="mb-2 flex items-center gap-2 bg-[var(--surface-2)] rounded-xl px-3 py-2 border border-[var(--surface-border)]">
      <svg class="w-3.5 h-3.5 text-[var(--text-4)] shrink-0" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" d="M9 15 3 9m0 0 6-6M3 9h12a6 6 0 0 1 0 12h-3" />
      </svg>
      <span class="text-sm text-[var(--text-3)]">Replying to</span>
      <span class="text-sm text-[var(--text-1)] font-medium">{{ replyingTo.author?.display_name || 'Unknown' }}</span>
      <span class="text-xs text-[var(--text-4)] truncate flex-1">{{ replyingTo.content?.slice(0, 50) }}</span>
      <button @click="emit('cancel-reply')" class="text-[var(--text-4)] hover:text-[#E8521A] cursor-pointer p-0.5 rounded transition-colors duration-100">
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
        </svg>
      </button>
    </div>

    <!-- File preview bar -->
    <div v-if="selectedFile" class="mb-2 flex items-center gap-2 bg-[var(--surface-2)] rounded-xl px-3 py-2 border border-[var(--surface-border)]">
      <svg class="w-4 h-4 text-[var(--text-3)] shrink-0" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z" />
      </svg>
      <span class="text-sm text-[var(--text-2)] truncate flex-1">{{ selectedFile.name }}</span>
      <span class="text-[11px] text-[var(--text-4)] shrink-0">{{ formatFileSize(selectedFile.size) }}</span>
      <button @click="removeFile" class="text-[var(--text-4)] hover:text-[#E8521A] cursor-pointer p-0.5 rounded transition-colors duration-100">
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
        </svg>
      </button>
    </div>

    <div class="bg-[var(--card)] rounded-2xl flex items-center px-4 shadow-lg shadow-black/[0.04] border border-[var(--surface-border)]">
      <!-- File attach button -->
      <button
        @click="openFilePicker"
        class="text-[var(--text-4)] hover:text-[var(--text-1)] cursor-pointer transition-colors duration-150 p-1.5 -ml-1 rounded-lg hover:bg-[var(--surface-2)]"
        title="Attach file"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="m18.375 12.739-7.693 7.693a4.5 4.5 0 0 1-6.364-6.364l10.94-10.94A3 3 0 1 1 19.5 7.372L8.552 18.32m.009-.01-.01.01m5.699-9.941-7.81 7.81a1.5 1.5 0 0 0 2.112 2.13" />
        </svg>
      </button>
      <input
        ref="fileInputRef"
        type="file"
        class="hidden"
        accept=".jpg,.jpeg,.png,.gif,.webp,.pdf,.txt,.zip"
        @change="onFileSelected"
      />

      <input
        ref="textInputRef"
        v-model="content"
        @keyup.enter="send"
        @input="handleInput"
        :placeholder="placeholder"
        class="flex-1 bg-transparent text-[var(--text-1)] py-3.5 outline-none placeholder-[var(--text-4)] text-sm border-none ml-1"
      />
      <button
        @click="send"
        :disabled="!canSend"
        class="text-[#E8521A] hover:text-[#D44818] disabled:text-[var(--text-disabled)] cursor-pointer transition-colors duration-150 ml-2 -mr-1 p-1.5 rounded-lg"
        :class="canSend ? 'hover:bg-[#E8521A]/10' : ''"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="M6 12L3.269 3.126A59.768 59.768 0 0121.485 12 59.77 59.77 0 013.27 20.876L5.999 12zm0 0h7.5" />
        </svg>
      </button>
    </div>
  </div>
</template>
