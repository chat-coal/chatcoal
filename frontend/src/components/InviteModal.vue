<script setup>
import { ref, onMounted } from 'vue'
import { useServersStore } from '@/stores/servers'
import { useInvitesStore } from '@/stores/invites'
import { useEscapeClose } from '@/composables/useEscapeClose'

const emit = defineEmits(['close'])
useEscapeClose(() => emit('close'))
const serversStore = useServersStore()
const invitesStore = useInvitesStore()

const expiresIn = ref(0)
const maxUses = ref(0)
const creating = ref(false)
const copiedId = ref(null)

const expiresOptions = [
  { label: 'Never', value: 0 },
  { label: '30 minutes', value: 1800 },
  { label: '1 hour', value: 3600 },
  { label: '6 hours', value: 21600 },
  { label: '12 hours', value: 43200 },
  { label: '1 day', value: 86400 },
  { label: '7 days', value: 604800 },
]

const maxUsesOptions = [
  { label: 'No limit', value: 0 },
  { label: '1 use', value: 1 },
  { label: '5 uses', value: 5 },
  { label: '10 uses', value: 10 },
  { label: '25 uses', value: 25 },
  { label: '50 uses', value: 50 },
  { label: '100 uses', value: 100 },
]

onMounted(() => {
  invitesStore.fetchInvites(serversStore.currentServer.id)
})

async function generate() {
  creating.value = true
  try {
    await invitesStore.createInvite(serversStore.currentServer.id, {
      max_uses: maxUses.value,
      expires_in: expiresIn.value,
    })
  } finally {
    creating.value = false
  }
}

async function remove(inviteId) {
  await invitesStore.deleteInvite(serversStore.currentServer.id, inviteId)
}

function copyLink(invite) {
  const url = `${window.location.origin}/invite/${invite.code}`
  navigator.clipboard.writeText(url)
  copiedId.value = invite.id
  setTimeout(() => (copiedId.value = null), 2000)
}

function formatExpiry(invite) {
  if (!invite.expires_at) return 'Never'
  const d = new Date(invite.expires_at)
  if (d < new Date()) return 'Expired'
  const diff = d - new Date()
  const hours = Math.floor(diff / 3600000)
  const mins = Math.floor((diff % 3600000) / 60000)
  if (hours > 24) return `${Math.floor(hours / 24)}d`
  if (hours > 0) return `${hours}h ${mins}m`
  return `${mins}m`
}
</script>

<template>
  <Teleport to="body">
    <div class="fixed inset-0 bg-[var(--backdrop)] backdrop-blur-sm flex items-center justify-center z-50" @click.self="emit('close')">
      <div class="bg-[var(--modal-bg)] rounded-2xl p-7 w-full max-w-lg max-h-[80vh] flex flex-col shadow-2xl shadow-black/10 animate-scale-in border border-[var(--modal-border)]">
        <h2 class="font-display text-2xl font-bold text-[var(--text-1)] mb-1">Invite People</h2>
        <p class="text-[var(--text-3)] text-sm mb-6">
          to <span class="font-semibold text-[var(--text-1)]">{{ serversStore.currentServer?.name }}</span>
        </p>

        <!-- Create section -->
        <div class="flex gap-3 mb-4">
          <div class="flex-1">
            <label class="text-[11px] font-bold text-[var(--text-4)] uppercase tracking-[0.1em]">Expire After</label>
            <div class="relative mt-1.5">
              <select
                v-model="expiresIn"
                class="w-full appearance-none bg-[var(--surface)] text-[var(--text-1)] px-3 py-2.5 pr-9 rounded-xl text-sm border border-[var(--surface-border)] cursor-pointer"
              >
                <option v-for="opt in expiresOptions" :key="opt.value" :value="opt.value">
                  {{ opt.label }}
                </option>
              </select>
              <svg class="pointer-events-none absolute right-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-3)]" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.23 7.21a.75.75 0 011.06.02L10 11.168l3.71-3.938a.75.75 0 111.08 1.04l-4.25 4.5a.75.75 0 01-1.08 0l-4.25-4.5a.75.75 0 01.02-1.06z" clip-rule="evenodd"/></svg>
            </div>
          </div>
          <div class="flex-1">
            <label class="text-[11px] font-bold text-[var(--text-4)] uppercase tracking-[0.1em]">Max Uses</label>
            <div class="relative mt-1.5">
              <select
                v-model="maxUses"
                class="w-full appearance-none bg-[var(--surface)] text-[var(--text-1)] px-3 py-2.5 pr-9 rounded-xl text-sm border border-[var(--surface-border)] cursor-pointer"
              >
                <option v-for="opt in maxUsesOptions" :key="opt.value" :value="opt.value">
                  {{ opt.label }}
                </option>
              </select>
              <svg class="pointer-events-none absolute right-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--text-3)]" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.23 7.21a.75.75 0 011.06.02L10 11.168l3.71-3.938a.75.75 0 111.08 1.04l-4.25 4.5a.75.75 0 01-1.08 0l-4.25-4.5a.75.75 0 01.02-1.06z" clip-rule="evenodd"/></svg>
            </div>
          </div>
        </div>

        <button
          @click="generate"
          :disabled="creating"
          class="w-full bg-[#E8521A] text-white py-2.5 rounded-xl hover:bg-[#D44818] disabled:opacity-40 mb-6 cursor-pointer text-sm font-semibold shadow-lg shadow-[#E8521A]/15 transition-all duration-200"
        >
          {{ creating ? 'Generating...' : 'Generate Invite Link' }}
        </button>

        <!-- Divider -->
        <div class="border-t border-[var(--surface-border)] mb-5"></div>

        <!-- Invites list -->
        <div class="flex-1 overflow-y-auto scrollbar-light min-h-0">
          <div v-if="invitesStore.invites.length === 0" class="text-[var(--text-4)] text-sm text-center py-6">
            No invites yet
          </div>
          <div
            v-for="invite in invitesStore.invites"
            :key="invite.id"
            class="flex items-center justify-between py-2.5 px-3 rounded-xl hover:bg-[var(--surface-2)] group transition-colors duration-100"
          >
            <div class="min-w-0 flex-1">
              <div class="text-[var(--text-1)] text-sm font-mono font-semibold truncate">{{ invite.code }}</div>
              <div class="text-[var(--text-4)] text-[11px] mt-0.5">
                {{ invite.creator?.display_name || 'Unknown' }}
                &middot; {{ invite.uses }}{{ invite.max_uses > 0 ? `/${invite.max_uses}` : '' }} uses
                &middot; {{ formatExpiry(invite) }}
              </div>
            </div>
            <div class="flex gap-1 ml-2 shrink-0">
              <button
                @click="copyLink(invite)"
                class="text-[var(--text-4)] hover:text-[var(--text-1)] p-1.5 rounded-lg hover:bg-[var(--surface-3)] cursor-pointer transition-colors duration-100"
                :title="copiedId === invite.id ? 'Copied!' : 'Copy link'"
              >
                <svg v-if="copiedId === invite.id" class="w-4 h-4 text-[#D4782A]" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                </svg>
                <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                  <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
                  <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
                </svg>
              </button>
              <button
                @click="remove(invite.id)"
                class="text-[var(--text-4)] hover:text-[#E8521A] p-1.5 rounded-lg hover:bg-[#E8521A]/10 cursor-pointer transition-colors duration-100"
                title="Delete invite"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
              </button>
            </div>
          </div>
        </div>

        <!-- Close button -->
        <div class="flex justify-end mt-5">
          <button @click="emit('close')" class="text-[var(--text-3)] hover:text-[var(--text-1)] px-4 py-2.5 rounded-xl cursor-pointer text-sm font-medium transition-colors duration-150">
            Close
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>
