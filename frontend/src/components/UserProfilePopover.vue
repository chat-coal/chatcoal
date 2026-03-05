<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useServersStore } from '@/stores/servers'
import { useDMsStore } from '@/stores/dms'
import api from '@/services/api.service'
import { getAvatarColor, getDefaultAvatarStyle, resolveFileUrl, cssBackgroundUrl } from '@/utils/avatar'

const props = defineProps({
  userId: { type: String, required: true },
  anchorEl: { type: Object, default: null }, // reference element for positioning
  serverMember: { type: Object, default: null }, // pre-loaded member data if available
})

const emit = defineEmits(['close'])

const router = useRouter()
const authStore = useAuthStore()
const serversStore = useServersStore()
const dmsStore = useDMsStore()

const profile = ref(null)
const loading = ref(true)
const popoverEl = ref(null)
const style = ref({})

onMounted(async () => {
  try {
    profile.value = await api.getUserProfile(props.userId)
  } catch { /* ignore */ }
  loading.value = false

  // Position popover near anchor
  if (props.anchorEl && popoverEl.value) {
    positionPopover()
  }

  document.addEventListener('mousedown', handleOutsideClick)
})

onUnmounted(() => {
  document.removeEventListener('mousedown', handleOutsideClick)
})

function positionPopover() {
  const anchor = props.anchorEl.getBoundingClientRect()
  const pop = popoverEl.value
  const viewW = window.innerWidth
  const viewH = window.innerHeight
  const popW = 240
  const popH = 280 // estimated

  let left = anchor.right + 8
  let top = anchor.top

  // Flip left if no room on right
  if (left + popW > viewW - 8) {
    left = anchor.left - popW - 8
  }
  // Flip up if no room below
  if (top + popH > viewH - 8) {
    top = viewH - popH - 8
  }
  if (top < 8) top = 8

  style.value = { left: left + 'px', top: top + 'px' }
}

function handleOutsideClick(e) {
  if (popoverEl.value && !popoverEl.value.contains(e.target)) {
    emit('close')
  }
}

const roleLabels = { owner: 'Owner', admin: 'Admin', member: 'Member' }
const roleColors = {
  owner: 'text-[#E8521A] bg-[#E8521A]/10',
  admin: 'text-[#D4782A] bg-[#D4782A]/10',
  member: 'text-[var(--text-3)] bg-[var(--surface-3)]',
}

function formatJoined(dateStr) {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })
}

async function startDM() {
  if (props.userId === authStore.dbUser?.id) return
  emit('close')
  const channel = await dmsStore.openDM(props.userId)
  serversStore.selectServer(null)
  dmsStore.selectDM(channel)
  router.push(`/channels/@me/${channel.id}`)
}
</script>

<template>
  <Teleport to="body">
    <div
      ref="popoverEl"
      class="fixed z-[70] w-[240px] bg-[var(--card)] border border-[var(--surface-border)] rounded-2xl shadow-2xl shadow-black/20 overflow-hidden animate-scale-in"
      :style="style"
    >
      <!-- Loading -->
      <div v-if="loading" class="flex items-center justify-center py-10">
        <div class="w-5 h-5 border-2 border-[var(--text-4)] border-t-[#E8521A] rounded-full animate-spin"></div>
      </div>

      <template v-else-if="profile">
        <!-- Banner / Avatar area -->
        <div class="h-16 bg-gradient-to-br from-[#E8521A]/20 to-[#D4782A]/10 relative">
          <div class="absolute -bottom-6 left-4">
            <div
              v-if="profile.avatar_url"
              class="w-14 h-14 rounded-full bg-cover bg-center border-4 border-[var(--card)]"
              :style="{ backgroundImage: cssBackgroundUrl(resolveFileUrl(profile.avatar_url)) }"
            ></div>
            <div
              v-else
              class="w-14 h-14 rounded-full flex items-center justify-center text-white text-xl font-bold border-4 border-[var(--card)]"
              :style="getDefaultAvatarStyle(userId)"
            >
              {{ (profile.display_name || '?')[0].toUpperCase() }}
            </div>
            <!-- Status dot -->
            <div
              class="absolute bottom-0.5 right-0.5 w-3.5 h-3.5 rounded-full border-2 border-[var(--card)]"
              :class="profile.status === 'online' ? 'bg-green-500' : 'bg-[var(--offline)]'"
            ></div>
          </div>
        </div>

        <!-- Profile info -->
        <div class="pt-8 pb-4 px-4">
          <h3 class="text-[var(--text-1)] font-bold text-base truncate">{{ profile.display_name }}</h3>
          <p v-if="profile.home_instance" class="text-[var(--text-4)] text-[11px] mb-1">@{{ profile.home_instance }}</p>
          <p class="text-[var(--text-4)] text-xs mb-3">
            {{ profile.status === 'online' ? 'Online' : 'Offline' }}
          </p>

          <!-- Server role (if in a server context) -->
          <div v-if="serverMember" class="mb-3 space-y-1.5">
            <div class="flex items-center justify-between text-[10px]">
              <span class="text-[var(--text-4)] uppercase font-bold tracking-wide">Server Role</span>
              <span
                class="px-2 py-0.5 rounded-full text-[10px] font-semibold"
                :class="roleColors[serverMember.role] || roleColors.member"
              >
                {{ roleLabels[serverMember.role] || serverMember.role }}
              </span>
            </div>
            <div class="flex items-center justify-between text-[10px]">
              <span class="text-[var(--text-4)] uppercase font-bold tracking-wide">Joined</span>
              <span class="text-[var(--text-3)]">{{ formatJoined(serverMember.joined_at) }}</span>
            </div>
          </div>

          <div class="text-[10px] text-[var(--text-4)] mb-3">
            Member since {{ formatJoined(profile.created_at) }}
          </div>

          <!-- DM button (not for self, not for deleted users) -->
          <button
            v-if="userId !== authStore.dbUser?.id && !profile.deleted"
            @click="startDM"
            class="w-full bg-[#E8521A] hover:bg-[#D44818] text-white text-xs font-semibold py-2 rounded-xl cursor-pointer transition-colors duration-150"
          >
            Send Message
          </button>
        </div>
      </template>
    </div>
  </Teleport>
</template>
