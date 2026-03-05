<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useServersStore } from '@/stores/servers'
import { useToastStore } from '@/stores/toast'
import api, { API_URL } from '@/services/api.service'
import { useEscapeClose } from '@/composables/useEscapeClose'

const props = defineProps({ initialTab: { type: String, default: 'invite' } })
const emit = defineEmits(['close'])
useEscapeClose(() => emit('close'))
const router = useRouter()
const serversStore = useServersStore()
const toastStore = useToastStore()

const activeTab = ref(props.initialTab)
const inviteCode = ref('')
const error = ref('')
const inputRef = ref(null)

const publicServers = ref([])
const hasMoreServers = ref(false)
const loadingPublic = ref(false)
const joiningId = ref(null)

async function loadPublicServers() {
  loadingPublic.value = true
  try {
    const data = await api.getPublicServers()
    publicServers.value = data.servers
    hasMoreServers.value = data.has_more
  } catch {
    toastStore.add('Failed to load public servers')
  } finally {
    loadingPublic.value = false
  }
}

onMounted(async () => {
  if (activeTab.value === 'explore') {
    await loadPublicServers()
  } else {
    inputRef.value?.focus()
  }
})

function extractCode(input) {
  const match = input.match(/\/invite\/([a-zA-Z0-9]+)/)
  return match ? match[1] : input
}

async function join() {
  if (!inviteCode.value.trim()) return
  error.value = ''
  try {
    const code = extractCode(inviteCode.value.trim())
    const server = await serversStore.joinServer(code)
    serversStore.selectServer(server)
    router.push(`/channels/${server.id}`)
    emit('close')
  } catch {
    error.value = 'Invalid invite code'
  }
}

async function switchToExplore() {
  activeTab.value = 'explore'
  if (publicServers.value.length > 0 || loadingPublic.value) return
  await loadPublicServers()
}

async function joinPublic(server) {
  if (joiningId.value) return
  joiningId.value = server.id
  try {
    const joined = await serversStore.joinPublicServer(server.id)
    serversStore.selectServer(joined)
    router.push(`/channels/${joined.id}`)
    emit('close')
  } catch {
    toastStore.add('Failed to join server')
  } finally {
    joiningId.value = null
  }
}

function serverIcon(s) {
  return s.icon_url ? `${API_URL}${s.icon_url}` : null
}

function getInitials(name) {
  return name.split(' ').map((w) => w[0]).join('').slice(0, 2).toUpperCase()
}

function formatCount(n) {
  if (n >= 1000) return (n / 1000).toFixed(1).replace(/\.0$/, '') + 'k'
  return String(n)
}
</script>

<template>
  <Teleport to="body">
    <div class="fixed inset-0 bg-[var(--backdrop)] backdrop-blur-sm flex items-center justify-center z-50" @click.self="emit('close')">
      <div class="bg-[var(--modal-bg)] rounded-2xl w-full max-w-md shadow-2xl shadow-black/10 animate-scale-in border border-[var(--modal-border)] flex flex-col" style="max-height: 80vh">

        <!-- Header -->
        <div class="p-7 pb-0 shrink-0">
          <h2 class="font-display text-2xl font-bold text-[var(--text-1)] mb-4">
            {{ activeTab === 'invite' ? 'Join a Server' : 'Explore Servers' }}
          </h2>

          <!-- Tabs -->
          <div class="flex gap-1 border-b border-[var(--surface-border)]">
            <button
              @click="activeTab = 'invite'; $nextTick(() => inputRef?.focus())"
              class="px-4 py-2.5 text-sm font-medium transition-colors duration-150 cursor-pointer border-b-2 -mb-px"
              :class="activeTab === 'invite' ? 'text-[var(--text-1)] border-[#E8521A]' : 'text-[var(--text-3)] border-transparent hover:text-[var(--text-1)]'"
            >Invite Code</button>
            <button
              @click="switchToExplore"
              class="px-4 py-2.5 text-sm font-medium transition-colors duration-150 cursor-pointer border-b-2 -mb-px"
              :class="activeTab === 'explore' ? 'text-[var(--text-1)] border-[#E8521A]' : 'text-[var(--text-3)] border-transparent hover:text-[var(--text-1)]'"
            >Explore</button>
          </div>
        </div>

        <!-- Invite Code Tab -->
        <div v-if="activeTab === 'invite'" class="p-7 pt-5">
          <p class="text-[var(--text-3)] text-sm mb-5">Enter an invite code or link to join</p>

          <label class="text-[11px] font-bold text-[var(--text-4)] uppercase tracking-[0.1em]">Invite Code</label>
          <input
            ref="inputRef"
            v-model="inviteCode"
            @keyup.enter="join"
            placeholder="abc123ef or https://..."
            class="w-full bg-[var(--surface)] text-[var(--text-1)] px-3.5 py-2.5 rounded-xl mt-2 mb-1 placeholder-[var(--text-4)] text-sm border border-[var(--surface-border)]"
          />
          <p v-if="error" class="text-[#E8521A] text-sm mb-5 mt-1.5 font-medium">{{ error }}</p>
          <div v-else class="mb-5"></div>

          <div class="flex justify-end gap-3">
            <button @click="emit('close')" class="text-[var(--text-3)] hover:text-[var(--text-1)] px-4 py-2.5 rounded-xl cursor-pointer font-medium transition-colors duration-150">
              Cancel
            </button>
            <button
              @click="join"
              :disabled="!inviteCode.trim()"
              class="bg-[#E8521A] text-white px-5 py-2.5 rounded-xl hover:bg-[#D44818] disabled:opacity-40 cursor-pointer font-semibold shadow-lg shadow-[#E8521A]/15 transition-all duration-200"
            >
              Join
            </button>
          </div>
        </div>

        <!-- Explore Tab -->
        <div v-else class="flex flex-col overflow-hidden flex-1">
          <!-- Loading -->
          <div v-if="loadingPublic" class="flex items-center justify-center py-16 text-[var(--text-4)]">
            <svg class="w-5 h-5 animate-spin mr-2" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8z"/>
            </svg>
            <span class="text-sm">Loading servers…</span>
          </div>

          <!-- Empty -->
          <div v-else-if="publicServers.length === 0" class="flex flex-col items-center justify-center py-16 px-7 text-center">
            <div class="w-12 h-12 rounded-2xl bg-[var(--surface)] flex items-center justify-center mb-3">
              <svg class="w-6 h-6 text-[var(--text-4)]" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 6A2.25 2.25 0 016 3.75h2.25A2.25 2.25 0 0110.5 6v2.25a2.25 2.25 0 01-2.25 2.25H6a2.25 2.25 0 01-2.25-2.25V6zM3.75 15.75A2.25 2.25 0 016 13.5h2.25a2.25 2.25 0 012.25 2.25V18a2.25 2.25 0 01-2.25 2.25H6A2.25 2.25 0 013.75 18v-2.25zM13.5 6a2.25 2.25 0 012.25-2.25H18A2.25 2.25 0 0120.25 6v2.25A2.25 2.25 0 0118 10.5h-2.25a2.25 2.25 0 01-2.25-2.25V6zM13.5 15.75a2.25 2.25 0 012.25-2.25H18a2.25 2.25 0 012.25 2.25V18A2.25 2.25 0 0118 20.25h-2.25A2.25 2.25 0 0113.5 18v-2.25z" />
              </svg>
            </div>
            <p class="text-sm font-medium text-[var(--text-2)]">
              {{ hasMoreServers ? 'No new servers to join' : 'No public servers yet' }}
            </p>
            <p class="text-xs text-[var(--text-4)] mt-1">
              {{ hasMoreServers ? "You're already in all available public servers" : 'Public servers will appear here once created' }}
            </p>
          </div>

          <!-- Server list -->
          <div v-else class="overflow-y-auto flex-1 p-4 space-y-2">
            <div
              v-for="s in publicServers"
              :key="s.id"
              class="flex items-center gap-3 p-3 rounded-xl hover:bg-[var(--surface-2)] transition-colors duration-100 group"
            >
              <!-- Icon -->
              <div class="w-11 h-11 rounded-[14px] shrink-0 overflow-hidden flex items-center justify-center bg-[var(--surface)] text-[var(--text-3)] font-bold text-sm">
                <img v-if="serverIcon(s)" :src="serverIcon(s)" class="w-full h-full object-cover" />
                <span v-else>{{ getInitials(s.name) }}</span>
              </div>

              <!-- Info -->
              <div class="flex-1 min-w-0">
                <p class="text-sm font-semibold text-[var(--text-1)] truncate">{{ s.name }}</p>
                <p class="text-xs text-[var(--text-4)]">{{ formatCount(s.member_count) }} {{ s.member_count === 1 ? 'member' : 'members' }}</p>
              </div>

              <!-- Join button -->
              <button
                @click="joinPublic(s)"
                :disabled="joiningId === s.id"
                class="shrink-0 bg-[#E8521A] text-white text-xs font-semibold px-3.5 py-1.5 rounded-lg hover:bg-[#D44818] disabled:opacity-60 cursor-pointer transition-all duration-150 shadow-sm shadow-[#E8521A]/20"
              >
                {{ joiningId === s.id ? 'Joining…' : 'Join' }}
              </button>
            </div>
          </div>

          <!-- Footer -->
          <div class="p-5 pt-3 shrink-0 flex justify-end border-t border-[var(--surface-border)]">
            <button @click="emit('close')" class="text-[var(--text-3)] hover:text-[var(--text-1)] px-4 py-2.5 rounded-xl cursor-pointer font-medium transition-colors duration-150">
              Close
            </button>
          </div>
        </div>

      </div>
    </div>
  </Teleport>
</template>
