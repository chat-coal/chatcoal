<script setup>
import { ref, nextTick, watch, onMounted, onUnmounted, computed, inject } from 'vue'
import { useRoute } from 'vue-router'
import { useChannelsStore } from '@/stores/channels'
import { useMessagesStore } from '@/stores/messages'
import { useVoiceStore } from '@/stores/voice'
import { useAuthStore } from '@/stores/auth'
import { useServersStore } from '@/stores/servers'
import { useDMsStore } from '@/stores/dms'
import { useUnreadStore } from '@/stores/unread'
import { useTypingStore } from '@/stores/typing'
import { useSearchStore } from '@/stores/search'
import api from '@/services/api.service'
import { getAvatarColor, getDefaultAvatarStyle } from '@/utils/avatar'
import { joinChannel, leaveChannel, setMuted, setDeafened } from '@/services/voice.service'
import { on } from '@/services/websocket.service'
import { usePushToTalk } from '@/composables/usePushToTalk'
import MessageItem from './MessageItem.vue'
import MessageInput from './MessageInput.vue'
import ForumView from './ForumView.vue'
import PinnedMessagesPanel from './PinnedMessagesPanel.vue'

const props = defineProps({
  showMembers: Boolean,
  initialLoading: Boolean,
})
const emit = defineEmits(['toggle-members', 'open-nav'])

const openMobileMembers = inject('openMobileMembers', () => {})

const route = useRoute()
const channelsStore = useChannelsStore()
const messagesStore = useMessagesStore()
const voiceStore = useVoiceStore()
const authStore = useAuthStore()
const serversStore = useServersStore()
const dmsStore = useDMsStore()
const unreadStore = useUnreadStore()
const typingStore = useTypingStore()
const searchStore = useSearchStore()
usePushToTalk()
const messagesContainer = ref(null)
const members = ref([])
const showPinnedPanel = ref(false)
const pinnedPanel = ref(null)
const pinnedMessageIds = ref(new Set())
const selecting = ref(false)

const headerTooltip = ref({ visible: false, text: '', top: 0, left: 0 })
let headerHideTimeout = null

function showHeaderTooltip(event, text) {
  if (window.matchMedia('(hover: none)').matches) return
  clearTimeout(headerHideTimeout)
  const rect = event.currentTarget.getBoundingClientRect()
  headerTooltip.value = {
    visible: true,
    text,
    top: rect.bottom + 8,
    left: rect.left + rect.width / 2,
  }
}

function hideHeaderTooltip() {
  headerHideTimeout = setTimeout(() => {
    headerTooltip.value.visible = false
  }, 50)
}
const selectedIds = ref(new Set())
const showBulkDeleteConfirm = ref(false)
const bulkDeleting = ref(false)

const isDMMode = computed(() => {
  return dmsStore.currentDMChannel != null
})

const isAudioChannel = computed(
  () => !isDMMode.value && channelsStore.currentChannel?.type === 'audio',
)

const isForumChannel = computed(
  () => !isDMMode.value && channelsStore.currentChannel?.type === 'forum',
)

const isInThisChannel = computed(
  () => voiceStore.currentVoiceChannelId === channelsStore.currentChannel?.id,
)

const connectedUsers = computed(() => {
  if (!channelsStore.currentChannel) return []
  return voiceStore.voiceStates[channelsStore.currentChannel.id] || []
})

const activeMessages = computed(() => {
  return isDMMode.value ? dmsStore.messages : messagesStore.messages
})

const activeLoading = computed(() => {
  return isDMMode.value ? dmsStore.loading : messagesStore.loading
})

const activeHasMore = computed(() => {
  return isDMMode.value ? dmsStore.hasMore : messagesStore.hasMore
})

const dmOtherUser = computed(() => {
  if (!isDMMode.value || !dmsStore.currentDMChannel || !authStore.dbUser) return null
  const dm = dmsStore.currentDMChannel
  return dm.user1?.id === authStore.dbUser.id ? dm.user2 : dm.user1
})

const headerTitle = computed(() => {
  if (isDMMode.value) {
    return dmOtherUser.value?.display_name || 'Direct Message'
  }
  return channelsStore.currentChannel?.name || 'Select a channel'
})

const hasActiveChannel = computed(() => {
  return isDMMode.value ? !!dmsStore.currentDMChannel : !!channelsStore.currentChannel
})

const typingText = computed(() => {
  let ids
  if (isDMMode.value && dmsStore.currentDMChannel) {
    ids = typingStore.getTyperIds('dm', dmsStore.currentDMChannel.id)
  } else if (channelsStore.currentChannel) {
    ids = typingStore.getTyperIds('channel', channelsStore.currentChannel.id)
  } else {
    return ''
  }
  if (ids.length === 0) return ''

  const names = ids.map((id) => {
    if (isDMMode.value) {
      return dmOtherUser.value?.display_name || dmOtherUser.value?.username || 'Someone'
    }
    const member = members.value.find((m) => m.user_id === id || m.user?.id === id)
    return member?.user?.display_name || member?.user?.username || 'Someone'
  })

  if (names.length === 1) return `${names[0]} is typing`
  if (names.length === 2) return `${names[0]} and ${names[1]} are typing`
  return 'Several people are typing'
})

watch(
  () => serversStore.currentServer,
  async (server) => {
    if (server) {
      try {
        members.value = await api.getServerMembers(server.id)
      } catch {
        members.value = []
      }
    }
  },
  { immediate: true },
)

// Keep members in sync with real-time events (mirrors MemberList.vue handlers)
const offMemberJoin = on('member_join', (member, serverId) => {
  if (serverId !== serversStore.currentServer?.id) return
  if (!members.value.find((m) => m.id === member.id)) {
    members.value.push(member)
  }
})
const offMemberLeave = on('member_leave', (member, serverId) => {
  if (serverId !== serversStore.currentServer?.id) return
  members.value = members.value.filter((m) => m.id !== member.id)
})
const offMemberUpdate = on('member_update', (updated) => {
  const idx = members.value.findIndex((m) => m.id === updated.id)
  if (idx !== -1) {
    members.value[idx] = { ...members.value[idx], ...updated }
  }
})
const offUserUpdate = on('user_update', (data) => {
  const idx = members.value.findIndex((m) => m.user_id === data.user_id)
  if (idx !== -1 && members.value[idx].user) {
    members.value[idx] = { ...members.value[idx], user: { ...members.value[idx].user, display_name: data.display_name, avatar_url: data.avatar_url } }
  }
})
const offConnected = on('connected', async () => {
  const server = serversStore.currentServer
  if (server) {
    try { members.value = await api.getServerMembers(server.id) } catch { /* keep existing */ }
  }
})
onUnmounted(() => {
  offMemberJoin()
  offMemberLeave()
  offMemberUpdate()
  offUserUpdate()
  offConnected()
})

function getMember(userId) {
  const member = members.value.find((m) => String(m.user_id) === String(userId) || String(m.user?.id) === String(userId))
  return member?.user
}

async function handleJoinVoice() {
  await joinChannel(channelsStore.currentChannel.id, serversStore.currentServer?.id)
  if (voiceStore.inputMode === 'push_to_talk') {
    voiceStore.isMuted = true
    setMuted(true)
  }
}

function handleLeaveVoice() {
  headerTooltip.value.visible = false
  leaveChannel()
}

function handleToggleMute() {
  voiceStore.toggleMute()
  setMuted(voiceStore.isMuted)
  if (headerTooltip.value.visible) {
    headerTooltip.value.text = voiceStore.isMuted ? 'Unmute' : 'Mute'
  }
}

function handleToggleDeafen() {
  voiceStore.toggleDeafen()
  setMuted(voiceStore.isMuted)
  setDeafened(voiceStore.isDeafened)
  if (headerTooltip.value.visible) {
    headerTooltip.value.text = voiceStore.isDeafened ? 'Undeafen' : 'Deafen'
  }
}

function handleReply(message) {
  messagesStore.setReplyingTo(message)
}

function handleCancelReply() {
  messagesStore.clearReplyingTo()
}

function scrollToMessage(messageId) {
  const el = messagesContainer.value?.querySelector(`[data-message-id="${messageId}"]`)
  if (el) {
    el.scrollIntoView({ behavior: 'smooth', block: 'center' })
    el.classList.add('bg-[var(--surface-2)]')
    setTimeout(() => el.classList.remove('bg-[var(--surface-2)]'), 1500)
  }
}

function scrollToBottom() {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

watch(() => activeMessages.value.length, scrollToBottom)

async function loadMore() {
  if (!activeHasMore.value || activeLoading.value) return
  const oldest = activeMessages.value[0]
  if (oldest) {
    if (isDMMode.value) {
      await dmsStore.fetchMessages(dmsStore.currentDMChannel.id, oldest.id)
    } else {
      await messagesStore.fetchMessages(channelsStore.currentChannel.id, oldest.id)
    }
  }
}

// Auto-mark as read when DM channel is active
watch(
  () => dmsStore.currentDMChannel,
  async (dm) => {
    if (dm) {
      unreadStore.markDMRead(dm.id)
      // Mark as read on server if there are messages
      if (dmsStore.messages.length > 0) {
        const lastMsg = dmsStore.messages[dmsStore.messages.length - 1]
        try { await api.markDMAsRead(dm.id, lastMsg.id) } catch {}
      }
    }
  },
)

// Auto-mark server channel as read when channel is active
watch(
  () => channelsStore.currentChannel,
  async (channel) => {
    showPinnedPanel.value = false
    pinnedMessageIds.value = new Set()
    if (channel && serversStore.currentServer) {
      unreadStore.markChannelRead(channel.id, serversStore.currentServer.id)
      if (messagesStore.messages.length > 0) {
        const lastMsg = messagesStore.messages[messagesStore.messages.length - 1]
        try { await api.markChannelAsRead(channel.id, lastMsg.id) } catch {}
      }
    }
  },
)

function handlePin(pin) {
  pinnedMessageIds.value = new Set([...pinnedMessageIds.value, pin.message_id])
  pinnedPanel.value?.addPin(pin)
}

function handleUnpin(messageId) {
  const next = new Set(pinnedMessageIds.value)
  next.delete(messageId)
  pinnedMessageIds.value = next
  pinnedPanel.value?.removePin(messageId)
}

function handlePinEvent(pin) {
  handlePin(pin)
}

function handleUnpinEvent(messageId) {
  handleUnpin(messageId)
}

function enterSelectMode() {
  selecting.value = true
  selectedIds.value = new Set()
}

function exitSelectMode() {
  selecting.value = false
  selectedIds.value = new Set()
}

function toggleSelect(id) {
  const next = new Set(selectedIds.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  selectedIds.value = next
}

async function bulkDelete() {
  bulkDeleting.value = true
  const ids = [...selectedIds.value]
  // Optimistically remove
  ids.forEach((id) => messagesStore.removeMessage(id))
  await api.bulkDeleteMessages(ids)
  showBulkDeleteConfirm.value = false
  bulkDeleting.value = false
  exitSelectMode()
}

// Exit selection mode when switching channels
watch(() => channelsStore.currentChannel, () => exitSelectMode())
watch(() => dmsStore.currentDMChannel, () => exitSelectMode())

defineExpose({ handlePinEvent, handleUnpinEvent })

onMounted(scrollToBottom)
</script>

<template>
  <div class="flex-1 flex flex-col min-w-0 bg-[var(--surface)]">
    <!-- Channel header -->
    <div class="h-13 px-4 flex items-center justify-between border-b border-[var(--surface-border)]">
      <div class="flex items-center min-w-0">
        <!-- Mobile menu button -->
        <button
          @click="emit('open-nav')"
          class="lg:hidden shrink-0 p-2 mr-1 text-[var(--text-4)] hover:text-[var(--text-1)] rounded-lg hover:bg-[var(--surface-3)] transition-colors duration-150"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5" />
          </svg>
        </button>
        <template v-if="isDMMode">
          <span class="text-[var(--text-4)] text-base mr-2 shrink-0">@</span>
        </template>
        <template v-else-if="isAudioChannel">
          <svg class="w-4 h-4 text-[var(--text-4)] mr-2 shrink-0" fill="currentColor" viewBox="0 0 24 24">
            <path d="M12 3a1 1 0 0 0-1 1v8a1 1 0 0 0 2 0V4a1 1 0 0 0-1-1ZM6.5 8A.5.5 0 0 0 6 8.5v3a6 6 0 0 0 5 5.91V20H8.5a.5.5 0 0 0 0 1h7a.5.5 0 0 0 0-1H13v-2.59A6 6 0 0 0 18 11.5v-3a.5.5 0 0 0-1 0v3a5 5 0 0 1-10 0v-3a.5.5 0 0 0-.5-.5Z"/>
          </svg>
        </template>
        <template v-else-if="isForumChannel">
          <svg class="w-4 h-4 text-[var(--text-4)] mr-2 shrink-0" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" d="M20 13V6a2 2 0 0 0-2-2H6a2 2 0 0 0-2 2v7m16 0v1a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2v-1m16 0h-2.586a1 1 0 0 0-.707.293l-2.414 2.414a1 1 0 0 1-.707.293h-3.172a1 1 0 0 1-.707-.293l-2.414-2.414A1 1 0 0 0 6.586 13H4" />
          </svg>
        </template>
        <template v-else>
          <span class="text-[var(--text-4)] text-base mr-2 shrink-0">#</span>
        </template>
        <div v-if="initialLoading && !hasActiveChannel" class="h-4 w-32 rounded bg-[var(--surface-3)] animate-pulse"></div>
        <h3 v-else class="text-[var(--text-1)] font-semibold truncate">
          {{ headerTitle }}
        </h3>
        <!-- Mobile: open member list sheet -->
        <button
          v-if="serversStore.currentServer && !isDMMode"
          @click="openMobileMembers()"
          class="lg:hidden shrink-0 ml-1 p-1 text-[var(--text-4)] hover:text-[var(--text-1)] rounded transition-colors duration-150"
          title="Members"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" d="M8.25 4.5l7.5 7.5-7.5 7.5" />
          </svg>
        </button>
        <span v-if="!isDMMode && channelsStore.currentChannel?.topic" class="hidden md:flex items-center gap-2 ml-2 min-w-0 shrink">
          <span class="text-[var(--text-4)] opacity-40 select-none">|</span>
          <span class="text-[var(--text-4)] text-sm truncate">{{ channelsStore.currentChannel.topic }}</span>
        </span>
      </div>

      <div class="flex items-center gap-1">
      <!-- Select messages button (admin/owner, server text channels only) -->
      <button
        v-if="serversStore.canManageMessages && !isDMMode && !isAudioChannel && !isForumChannel && !selecting"
        @click="enterSelectMode"
        @mouseenter="showHeaderTooltip($event, 'Select Messages')"
        @mouseleave="hideHeaderTooltip"
        class="hidden lg:flex items-center gap-1.5 text-sm px-3 py-1.5 rounded-lg cursor-pointer text-[var(--text-4)] hover:text-[var(--text-1)] hover:bg-[var(--surface-3)] transition-colors duration-150"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
        </svg>
      </button>

      <!-- Cancel selection button -->
      <button
        v-if="selecting"
        @click="exitSelectMode"
        class="hidden lg:flex items-center gap-1.5 text-sm px-3 py-1.5 rounded-lg cursor-pointer text-[#E8521A] hover:bg-[#E8521A]/10 transition-colors duration-150 font-medium"
      >
        Cancel
      </button>

      <!-- Search button (server mode only) -->
      <button
        v-if="serversStore.currentServer && !isDMMode"
        @click="searchStore.isOpen ? searchStore.close() : searchStore.open()"
        @mouseenter="showHeaderTooltip($event, 'Search Messages')"
        @mouseleave="hideHeaderTooltip"
        class="hidden lg:flex items-center gap-1.5 text-sm px-3 py-1.5 rounded-lg cursor-pointer transition-colors duration-150"
        :class="searchStore.isOpen ? 'text-[var(--text-1)] bg-[var(--surface-3)]' : 'text-[var(--text-4)] hover:text-[var(--text-1)] hover:bg-[var(--surface-3)]'"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" />
        </svg>
      </button>

      <!-- Pinned messages button (server text channels only) -->
      <button
        v-if="serversStore.currentServer && !isDMMode && !isAudioChannel && !isForumChannel"
        @click="showPinnedPanel = !showPinnedPanel"
        @mouseenter="showHeaderTooltip($event, 'Pinned Messages')"
        @mouseleave="hideHeaderTooltip"
        class="hidden lg:flex items-center gap-1.5 text-sm px-3 py-1.5 rounded-lg cursor-pointer transition-colors duration-150"
        :class="showPinnedPanel ? 'text-[var(--text-1)] bg-[var(--surface-3)]' : 'text-[var(--text-4)] hover:text-[var(--text-1)] hover:bg-[var(--surface-3)]'"
      >
        <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
          <path d="M16 3a1 1 0 0 1 .117 1.993L16 5h-.08l-.349 2.792A4.5 4.5 0 0 1 17 11v2.5h.5a1 1 0 0 1 .117 1.993L17.5 15.5h-4.5V21a1 1 0 0 1-1.993.117L11 21v-5.5H6.5a1 1 0 0 1-.117-1.993L6.5 13.5H7V11a4.5 4.5 0 0 1 1.429-3.208L8.08 5H8a1 1 0 0 1-.117-1.993L8 3h8Z" />
        </svg>
      </button>

      <!-- Members toggle (server mode only) -->
      <button
        v-if="serversStore.currentServer && !isDMMode"
        @click="emit('toggle-members')"
        @mouseenter="showHeaderTooltip($event, 'Toggle Members')"
        @mouseleave="hideHeaderTooltip"
        class="hidden lg:flex items-center gap-1.5 text-sm px-3 py-1.5 rounded-lg cursor-pointer transition-colors duration-150"
        :class="showMembers ? 'text-[var(--text-1)] bg-[var(--surface-3)]' : 'text-[var(--text-4)] hover:text-[var(--text-1)] hover:bg-[var(--surface-3)]'"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="M15 19.128a9.38 9.38 0 002.625.372 9.337 9.337 0 004.121-.952 4.125 4.125 0 00-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 018.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0111.964-3.07M12 6.375a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zm8.25 2.25a2.625 2.625 0 11-5.25 0 2.625 2.625 0 015.25 0z" />
        </svg>
      </button>
      </div>
    </div>

    <!-- Forum channel view -->
    <ForumView v-if="isForumChannel" />

    <!-- Audio channel view -->
    <div v-if="isAudioChannel" class="flex-1 flex items-center justify-center">
      <div class="text-center">
        <div class="w-20 h-20 rounded-2xl bg-[#E8521A]/10 flex items-center justify-center mx-auto mb-5">
          <svg class="w-10 h-10 text-[#E8521A]" fill="currentColor" viewBox="0 0 24 24">
            <path d="M12 3a1 1 0 0 0-1 1v8a1 1 0 0 0 2 0V4a1 1 0 0 0-1-1ZM6.5 8A.5.5 0 0 0 6 8.5v3a6 6 0 0 0 5 5.91V20H8.5a.5.5 0 0 0 0 1h7a.5.5 0 0 0 0-1H13v-2.59A6 6 0 0 0 18 11.5v-3a.5.5 0 0 0-1 0v3a5 5 0 0 1-10 0v-3a.5.5 0 0 0-.5-.5Z"/>
          </svg>
        </div>
        <div class="flex items-center justify-center gap-2 mb-2">
          <h3 class="font-display text-[var(--text-1)] text-2xl font-bold">
            {{ channelsStore.currentChannel?.name }}
          </h3>
          <span v-if="isInThisChannel && voiceStore.connectionMode" class="relative group flex items-center">
            <span class="w-2.5 h-2.5 rounded-full bg-emerald-500 cursor-default"></span>
            <span class="absolute left-1/2 -translate-x-1/2 top-full mt-2 px-2 py-1 rounded-lg bg-[var(--surface-2)] text-[var(--text-3)] text-xs font-medium whitespace-nowrap opacity-0 pointer-events-none group-hover:opacity-100 transition-opacity duration-150 shadow-lg border border-[var(--surface-border)] z-10">
              {{ voiceStore.connectionMode === 'livekit' ? 'LiveKit' : 'WebRTC P2P' }}
            </span>
          </span>
        </div>

        <!-- Connected users list -->
        <div v-if="connectedUsers.length > 0" class="mb-8">
          <div class="flex flex-wrap justify-center gap-4">
            <div
              v-for="userId in connectedUsers"
              :key="userId"
              class="flex flex-col items-center gap-1.5"
            >
              <div
                class="w-14 h-14 rounded-full flex items-center justify-center text-white text-lg font-bold shadow-lg"
                :style="getDefaultAvatarStyle(userId)"
              >
                {{ (getMember(userId)?.display_name || getMember(userId)?.username || '?')[0].toUpperCase() }}
              </div>
              <span class="text-[var(--text-1)] text-xs font-medium">
                {{ getMember(userId)?.display_name || getMember(userId)?.username || 'User' }}
                <span v-if="String(userId) === String(authStore.dbUser?.id)" class="text-[var(--text-4)]">(you)</span>
              </span>
            </div>
          </div>
        </div>
        <p v-else class="text-[var(--text-4)] text-sm mb-8">No one is currently in this channel.</p>

        <!-- Voice controls -->
        <div v-if="isInThisChannel" class="flex flex-wrap items-center justify-center gap-3">
            <button
              @click="handleToggleMute"
              @mouseenter="showHeaderTooltip($event, voiceStore.isMuted ? 'Unmute' : 'Mute')"
              @mouseleave="hideHeaderTooltip"
              class="w-11 h-11 rounded-full flex items-center justify-center cursor-pointer transition-all duration-200"
              :class="voiceStore.isMuted ? 'bg-[#E8521A] hover:bg-[#D44818] shadow-lg shadow-[#E8521A]/20 text-white' : 'bg-[var(--surface-3)] hover:bg-[var(--surface-border)] text-[var(--text-1)]'"
            >
              <svg v-if="!voiceStore.isMuted" class="w-5 h-5" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" viewBox="0 0 24 24">
                <path d="M9 5a3 3 0 0 1 3 -3a3 3 0 0 1 3 3v5a3 3 0 0 1 -3 3a3 3 0 0 1 -3 -3l0 -5" /><path d="M5 10a7 7 0 0 0 14 0" /><path d="M8 21l8 0" /><path d="M12 17l0 4" />
              </svg>
              <svg v-else class="w-5 h-5" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" viewBox="0 0 24 24">
                <path d="M3 3l18 18" /><path d="M9 5a3 3 0 0 1 6 0v5a3 3 0 0 1 -.13 .874m-2 2a3 3 0 0 1 -3.87 -2.872v-1" /><path d="M5 10a7 7 0 0 0 10.846 5.85m2 -2a6.967 6.967 0 0 0 1.152 -3.85" /><path d="M8 21l8 0" /><path d="M12 17l0 4" />
              </svg>
            </button>

            <button
              @click="handleToggleDeafen"
              @mouseenter="showHeaderTooltip($event, voiceStore.isDeafened ? 'Undeafen' : 'Deafen')"
              @mouseleave="hideHeaderTooltip"
              class="w-11 h-11 rounded-full flex items-center justify-center cursor-pointer transition-all duration-200"
              :class="voiceStore.isDeafened ? 'bg-[#E8521A] hover:bg-[#D44818] shadow-lg shadow-[#E8521A]/20 text-white' : 'bg-[var(--surface-3)] hover:bg-[var(--surface-border)] text-[var(--text-1)]'"
            >
              <svg v-if="!voiceStore.isDeafened" class="w-5 h-5" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" viewBox="0 0 24 24">
                <path d="M4 14v-3a8 8 0 1 1 16 0v3" /><path d="M18 19c0 1.657 -2.686 3 -6 3" /><path d="M4 14a2 2 0 0 1 2 -2h1a2 2 0 0 1 2 2v3a2 2 0 0 1 -2 2h-1a2 2 0 0 1 -2 -2v-3" /><path d="M15 14a2 2 0 0 1 2 -2h1a2 2 0 0 1 2 2v3a2 2 0 0 1 -2 2h-1a2 2 0 0 1 -2 -2v-3" />
              </svg>
              <svg v-else class="w-5 h-5" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" viewBox="0 0 24 24">
                <path d="M4 14v-3c0 -1.953 .7 -3.742 1.862 -5.13m2.182 -1.825a8 8 0 0 1 11.956 6.955v3" /><path d="M18 19c0 1.657 -2.686 3 -6 3" /><path d="M4 14a2 2 0 0 1 2 -2h1a2 2 0 0 1 2 2v3a2 2 0 0 1 -2 2h-1a2 2 0 0 1 -2 -2v-3" /><path d="M16.169 12.18c.253 -.115 .534 -.18 .831 -.18h1a2 2 0 0 1 2 2v2m-1.183 2.826c-.25 .112 -.526 .174 -.817 .174h-1a2 2 0 0 1 -2 -2v-2" /><path d="M3 3l18 18" />
              </svg>
            </button>

            <button
              @click="handleLeaveVoice"
              @mouseenter="showHeaderTooltip($event, 'Disconnect')"
              @mouseleave="hideHeaderTooltip"
              class="w-11 h-11 rounded-full bg-red-600 hover:bg-red-700 flex items-center justify-center cursor-pointer transition-all duration-200 shadow-lg shadow-red-600/25"
            >
              <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" viewBox="0 0 24 24">
                <path d="M5 4h4l2 5l-2.5 1.5a11 11 0 0 0 5 5l1.5 -2.5l5 2v4a2 2 0 0 1 -2 2c-8.072 -.49 -14.51 -6.928 -15 -15a2 2 0 0 1 2 -2" /><path d="M17 3l4 4" /><path d="M21 3l-4 4" />
              </svg>
            </button>

          <!-- PTT indicator -->
          <p v-if="voiceStore.inputMode === 'push_to_talk'" class="hidden md:block text-[var(--text-4)] text-xs mt-3 w-full text-center">
            Hold <kbd class="px-1.5 py-0.5 rounded bg-[var(--surface-3)] text-[var(--text-3)] font-mono text-[11px]">{{ voiceStore.pttKey === 'Space' ? 'Space' : voiceStore.pttKey.replace('Key', '') }}</kbd> to talk
          </p>
        </div>

        <!-- Joining state -->
        <button
          v-else-if="voiceStore.isJoining"
          disabled
          class="w-[140px] mx-auto bg-[#D4782A]/70 text-white font-semibold py-2.5 rounded-xl transition-all duration-200 shadow-lg shadow-[#D4782A]/20 flex items-center justify-center gap-2"
        >
          <svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
          </svg>
          Joining...
        </button>

        <button
          v-else
          @click="handleJoinVoice"
          class="w-[140px] bg-[#D4782A] hover:bg-[#C06D26] text-white font-semibold py-2.5 rounded-xl cursor-pointer transition-all duration-200 shadow-lg shadow-[#D4782A]/20"
        >
          Join Voice
        </button>
      </div>
    </div>

    <!-- Text channel / DM view -->
    <template v-else-if="!isForumChannel">
      <div class="flex flex-1 min-h-0">
        <!-- Main column: messages + typing + input -->
        <div class="flex-1 flex flex-col min-w-0">
          <!-- Messages -->
          <div ref="messagesContainer" class="flex-1 overflow-y-auto px-5 py-5 scrollbar-light">
            <div v-if="hasActiveChannel">
              <!-- Loading skeleton for initial channel load -->
              <div v-if="activeLoading && activeMessages.length === 0" class="space-y-5">
                <div v-for="i in 7" :key="i" class="flex items-start gap-3 animate-pulse">
                  <div class="w-10 h-10 rounded-full bg-[var(--surface-3)] shrink-0" />
                  <div class="flex-1 min-w-0 pt-0.5">
                    <div class="flex items-center gap-2 mb-1.5">
                      <div class="h-3.5 rounded bg-[var(--surface-3)]" :style="{ width: [90, 70, 110, 80, 100, 60, 95][i - 1] + 'px' }" />
                      <div class="h-2.5 w-10 rounded bg-[var(--surface-3)] opacity-50" />
                    </div>
                    <div class="space-y-1.5">
                      <div class="h-3 rounded bg-[var(--surface-3)] opacity-70" :style="{ width: [85, 60, 95, 45, 75, 90, 55][i - 1] + '%' }" />
                      <div v-if="i % 3 !== 0" class="h-3 rounded bg-[var(--surface-3)] opacity-50" :style="{ width: [50, 70, 0, 40, 60, 0, 35][i - 1] + '%' }" />
                    </div>
                  </div>
                </div>
              </div>

              <template v-else>
                <button
                  v-if="activeHasMore"
                  @click="loadMore"
                  class="text-[#E8521A] text-sm hover:text-[#D44818] mb-5 cursor-pointer font-medium"
                >
                  Load older messages
                </button>

                <div
                  v-for="message in activeMessages"
                  :key="message.id"
                  :data-message-id="message.id"
                >
                  <MessageItem
                    :message="message"
                    :mode="isDMMode ? 'dm' : 'server'"
                    :pinned-message-ids="pinnedMessageIds"
                    :selecting="selecting"
                    :selected="selectedIds.has(message.id)"
                    @reply="handleReply"
                    @scroll-to-message="scrollToMessage"
                    @pin="handlePin"
                    @unpin="handleUnpin"
                    @toggle-select="toggleSelect"
                  />
                </div>

                <div v-if="activeMessages.length === 0 && !activeLoading" class="text-center mt-16">
                <div class="w-16 h-16 rounded-2xl bg-[#E8521A]/10 flex items-center justify-center mx-auto mb-4">
                  <svg class="w-7 h-7 text-[#E8521A]" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                  </svg>
                </div>
                <p class="text-[var(--text-3)] font-medium">No messages yet</p>
                <p class="text-[var(--text-4)] text-sm mt-1">Start the conversation!</p>
              </div>
              </template>
            </div>

            <!-- Loading skeleton while initial data is being fetched -->
            <div v-else-if="initialLoading" class="flex flex-col h-full">
              <!-- Fake header skeleton -->
              <div class="px-5 pt-5 pb-3">
                <div class="h-5 w-48 rounded bg-[var(--surface-3)] animate-pulse"></div>
              </div>
              <!-- Fake messages skeleton -->
              <div class="flex-1 px-5 space-y-5">
                <div v-for="i in 8" :key="i" class="flex items-start gap-3 animate-pulse">
                  <div class="w-10 h-10 rounded-full bg-[var(--surface-3)] shrink-0" />
                  <div class="flex-1 min-w-0 pt-0.5">
                    <div class="flex items-center gap-2 mb-1.5">
                      <div class="h-3.5 rounded bg-[var(--surface-3)]" :style="{ width: [100, 75, 120, 85, 95, 110, 70, 90][i - 1] + 'px' }" />
                      <div class="h-2.5 w-10 rounded bg-[var(--surface-3)] opacity-50" />
                    </div>
                    <div class="space-y-1.5">
                      <div class="h-3 rounded bg-[var(--surface-3)] opacity-70" :style="{ width: [80, 55, 90, 40, 70, 85, 60, 50][i - 1] + '%' }" />
                      <div v-if="i % 3 !== 0" class="h-3 rounded bg-[var(--surface-3)] opacity-50" :style="{ width: [45, 65, 0, 35, 55, 0, 40, 30][i - 1] + '%' }" />
                    </div>
                  </div>
                </div>
              </div>
              <!-- Fake input skeleton -->
              <div class="px-5 py-4">
                <div class="h-11 rounded-xl bg-[var(--surface-3)] animate-pulse"></div>
              </div>
            </div>

            <div v-else class="flex flex-col items-center justify-center h-full text-center">
              <template v-if="route.path === '/channels/@me'">
                <div class="w-16 h-16 rounded-2xl bg-[#E8521A]/10 flex items-center justify-center mx-auto mb-4">
                  <svg class="w-7 h-7 text-[#E8521A]" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M21.75 6.75v10.5a2.25 2.25 0 0 1-2.25 2.25h-15a2.25 2.25 0 0 1-2.25-2.25V6.75m19.5 0A2.25 2.25 0 0 0 19.5 4.5h-15a2.25 2.25 0 0 0-2.25 2.25m19.5 0v.243a2.25 2.25 0 0 1-1.07 1.916l-7.5 4.615a2.25 2.25 0 0 1-2.36 0L3.32 8.91a2.25 2.25 0 0 1-1.07-1.916V6.75" />
                  </svg>
                </div>
                <h2 class="font-display text-3xl font-bold text-[var(--text-1)] mb-2">Direct Messages</h2>
                <p class="text-[var(--text-4)]">Select a conversation to start chatting</p>
              </template>
              <template v-else>
                <h2 class="font-display text-3xl font-bold text-[var(--text-1)] mb-2">Welcome</h2>
                <p class="text-[var(--text-4)]">Select a channel or conversation to start chatting</p>
              </template>
            </div>
          </div>

          <!-- Typing indicator -->
          <div v-if="typingText" class="typing-indicator px-5 py-1 flex items-center gap-1.5 text-[var(--text-4)] text-xs">
            <span class="typing-dots">
              <span class="dot"></span>
              <span class="dot"></span>
              <span class="dot"></span>
            </span>
            <span>{{ typingText }}</span>
          </div>

          <!-- Bulk select toolbar -->
          <div
            v-if="selecting && selectedIds.size > 0"
            class="px-5 py-3 border-t border-[var(--surface-border)] bg-[var(--surface)] flex items-center justify-between"
          >
            <span class="text-sm text-[var(--text-2)] font-medium">
              {{ selectedIds.size }} message{{ selectedIds.size === 1 ? '' : 's' }} selected
            </span>
            <div class="flex items-center gap-2">
              <button
                @click="exitSelectMode"
                class="text-[var(--text-3)] hover:text-[var(--text-1)] px-3 py-1.5 rounded-lg text-sm cursor-pointer font-medium transition-colors duration-150"
              >
                Cancel
              </button>
              <button
                @click="showBulkDeleteConfirm = true"
                class="bg-[#E8521A] text-white px-4 py-1.5 rounded-lg hover:bg-[#D44818] cursor-pointer text-sm font-semibold shadow-lg shadow-[#E8521A]/15 transition-all duration-200"
              >
                Delete Selected
              </button>
            </div>
          </div>

          <!-- Input -->
          <MessageInput
            v-if="hasActiveChannel && !selecting"
            :mode="isDMMode ? 'dm' : 'server'"
            :replying-to="messagesStore.replyingTo"
            @cancel-reply="handleCancelReply"
          />
        </div>

        <!-- Pinned messages panel -->
        <PinnedMessagesPanel
          v-if="showPinnedPanel && channelsStore.currentChannel"
          ref="pinnedPanel"
          :channel-id="channelsStore.currentChannel.id"
          @close="showPinnedPanel = false"
          @scroll-to-message="scrollToMessage"
        />
      </div>
    </template>

    <!-- Header tooltip (below buttons) -->
    <Teleport to="body">
      <Transition
        enter-active-class="transition duration-150 ease-out"
        enter-from-class="opacity-0 -translate-y-1"
        enter-to-class="opacity-100 translate-y-0"
        leave-active-class="transition duration-100 ease-in"
        leave-from-class="opacity-100 translate-y-0"
        leave-to-class="opacity-0 -translate-y-1"
      >
        <div
          v-if="headerTooltip.visible"
          class="fixed z-[9999] pointer-events-none flex flex-col items-center -translate-x-1/2"
          :style="{ top: headerTooltip.top + 'px', left: headerTooltip.left + 'px' }"
        >
          <div class="w-2 h-2 bg-[#111214] rotate-45 rounded-[2px] shrink-0 -mb-1.5"></div>
          <div class="bg-[#111214] text-white text-sm font-semibold px-3 py-1.5 rounded-md shadow-lg shadow-black/30 whitespace-nowrap">{{ headerTooltip.text }}</div>
        </div>
      </Transition>
    </Teleport>

    <!-- Bulk delete confirmation modal -->
    <Teleport to="body">
      <div v-if="showBulkDeleteConfirm" class="fixed inset-0 bg-[var(--backdrop)] backdrop-blur-sm flex items-center justify-center z-[60]" @click.self="showBulkDeleteConfirm = false">
        <div class="bg-[var(--modal-bg)] rounded-2xl p-7 w-full max-w-sm shadow-2xl shadow-black/10 animate-scale-in border border-[var(--modal-border)]">
          <h2 class="font-display text-xl font-bold text-[var(--text-1)] mb-2">Delete Messages</h2>
          <p class="text-[var(--text-3)] text-sm mb-5">
            Are you sure you want to delete {{ selectedIds.size }} message{{ selectedIds.size === 1 ? '' : 's' }}? This cannot be undone.
          </p>
          <div class="flex justify-end gap-3">
            <button @click="showBulkDeleteConfirm = false" class="text-[var(--text-3)] hover:text-[var(--text-1)] px-4 py-2.5 rounded-xl cursor-pointer font-medium transition-colors duration-150">
              Cancel
            </button>
            <button
              @click="bulkDelete"
              :disabled="bulkDeleting"
              class="bg-[#E8521A] text-white px-5 py-2.5 rounded-xl hover:bg-[#D44818] cursor-pointer font-semibold shadow-lg shadow-[#E8521A]/15 transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {{ bulkDeleting ? 'Deleting...' : 'Delete' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.typing-dots {
  display: inline-flex;
  gap: 2px;
  align-items: center;
}
.typing-dots .dot {
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: var(--text-4);
  animation: typing-bounce 1.4s infinite ease-in-out;
}
.typing-dots .dot:nth-child(2) {
  animation-delay: 0.2s;
}
.typing-dots .dot:nth-child(3) {
  animation-delay: 0.4s;
}
@keyframes typing-bounce {
  0%, 60%, 100% {
    transform: translateY(0);
  }
  30% {
    transform: translateY(-4px);
  }
}
</style>
