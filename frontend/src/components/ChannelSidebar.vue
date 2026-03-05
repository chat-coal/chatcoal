<script setup>
import { ref, computed, watch, nextTick, inject, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useServersStore } from '@/stores/servers'
import { useChannelsStore } from '@/stores/channels'
import { useAuthStore } from '@/stores/auth'
import { useVoiceStore } from '@/stores/voice'
import { useDMsStore } from '@/stores/dms'
import { useUnreadStore } from '@/stores/unread'
import { useNotificationSettingsStore } from '@/stores/notificationSettings'
import { useToastStore } from '@/stores/toast'
import api from '@/services/api.service'
import { getAvatarColor, getDefaultAvatarStyle, resolveFileUrl, cssBackgroundUrl } from '@/utils/avatar'
import { on } from '@/services/websocket.service'
import UserStatusBar from './UserStatusBar.vue'
import InviteModal from './InviteModal.vue'
import ServerSettingsModal from './ServerSettingsModal.vue'
import ChannelFederationModal from './ChannelFederationModal.vue'

const router = useRouter()
const route = useRoute()
const serversStore = useServersStore()
const channelsStore = useChannelsStore()
const authStore = useAuthStore()
const voiceStore = useVoiceStore()
const dmsStore = useDMsStore()
const unreadStore = useUnreadStore()
const notifStore = useNotificationSettingsStore()
const toastStore = useToastStore()

const members = ref([])
const closeMobileNav = inject('closeMobileNav', () => {})
const initialLoading = inject('initialLoading', ref(false))

const isDMMode = computed(() => {
  return !serversStore.currentServer && route.path.startsWith('/channels/@me')
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
    } else {
      members.value = []
    }
  },
  { immediate: true },
)

function getVoiceUsers(channelId) {
  return voiceStore.voiceStates[channelId] || []
}

// Keep members in sync with real-time events
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

const isOwner = computed(() => serversStore.isOwner)
const isAdmin = computed(() => serversStore.isAdmin)

const showInviteModal = ref(false)
const showSettingsModal = ref(false)
const showLeaveModal = ref(false)
const showUserSettingsModal = ref(false)

const isServerMuted = computed(() =>
  serversStore.currentServer ? notifStore.mutedServers.has(serversStore.currentServer.id) : false,
)

function toggleServerMute() {
  if (serversStore.currentServer) {
    notifStore.toggleMute('server', serversStore.currentServer.id)
  }
}
const newChannelName = ref('')
const showAddChannel = ref(false)
const newChannelType = ref('text')
const channelNameInput = ref(null)

const showEditModal = ref(false)
const editingChannel = ref(null)
const editName = ref('')
const editTopic = ref('')
const showFederationModal = ref(false)
const federationChannel = ref(null)

function onEscapeKey(e) {
  if (e.key !== 'Escape') return
  if (showEditModal.value) { showEditModal.value = false; return }
  if (showLeaveModal.value) { showLeaveModal.value = false; return }
  if (showUserSettingsModal.value) { showUserSettingsModal.value = false; return }
}
onMounted(() => window.addEventListener('keydown', onEscapeKey))
onUnmounted(() => window.removeEventListener('keydown', onEscapeKey))

function openFederationModal(channel) {
  federationChannel.value = channel
  showFederationModal.value = true
}
function closeFederationModal() {
  showFederationModal.value = false
  federationChannel.value = null
}

function openEditModal(channel) {
  editingChannel.value = channel
  editName.value = channel.name
  editTopic.value = channel.topic || ''
  showEditModal.value = true
}

async function saveEditChannel() {
  if (!editName.value.trim()) return
  try {
    await channelsStore.updateChannel(editingChannel.value.id, {
      name: editName.value.trim(),
      topic: editTopic.value.trim(),
    })
    showEditModal.value = false
  } catch {
    toastStore.add('Failed to update channel')
  }
}

function cancelAddChannel() {
  showAddChannel.value = false
  newChannelName.value = ''
  newChannelType.value = 'text'
}

async function toggleAddChannel() {
  showAddChannel.value = !showAddChannel.value
  if (showAddChannel.value) {
    await nextTick()
    channelNameInput.value?.focus()
  } else {
    newChannelName.value = ''
    newChannelType.value = 'text'
  }
}

const textChannels = computed(() =>
  channelsStore.channels.filter((c) => c.type === 'text'),
)
const audioChannels = computed(() =>
  channelsStore.channels.filter((c) => c.type === 'audio'),
)
const forumChannels = computed(() =>
  channelsStore.channels.filter((c) => c.type === 'forum'),
)

function selectChannel(channel) {
  channelsStore.selectChannel(channel)
  router.push(`/channels/${serversStore.currentServer.id}/${channel.id}`)
  closeMobileNav()
}

function selectDMChannel(channel) {
  dmsStore.selectDM(channel)
  router.push(`/channels/@me/${channel.id}`)
  closeMobileNav()
}

function getOtherUser(dmChannel) {
  if (!authStore.dbUser) return null
  if (dmChannel.user1?.id === authStore.dbUser.id) return dmChannel.user2
  return dmChannel.user1
}

async function leaveServer() {
  try {
    await serversStore.leaveServer(serversStore.currentServer.id)
    showLeaveModal.value = false
    router.push('/channels/@me')
  } catch {
    showLeaveModal.value = false
    toastStore.add('Failed to leave server')
  }
}

async function addChannel() {
  if (!newChannelName.value.trim()) return
  try {
    const channel = await channelsStore.createChannel(
      serversStore.currentServer.id,
      newChannelName.value.trim(),
      newChannelType.value,
    )
    newChannelName.value = ''
    newChannelType.value = 'text'
    showAddChannel.value = false
    selectChannel(channel)
  } catch {
    toastStore.add('Failed to create channel')
  }
}

function formatLastMessage(dm) {
  if (!dm.last_message) return ''
  const content = dm.last_message.content
  return content.length > 30 ? content.slice(0, 30) + '...' : content
}

// Drag-and-drop reorder
const canManageChannels = computed(() => serversStore.canManageChannels)
const dragState = ref(null) // { type, index }
const dropIndex = ref(null) // insertion point: 0..length (null = none)

function getGroupChannels(type) {
  if (type === 'text') return textChannels.value
  if (type === 'audio') return audioChannels.value
  if (type === 'forum') return forumChannels.value
  return []
}

function onDragStart(e, type, index) {
  if (!canManageChannels.value) return
  dragState.value = { type, index }
  e.dataTransfer.effectAllowed = 'move'
  e.dataTransfer.setData('text/plain', '') // required for Firefox
}

function onDragOver(e, type, index) {
  if (!dragState.value || dragState.value.type !== type) return
  e.preventDefault()
  e.dataTransfer.dropEffect = 'move'
  // Determine if cursor is in the top or bottom half of the element
  const rect = e.currentTarget.getBoundingClientRect()
  const midY = rect.top + rect.height / 2
  dropIndex.value = e.clientY < midY ? index : index + 1
}

function onDragLeave() {
  dropIndex.value = null
}

function onDragEnd() {
  dragState.value = null
  dropIndex.value = null
}

function dropToMoveIndex(fromIndex, insertionIndex) {
  // Convert insertion point to the target index after removal
  if (insertionIndex > fromIndex) return insertionIndex - 1
  return insertionIndex
}

async function onDrop(e, type, index) {
  e.preventDefault()
  if (!dragState.value || dragState.value.type !== type) return

  const fromIndex = dragState.value.index
  // Use dropIndex (insertion point with top/bottom half detection) instead of raw item index
  const insertAt = dropIndex.value ?? index
  dragState.value = null
  dropIndex.value = null

  const toIndex = dropToMoveIndex(fromIndex, insertAt)
  if (fromIndex === toIndex) return

  const group = getGroupChannels(type)
  const reordered = [...group]
  const [moved] = reordered.splice(fromIndex, 1)
  reordered.splice(toIndex, 0, moved)

  // Optimistically apply reorder — update positions of all channels in this type group
  const posMap = new Map()
  reordered.forEach((ch, i) => { posMap.set(String(ch.id), i) })
  // Update positions for channels in this group and re-sort the full list
  for (const ch of channelsStore.channels) {
    if (ch.type === type) {
      ch.position = posMap.get(String(ch.id)) ?? ch.position
    }
  }
  channelsStore.channels.sort((a, b) => a.position - b.position || String(a.id).localeCompare(String(b.id)))

  // Build the full ordered channel ID list (all types) to send to backend
  const allIds = channelsStore.channels.map((ch) => ch.id)

  try {
    await channelsStore.reorderChannels(serversStore.currentServer.id, allIds)
  } catch {
    // Revert on error by refetching
    toastStore.add('Failed to reorder channels', 'error')
    await channelsStore.fetchChannels(serversStore.currentServer.id)
  }
}
</script>

<template>
  <div class="w-[248px] bg-[var(--sb-bg-2)] flex flex-col shrink-0 relative noise-texture">
    <!-- Header -->
    <div class="relative z-10 h-13 px-4 flex items-center justify-between border-b border-[var(--sb-border)]">
      <template v-if="isDMMode">
        <h2 class="font-display text-[var(--sb-text)] text-lg font-semibold truncate">Direct Messages</h2>
        <!-- Mobile close button -->
        <button
          @click="closeMobileNav()"
          class="lg:hidden ml-auto text-[var(--sb-text-3)] hover:text-[var(--sb-text)] p-1.5 rounded-lg hover:bg-[var(--sb-hover)] transition-colors duration-150"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" d="M18 6L6 18M6 6l12 12" />
          </svg>
        </button>
      </template>
      <template v-else>
        <div v-if="initialLoading && !serversStore.currentServer" class="h-4 w-28 rounded bg-[var(--sb-hover)] animate-pulse"></div>
        <h2 v-else class="font-display text-[var(--sb-text)] text-lg font-semibold truncate">
          {{ serversStore.currentServer?.name || 'chatcoal' }}
        </h2>
        <div class="flex items-center gap-0.5 shrink-0">
          <button
            v-if="serversStore.currentServer && serversStore.canManageInvites"
            @click="showInviteModal = true"
            class="text-[var(--sb-text-2)] hover:text-[var(--sb-text)] p-1.5 rounded-lg hover:bg-[var(--sb-hover)] cursor-pointer transition-colors duration-150"
            title="Invite People"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
              <circle cx="9" cy="7" r="4" />
              <line x1="19" y1="8" x2="19" y2="14" />
              <line x1="22" y1="11" x2="16" y2="11" />
            </svg>
          </button>
          <button
            v-if="serversStore.currentServer && isAdmin"
            @click="showSettingsModal = true"
            class="text-[var(--sb-text-2)] hover:text-[var(--sb-text)] p-1.5 rounded-lg hover:bg-[var(--sb-hover)] cursor-pointer transition-colors duration-150"
            title="Server Settings"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
              <circle cx="12" cy="12" r="3" />
            </svg>
          </button>
          <button
            v-if="serversStore.currentServer"
            @click="showUserSettingsModal = true"
            class="p-1.5 rounded-lg hover:bg-[var(--sb-hover)] cursor-pointer transition-colors duration-150"
            :class="isServerMuted ? 'text-[#E8521A]' : 'text-[var(--sb-text-2)] hover:text-[var(--sb-text)]'"
            title="Notification Settings"
          >
            <svg v-if="isServerMuted" class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9.143 17.082a24.248 24.248 0 0 0 3.844.148m-3.844-.148a23.856 23.856 0 0 1-5.455-1.31 8.964 8.964 0 0 0 2.3-5.542m3.155 6.852a3 3 0 0 0 5.667 1.97m1.965-2.277L21 21m-4.225-4.225a23.9 23.9 0 0 0 3.882-1.495l-.105-.132A8.966 8.966 0 0 1 15.5 5.702 8.959 8.959 0 0 1 12 5c-1.157 0-2.27.218-3.29.618M3.097 3.097l17.866 17.866" />
            </svg>
            <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M14.857 17.082a23.848 23.848 0 0 0 5.454-1.31A8.967 8.967 0 0 1 18 9.75V9A6 6 0 0 0 6 9v.75a8.967 8.967 0 0 1-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 0 1-5.714 0m5.714 0a3 3 0 1 1-5.714 0" />
            </svg>
          </button>
          <button
            v-if="serversStore.currentServer && !isOwner"
            @click="showLeaveModal = true"
            class="text-[var(--sb-text-2)] hover:text-[#E8521A] p-1.5 rounded-lg hover:bg-[var(--sb-hover)] cursor-pointer transition-colors duration-150"
            title="Leave Server"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
            </svg>
          </button>
          <!-- Mobile close button -->
          <button
            @click="closeMobileNav()"
            class="lg:hidden text-[var(--sb-text-3)] hover:text-[var(--sb-text)] p-1.5 rounded-lg hover:bg-[var(--sb-hover)] cursor-pointer transition-colors duration-150"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M18 6L6 18M6 6l12 12" />
            </svg>
          </button>
        </div>
      </template>
    </div>

    <!-- Content -->
    <div class="relative z-10 flex-1 overflow-y-auto px-2.5 py-4 scrollbar-thin">
      <!-- DM Mode: show DM conversations -->
      <template v-if="isDMMode">
        <div v-if="dmsStore.dmChannels.length > 0" class="space-y-0.5">
          <button
            v-for="dm in dmsStore.dmChannels"
            :key="dm.id"
            @click="selectDMChannel(dm)"
            class="w-full flex items-center gap-2.5 px-2.5 py-2 rounded-lg cursor-pointer transition-all duration-150"
            :class="
              dmsStore.currentDMChannel?.id === dm.id
                ? 'bg-[var(--sb-hover)] text-[var(--sb-text)]'
                : 'text-[var(--sb-text-2)] hover:bg-[var(--sb-hover)] hover:text-[var(--sb-text)]'
            "
          >
            <!-- Avatar -->
            <div class="relative shrink-0">
              <div
                v-if="getOtherUser(dm)?.avatar_url"
                class="w-8 h-8 rounded-full bg-cover bg-center"
                :style="{ backgroundImage: cssBackgroundUrl(resolveFileUrl(getOtherUser(dm).avatar_url)) }"
              ></div>
              <div
                v-else
                class="w-8 h-8 rounded-full flex items-center justify-center text-white text-xs font-bold"
                :style="getDefaultAvatarStyle(getOtherUser(dm)?.id || 0)"
              >
                {{ (getOtherUser(dm)?.display_name || '?')[0].toUpperCase() }}
              </div>
            </div>
            <!-- Name + preview -->
            <div class="min-w-0 flex-1 text-left">
              <div class="flex items-center justify-between">
                <span class="text-sm font-medium truncate">{{ getOtherUser(dm)?.display_name || 'Unknown' }}</span>
                <span
                  v-if="unreadStore.dmUnread[dm.id]"
                  class="ml-1 min-w-[18px] h-[18px] bg-[#E8521A] rounded-full flex items-center justify-center text-white text-[10px] font-bold px-1 shrink-0"
                >
                  {{ unreadStore.dmUnread[dm.id] > 99 ? '99+' : unreadStore.dmUnread[dm.id] }}
                </span>
              </div>
              <p v-if="dm.last_message" class="text-xs text-[var(--sb-text-3)] truncate mt-0.5">
                {{ formatLastMessage(dm) }}
              </p>
            </div>
          </button>
        </div>
        <div v-else class="text-[var(--sb-text-3)] text-sm px-2 py-4 text-center">
          No conversations yet
        </div>
      </template>

      <!-- Server Mode: show channels -->
      <template v-else>
        <div v-if="serversStore.currentServer && !channelsStore.loading" class="space-y-5">
          <!-- Add channel input -->
          <div v-if="showAddChannel" class="px-1 mb-3 animate-fade-in-up">
            <div class="flex items-center gap-1.5 mb-2">
              <button
                @click="newChannelType = 'text'"
                class="text-xs px-2.5 py-1 rounded-lg cursor-pointer font-medium transition-colors duration-150"
                :class="
                  newChannelType === 'text'
                    ? 'bg-[#E8521A] text-white'
                    : 'bg-[var(--sb-hover)] text-[var(--sb-text-2)] hover:text-[var(--sb-text)]'
                "
              >
                # Text
              </button>
              <button
                @click="newChannelType = 'audio'"
                class="text-xs px-2.5 py-1 rounded-lg cursor-pointer font-medium transition-colors duration-150"
                :class="
                  newChannelType === 'audio'
                    ? 'bg-[#E8521A] text-white'
                    : 'bg-[var(--sb-hover)] text-[var(--sb-text-2)] hover:text-[var(--sb-text)]'
                "
              >
                Audio
              </button>
              <button
                @click="newChannelType = 'forum'"
                class="text-xs px-2.5 py-1 rounded-lg cursor-pointer font-medium transition-colors duration-150"
                :class="
                  newChannelType === 'forum'
                    ? 'bg-[#E8521A] text-white'
                    : 'bg-[var(--sb-hover)] text-[var(--sb-text-2)] hover:text-[var(--sb-text)]'
                "
              >
                Forum
              </button>
              <button
                @click="cancelAddChannel"
                class="ml-auto text-[var(--sb-text-3)] hover:text-[var(--sb-text)] cursor-pointer transition-colors duration-150"
                title="Cancel"
              >
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M18 6L6 18M6 6l12 12" />
                </svg>
              </button>
            </div>
            <input
              ref="channelNameInput"
              v-model="newChannelName"
              @keyup.enter="addChannel"
              @keyup.escape="cancelAddChannel"
              placeholder="channel-name"
              class="w-full bg-[var(--sb-bg)] text-[var(--sb-text)] text-sm px-3 py-2 rounded-lg placeholder-[var(--sb-text-3)] border border-[var(--sb-border)]"
            />
          </div>

          <!-- Text Channels -->
          <div>
            <div class="flex items-center justify-between px-2 mb-2">
              <span class="text-[var(--sb-text-3)] text-[11px] font-bold uppercase tracking-[0.1em]">Text Channels</span>
              <button
                v-if="serversStore.canManageChannels && !showAddChannel"
                @click="toggleAddChannel"
                class="text-[var(--sb-text-3)] hover:text-[var(--sb-text)] cursor-pointer transition-colors duration-150"
              >
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
                </svg>
              </button>
            </div>

            <div
              v-for="(channel, idx) in textChannels"
              :key="channel.id"
              class="group relative"
              :draggable="canManageChannels"
              @dragstart="onDragStart($event, 'text', idx)"
              @dragover="onDragOver($event, 'text', idx)"
              @dragleave="onDragLeave"
              @drop="onDrop($event, 'text', idx)"
              @dragend="onDragEnd"
            >
              <div v-if="dragState?.type === 'text' && dropIndex === idx && dropIndex !== dragState.index && dropIndex !== dragState.index + 1" class="h-0.5 bg-[#E8521A] rounded-full mx-2 -mt-px mb-px"></div>
              <button
                @click="selectChannel(channel)"
                class="w-full flex items-center gap-2 px-2.5 py-[7px] rounded-lg text-sm transition-all duration-150"
                :class="[
                  channelsStore.currentChannel?.id === channel.id
                    ? 'bg-[var(--sb-hover)] text-[var(--sb-text)]'
                    : 'text-[var(--sb-text-2)] hover:bg-[var(--sb-hover)] hover:text-[var(--sb-text)]',
                  canManageChannels ? 'cursor-grab active:cursor-grabbing' : 'cursor-pointer',
                ]"
              >
                <span v-if="channel.federation_id" class="opacity-50 shrink-0">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M12 21a9.004 9.004 0 0 0 8.716-6.747M12 21a9.004 9.004 0 0 1-8.716-6.747M12 21c2.485 0 4.5-4.03 4.5-9S14.485 3 12 3m0 18c-2.485 0-4.5-4.03-4.5-9S9.515 3 12 3m0 0a8.997 8.997 0 0 1 7.843 4.582M12 3a8.997 8.997 0 0 0-7.843 4.582m15.686 0A11.953 11.953 0 0 1 12 10.5c-2.998 0-5.74-1.1-7.843-2.918m15.686 0A8.959 8.959 0 0 1 21 12c0 .778-.099 1.533-.284 2.253m0 0A17.919 17.919 0 0 1 12 16.5c-3.162 0-6.133-.815-8.716-2.247m0 0A9.015 9.015 0 0 1 3 12c0-1.605.42-3.113 1.157-4.418" />
                  </svg>
                </span>
                <span v-else class="text-base opacity-50 font-light">#</span>
                <span class="truncate font-medium flex-1 text-left">{{ channel.name }}</span>
                <!-- Muted icon -->
                <svg
                  v-if="notifStore.isEffectivelyMuted(channel.id, serversStore.currentServer?.id)"
                  class="w-3.5 h-3.5 text-[var(--sb-text-3)] shrink-0 opacity-60 group-hover:hidden"
                  fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"
                  title="Muted"
                >
                  <path stroke-linecap="round" stroke-linejoin="round" d="M17.25 9.75 19.5 12m0 0 2.25 2.25M19.5 12l2.25-2.25M19.5 12l-2.25 2.25m-10.5-6 4.72-4.72a.75.75 0 0 1 1.28.53v15.88a.75.75 0 0 1-1.28.53l-4.72-4.72H4.51c-.88 0-1.704-.507-1.938-1.354A9.009 9.009 0 0 1 2.25 12c0-.83.112-1.633.322-2.396C2.806 8.756 3.63 8.25 4.51 8.25H6.75Z" />
                </svg>
                <span
                  v-else-if="unreadStore.channelUnread[channel.id] && channelsStore.currentChannel?.id !== channel.id"
                  class="min-w-[18px] h-[18px] bg-[#E8521A] rounded-full flex items-center justify-center text-white text-[10px] font-bold px-1 shrink-0"
                  :class="serversStore.canManageChannels ? 'group-hover:hidden' : 'group-hover:hidden'"
                >
                  {{ unreadStore.channelUnread[channel.id] > 99 ? '99+' : unreadStore.channelUnread[channel.id] }}
                </span>
              </button>
              <div
                class="absolute right-1.5 top-1/2 -translate-y-1/2 opacity-0 group-hover:opacity-100 flex items-center gap-0.5 transition-all duration-150"
              >
                <!-- Mute toggle -->
                <button
                  @click.stop="notifStore.toggleMute('channel', channel.id)"
                  class="text-[var(--sb-text-3)] hover:text-[var(--sb-text)] p-0.5 rounded"
                  :title="notifStore.mutedChannels.has(channel.id) ? 'Unmute channel' : 'Mute channel'"
                >
                  <svg v-if="notifStore.mutedChannels.has(channel.id)" class="w-3.5 h-3.5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M19.114 5.636a9 9 0 0 1 0 12.728M16.463 8.288a5.25 5.25 0 0 1 0 7.424M6.75 8.25l4.72-4.72a.75.75 0 0 1 1.28.53v15.88a.75.75 0 0 1-1.28.53l-4.72-4.72H4.51c-.88 0-1.704-.507-1.938-1.354A9.009 9.009 0 0 1 2.25 12c0-.83.112-1.633.322-2.396C2.806 8.756 3.63 8.25 4.51 8.25H6.75Z" />
                  </svg>
                  <svg v-else class="w-3.5 h-3.5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M17.25 9.75 19.5 12m0 0 2.25 2.25M19.5 12l2.25-2.25M19.5 12l-2.25 2.25m-10.5-6 4.72-4.72a.75.75 0 0 1 1.28.53v15.88a.75.75 0 0 1-1.28.53l-4.72-4.72H4.51c-.88 0-1.704-.507-1.938-1.354A9.009 9.009 0 0 1 2.25 12c0-.83.112-1.633.322-2.396C2.806 8.756 3.63 8.25 4.51 8.25H6.75Z" />
                  </svg>
                </button>
                <button
                  v-if="serversStore.canManageChannels"
                  @click.stop="openFederationModal(channel)"
                  class="text-[var(--sb-text-3)] hover:text-[var(--sb-text)] p-0.5 rounded"
                  title="Channel Federation"
                >
                  <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M12 21a9.004 9.004 0 0 0 8.716-6.747M12 21a9.004 9.004 0 0 1-8.716-6.747M12 21c2.485 0 4.5-4.03 4.5-9S14.485 3 12 3m0 18c-2.485 0-4.5-4.03-4.5-9S9.515 3 12 3m0 0a8.997 8.997 0 0 1 7.843 4.582M12 3a8.997 8.997 0 0 0-7.843 4.582m15.686 0A11.953 11.953 0 0 1 12 10.5c-2.998 0-5.74-1.1-7.843-2.918m15.686 0A8.959 8.959 0 0 1 21 12c0 .778-.099 1.533-.284 2.253m0 0A17.919 17.919 0 0 1 12 16.5c-3.162 0-6.133-.815-8.716-2.247m0 0A9.015 9.015 0 0 1 3 12c0-1.605.42-3.113 1.157-4.418" />
                  </svg>
                </button>
                <button
                  v-if="serversStore.canManageChannels"
                  @click.stop="openEditModal(channel)"
                  class="text-[var(--sb-text-3)] hover:text-[var(--sb-text)] p-0.5 rounded"
                  title="Edit channel"
                >
                  <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.324.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 011.37.49l1.296 2.247a1.125 1.125 0 01-.26 1.431l-1.003.827c-.293.24-.438.613-.431.992a6.759 6.759 0 010 .255c-.007.378.138.75.43.99l1.005.828c.424.35.534.954.26 1.43l-1.298 2.247a1.125 1.125 0 01-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.57 6.57 0 01-.22.128c-.331.183-.581.495-.644.869l-.213 1.28c-.09.543-.56.941-1.11.941h-2.594c-.55 0-1.02-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 01-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 01-1.369-.49l-1.297-2.247a1.125 1.125 0 01.26-1.431l1.004-.827c.292-.24.437-.613.43-.992a6.932 6.932 0 010-.255c.007-.378-.138-.75-.43-.99l-1.004-.828a1.125 1.125 0 01-.26-1.43l1.297-2.247a1.125 1.125 0 011.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.087.22-.128.332-.183.582-.495.644-.869l.214-1.281z" />
                    <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                  </svg>
                </button>
              </div>
            </div>
            <div v-if="dragState?.type === 'text' && dropIndex === textChannels.length && dropIndex !== dragState.index && dropIndex !== dragState.index + 1" class="h-0.5 bg-[#E8521A] rounded-full mx-2 mt-px"></div>
          </div>

          <!-- Audio Channels -->
          <div v-if="audioChannels.length > 0">
            <div class="flex items-center justify-between px-2 mb-2">
              <span class="text-[var(--sb-text-3)] text-[11px] font-bold uppercase tracking-[0.1em]">Voice Channels</span>
            </div>

            <div
              v-for="(channel, idx) in audioChannels"
              :key="channel.id"
              :draggable="canManageChannels"
              @dragstart="onDragStart($event, 'audio', idx)"
              @dragover="onDragOver($event, 'audio', idx)"
              @dragleave="onDragLeave"
              @drop="onDrop($event, 'audio', idx)"
              @dragend="onDragEnd"
            >
              <div v-if="dragState?.type === 'audio' && dropIndex === idx && dropIndex !== dragState.index && dropIndex !== dragState.index + 1" class="h-0.5 bg-[#E8521A] rounded-full mx-2 -mt-px mb-px"></div>
              <button
                @click="selectChannel(channel)"
                class="w-full flex items-center gap-2 px-2.5 py-[7px] rounded-lg text-sm transition-all duration-150"
                :class="[
                  channelsStore.currentChannel?.id === channel.id
                    ? 'bg-[var(--sb-hover)] text-[var(--sb-text)]'
                    : 'text-[var(--sb-text-2)] hover:bg-[var(--sb-hover)] hover:text-[var(--sb-text)]',
                  canManageChannels ? 'cursor-grab active:cursor-grabbing' : 'cursor-pointer',
                ]"
              >
                <svg class="w-4 h-4 opacity-50 shrink-0" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M12 3a1 1 0 0 0-1 1v8a1 1 0 0 0 2 0V4a1 1 0 0 0-1-1ZM6.5 8A.5.5 0 0 0 6 8.5v3a6 6 0 0 0 5 5.91V20H8.5a.5.5 0 0 0 0 1h7a.5.5 0 0 0 0-1H13v-2.59A6 6 0 0 0 18 11.5v-3a.5.5 0 0 0-1 0v3a5 5 0 0 1-10 0v-3a.5.5 0 0 0-.5-.5Z"/>
                </svg>
                <span class="truncate font-medium">{{ channel.name }}</span>
              </button>
              <!-- Connected users -->
              <div v-if="getVoiceUsers(channel.id).length > 0" class="ml-7 mt-1 mb-1.5 space-y-0.5">
                <div
                  v-for="userId in getVoiceUsers(channel.id)"
                  :key="userId"
                  class="flex items-center gap-2 px-1.5 py-1 text-xs text-[var(--sb-text-2)]"
                >
                  <div
                    class="w-5 h-5 rounded-full flex items-center justify-center text-white text-[10px] font-bold shrink-0"
                    :style="getDefaultAvatarStyle(userId)"
                  >
                    {{ (getMember(userId)?.display_name || getMember(userId)?.username || '?')[0].toUpperCase() }}
                  </div>
                  <span class="truncate">{{ getMember(userId)?.display_name || getMember(userId)?.username || 'User' }}</span>
                </div>
              </div>
            </div>
            <div v-if="dragState?.type === 'audio' && dropIndex === audioChannels.length && dropIndex !== dragState.index && dropIndex !== dragState.index + 1" class="h-0.5 bg-[#E8521A] rounded-full mx-2 mt-px"></div>
          </div>

          <!-- Forum Channels -->
          <div v-if="forumChannels.length > 0">
            <div class="flex items-center justify-between px-2 mb-2">
              <span class="text-[var(--sb-text-3)] text-[11px] font-bold uppercase tracking-[0.1em]">Forum Channels</span>
            </div>

            <div
              v-for="(channel, idx) in forumChannels"
              :key="channel.id"
              class="group relative"
              :draggable="canManageChannels"
              @dragstart="onDragStart($event, 'forum', idx)"
              @dragover="onDragOver($event, 'forum', idx)"
              @dragleave="onDragLeave"
              @drop="onDrop($event, 'forum', idx)"
              @dragend="onDragEnd"
            >
              <div v-if="dragState?.type === 'forum' && dropIndex === idx && dropIndex !== dragState.index && dropIndex !== dragState.index + 1" class="h-0.5 bg-[#E8521A] rounded-full mx-2 -mt-px mb-px"></div>
              <button
                @click="selectChannel(channel)"
                class="w-full flex items-center gap-2 px-2.5 py-[7px] rounded-lg text-sm transition-all duration-150"
                :class="[
                  channelsStore.currentChannel?.id === channel.id
                    ? 'bg-[var(--sb-hover)] text-[var(--sb-text)]'
                    : 'text-[var(--sb-text-2)] hover:bg-[var(--sb-hover)] hover:text-[var(--sb-text)]',
                  canManageChannels ? 'cursor-grab active:cursor-grabbing' : 'cursor-pointer',
                ]"
              >
                <svg class="w-4 h-4 opacity-50 shrink-0" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M20 13V6a2 2 0 0 0-2-2H6a2 2 0 0 0-2 2v7m16 0v1a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2v-1m16 0h-2.586a1 1 0 0 0-.707.293l-2.414 2.414a1 1 0 0 1-.707.293h-3.172a1 1 0 0 1-.707-.293l-2.414-2.414A1 1 0 0 0 6.586 13H4" />
                </svg>
                <span class="truncate font-medium">{{ channel.name }}</span>
              </button>
              <button
                v-if="serversStore.canManageChannels"
                @click.stop="openEditModal(channel)"
                class="absolute right-1.5 top-1/2 -translate-y-1/2 opacity-0 group-hover:opacity-100 text-[var(--sb-text-3)] hover:text-[var(--sb-text)] p-0.5 rounded transition-all duration-150"
                title="Edit channel"
              >
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.324.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 011.37.49l1.296 2.247a1.125 1.125 0 01-.26 1.431l-1.003.827c-.293.24-.438.613-.431.992a6.759 6.759 0 010 .255c-.007.378.138.75.43.99l1.005.828c.424.35.534.954.26 1.43l-1.298 2.247a1.125 1.125 0 01-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.57 6.57 0 01-.22.128c-.331.183-.581.495-.644.869l-.213 1.28c-.09.543-.56.941-1.11.941h-2.594c-.55 0-1.02-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 01-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 01-1.369-.49l-1.297-2.247a1.125 1.125 0 01.26-1.431l1.004-.827c.292-.24.437-.613.43-.992a6.932 6.932 0 010-.255c.007-.378-.138-.75-.43-.99l-1.004-.828a1.125 1.125 0 01-.26-1.43l1.297-2.247a1.125 1.125 0 011.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.087.22-.128.332-.183.582-.495.644-.869l.214-1.281z" />
                  <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
              </button>
            </div>
            <div v-if="dragState?.type === 'forum' && dropIndex === forumChannels.length && dropIndex !== dragState.index && dropIndex !== dragState.index + 1" class="h-0.5 bg-[#E8521A] rounded-full mx-2 mt-px"></div>
          </div>
        </div>

        <!-- Loading skeleton while channels are being fetched -->
        <div v-else-if="initialLoading || channelsStore.loading" class="space-y-5 px-1">
          <div>
            <div class="px-1 mb-2">
              <div class="h-2.5 w-24 rounded bg-[var(--sb-hover)] animate-pulse"></div>
            </div>
            <div class="space-y-1">
              <div v-for="i in 4" :key="i" class="flex items-center gap-2 px-2.5 py-[7px] animate-pulse">
                <div class="w-4 h-3 rounded bg-[var(--sb-hover)] opacity-60 shrink-0"></div>
                <div class="h-3 rounded bg-[var(--sb-hover)]" :style="{ width: [80, 60, 100, 70][i - 1] + 'px' }"></div>
              </div>
            </div>
          </div>
          <div>
            <div class="px-1 mb-2">
              <div class="h-2.5 w-28 rounded bg-[var(--sb-hover)] animate-pulse"></div>
            </div>
            <div class="space-y-1">
              <div v-for="i in 2" :key="i" class="flex items-center gap-2 px-2.5 py-[7px] animate-pulse">
                <div class="w-4 h-3 rounded bg-[var(--sb-hover)] opacity-60 shrink-0"></div>
                <div class="h-3 rounded bg-[var(--sb-hover)]" :style="{ width: [75, 90][i - 1] + 'px' }"></div>
              </div>
            </div>
          </div>
        </div>

        <div v-else class="text-[var(--sb-text-3)] text-sm px-2 py-4">
          Select a server to begin
        </div>
      </template>
    </div>

    <UserStatusBar />

    <InviteModal v-if="showInviteModal" @close="showInviteModal = false" />
    <ServerSettingsModal v-if="showSettingsModal" @close="showSettingsModal = false" />
    <ChannelFederationModal v-if="showFederationModal && federationChannel" :channel="federationChannel" @close="closeFederationModal" />

    <!-- Edit channel modal -->
    <Teleport v-if="showEditModal" to="body">
      <div class="fixed inset-0 bg-[var(--backdrop)] backdrop-blur-sm flex items-center justify-center z-50" @click.self="showEditModal = false">
        <div class="bg-[var(--modal-bg)] rounded-2xl p-6 w-full max-w-md shadow-2xl shadow-black/10 animate-scale-in border border-[var(--modal-border)]">
          <h2 class="text-xl font-bold text-[var(--text-1)] mb-5">Edit Channel</h2>
          <div class="space-y-4">
            <div>
              <label class="block text-xs font-semibold text-[var(--text-3)] uppercase tracking-wider mb-1.5">Channel Name</label>
              <input
                v-model="editName"
                @keyup.enter="saveEditChannel"
                @keyup.escape="showEditModal = false"
                placeholder="channel-name"
                class="w-full bg-[var(--surface-3)] text-[var(--text-1)] text-sm px-3 py-2.5 rounded-xl placeholder-[var(--text-4)] border border-[var(--modal-border)] focus:outline-none focus:border-[#E8521A] transition-colors"
              />
            </div>
            <div>
              <label class="block text-xs font-semibold text-[var(--text-3)] uppercase tracking-wider mb-1.5">Channel Topic <span class="normal-case font-normal text-[var(--text-4)]">— optional</span></label>
              <input
                v-model="editTopic"
                @keyup.enter="saveEditChannel"
                @keyup.escape="showEditModal = false"
                placeholder="Add a topic…"
                class="w-full bg-[var(--surface-3)] text-[var(--text-1)] text-sm px-3 py-2.5 rounded-xl placeholder-[var(--text-4)] border border-[var(--modal-border)] focus:outline-none focus:border-[#E8521A] transition-colors"
              />
            </div>
          </div>
          <div class="flex justify-end gap-3 mt-6">
            <button @click="showEditModal = false" class="text-[var(--text-3)] hover:text-[var(--text-1)] px-4 py-2.5 rounded-xl cursor-pointer font-medium transition-colors duration-150">
              Cancel
            </button>
            <button
              @click="saveEditChannel"
              :disabled="!editName.trim()"
              class="bg-[#E8521A] text-white px-5 py-2.5 rounded-xl hover:bg-[#D44818] cursor-pointer font-semibold transition-colors duration-150 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Save
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Leave server confirmation modal -->
    <Teleport v-if="showLeaveModal" to="body">
      <div class="fixed inset-0 bg-[var(--backdrop)] backdrop-blur-sm flex items-center justify-center z-50" @click.self="showLeaveModal = false">
        <div class="bg-[var(--modal-bg)] rounded-2xl p-6 w-full max-w-md shadow-2xl shadow-black/10 animate-scale-in border border-[var(--modal-border)]">
          <h2 class="text-xl font-bold text-[var(--text-1)] mb-2">Leave Server</h2>
          <p class="text-[var(--text-3)] text-sm mb-6">
            Are you sure you want to leave <span class="font-semibold text-[var(--text-1)]">{{ serversStore.currentServer?.name }}</span>?
          </p>
          <div class="flex justify-end gap-3">
            <button @click="showLeaveModal = false" class="text-[var(--text-3)] hover:text-[var(--text-1)] px-4 py-2.5 rounded-xl cursor-pointer font-medium transition-colors duration-150">
              Cancel
            </button>
            <button
              @click="leaveServer"
              class="bg-[#E8521A] text-white px-5 py-2.5 rounded-xl hover:bg-[#D44818] cursor-pointer font-semibold transition-colors duration-150"
            >
              Leave Server
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- User notification settings modal -->
    <Teleport v-if="showUserSettingsModal" to="body">
      <div class="fixed inset-0 bg-[var(--backdrop)] backdrop-blur-sm flex items-center justify-center z-50" @click.self="showUserSettingsModal = false">
        <div class="bg-[var(--modal-bg)] rounded-2xl p-6 w-full max-w-sm shadow-2xl shadow-black/10 animate-scale-in border border-[var(--modal-border)]">
          <h2 class="text-xl font-bold text-[var(--text-1)] mb-1">Notification Settings</h2>
          <p class="text-[var(--text-4)] text-xs mb-5">{{ serversStore.currentServer?.name }}</p>

          <div class="space-y-3">
            <button
              @click="toggleServerMute"
              class="w-full flex items-center justify-between gap-3 px-4 py-3 rounded-xl border transition-colors duration-150"
              :class="isServerMuted
                ? 'bg-[#E8521A]/8 border-[#E8521A]/30'
                : 'bg-[var(--surface-2)] border-[var(--surface-border)] hover:bg-[var(--surface-3)]'"
            >
              <div class="flex items-center gap-3">
                <svg v-if="isServerMuted" class="w-5 h-5 text-[#E8521A] shrink-0" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M17.25 9.75 19.5 12m0 0 2.25 2.25M19.5 12l2.25-2.25M19.5 12l-2.25 2.25m-10.5-6 4.72-4.72a.75.75 0 0 1 1.28.53v15.88a.75.75 0 0 1-1.28.53l-4.72-4.72H4.51c-.88 0-1.704-.507-1.938-1.354A9.009 9.009 0 0 1 2.25 12c0-.83.112-1.633.322-2.396C2.806 8.756 3.63 8.25 4.51 8.25H6.75Z" />
                </svg>
                <svg v-else class="w-5 h-5 text-[var(--text-3)] shrink-0" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M19.114 5.636a9 9 0 0 1 0 12.728M16.463 8.288a5.25 5.25 0 0 1 0 7.424M6.75 8.25l4.72-4.72a.75.75 0 0 1 1.28.53v15.88a.75.75 0 0 1-1.28.53l-4.72-4.72H4.51c-.88 0-1.704-.507-1.938-1.354A9.009 9.009 0 0 1 2.25 12c0-.83.112-1.633.322-2.396C2.806 8.756 3.63 8.25 4.51 8.25H6.75Z" />
                </svg>
                <div class="text-left">
                  <div class="text-sm font-medium text-[var(--text-1)]">Mute Server</div>
                  <div class="text-xs text-[var(--text-4)]">Suppress all unread badges for this server</div>
                </div>
              </div>
              <div
                class="w-9 h-5 rounded-full transition-colors duration-200 shrink-0 relative"
                :class="isServerMuted ? 'bg-[#E8521A]' : 'bg-[var(--surface-border)]'"
              >
                <div
                  class="absolute top-0.5 w-4 h-4 rounded-full bg-white shadow transition-transform duration-200"
                  :class="isServerMuted ? 'translate-x-4' : 'translate-x-0.5'"
                ></div>
              </div>
            </button>
          </div>

          <div class="flex justify-end mt-6">
            <button
              @click="showUserSettingsModal = false"
              class="text-[var(--text-3)] hover:text-[var(--text-1)] px-4 py-2.5 rounded-xl cursor-pointer font-medium transition-colors duration-150"
            >
              Done
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
