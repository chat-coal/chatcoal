<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { API_URL } from '@/services/api.service'
import { useServersStore } from '@/stores/servers'
import { useChannelsStore } from '@/stores/channels'
import { useUnreadStore } from '@/stores/unread'
import { useNotificationSettingsStore } from '@/stores/notificationSettings'
import { useDMsStore } from '@/stores/dms'
import CreateServerModal from './CreateServerModal.vue'
import JoinServerModal from './JoinServerModal.vue'
import logoSvg from '@/assets/logo.svg'

const router = useRouter()
const serversStore = useServersStore()
const channelsStore = useChannelsStore()
const unreadStore = useUnreadStore()
const notifStore = useNotificationSettingsStore()
const dmsStore = useDMsStore()

const showCreate = ref(false)
const showJoin = ref(false)
const showExplore = ref(false)
const contextMenu = ref({ visible: false, serverId: null, top: 0, left: 0 })

function showContextMenu(event, server) {
  event.preventDefault()
  contextMenu.value = {
    visible: true,
    serverId: server.id,
    serverName: server.name,
    top: event.clientY,
    left: event.clientX,
  }
}

function closeContextMenu() {
  contextMenu.value.visible = false
}

function toggleServerMute() {
  notifStore.toggleMute('server', contextMenu.value.serverId)
  closeContextMenu()
}

const tooltip = ref({ visible: false, text: '', top: 0, left: 0 })
let hideTimeout = null

function showTooltip(event, text) {
  if (window.matchMedia('(hover: none)').matches) return
  clearTimeout(hideTimeout)
  const rect = event.currentTarget.getBoundingClientRect()
  tooltip.value = {
    visible: true,
    text,
    top: rect.top + rect.height / 2,
    left: rect.right + 12,
  }
}

function hideTooltip() {
  hideTimeout = setTimeout(() => {
    tooltip.value.visible = false
  }, 50)
}

function selectServer(server) {
  if (serversStore.currentServer?.id === server.id) return
  dmsStore.selectDM(null)
  serversStore.selectServer(server)
  channelsStore.clear()
  router.push(`/channels/${server.id}`)
}

function goHome() {
  serversStore.selectServer(null)
  dmsStore.selectDM(null)
  channelsStore.clear()
  router.push('/channels/@me')
}

function getInitials(name) {
  return name
    .split(' ')
    .map((w) => w[0])
    .join('')
    .slice(0, 2)
    .toUpperCase()
}
</script>

<template>
  <div class="w-[68px] bg-[var(--sb-bg)] flex flex-col items-center py-3 gap-2 overflow-y-auto scrollbar-hide shrink-0 relative noise-texture">
    <!-- Home button -->
    <div class="relative w-full flex items-center justify-center">
      <!-- Active pill indicator -->
      <div
        class="absolute left-0 w-1 rounded-r-full bg-white transition-all duration-200"
        :class="!serversStore.currentServer ? 'h-10 opacity-100' : unreadStore.totalDMUnread > 0 ? 'h-2 opacity-100' : 'h-0 opacity-0'"
      ></div>
      <button
        class="relative z-10 w-11 h-11 rounded-full bg-[#1E1F22] text-white flex items-center justify-center hover:rounded-[14px] transition-all duration-200 cursor-pointer shadow-lg shadow-black/20"
        @click="goHome"
        @mouseenter="showTooltip($event, 'Direct Messages')"
        @mouseleave="hideTooltip"
      >
        <img :src="logoSvg" alt="Home" class="w-6 h-6" />
        <!-- DM unread badge -->
        <span
          v-if="unreadStore.totalDMUnread > 0"
          class="absolute -bottom-0.5 -right-0.5 min-w-[18px] h-[18px] bg-[#E8521A] border-2 border-[var(--sb-bg)] rounded-full flex items-center justify-center text-white text-[10px] font-bold px-1"
        >
          {{ unreadStore.totalDMUnread > 99 ? '99+' : unreadStore.totalDMUnread }}
        </span>
      </button>
    </div>

    <div class="relative z-10 w-8 h-px bg-[var(--sb-border)] my-1"></div>

    <!-- Server icons -->
    <div
      v-for="server in serversStore.servers"
      :key="server.id"
      class="relative w-full flex items-center justify-center"
    >
      <!-- Active pill indicator -->
      <div
        class="absolute left-0 w-1 rounded-r-full bg-white transition-all duration-200"
        :class="
          serversStore.currentServer?.id === server.id
            ? 'h-10 opacity-100'
            : (!notifStore.mutedServers.has(server.id) && unreadStore.getServerUnread(server.id) > 0)
              ? 'h-2 opacity-100'
              : 'h-0 opacity-0'
        "
      ></div>
      <button
        @click="selectServer(server)"
        @contextmenu="showContextMenu($event, server)"
        @mouseenter="showTooltip($event, server.name + (notifStore.mutedServers.has(server.id) ? ' (muted)' : ''))"
        @mouseleave="hideTooltip"
        class="relative z-10 w-11 h-11 rounded-full flex items-center justify-center text-[var(--sb-text)] font-semibold text-xs transition-all duration-200 cursor-pointer shrink-0 overflow-hidden"
        :class="
          serversStore.currentServer?.id === server.id
            ? 'rounded-[14px] bg-[#E8521A] shadow-lg shadow-[#E8521A]/20'
            : 'bg-[var(--sb-hover)] hover:rounded-[14px] hover:bg-[#E8521A]/70'
        "
      >
        <img v-if="server.icon_url" :src="`${API_URL}${server.icon_url}`" class="w-full h-full rounded-[inherit] object-cover" />
        <span v-else>{{ getInitials(server.name) }}</span>
      </button>
    </div>

    <!-- Explore button -->
    <button
      @click="showExplore = true"
      @mouseenter="showTooltip($event, 'Explore Servers')"
      @mouseleave="hideTooltip"
      class="relative z-10 w-11 h-11 rounded-full bg-[var(--sb-hover)] text-[#D4782A] flex items-center justify-center hover:rounded-[14px] hover:bg-[#D4782A] hover:text-white transition-all duration-200 cursor-pointer shrink-0"
    >
      <svg class="w-5 h-5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" d="M12 21a9.004 9.004 0 0 0 8.716-6.747M12 21a9.004 9.004 0 0 1-8.716-6.747M12 21c2.485 0 4.5-4.03 4.5-9S14.485 3 12 3m0 18c-2.485 0-4.5-4.03-4.5-9S9.515 3 12 3m0 0a8.997 8.997 0 0 1 7.843 4.582M12 3a8.997 8.997 0 0 0-7.843 4.582m15.686 0A11.953 11.953 0 0 1 12 10.5c-2.998 0-5.74-1.1-7.843-2.918m15.686 0A8.959 8.959 0 0 1 21 12c0 .778-.099 1.533-.284 2.253m0 0A17.919 17.919 0 0 1 12 16.5c-3.162 0-6.133-.815-8.716-2.247m0 0A9.015 9.015 0 0 1 3 12c0-1.605.42-3.113 1.157-4.418" />
      </svg>
    </button>

    <!-- Add server button -->
    <button
      @click="showCreate = true"
      @mouseenter="showTooltip($event, 'Create Server')"
      @mouseleave="hideTooltip"
      class="relative z-10 w-11 h-11 rounded-full bg-[var(--sb-hover)] text-[#D4782A] flex items-center justify-center hover:rounded-[14px] hover:bg-[#D4782A] hover:text-white transition-all duration-200 cursor-pointer shrink-0"
    >
      <svg class="w-5 h-5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
      </svg>
    </button>

    <!-- Join server button -->
    <button
      @click="showJoin = true"
      @mouseenter="showTooltip($event, 'Join Server')"
      @mouseleave="hideTooltip"
      class="relative z-10 w-11 h-11 rounded-full bg-[var(--sb-hover)] text-[#D4782A] flex items-center justify-center hover:rounded-[14px] hover:bg-[#D4782A] hover:text-white transition-all duration-200 cursor-pointer shrink-0"
    >
      <svg class="w-5 h-5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" d="M8.25 9V5.25A2.25 2.25 0 0 1 10.5 3h6a2.25 2.25 0 0 1 2.25 2.25v13.5A2.25 2.25 0 0 1 16.5 21h-6a2.25 2.25 0 0 1-2.25-2.25V15M12 9l3 3m0 0-3 3m3-3H2.25" />
      </svg>
    </button>

    <CreateServerModal v-if="showCreate" @close="showCreate = false" />
    <JoinServerModal v-if="showJoin" @close="showJoin = false" />
    <JoinServerModal v-if="showExplore" initial-tab="explore" @close="showExplore = false" />
  </div>

  <!-- Sidebar tooltip (teleported to escape overflow clipping) -->
  <Teleport to="body">
    <Transition
      enter-active-class="transition duration-150 ease-out"
      enter-from-class="opacity-0 translate-x-0"
      enter-to-class="opacity-100 translate-x-1"
      leave-active-class="transition duration-100 ease-in"
      leave-from-class="opacity-100 translate-x-1"
      leave-to-class="opacity-0 translate-x-0"
    >
      <div
        v-if="tooltip.visible"
        class="fixed z-[9999] pointer-events-none flex items-center -translate-y-1/2"
        :style="{ top: tooltip.top + 'px', left: tooltip.left + 'px' }"
      >
        <div class="w-2 h-2 bg-[#111214] rotate-45 rounded-[2px] shrink-0 -mr-1"></div>
        <div class="bg-[#111214] text-white text-sm font-semibold px-3 py-1.5 rounded-md shadow-lg shadow-black/30 whitespace-nowrap">{{ tooltip.text }}</div>
      </div>
    </Transition>
  </Teleport>

  <!-- Server context menu -->
  <Teleport to="body">
    <div v-if="contextMenu.visible" class="fixed inset-0 z-[9998]" @click="closeContextMenu" @contextmenu.prevent="closeContextMenu">
      <div
        class="fixed z-[9999] bg-[#111214] rounded-lg shadow-xl shadow-black/40 border border-white/5 py-1.5 min-w-[160px]"
        :style="{ top: contextMenu.top + 'px', left: contextMenu.left + 'px' }"
      >
        <button
          @click="toggleServerMute"
          class="w-full flex items-center gap-2.5 px-3 py-1.5 text-sm text-[#dcddde] hover:bg-[#E8521A] hover:text-white cursor-pointer transition-colors"
        >
          <svg v-if="notifStore.mutedServers.has(contextMenu.serverId)" class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" d="M19.114 5.636a9 9 0 0 1 0 12.728M16.463 8.288a5.25 5.25 0 0 1 0 7.424M6.75 8.25l4.72-4.72a.75.75 0 0 1 1.28.53v15.88a.75.75 0 0 1-1.28.53l-4.72-4.72H4.51c-.88 0-1.704-.507-1.938-1.354A9.009 9.009 0 0 1 2.25 12c0-.83.112-1.633.322-2.396C2.806 8.756 3.63 8.25 4.51 8.25H6.75Z" />
          </svg>
          <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" d="M17.25 9.75 19.5 12m0 0 2.25 2.25M19.5 12l2.25-2.25M19.5 12l-2.25 2.25m-10.5-6 4.72-4.72a.75.75 0 0 1 1.28.53v15.88a.75.75 0 0 1-1.28.53l-4.72-4.72H4.51c-.88 0-1.704-.507-1.938-1.354A9.009 9.009 0 0 1 2.25 12c0-.83.112-1.633.322-2.396C2.806 8.756 3.63 8.25 4.51 8.25H6.75Z" />
          </svg>
          {{ notifStore.mutedServers.has(contextMenu.serverId) ? 'Unmute Server' : 'Mute Server' }}
        </button>
      </div>
    </div>
  </Teleport>
</template>
