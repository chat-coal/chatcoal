<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useServersStore } from '@/stores/servers'
import { useAuthStore } from '@/stores/auth'
import { useToastStore } from '@/stores/toast'
import api, { API_URL } from '@/services/api.service'
import { getAvatarColor, getDefaultAvatarStyle, resolveFileUrl, cssBackgroundUrl, gifFirstFrame } from '@/utils/avatar'
import { useEscapeClose } from '@/composables/useEscapeClose'

const emit = defineEmits(['close'])
const router = useRouter()
const serversStore = useServersStore()
const authStore = useAuthStore()
const toastStore = useToastStore()

const server = computed(() => serversStore.currentServer)
const activeTab = ref('general')
const name = ref(server.value?.name || '')
const isPublic = ref(server.value?.is_public ?? false)
const iconFile = ref(null)
const clearIcon = ref(false)
const iconPreview = ref(server.value?.icon_url ? `${API_URL}${server.value.icon_url}` : '')
const saving = ref(false)
const showDeleteConfirm = ref(false)
const deleteConfirmName = ref('')
const deleting = ref(false)

// System announcements
const showJoinLeave = ref(server.value?.show_join_leave ?? false)
const systemChannelId = ref(server.value?.system_channel_id ?? '')
const textChannels = ref([])

// Members tab
const members = ref([])
const pendingRoles = ref({}) // { memberId: newRole }
const savingRoles = ref(false)
const showKickConfirm = ref(false)
const kickTarget = ref(null)
const showTransferConfirm = ref(false)
const transferTarget = ref(null)

// Bans tab
const bans = ref([])
const loadingBans = ref(false)
const showUnbanConfirm = ref(false)
const unbanTarget = ref(null)

useEscapeClose(() => {
  if (showDeleteConfirm.value) { showDeleteConfirm.value = false; return }
  if (showKickConfirm.value) { showKickConfirm.value = false; return }
  if (showTransferConfirm.value) { showTransferConfirm.value = false; return }
  if (showUnbanConfirm.value) { showUnbanConfirm.value = false; return }
  emit('close')
})

onMounted(async () => {
  if (server.value) {
    members.value = await api.getServerMembers(server.value.id)
    const channels = await api.getChannels(server.value.id)
    textChannels.value = channels.filter((c) => c.type === 'text')
  }
})

async function onFileSelect(e) {
  const raw = e.target.files[0]
  if (!raw) return
  const file = await gifFirstFrame(raw)
  iconFile.value = file
  if (iconPreview.value && iconPreview.value.startsWith('blob:')) {
    URL.revokeObjectURL(iconPreview.value)
  }
  iconPreview.value = URL.createObjectURL(file)
  clearIcon.value = false
}

function removeIcon() {
  iconFile.value = null
  iconPreview.value = ''
  clearIcon.value = true
}

async function save() {
  if (saving.value) return
  saving.value = true
  try {
    await serversStore.updateServer(server.value.id, {
      name: name.value.trim() || undefined,
      iconFile: iconFile.value || undefined,
      clearIcon: clearIcon.value || undefined,
      isPublic: isPublic.value,
      showJoinLeave: showJoinLeave.value,
      systemChannelId: systemChannelId.value || 0,
    })
    emit('close')
  } catch {
    toastStore.add('Failed to save server settings')
  } finally {
    saving.value = false
  }
}

async function deleteServer() {
  if (deleting.value) return
  deleting.value = true
  try {
    await serversStore.deleteServer(server.value.id)
    router.push('/channels/@me')
    emit('close')
  } catch {
    toastStore.add('Failed to delete server')
  } finally {
    deleting.value = false
  }
}

function stageRole(member, newRole) {
  if (newRole === member.role) {
    const { [member.id]: _, ...rest } = pendingRoles.value
    pendingRoles.value = rest
  } else {
    pendingRoles.value = { ...pendingRoles.value, [member.id]: newRole }
  }
}

async function saveRoles() {
  if (savingRoles.value) return
  savingRoles.value = true
  try {
    for (const [memberId, newRole] of Object.entries(pendingRoles.value)) {
      const member = members.value.find((m) => String(m.id) === memberId)
      if (!member) continue
      const updated = await api.updateMemberRole(server.value.id, member.user_id, newRole)
      const idx = members.value.findIndex((m) => m.id === updated.id)
      if (idx !== -1) members.value[idx] = { ...members.value[idx], ...updated }
    }
    pendingRoles.value = {}
  } catch {
    toastStore.add('Failed to update member roles')
  } finally {
    savingRoles.value = false
  }
}

function confirmKick(member) {
  kickTarget.value = member
  showKickConfirm.value = true
}

async function kickMember() {
  if (!kickTarget.value) return
  try {
    await api.kickMember(server.value.id, kickTarget.value.user_id)
    members.value = members.value.filter((m) => m.id !== kickTarget.value.id)
  } catch {
    toastStore.add('Failed to kick member')
  }
  showKickConfirm.value = false
  kickTarget.value = null
}

function confirmTransfer(member) {
  transferTarget.value = member
  showTransferConfirm.value = true
}

async function loadBans() {
  if (!server.value) return
  loadingBans.value = true
  try {
    bans.value = await api.getServerBans(server.value.id)
  } catch {
    toastStore.add('Failed to load bans')
  } finally {
    loadingBans.value = false
  }
}

function confirmUnban(ban) {
  unbanTarget.value = ban
  showUnbanConfirm.value = true
}

async function unbanUser() {
  if (!unbanTarget.value) return
  try {
    await api.unbanUser(server.value.id, unbanTarget.value.user_id)
    bans.value = bans.value.filter((b) => b.id !== unbanTarget.value.id)
  } catch {
    toastStore.add('Failed to unban user')
  }
  showUnbanConfirm.value = false
  unbanTarget.value = null
}

async function transferOwnership() {
  if (!transferTarget.value) return
  try {
    const updated = await api.transferOwnership(server.value.id, transferTarget.value.user_id)
    // Refresh members and server
    serversStore.patchCurrentServer({ owner_id: updated.owner_id })
    members.value = await api.getServerMembers(server.value.id)
    const me = members.value.find((m) => m.user_id === authStore.dbUser?.id)
    serversStore.setMemberRole(me?.role || null)
    showTransferConfirm.value = false
    transferTarget.value = null
  } catch {
    toastStore.add('Failed to transfer ownership')
  }
}
</script>

<template>
  <Teleport to="body">
    <div class="fixed inset-0 bg-[var(--backdrop)] backdrop-blur-sm flex items-center justify-center z-50" @click.self="emit('close')">
      <div class="bg-[var(--modal-bg)] rounded-2xl w-full max-w-lg shadow-2xl shadow-black/10 animate-scale-in border border-[var(--modal-border)] max-h-[85vh] flex flex-col">
        <div class="p-7 pb-0">
          <h2 class="font-display text-2xl font-bold text-[var(--text-1)] mb-4">Server Settings</h2>

          <!-- Tabs -->
          <div class="flex gap-1 border-b border-[var(--surface-border)]">
            <button
              @click="activeTab = 'general'"
              class="px-4 py-2.5 text-sm font-medium transition-colors duration-150 cursor-pointer border-b-2 -mb-px"
              :class="activeTab === 'general' ? 'text-[var(--text-1)] border-[#E8521A]' : 'text-[var(--text-3)] border-transparent hover:text-[var(--text-1)]'"
            >General</button>
            <button
              @click="activeTab = 'members'"
              class="px-4 py-2.5 text-sm font-medium transition-colors duration-150 cursor-pointer border-b-2 -mb-px"
              :class="activeTab === 'members' ? 'text-[var(--text-1)] border-[#E8521A]' : 'text-[var(--text-3)] border-transparent hover:text-[var(--text-1)]'"
            >Members</button>
            <button
              v-if="serversStore.isAdmin"
              @click="activeTab = 'bans'; loadBans()"
              class="px-4 py-2.5 text-sm font-medium transition-colors duration-150 cursor-pointer border-b-2 -mb-px"
              :class="activeTab === 'bans' ? 'text-[var(--text-1)] border-[#E8521A]' : 'text-[var(--text-3)] border-transparent hover:text-[var(--text-1)]'"
            >Bans</button>
          </div>
        </div>

        <div class="p-7 pt-5 overflow-y-auto flex-1">
          <!-- General Tab -->
          <template v-if="activeTab === 'general'">
            <!-- Icon upload -->
            <div class="flex flex-col items-center mb-6 gap-2">
              <label
                class="w-20 h-20 rounded-2xl bg-[var(--surface)] border-2 border-dashed border-[var(--surface-border)] flex items-center justify-center cursor-pointer hover:border-[#E8521A]/40 transition-colors duration-200 overflow-hidden relative group"
              >
                <img v-if="iconPreview" :src="iconPreview" class="w-full h-full object-cover" />
                <span v-else class="text-[var(--text-4)] text-2xl font-bold">{{ (server?.name || '?')[0].toUpperCase() }}</span>
                <div class="absolute inset-0 bg-black/40 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity duration-200 rounded-2xl">
                  <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z" />
                    <circle cx="12" cy="13" r="3" />
                  </svg>
                </div>
                <input type="file" accept="image/png,image/jpeg,image/gif,image/webp" class="hidden" @change="onFileSelect" />
              </label>
              <button
                v-if="iconPreview"
                type="button"
                @click="removeIcon"
                class="text-xs text-[var(--text-4)] hover:text-red-400 transition-colors duration-150 cursor-pointer"
              >
                Remove icon
              </button>
            </div>

            <!-- Server name -->
            <label class="text-[11px] font-bold text-[var(--text-4)] uppercase tracking-[0.1em]">Server Name</label>
            <input
              v-model="name"
              @keyup.enter="save"
              class="w-full bg-[var(--surface)] text-[var(--text-1)] px-3.5 py-2.5 rounded-xl mt-2 mb-5 placeholder-[var(--text-4)] text-sm border border-[var(--surface-border)]"
            />

            <!-- Visibility (owner only) -->
            <div v-if="serversStore.isOwner" class="flex items-center justify-between mb-6 p-3.5 rounded-xl bg-[var(--surface)] border border-[var(--surface-border)]">
              <div>
                <p class="text-sm font-medium text-[var(--text-1)]">{{ isPublic ? 'Public' : 'Private' }} Server</p>
                <p class="text-xs text-[var(--text-4)] mt-0.5">
                  {{ isPublic ? 'Anyone can discover and join in Explore' : 'Only accessible via invite link' }}
                </p>
              </div>
              <button
                type="button"
                @click="isPublic = !isPublic"
                class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors duration-200 cursor-pointer shrink-0 ml-4"
                :class="isPublic ? 'bg-[#E8521A]' : 'bg-[var(--surface-border)]'"
              >
                <span
                  class="inline-block h-4 w-4 transform rounded-full bg-white shadow transition-transform duration-200"
                  :class="isPublic ? 'translate-x-6' : 'translate-x-1'"
                />
              </button>
            </div>

            <!-- Join/Leave Announcements (owner only) -->
            <div v-if="serversStore.isOwner" class="flex items-center justify-between mb-4 p-3.5 rounded-xl bg-[var(--surface)] border border-[var(--surface-border)]">
              <div>
                <p class="text-sm font-medium text-[var(--text-1)]">Join/Leave Announcements</p>
                <p class="text-xs text-[var(--text-4)] mt-0.5">Post a message when members join or leave</p>
              </div>
              <button
                type="button"
                @click="showJoinLeave = !showJoinLeave"
                class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors duration-200 cursor-pointer shrink-0 ml-4"
                :class="showJoinLeave ? 'bg-[#E8521A]' : 'bg-[var(--surface-border)]'"
              >
                <span
                  class="inline-block h-4 w-4 transform rounded-full bg-white shadow transition-transform duration-200"
                  :class="showJoinLeave ? 'translate-x-6' : 'translate-x-1'"
                />
              </button>
            </div>

            <!-- System channel picker (owner only, when enabled) -->
            <div v-if="serversStore.isOwner && showJoinLeave" class="mb-6">
              <label class="text-[11px] font-bold text-[var(--text-4)] uppercase tracking-[0.1em]">System Channel</label>
              <div class="relative mt-2">
                <select
                  v-model="systemChannelId"
                  class="w-full appearance-none bg-[var(--surface)] text-[var(--text-1)] px-3.5 py-2.5 pr-10 rounded-xl text-sm border border-[var(--surface-border)] cursor-pointer"
                >
                  <option value="">First text channel (default)</option>
                  <option v-for="ch in textChannels" :key="ch.id" :value="ch.id">#{{ ch.name }}</option>
                </select>
                <svg class="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-3)]" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.23 7.21a.75.75 0 011.06.02L10 11.168l3.71-3.938a.75.75 0 111.08 1.04l-4.25 4.5a.75.75 0 01-1.08 0l-4.25-4.5a.75.75 0 01.02-1.06z" clip-rule="evenodd"/></svg>
              </div>
            </div>

            <!-- Actions -->
            <div class="flex justify-end gap-3 mb-6">
              <button @click="emit('close')" class="text-[var(--text-3)] hover:text-[var(--text-1)] px-4 py-2.5 rounded-xl cursor-pointer font-medium transition-colors duration-150">
                Cancel
              </button>
              <button
                @click="save"
                :disabled="saving"
                class="bg-[#E8521A] text-white px-5 py-2.5 rounded-xl hover:bg-[#D44818] disabled:opacity-40 cursor-pointer font-semibold shadow-lg shadow-[#E8521A]/15 transition-all duration-200"
              >
                Save
              </button>
            </div>

            <!-- Delete server (owner only) -->
            <div v-if="serversStore.isOwner" class="border-t border-[var(--surface-border)] pt-5">
              <button
                @click="showDeleteConfirm = true; deleteConfirmName = ''"
                class="text-[#E8521A] hover:bg-[#E8521A] hover:text-white px-4 py-2 rounded-xl text-sm cursor-pointer transition-all duration-200 font-medium"
              >
                Delete Server
              </button>
            </div>
          </template>

          <!-- Members Tab -->
          <template v-if="activeTab === 'members'">
            <div class="space-y-1 mb-4">
              <div
                v-for="member in members"
                :key="member.id"
                class="flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-[var(--surface-2)] transition-colors duration-100"
              >
                <!-- Avatar -->
                <div
                  v-if="member.user?.avatar_url"
                  class="w-8 h-8 rounded-full bg-cover bg-center shrink-0"
                  :style="{ backgroundImage: cssBackgroundUrl(resolveFileUrl(member.user.avatar_url)) }"
                ></div>
                <div
                  v-else
                  class="w-8 h-8 rounded-full flex items-center justify-center text-white text-xs font-bold shrink-0"
                  :style="getDefaultAvatarStyle(member.user_id)"
                >
                  {{ (member.user?.display_name || '?')[0].toUpperCase() }}
                </div>

                <!-- Name -->
                <div class="flex-1 min-w-0">
                  <span class="text-sm text-[var(--text-1)] font-medium truncate block">{{ member.user?.display_name || 'Unknown' }}</span>
                  <span class="text-[11px] text-[var(--text-4)] capitalize">{{ member.role }}</span>
                </div>

                <!-- Actions (not for self, not for owner target unless transferring) -->
                <div v-if="member.user_id !== authStore.dbUser?.id && member.role !== 'owner'" class="flex items-center gap-1.5 shrink-0">
                  <!-- Role dropdown (owner only) -->
                  <div v-if="serversStore.isOwner" class="relative">
                    <select
                      :value="pendingRoles[member.id] ?? member.role"
                      @change="stageRole(member, $event.target.value)"
                      class="appearance-none bg-[var(--surface)] text-[var(--text-2)] text-xs px-2 py-1.5 pr-7 rounded-lg border cursor-pointer"
                      :class="pendingRoles[member.id] ? 'border-[#E8521A]' : 'border-[var(--surface-border)]'"
                    >
                      <option value="admin">Admin</option>
                      <option value="member">Member</option>
                    </select>
                    <svg class="pointer-events-none absolute right-1.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-[var(--text-3)]" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.23 7.21a.75.75 0 011.06.02L10 11.168l3.71-3.938a.75.75 0 111.08 1.04l-4.25 4.5a.75.75 0 01-1.08 0l-4.25-4.5a.75.75 0 01.02-1.06z" clip-rule="evenodd"/></svg>
                  </div>
                  <!-- Kick button -->
                  <button
                    v-if="serversStore.canKick(member.role)"
                    @click="confirmKick(member)"
                    class="text-[var(--text-4)] hover:text-[#E8521A] p-1.5 rounded-lg hover:bg-[#E8521A]/10 cursor-pointer transition-colors duration-150"
                    title="Kick"
                  >
                    <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </div>

                <!-- Transfer ownership button (owner only, target not owner) -->
                <button
                  v-if="serversStore.isOwner && member.role !== 'owner' && member.user_id !== authStore.dbUser?.id"
                  @click="confirmTransfer(member)"
                  class="text-[var(--text-4)] hover:text-[var(--text-1)] p-1.5 rounded-lg hover:bg-[var(--surface-2)] cursor-pointer transition-colors duration-150"
                  title="Transfer Ownership"
                >
                  <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M7.5 21L3 16.5m0 0L7.5 12M3 16.5h13.5m0-13.5L21 7.5m0 0L16.5 12M21 7.5H7.5" />
                  </svg>
                </button>
              </div>
            </div>
            <!-- Save roles button -->
            <div v-if="Object.keys(pendingRoles).length" class="flex justify-end mt-4">
              <button
                @click="saveRoles"
                :disabled="savingRoles"
                class="bg-[#E8521A] text-white px-5 py-2.5 rounded-xl hover:bg-[#D44818] disabled:opacity-40 cursor-pointer font-semibold shadow-lg shadow-[#E8521A]/15 transition-all duration-200 text-sm"
              >
                {{ savingRoles ? 'Saving...' : 'Save Changes' }}
              </button>
            </div>
          </template>

          <!-- Bans Tab -->
          <template v-if="activeTab === 'bans'">
            <div v-if="loadingBans" class="flex items-center justify-center py-10">
              <div class="w-5 h-5 border-2 border-[var(--text-4)] border-t-transparent rounded-full animate-spin"></div>
            </div>
            <div v-else-if="!bans.length" class="text-center py-10">
              <p class="text-[var(--text-4)] text-sm">No banned users</p>
            </div>
            <div v-else class="space-y-1">
              <div
                v-for="ban in bans"
                :key="ban.id"
                class="flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-[var(--surface-2)] transition-colors duration-100"
              >
                <!-- Avatar -->
                <div
                  v-if="ban.user?.avatar_url"
                  class="w-8 h-8 rounded-full bg-cover bg-center shrink-0"
                  :style="{ backgroundImage: cssBackgroundUrl(resolveFileUrl(ban.user.avatar_url)) }"
                ></div>
                <div
                  v-else
                  class="w-8 h-8 rounded-full flex items-center justify-center text-white text-xs font-bold shrink-0"
                  :style="getDefaultAvatarStyle(ban.user_id)"
                >
                  {{ (ban.user?.display_name || '?')[0].toUpperCase() }}
                </div>

                <!-- Info -->
                <div class="flex-1 min-w-0">
                  <span class="text-sm text-[var(--text-1)] font-medium truncate block">{{ ban.user?.display_name || 'Unknown' }}</span>
                  <span v-if="ban.reason" class="text-[11px] text-[var(--text-4)] truncate block">{{ ban.reason }}</span>
                </div>

                <!-- Unban button -->
                <button
                  @click="confirmUnban(ban)"
                  class="text-xs text-[var(--text-3)] hover:text-[var(--text-1)] px-3 py-1.5 rounded-lg border border-[var(--surface-border)] hover:bg-[var(--surface-2)] cursor-pointer transition-colors duration-150 shrink-0"
                >
                  Unban
                </button>
              </div>
            </div>
          </template>
        </div>
      </div>
    </div>
  </Teleport>

  <!-- Delete confirmation modal -->
  <Teleport to="body">
    <div v-if="showDeleteConfirm" class="fixed inset-0 bg-[var(--backdrop)] backdrop-blur-sm flex items-center justify-center z-[60]" @click.self="showDeleteConfirm = false">
      <div class="bg-[var(--modal-bg)] rounded-2xl p-7 w-full max-w-sm shadow-2xl shadow-black/10 animate-scale-in border border-[var(--modal-border)]">
        <h2 class="font-display text-xl font-bold text-[var(--text-1)] mb-2">Delete Server</h2>
        <p class="text-[var(--text-3)] text-sm mb-5">
          This action is irreversible. All channels, messages, and members will be permanently removed.
        </p>
        <label class="text-[11px] font-bold text-[var(--text-4)] uppercase tracking-[0.1em]">
          Type <span class="text-[var(--text-1)] normal-case tracking-normal">{{ server?.name }}</span> to confirm
        </label>
        <input
          v-model="deleteConfirmName"
          @keyup.enter="deleteConfirmName === server?.name && deleteServer()"
          class="w-full bg-[var(--surface)] text-[var(--text-1)] px-3.5 py-2.5 rounded-xl mt-2 mb-5 placeholder-[var(--text-4)] text-sm border border-[var(--surface-border)]"
          :placeholder="server?.name"
        />
        <div class="flex justify-end gap-3">
          <button @click="showDeleteConfirm = false" class="text-[var(--text-3)] hover:text-[var(--text-1)] px-4 py-2.5 rounded-xl cursor-pointer font-medium transition-colors duration-150">
            Cancel
          </button>
          <button
            @click="deleteServer"
            :disabled="deleteConfirmName !== server?.name || deleting"
            class="bg-[#E8521A] text-white px-5 py-2.5 rounded-xl hover:bg-[#D44818] disabled:opacity-40 disabled:cursor-not-allowed cursor-pointer font-semibold shadow-lg shadow-[#E8521A]/15 transition-all duration-200"
          >
            {{ deleting ? 'Deleting...' : 'Delete Server' }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>

  <!-- Kick confirmation modal -->
  <Teleport to="body">
    <div v-if="showKickConfirm" class="fixed inset-0 bg-[var(--backdrop)] backdrop-blur-sm flex items-center justify-center z-[60]" @click.self="showKickConfirm = false">
      <div class="bg-[var(--modal-bg)] rounded-2xl p-7 w-full max-w-sm shadow-2xl shadow-black/10 animate-scale-in border border-[var(--modal-border)]">
        <h2 class="font-display text-xl font-bold text-[var(--text-1)] mb-2">Kick Member</h2>
        <p class="text-[var(--text-3)] text-sm mb-5">
          Are you sure you want to kick <span class="font-semibold text-[var(--text-1)]">{{ kickTarget?.user?.display_name || 'this member' }}</span>? They will be removed from the server and can rejoin with a new invite.
        </p>
        <div class="flex justify-end gap-3">
          <button @click="showKickConfirm = false" class="text-[var(--text-3)] hover:text-[var(--text-1)] px-4 py-2.5 rounded-xl cursor-pointer font-medium transition-colors duration-150">
            Cancel
          </button>
          <button
            @click="kickMember"
            class="bg-[#E8521A] text-white px-5 py-2.5 rounded-xl hover:bg-[#D44818] cursor-pointer font-semibold shadow-lg shadow-[#E8521A]/15 transition-all duration-200"
          >
            Kick
          </button>
        </div>
      </div>
    </div>
  </Teleport>

  <!-- Transfer ownership confirmation -->
  <Teleport to="body">
    <div v-if="showTransferConfirm" class="fixed inset-0 bg-[var(--backdrop)] backdrop-blur-sm flex items-center justify-center z-[60]" @click.self="showTransferConfirm = false">
      <div class="bg-[var(--modal-bg)] rounded-2xl p-7 w-full max-w-sm shadow-2xl shadow-black/10 animate-scale-in border border-[var(--modal-border)]">
        <h2 class="font-display text-xl font-bold text-[var(--text-1)] mb-2">Transfer Ownership</h2>
        <p class="text-[var(--text-3)] text-sm mb-5">
          Transfer server ownership to <span class="font-semibold text-[var(--text-1)]">{{ transferTarget?.user?.display_name }}</span>? You will become an admin.
        </p>
        <div class="flex justify-end gap-3">
          <button @click="showTransferConfirm = false" class="text-[var(--text-3)] hover:text-[var(--text-1)] px-4 py-2.5 rounded-xl cursor-pointer font-medium transition-colors duration-150">
            Cancel
          </button>
          <button
            @click="transferOwnership"
            class="bg-[#E8521A] text-white px-5 py-2.5 rounded-xl hover:bg-[#D44818] cursor-pointer font-semibold shadow-lg shadow-[#E8521A]/15 transition-all duration-200"
          >
            Transfer
          </button>
        </div>
      </div>
    </div>
  </Teleport>

  <!-- Unban confirmation modal -->
  <Teleport to="body">
    <div v-if="showUnbanConfirm" class="fixed inset-0 bg-[var(--backdrop)] backdrop-blur-sm flex items-center justify-center z-[60]" @click.self="showUnbanConfirm = false">
      <div class="bg-[var(--modal-bg)] rounded-2xl p-7 w-full max-w-sm shadow-2xl shadow-black/10 animate-scale-in border border-[var(--modal-border)]">
        <h2 class="font-display text-xl font-bold text-[var(--text-1)] mb-2">Unban User</h2>
        <p class="text-[var(--text-3)] text-sm mb-5">
          Are you sure you want to unban <span class="font-semibold text-[var(--text-1)]">{{ unbanTarget?.user?.display_name || 'this user' }}</span>? They will be able to rejoin the server.
        </p>
        <div class="flex justify-end gap-3">
          <button @click="showUnbanConfirm = false" class="text-[var(--text-3)] hover:text-[var(--text-1)] px-4 py-2.5 rounded-xl cursor-pointer font-medium transition-colors duration-150">
            Cancel
          </button>
          <button
            @click="unbanUser"
            class="bg-[#E8521A] text-white px-5 py-2.5 rounded-xl hover:bg-[#D44818] cursor-pointer font-semibold shadow-lg shadow-[#E8521A]/15 transition-all duration-200"
          >
            Unban
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>
