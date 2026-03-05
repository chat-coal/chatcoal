<script setup>
import { ref, watch, onUnmounted, defineProps } from 'vue'

const props = defineProps({
  inBottomSheet: { type: Boolean, default: false },
})
import { useRouter } from 'vue-router'
import { useServersStore } from '@/stores/servers'
import { useAuthStore } from '@/stores/auth'
import { useDMsStore } from '@/stores/dms'
import { useToastStore } from '@/stores/toast'
import api from '@/services/api.service'
import { on } from '@/services/websocket.service'
import { getAvatarColor, getDefaultAvatarStyle, resolveFileUrl, cssBackgroundUrl } from '@/utils/avatar'
import UserProfilePopover from './UserProfilePopover.vue'

const router = useRouter()
const serversStore = useServersStore()
const authStore = useAuthStore()
const dmsStore = useDMsStore()
const toastStore = useToastStore()
const members = ref([])
const contextMenu = ref(null) // { memberId, x, y }
const confirmModal = ref(null) // { title, message, action, showReasonInput }
const banReason = ref('')

watch(
  () => serversStore.currentServer,
  async (server) => {
    if (server) {
      members.value = await api.getServerMembers(server.id)
      // Set current user's role
      const me = members.value.find((m) => m.user_id === authStore.dbUser?.id)
      serversStore.setMemberRole(me?.role || null)
    } else {
      members.value = []
      serversStore.setMemberRole(null)
    }
  },
  { immediate: true },
)

const offJoin = on('member_join', (member, serverId) => {
  if (serverId !== serversStore.currentServer?.id) return
  if (!members.value.find((m) => m.id === member.id)) {
    // Set online status — user just joined so they're connected
    if (member.user) member.user.status = 'online'
    members.value.push(member)
  }
})

const offLeave = on('member_leave', (member, serverId) => {
  if (serverId !== serversStore.currentServer?.id) return
  members.value = members.value.filter((m) => m.id !== member.id)
})

const offPresence = on('presence_update', (data) => {
  const idx = members.value.findIndex((m) => m.user_id === data.user_id)
  if (idx !== -1 && members.value[idx].user) {
    members.value[idx] = { ...members.value[idx], user: { ...members.value[idx].user, status: data.status } }
  }
})

const offMemberUpdate = on('member_update', (updated) => {
  const idx = members.value.findIndex((m) => m.id === updated.id)
  if (idx !== -1) {
    members.value[idx] = { ...members.value[idx], ...updated }
  }
  // Update own role if it's us
  if (updated.user_id === authStore.dbUser?.id) {
    serversStore.setMemberRole(updated.role)
  }
})

const offUserUpdate = on('user_update', (data) => {
  const idx = members.value.findIndex((m) => m.user_id === data.user_id)
  if (idx !== -1 && members.value[idx].user) {
    members.value[idx] = { ...members.value[idx], user: { ...members.value[idx].user, display_name: data.display_name, avatar_url: data.avatar_url } }
  }
})

// Refetch members on WebSocket reconnect to catch events missed during disconnection
const offConnected = on('connected', async () => {
  const server = serversStore.currentServer
  if (server) {
    members.value = await api.getServerMembers(server.id)
  }
})

onUnmounted(() => {
  offJoin()
  offLeave()
  offPresence()
  offMemberUpdate()
  offUserUpdate()
  offConnected()
})

function groupedMembers() {
  const owners = members.value.filter((m) => m.role === 'owner')
  const admins = members.value.filter((m) => m.role === 'admin')
  const rest = members.value.filter((m) => m.role === 'member')
  const groups = []
  if (owners.length) groups.push({ role: 'Owner', members: owners })
  if (admins.length) groups.push({ role: 'Admin', members: admins })
  if (rest.length) groups.push({ role: 'Members', members: rest })
  return groups
}

async function startDM(userId) {
  if (userId === authStore.dbUser?.id) return
  try {
    const channel = await dmsStore.openDM(userId)
    serversStore.selectServer(null)
    dmsStore.selectDM(channel)
    router.push(`/channels/@me/${channel.id}`)
  } catch {
    toastStore.add('Failed to open conversation')
  }
}

function openContextMenu(e, member) {
  if (member.user_id === authStore.dbUser?.id) return
  // Only show if user has some action available
  if (!serversStore.isOwner && !serversStore.canKick(member.role) && !serversStore.canBan(member.role)) return
  contextMenu.value = { memberId: member.id, userId: member.user_id, role: member.role, x: e.clientX, y: e.clientY }
}

function closeContextMenu() {
  contextMenu.value = null
}

function toggleAdmin(member) {
  const newRole = member.role === 'admin' ? 'member' : 'admin'
  const title = newRole === 'admin' ? 'Make Admin' : 'Remove Admin'
  const message = newRole === 'admin'
    ? 'This member will be able to manage channels, messages, and invites.'
    : 'This member will lose all admin privileges.'
  closeContextMenu()
  confirmModal.value = {
    title,
    message,
    action: async () => {
      try {
        const updated = await api.updateMemberRole(serversStore.currentServer.id, member.userId, newRole)
        const idx = members.value.findIndex((m) => m.id === updated.id)
        if (idx !== -1) members.value[idx] = { ...members.value[idx], ...updated }
      } catch {
        toastStore.add('Failed to update member role')
      }
      confirmModal.value = null
    },
  }
}

const profileMember = ref(null)
const profileAnchor = ref(null)

function openProfile(event, member) {
  profileAnchor.value = event.currentTarget
  profileMember.value = member
}

function kickMember(member) {
  closeContextMenu()
  confirmModal.value = {
    title: 'Kick Member',
    message: 'This member will be removed from the server. They can rejoin with a new invite.',
    action: async () => {
      try {
        await api.kickMember(serversStore.currentServer.id, member.userId)
        members.value = members.value.filter((m) => m.id !== member.memberId)
      } catch {
        toastStore.add('Failed to kick member')
      }
      confirmModal.value = null
    },
  }
}

function banMember(member) {
  closeContextMenu()
  banReason.value = ''
  confirmModal.value = {
    title: 'Ban Member',
    message: 'This member will be permanently removed from the server and cannot rejoin until unbanned.',
    showReasonInput: true,
    action: async () => {
      try {
        await api.banMember(serversStore.currentServer.id, member.userId, banReason.value.trim() || null)
        members.value = members.value.filter((m) => m.id !== member.memberId)
      } catch {
        toastStore.add('Failed to ban member')
      }
      confirmModal.value = null
    },
  }
}
</script>

<template>
  <div
    :class="inBottomSheet
      ? 'overflow-y-auto px-3 py-4 scrollbar-light flex-1 min-h-0'
      : 'w-[220px] bg-[var(--surface)] overflow-y-auto px-3 py-5 shrink-0 hidden lg:block scrollbar-light border-l border-[var(--surface-border)]'"
    @click="closeContextMenu"
  >
    <template v-if="serversStore.currentServer">
      <div v-for="group in groupedMembers()" :key="group.role" class="mb-5">
        <h4 class="text-[var(--text-4)] text-[11px] font-bold uppercase tracking-[0.1em] px-2 mb-2.5">
          {{ group.role }} &mdash; {{ group.members.length }}
        </h4>
        <div
          v-for="member in group.members"
          :key="member.id"
          class="group flex items-center gap-2.5 px-2 py-1.5 rounded-lg hover:bg-[var(--surface-3)] cursor-pointer transition-colors duration-100"
          @click.stop="openProfile($event, member)"
          @contextmenu.prevent="openContextMenu($event, member)"
        >
          <div class="relative">
            <div
              v-if="member.user?.avatar_url"
              class="w-7 h-7 rounded-full bg-cover bg-center"
              :style="{ backgroundImage: cssBackgroundUrl(resolveFileUrl(member.user.avatar_url)) }"
            ></div>
            <div
              v-else
              class="w-7 h-7 rounded-full flex items-center justify-center text-white text-[10px] font-bold"
              :style="getDefaultAvatarStyle(member.user_id)"
            >
              {{ (member.user?.display_name || '?')[0].toUpperCase() }}
            </div>
            <div
              class="absolute -bottom-0.5 -right-0.5 w-2.5 h-2.5 rounded-full border-[1.5px] border-[var(--surface)]"
              :class="member.user?.status === 'online' ? 'bg-green-500' : 'bg-[var(--offline)]'"
            ></div>
          </div>
          <span class="flex-1 min-w-0 flex flex-col leading-tight">
            <span
              class="text-sm truncate"
              :class="member.user?.is_anonymous ? 'text-[var(--text-4)]' : 'text-[var(--text-2)]'"
            >{{ member.user?.display_name || 'Unknown' }}</span>
            <span
              v-if="member.user?.home_instance"
              class="text-[10px] text-[var(--text-4)] truncate"
            >@{{ member.user.home_instance }}</span>
          </span>
          <!-- Message button (not shown for self) -->
          <button
            v-if="member.user_id !== authStore.dbUser?.id"
            @click.stop="startDM(member.user_id)"
            class="opacity-0 group-hover:opacity-100 text-[var(--text-4)] hover:text-[var(--text-1)] p-1 rounded transition-all duration-150 shrink-0"
            title="Message"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
            </svg>
          </button>
        </div>
      </div>
    </template>

    <!-- Context menu -->
    <Teleport to="body">
      <div
        v-if="contextMenu"
        class="fixed z-50 bg-[var(--card)] border border-[var(--surface-border)] rounded-xl shadow-xl py-1.5 min-w-[160px] animate-fade-in"
        :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }"
        @click.stop
      >
        <button
          v-if="serversStore.isOwner && contextMenu.role !== 'owner'"
          @click="toggleAdmin(contextMenu)"
          class="w-full text-left px-3.5 py-2 text-sm text-[var(--text-2)] hover:bg-[var(--surface-2)] transition-colors duration-100 cursor-pointer"
        >
          {{ contextMenu.role === 'admin' ? 'Remove Admin' : 'Make Admin' }}
        </button>
        <button
          v-if="serversStore.canKick(contextMenu.role)"
          @click="kickMember(contextMenu)"
          class="w-full text-left px-3.5 py-2 text-sm text-[#E8521A] hover:bg-[#E8521A]/10 transition-colors duration-100 cursor-pointer"
        >
          Kick Member
        </button>
        <button
          v-if="serversStore.canBan(contextMenu.role)"
          @click="banMember(contextMenu)"
          class="w-full text-left px-3.5 py-2 text-sm text-[#E8521A] hover:bg-[#E8521A]/10 transition-colors duration-100 cursor-pointer"
        >
          Ban Member
        </button>
      </div>
    </Teleport>

    <!-- Confirmation modal -->
    <Teleport to="body">
      <div
        v-if="confirmModal"
        class="fixed inset-0 bg-[var(--backdrop)] backdrop-blur-sm flex items-center justify-center z-[60]"
        @click.self="confirmModal = null"
      >
        <div class="bg-[var(--modal-bg)] rounded-2xl p-7 w-full max-w-sm shadow-2xl shadow-black/10 animate-scale-in border border-[var(--modal-border)]">
          <h2 class="font-display text-xl font-bold text-[var(--text-1)] mb-2">{{ confirmModal.title }}</h2>
          <p class="text-[var(--text-3)] text-sm mb-4">{{ confirmModal.message }}</p>
          <div v-if="confirmModal.showReasonInput" class="mb-5">
            <label class="text-[11px] font-bold text-[var(--text-4)] uppercase tracking-[0.1em]">Reason (optional)</label>
            <input
              v-model="banReason"
              class="w-full bg-[var(--surface)] text-[var(--text-1)] px-3.5 py-2.5 rounded-xl mt-2 placeholder-[var(--text-4)] text-sm border border-[var(--surface-border)]"
              placeholder="Why is this member being banned?"
              maxlength="512"
            />
          </div>
          <div v-else class="mb-1"></div>
          <div class="flex justify-end gap-3">
            <button
              @click="confirmModal = null"
              class="text-[var(--text-3)] hover:text-[var(--text-1)] px-4 py-2.5 rounded-xl cursor-pointer font-medium transition-colors duration-150"
            >
              Cancel
            </button>
            <button
              @click="confirmModal.action()"
              class="bg-[#E8521A] text-white px-5 py-2.5 rounded-xl hover:bg-[#D44818] cursor-pointer font-semibold shadow-lg shadow-[#E8521A]/15 transition-all duration-200"
            >
              Confirm
            </button>
          </div>
        </div>
      </div>
    </Teleport>
    <!-- User profile popover -->
    <UserProfilePopover
      v-if="profileMember"
      :user-id="profileMember.user_id"
      :anchor-el="profileAnchor"
      :server-member="profileMember"
      @close="profileMember = null"
    />
  </div>
</template>
