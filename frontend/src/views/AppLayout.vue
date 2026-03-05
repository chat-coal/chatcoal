<script setup>
import { ref, onMounted, onUnmounted, watch, watchEffect, provide } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useServersStore } from '@/stores/servers'
import { useChannelsStore } from '@/stores/channels'
import { useMessagesStore } from '@/stores/messages'
import { useVoiceStore } from '@/stores/voice'
import { useDMsStore } from '@/stores/dms'
import { useUnreadStore } from '@/stores/unread'
import { useTypingStore } from '@/stores/typing'
import { useSearchStore } from '@/stores/search'
import { useForumStore } from '@/stores/forum'
import { useNotificationSettingsStore } from '@/stores/notificationSettings'
import { connect, disconnect, on } from '@/services/websocket.service'
import { useVersionCheck } from '@/composables/useVersionCheck'
import { cleanup as voiceCleanup } from '@/services/voice.service'
import api from '@/services/api.service'
import ServerSidebar from '@/components/ServerSidebar.vue'
import ChannelSidebar from '@/components/ChannelSidebar.vue'
import ChatArea from '@/components/ChatArea.vue'
import MemberList from '@/components/MemberList.vue'
import SearchPanel from '@/components/SearchPanel.vue'
import ToastContainer from '@/components/ToastContainer.vue'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const serversStore = useServersStore()
const channelsStore = useChannelsStore()
const messagesStore = useMessagesStore()
const voiceStore = useVoiceStore()
const dmsStore = useDMsStore()
const unreadStore = useUnreadStore()
const typingStore = useTypingStore()
const searchStore = useSearchStore()
const forumStore = useForumStore()
const notifStore = useNotificationSettingsStore()

const isElectron = !!window.electronAPI?.isElectron
const showMembers = ref(true)
provide('showMembers', showMembers)

const mobileNavOpen = ref(false)
provide('closeMobileNav', () => { mobileNavOpen.value = false })

const mobileShowMembers = ref(false)
provide('openMobileMembers', () => { mobileShowMembers.value = true })
const pendingChannelId = ref(null)
const kickedModal = ref(null)
const bannedModal = ref(null)

const wsConnected = ref(true)
const { updateAvailable, electronUpdate, reloading, reload, openDownload, dismissUpdate } = useVersionCheck()

// Email verification banner state
const resendingEmail = ref(false)
const refreshingVerification = ref(false)

async function handleResendEmail() {
  resendingEmail.value = true
  try {
    await authStore.resendVerificationEmail()
  } catch { /* ignore */ }
  finally { resendingEmail.value = false }
}

async function handleRefreshVerification() {
  refreshingVerification.value = true
  try {
    await authStore.refreshEmailVerification()
    // Start WebSocket if email is now verified (wsConnected defaults to
    // true so we can't rely on it — just check whether we need to connect)
    if (authStore.dbUser?.email_verified) {
      connect(authStore.dbUser.id)
    }
  } catch { /* ignore */ }
  finally { refreshingVerification.value = false }
}

let unsubConnected, unsubDisconnected
let unsubMessage, unsubEdit, unsubDelete, unsubVoiceState
let unsubDMMessage, unsubDMEdit, unsubDMDelete
let unsubReaction, unsubDMReaction
let unsubMemberUpdate, unsubKicked, unsubBanned
let unsubTyping, unsubDMTyping
let unsubForumPost, unsubForumPostEdit, unsubForumPostDelete, unsubForumMessage, unsubForumMessageEdit, unsubForumMessageDelete
let unsubMessagePin, unsubMessageUnpin
let unsubChannelCreate, unsubChannelDelete
let unsubServerUpdate
let unsubPresence
let unsubEmbedUpdate, unsubDMEmbedUpdate

const chatArea = ref(null)
const initialLoading = ref(true)
provide('initialLoading', initialLoading)
const minLoadingTimer = new Promise((r) => setTimeout(r, 500))

onMounted(async () => {
  await serversStore.fetchServers()
  await dmsStore.fetchDMChannels()
  await unreadStore.fetchUnreadCounts()
  await notifStore.fetchSettings()

  // Restore last visited location on fresh page load
  // Either hydrate from current URL params, or fall back to localStorage
  const serverId = route.params.serverId || null
  const channelId = route.params.channelId || null
  const dmChannelId = route.params.dmChannelId || null

  if (serverId) {
    // URL already has server/channel — hydrate stores from route params
    const server = serversStore.servers.find((s) => s.id === serverId)
    if (server) {
      if (channelId) pendingChannelId.value = channelId
      serversStore.selectServer(server)
    }
  } else if (dmChannelId) {
    const dmChannel = dmsStore.dmChannels.find((c) => c.id === dmChannelId)
    if (dmChannel) dmsStore.selectDM(dmChannel)
  } else if (route.path === '/channels/@me' || route.path === '/channels/') {
    // No params in URL — try restoring from localStorage
    const saved = localStorage.getItem('lastChannel')
    if (saved && saved !== '/channels/@me') {
      const match = saved.match(/^\/channels\/(\d+)(?:\/(\d+))?$/)
      const dmMatch = saved.match(/^\/channels\/@me\/(\d+)$/)
      if (match) {
        const server = serversStore.servers.find((s) => s.id === match[1])
        if (server) {
          if (match[2]) pendingChannelId.value = match[2]
          serversStore.selectServer(server)
        }
      } else if (dmMatch) {
        const dmChannel = dmsStore.dmChannels.find((c) => c.id === dmMatch[1])
        if (dmChannel) {
          dmsStore.selectDM(dmChannel)
          router.replace(saved)
        }
      }
    }
  }

  // Keep skeleton visible for at least 500ms to avoid a flash
  await minLoadingTimer
  initialLoading.value = false

  // Single multiplexed WebSocket connection for all servers + DMs
  // Skip for unverified email users — they can't interact yet and the
  // server would close the socket, causing a reconnect loop.
  if (authStore.dbUser && authStore.dbUser.email_verified !== false) {
    connect(authStore.dbUser.id)
  }

  unsubConnected = on('connected', () => { wsConnected.value = true })
  unsubDisconnected = on('disconnected', () => { wsConnected.value = false })

  // Server channel messages
  unsubMessage = on('message', (msg, serverId) => {
    typingStore.removeTyper('channel', msg.channel_id, msg.author_id)
    if (channelsStore.currentChannel && msg.channel_id === channelsStore.currentChannel.id) {
      messagesStore.addMessage(msg)
      // Mark as read since we're viewing this channel
      api.markChannelAsRead(msg.channel_id, msg.id).catch(() => {})
    } else if (msg.author_id !== authStore.dbUser?.id) {
      // Increment unread for non-active channel (unless muted)
      if (!notifStore.isEffectivelyMuted(msg.channel_id, serverId)) {
        unreadStore.incrementChannel(msg.channel_id, serverId)
      }
    }
  })
  unsubEdit = on('message_edit', (msg) => {
    messagesStore.updateMessage(msg)
  })
  unsubDelete = on('message_delete', (msg) => {
    messagesStore.removeMessage(msg.id)
  })
  unsubVoiceState = on('voice_state_update', (data) => {
    voiceStore.handleVoiceStateUpdate(data)
  })
  unsubReaction = on('reaction_update', (data) => {
    const msg = messagesStore.messages.find((m) => m.id === data.message_id)
    if (msg) {
      const idx = messagesStore.messages.indexOf(msg)
      messagesStore.messages[idx] = { ...msg, reactions: data.reactions }
    }
  })

  // DM messages
  unsubDMMessage = on('dm_message', (msg) => {
    typingStore.removeTyper('dm', msg.dm_channel_id, msg.author_id)
    if (dmsStore.currentDMChannel && msg.dm_channel_id === dmsStore.currentDMChannel.id) {
      dmsStore.addMessage(msg)
      // Mark as read since we're viewing this DM
      api.markDMAsRead(msg.dm_channel_id, msg.id).catch(() => {})
    } else {
      // Add to channel list / move to top
      dmsStore.addMessage(msg)
      if (msg.author_id !== authStore.dbUser?.id) {
        unreadStore.incrementDM(msg.dm_channel_id)
      }
    }

    // If channel not in our list yet, refetch
    if (!dmsStore.dmChannels.find((c) => c.id === msg.dm_channel_id)) {
      dmsStore.fetchDMChannels()
    }
  })
  unsubDMEdit = on('dm_message_edit', (msg) => {
    dmsStore.updateMessage(msg)
  })
  unsubDMDelete = on('dm_message_delete', (msg) => {
    if (dmsStore.currentDMChannel && msg.dm_channel_id === dmsStore.currentDMChannel.id) {
      dmsStore.removeMessage(msg.id)
    }
  })
  unsubMemberUpdate = on('member_update', (data) => {
    if (data.user_id === authStore.dbUser?.id) {
      serversStore.setMemberRole(data.role)
    }
  })
  unsubKicked = on('kicked', (data) => {
    serversStore.removeServer(data.server_id)
    channelsStore.clear()
    messagesStore.clear()
    router.replace('/channels/@me')
    kickedModal.value = data.server_name || 'a server'
  })
  unsubBanned = on('banned', (data) => {
    serversStore.removeServer(data.server_id)
    channelsStore.clear()
    messagesStore.clear()
    router.replace('/channels/@me')
    bannedModal.value = data.server_name || 'a server'
  })
  unsubDMReaction = on('dm_reaction_update', (data) => {
    const msg = dmsStore.messages.find((m) => m.id === data.message_id)
    if (msg) {
      const idx = dmsStore.messages.indexOf(msg)
      dmsStore.messages[idx] = { ...msg, reactions: data.reactions }
    }
  })

  // Embed updates
  unsubEmbedUpdate = on('message_embed_update', (data) => {
    const msg = messagesStore.messages.find((m) => m.id === data.id)
    if (msg) {
      const idx = messagesStore.messages.indexOf(msg)
      messagesStore.messages[idx] = { ...msg, embeds: data.embeds }
    }
  })
  unsubDMEmbedUpdate = on('dm_embed_update', (data) => {
    const msg = dmsStore.messages.find((m) => m.id === data.id)
    if (msg) {
      const idx = dmsStore.messages.indexOf(msg)
      dmsStore.messages[idx] = { ...msg, embeds: data.embeds }
    }
  })

  // Typing indicators
  unsubTyping = on('typing', (data) => {
    if (data.user_id !== authStore.dbUser?.id) {
      typingStore.addTyper('channel', data.channel_id, data.user_id)
    }
  })
  unsubDMTyping = on('dm_typing', (data) => {
    if (data.user_id !== authStore.dbUser?.id) {
      typingStore.addTyper('dm', data.dm_channel_id, data.user_id)
    }
  })

  // Forum events
  unsubForumPost = on('forum_post', (post) => {
    if (post.channel_id === channelsStore.currentChannel?.id) {
      forumStore.addPost(post)
    }
  })
  unsubForumPostEdit = on('forum_post_edit', (post) => {
    if (post.channel_id === channelsStore.currentChannel?.id) {
      forumStore.updatePost(post)
    }
  })
  unsubForumPostDelete = on('forum_post_delete', (data) => {
    if (data.channel_id === channelsStore.currentChannel?.id) {
      forumStore.removePost(data.id)
    }
  })
  unsubForumMessage = on('forum_message', (msg) => {
    // Update reply count in the posts list regardless of current post
    forumStore.incrementPostReplyCount(msg.forum_post_id, msg.created_at)
    // Only add to messages array if viewing this post
    if (forumStore.currentPost && msg.forum_post_id === forumStore.currentPost.id) {
      forumStore.addMessage(msg)
    }
  })
  unsubForumMessageEdit = on('forum_message_edit', (msg) => {
    forumStore.updateMessage(msg)
  })
  unsubForumMessageDelete = on('forum_message_delete', (msg) => {
    forumStore.removeMessage(msg.id, msg.forum_post_id)
  })

  unsubMessagePin = on('message_pin', (data) => {
    if (data.channel_id === channelsStore.currentChannel?.id) {
      chatArea.value?.handlePinEvent(data.pin)
    }
  })
  unsubMessageUnpin = on('message_unpin', (data) => {
    if (data.channel_id === channelsStore.currentChannel?.id) {
      chatArea.value?.handleUnpinEvent(data.message_id)
    }
  })

  // Channel create/delete
  unsubChannelCreate = on('channel_create', (channel, serverId) => {
    if (serverId === serversStore.currentServer?.id) {
      channelsStore.addChannel(channel)
    }
  })
  unsubChannelDelete = on('channel_delete', (data, serverId) => {
    if (serverId === serversStore.currentServer?.id) {
      const wasCurrentChannel = channelsStore.currentChannel?.id === data.id
      channelsStore.removeChannel(data.id)
      if (wasCurrentChannel) {
        const next = channelsStore.channels[0]
        if (next) {
          channelsStore.selectChannel(next)
          router.replace(`/channels/${serverId}/${next.id}`)
        } else {
          router.replace(`/channels/${serverId}`)
        }
      }
    }
  })

  // Keep own status bar in sync with backend presence changes
  unsubPresence = on('presence_update', (data) => {
    if (data.user_id === authStore.dbUser?.id) {
      authStore.dbUser.status = data.status
    }
  })

  // Server settings update (e.g. name, icon, visibility, system announcements)
  unsubServerUpdate = on('server_update', (data) => {
    const idx = serversStore.servers.findIndex((s) => s.id === data.id)
    if (idx !== -1) serversStore.servers[idx] = data
    if (serversStore.currentServer?.id === data.id) serversStore.currentServer = data
  })
})

onUnmounted(() => {
  voiceCleanup()
  voiceStore.clear()
  serversStore.clear()
  channelsStore.clear()
  messagesStore.clear()
  dmsStore.clearAll()
  unreadStore.clear()
  disconnect()
  unsubConnected?.()
  unsubDisconnected?.()
  unsubMessage?.()
  unsubEdit?.()
  unsubDelete?.()
  unsubVoiceState?.()
  unsubDMMessage?.()
  unsubDMEdit?.()
  unsubDMDelete?.()
  unsubReaction?.()
  unsubDMReaction?.()
  unsubMemberUpdate?.()
  unsubKicked?.()
  unsubBanned?.()
  unsubTyping?.()
  unsubDMTyping?.()
  unsubForumPost?.()
  unsubForumPostEdit?.()
  unsubForumPostDelete?.()
  unsubForumMessage?.()
  unsubForumMessageEdit?.()
  unsubForumMessageDelete?.()
  unsubMessagePin?.()
  unsubMessageUnpin?.()
  unsubChannelCreate?.()
  unsubChannelDelete?.()
  unsubServerUpdate?.()
  unsubPresence?.()
  unsubEmbedUpdate?.()
  unsubDMEmbedUpdate?.()
})

watch(
  () => serversStore.currentServer,
  async (server) => {
    if (server) {
      dmsStore.selectDM(null)
      dmsStore.clear()
      voiceCleanup()
      voiceStore.clear()
      channelsStore.clear()
      messagesStore.clear()
      await channelsStore.fetchChannels(server.id)
      voiceStore.fetchVoiceStates(server.id)
      if (channelsStore.channels.length > 0 && !channelsStore.currentChannel) {
        const targetId = pendingChannelId.value
        pendingChannelId.value = null
        const channel = targetId
          ? channelsStore.channels.find((c) => c.id === targetId)
          : null
        const selected = channel || channelsStore.channels[0]
        channelsStore.selectChannel(selected)
        router.replace(`/channels/${server.id}/${selected.id}`)
      }
    } else {
      voiceCleanup()
      voiceStore.clear()
      channelsStore.clear()
      messagesStore.clear()
    }
  },
)

watch(
  () => channelsStore.currentChannel,
  async (channel) => {
    if (channel) {
      messagesStore.clear()
      forumStore.clear()
      searchStore.close()
      if (channel.type === 'forum') {
        // Forum channels load posts, not messages
        await forumStore.fetchPosts(channel.id)
      } else {
        await messagesStore.fetchMessages(channel.id)
        // Mark as read
        if (serversStore.currentServer && messagesStore.messages.length > 0) {
          const lastMsg = messagesStore.messages[messagesStore.messages.length - 1]
          unreadStore.markChannelRead(channel.id, serversStore.currentServer.id)
          api.markChannelAsRead(channel.id, lastMsg.id).catch(() => {})
        }
      }
    }
  },
)

// Handle browser back/forward navigation between DM and server contexts.
// The route URL changes but the existing store watchers don't re-fire because
// the store values haven't changed (e.g. currentServer is still set, so its
// watcher never triggers to reload the cleared channel list).
watch(
  () => route.params,
  async (params) => {
    const { serverId, channelId, dmChannelId } = params

    // Back from DM → server channel
    if (serverId && dmsStore.currentDMChannel) {
      dmsStore.selectDM(null)
      const server = serversStore.servers.find((s) => s.id === serverId)
      if (!server) return
      if (serversStore.currentServer?.id !== serverId) {
        if (channelId) pendingChannelId.value = channelId
        serversStore.selectServer(server)
      } else {
        // Same server but channels were cleared when DM was selected — reload them
        if (channelId) pendingChannelId.value = channelId
        await channelsStore.fetchChannels(serverId)
        voiceStore.fetchVoiceStates(serverId)
        if (channelsStore.channels.length > 0 && !channelsStore.currentChannel) {
          const targetId = pendingChannelId.value
          pendingChannelId.value = null
          const ch = targetId ? channelsStore.channels.find((c) => c.id === targetId) : null
          const selected = ch || channelsStore.channels[0]
          channelsStore.selectChannel(selected)
          router.replace(`/channels/${serverId}/${selected.id}`)
        }
      }
    }

    // Back from server → DM
    else if (dmChannelId && !dmsStore.currentDMChannel) {
      const dm = dmsStore.dmChannels.find((c) => c.id === dmChannelId)
      if (dm) dmsStore.selectDM(dm)
    }
  }
)

watch(
  () => dmsStore.currentDMChannel,
  async (dm) => {
    if (dm) {
      channelsStore.clear()
      messagesStore.clear()
      dmsStore.clear()
      await dmsStore.fetchMessages(dm.id)
      // Mark as read
      unreadStore.markDMRead(dm.id)
      if (dmsStore.messages.length > 0) {
        const lastMsg = dmsStore.messages[dmsStore.messages.length - 1]
        api.markDMAsRead(dm.id, lastMsg.id).catch(() => {})
      }
    }
  },
)

watchEffect(() => {
  const dm = dmsStore.currentDMChannel
  if (dm) {
    const other = dm.user1?.id === authStore.dbUser?.id ? dm.user2 : dm.user1
    const name = other?.display_name || 'Direct Message'
    document.title = `chatcoal | ${name}`
    return
  }
  const channel = channelsStore.currentChannel
  const server = serversStore.currentServer
  if (channel && server) {
    document.title = `chatcoal | ${channel.name} | ${server.name}`
    return
  }
  document.title = 'chatcoal'
})
</script>

<template>
  <div class="flex flex-col overflow-hidden bg-[var(--surface)]" :class="isElectron ? 'h-[calc(100dvh-1.75rem)]' : 'h-dvh'">
    <!-- Connection lost banner -->
    <Transition name="slide-banner">
      <div v-if="!wsConnected" class="shrink-0 flex items-center justify-center gap-2 bg-amber-600 text-white text-xs font-medium py-1.5 px-4">
        <div class="w-3 h-3 border-[1.5px] border-white border-t-transparent rounded-full animate-spin"></div>
        Reconnecting&hellip;
      </div>
    </Transition>

    <!-- New version available banner -->
    <Transition name="slide-banner">
      <div v-if="updateAvailable" class="shrink-0 flex items-center justify-center gap-2 bg-blue-600 text-white text-xs font-medium py-1.5 px-4">
        <template v-if="electronUpdate">
          Update available: v{{ electronUpdate.latestVersion }}
          <button @click="openDownload" class="underline underline-offset-2 hover:text-blue-200 transition-colors cursor-pointer">Download</button>
          <button @click="dismissUpdate" class="text-blue-200 hover:text-white transition-colors cursor-pointer ml-1">&times;</button>
        </template>
        <template v-else-if="reloading">
          <div class="w-3 h-3 border-[1.5px] border-white border-t-transparent rounded-full animate-spin"></div>
          Updating&hellip;
        </template>
        <template v-else>
          A new version is available.
          <button @click="reload" class="underline underline-offset-2 hover:text-blue-200 transition-colors cursor-pointer">Refresh</button>
        </template>
      </div>
    </Transition>

    <!-- Email verification banner -->
    <Transition name="slide-banner">
      <div
        v-if="authStore.dbUser && !authStore.dbUser.is_anonymous && authStore.dbUser.email_verified === false"
        class="shrink-0 flex items-center justify-center gap-3 bg-[#E8521A] text-white text-xs font-medium py-1.5 px-4"
      >
        Verify your email to unlock all features.
        <button @click="handleResendEmail" :disabled="resendingEmail" class="underline underline-offset-2 hover:text-orange-200 transition-colors cursor-pointer disabled:opacity-50">
          {{ resendingEmail ? 'Sending...' : 'Resend email' }}
        </button>
        <button @click="handleRefreshVerification" :disabled="refreshingVerification" class="underline underline-offset-2 hover:text-orange-200 transition-colors cursor-pointer disabled:opacity-50">
          {{ refreshingVerification ? 'Checking...' : "I've verified" }}
        </button>
      </div>
    </Transition>

    <div class="flex flex-1 min-h-0">
    <!-- Mobile backdrop -->
    <div
      v-if="mobileNavOpen"
      class="fixed inset-0 z-30 bg-black/50 lg:hidden"
      @click="mobileNavOpen = false"
    ></div>

    <!-- Navigation drawer (server + channel sidebars) -->
    <div
      class="fixed inset-y-0 left-0 z-40 flex lg:static lg:z-auto transition-transform duration-300 ease-out"
      :class="mobileNavOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'"
    >
      <ServerSidebar />
      <ChannelSidebar />
    </div>

    <ChatArea ref="chatArea" @toggle-members="showMembers = !showMembers" :show-members="showMembers" :initial-loading="initialLoading" @open-nav="mobileNavOpen = true" />
    <SearchPanel v-if="searchStore.isOpen" />
    <Transition v-else name="slide-members">
      <MemberList v-show="showMembers && serversStore.currentServer" />
    </Transition>
    </div>
  </div>

  <!-- Mobile members bottom sheet -->
  <Teleport to="body">
    <div v-if="mobileShowMembers" class="fixed inset-0 z-50 lg:hidden flex flex-col justify-end">
      <div class="absolute inset-0 bg-black/50 backdrop-blur-sm" @click="mobileShowMembers = false"></div>
      <div class="relative bg-[var(--surface)] rounded-t-2xl max-h-[70vh] flex flex-col shadow-2xl">
        <div class="flex items-center justify-between px-5 py-4 border-b border-[var(--surface-border)] shrink-0">
          <h3 class="font-semibold text-[var(--text-1)]">Members</h3>
          <button
            @click="mobileShowMembers = false"
            class="text-[var(--text-4)] hover:text-[var(--text-1)] p-1.5 rounded-lg hover:bg-[var(--surface-3)] transition-colors duration-150"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M18 6L6 18M6 6l12 12" />
            </svg>
          </button>
        </div>
        <MemberList :in-bottom-sheet="true" />
      </div>
    </div>
  </Teleport>

  <ToastContainer />

  <Teleport to="body">
    <div v-if="kickedModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" @click.self="kickedModal = null">
      <div class="bg-[var(--surface-overlay)] rounded-2xl p-6 w-full max-w-sm shadow-xl text-center">
        <h2 class="text-lg font-semibold text-[var(--text-primary)] mb-2">You were kicked</h2>
        <p class="text-sm text-[var(--text-secondary)] mb-6">You have been removed from <strong>{{ kickedModal }}</strong>.</p>
        <button @click="kickedModal = null" class="px-5 py-2 bg-[var(--accent)] text-white rounded-lg hover:opacity-90 transition-opacity">
          OK
        </button>
      </div>
    </div>
  </Teleport>

  <Teleport to="body">
    <div v-if="bannedModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" @click.self="bannedModal = null">
      <div class="bg-[var(--surface-overlay)] rounded-2xl p-6 w-full max-w-sm shadow-xl text-center">
        <h2 class="text-lg font-semibold text-[var(--text-primary)] mb-2">You were banned</h2>
        <p class="text-sm text-[var(--text-secondary)] mb-6">You have been banned from <strong>{{ bannedModal }}</strong>.</p>
        <button @click="bannedModal = null" class="px-5 py-2 bg-[var(--accent)] text-white rounded-lg hover:opacity-90 transition-opacity">
          OK
        </button>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.slide-members-enter-active,
.slide-members-leave-active {
  transition: width 0.2s ease, opacity 0.2s ease;
  overflow: hidden;
}
.slide-members-enter-from,
.slide-members-leave-to {
  width: 0 !important;
  opacity: 0;
}
.slide-banner-enter-active,
.slide-banner-leave-active {
  transition: max-height 0.25s ease, opacity 0.25s ease;
  overflow: hidden;
}
.slide-banner-enter-from,
.slide-banner-leave-to {
  max-height: 0;
  opacity: 0;
}
.slide-banner-enter-to,
.slide-banner-leave-from {
  max-height: 2rem;
}
</style>
