<script setup>
import { ref, computed, onMounted } from 'vue'
import { useServersStore } from '@/stores/servers'
import { useChannelsStore } from '@/stores/channels'
import { useToastStore } from '@/stores/toast'
import api from '@/services/api.service'
import { useEscapeClose } from '@/composables/useEscapeClose'

const props = defineProps({
  channel: { type: Object, required: true },
})
const emit = defineEmits(['close'])

const serversStore = useServersStore()
const channelsStore = useChannelsStore()
const toastStore = useToastStore()

const federationId = ref(props.channel.federation_id || null)
const links = ref([])
const loadingLinks = ref(false)
const enabling = ref(false)

// Link form
const remoteAddress = ref('')
const linking = ref(false)

// Confirm disable
const showDisableConfirm = ref(false)

useEscapeClose(() => {
  if (showDisableConfirm.value) { showDisableConfirm.value = false; return }
  emit('close')
})

const federationAddress = computed(() => {
  if (!federationId.value) return ''
  const domain = window.location.hostname
  return `${domain}/fed/${federationId.value}`
})

const copied = ref(false)
function copyAddress() {
  navigator.clipboard.writeText(federationAddress.value)
  copied.value = true
  setTimeout(() => (copied.value = false), 2000)
}

onMounted(async () => {
  if (federationId.value) {
    await loadLinks()
  }
})

async function loadLinks() {
  // Links are loaded from channel federation link list — we need a server-side endpoint
  // For now, links are tracked locally. We'll fetch by enabling federation and reading links.
  loadingLinks.value = true
  try {
    // Use the channel federation endpoint to get link info
    const ch = await api.enableChannelFederation(serversStore.currentServer.id, props.channel.id)
    federationId.value = ch.federation_id
    // Links are not returned directly from enable; we rely on the channel data
  } catch {
    // ignore
  }
  loadingLinks.value = false
}

async function enableFederation() {
  enabling.value = true
  try {
    const ch = await api.enableChannelFederation(serversStore.currentServer.id, props.channel.id)
    federationId.value = ch.federation_id
    // Update channel in store
    channelsStore.updateChannelLocal(props.channel.id, { federation_id: ch.federation_id })
    toastStore.add('Federation enabled')
  } catch (e) {
    toastStore.add(e.response?.data?.error || 'Failed to enable federation')
  } finally {
    enabling.value = false
  }
}

async function disableFederation() {
  try {
    await api.disableChannelFederation(serversStore.currentServer.id, props.channel.id)
    federationId.value = null
    links.value = []
    showDisableConfirm.value = false
    channelsStore.updateChannelLocal(props.channel.id, { federation_id: null })
    toastStore.add('Federation disabled')
  } catch (e) {
    toastStore.add(e.response?.data?.error || 'Failed to disable federation')
  }
}

async function addLink() {
  if (!remoteAddress.value.trim()) return
  linking.value = true
  try {
    const link = await api.linkRemoteChannel(serversStore.currentServer.id, props.channel.id, remoteAddress.value.trim())
    links.value.push(link)
    remoteAddress.value = ''
    toastStore.add('Channel linked')
  } catch (e) {
    toastStore.add(e.response?.data?.error || 'Failed to link channel')
  } finally {
    linking.value = false
  }
}

async function removeLink(linkId) {
  try {
    await api.unlinkRemoteChannel(serversStore.currentServer.id, props.channel.id, linkId)
    links.value = links.value.filter(l => l.id !== linkId)
    toastStore.add('Link removed')
  } catch {
    toastStore.add('Failed to remove link')
  }
}
</script>

<template>
  <Teleport to="body">
    <div class="fixed inset-0 bg-[var(--backdrop)] backdrop-blur-sm flex items-center justify-center z-50" @click.self="emit('close')">
      <div class="bg-[var(--modal-bg)] rounded-2xl p-6 w-full max-w-lg shadow-2xl shadow-black/10 animate-scale-in border border-[var(--modal-border)]">
        <div class="flex items-center justify-between mb-5">
          <h2 class="text-xl font-bold text-[var(--text-1)]">Channel Federation</h2>
          <button @click="emit('close')" class="text-[var(--text-4)] hover:text-[var(--text-2)] p-1 cursor-pointer transition-colors duration-150">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <p class="text-[var(--text-3)] text-sm mb-5">
          <span class="font-semibold text-[var(--text-2)]">#{{ channel.name }}</span> — Bridge this channel with remote instances.
        </p>

        <!-- Not enabled yet -->
        <div v-if="!federationId" class="text-center py-6">
          <p class="text-[var(--text-3)] text-sm mb-4">Federation is not enabled on this channel.</p>
          <button
            @click="enableFederation"
            :disabled="enabling"
            class="px-5 py-2.5 rounded-xl text-sm font-semibold text-white bg-[#E8521A] hover:bg-[#D44818] shadow-lg shadow-[#E8521A]/15 transition-all duration-150 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
          >{{ enabling ? 'Enabling...' : 'Enable Federation' }}</button>
        </div>

        <!-- Enabled -->
        <div v-else class="space-y-5">
          <!-- Federation address -->
          <div>
            <label class="block text-xs font-semibold text-[var(--text-3)] uppercase tracking-wider mb-1.5">Federation Address</label>
            <div class="flex items-center gap-2">
              <code class="flex-1 bg-[var(--surface-3)] text-[var(--text-2)] text-sm px-3 py-2.5 rounded-xl border border-[var(--surface-border)] truncate select-all">{{ federationAddress }}</code>
              <button
                @click="copyAddress"
                class="shrink-0 px-3 py-2.5 rounded-xl text-sm font-medium border transition-all duration-150 cursor-pointer"
                :class="copied
                  ? 'bg-emerald-500/15 text-emerald-400 border-emerald-500/30'
                  : 'bg-[var(--surface-2)] text-[var(--text-2)] border-[var(--surface-border)] hover:border-[var(--text-4)]'"
              >{{ copied ? 'Copied' : 'Copy' }}</button>
            </div>
          </div>

          <!-- Linked channels -->
          <div>
            <label class="block text-xs font-semibold text-[var(--text-3)] uppercase tracking-wider mb-1.5">Linked Remote Channels</label>
            <div v-if="links.length === 0" class="text-[var(--text-4)] text-sm py-3 text-center">
              No linked channels yet.
            </div>
            <div v-else class="space-y-1.5">
              <div
                v-for="link in links"
                :key="link.id"
                class="flex items-center gap-2 bg-[var(--surface-3)] px-3 py-2 rounded-xl border border-[var(--surface-border)]"
              >
                <svg class="w-4 h-4 text-[var(--text-4)] shrink-0" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M12 21a9.004 9.004 0 0 0 8.716-6.747M12 21a9.004 9.004 0 0 1-8.716-6.747M12 21c2.485 0 4.5-4.03 4.5-9S14.485 3 12 3m0 18c-2.485 0-4.5-4.03-4.5-9S9.515 3 12 3m0 0a8.997 8.997 0 0 1 7.843 4.582M12 3a8.997 8.997 0 0 0-7.843 4.582m15.686 0A11.953 11.953 0 0 1 12 10.5c-2.998 0-5.74-1.1-7.843-2.918m15.686 0A8.959 8.959 0 0 1 21 12c0 .778-.099 1.533-.284 2.253m0 0A17.919 17.919 0 0 1 12 16.5c-3.162 0-6.133-.815-8.716-2.247m0 0A9.015 9.015 0 0 1 3 12c0-1.605.42-3.113 1.157-4.418" />
                </svg>
                <span class="text-sm text-[var(--text-2)] truncate flex-1">{{ link.remote_domain }}/fed/{{ link.remote_federation_id }}</span>
                <button
                  @click="removeLink(link.id)"
                  class="shrink-0 text-[var(--text-4)] hover:text-red-400 p-0.5 cursor-pointer transition-colors duration-150"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
            </div>
          </div>

          <!-- Add link -->
          <div>
            <label class="block text-xs font-semibold text-[var(--text-3)] uppercase tracking-wider mb-1.5">Link Remote Channel</label>
            <div class="flex gap-2">
              <input
                v-model="remoteAddress"
                placeholder="domain.tld/fed/federation_id"
                @keyup.enter="addLink"
                class="flex-1 bg-[var(--surface-3)] text-[var(--text-1)] text-sm px-3 py-2.5 rounded-xl placeholder-[var(--text-4)] border border-[var(--surface-border)] focus:outline-none focus:border-[#E8521A]/50 transition-all duration-150"
              />
              <button
                @click="addLink"
                :disabled="linking || !remoteAddress.trim()"
                class="px-4 py-2.5 rounded-xl text-sm font-semibold text-white bg-[#E8521A] hover:bg-[#D44818] shadow-lg shadow-[#E8521A]/15 transition-all duration-150 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
              >{{ linking ? 'Linking...' : 'Link' }}</button>
            </div>
          </div>

          <!-- Disable -->
          <div class="pt-2 border-t border-[var(--modal-border)]">
            <div v-if="!showDisableConfirm">
              <button
                @click="showDisableConfirm = true"
                class="text-sm text-red-400 hover:text-red-300 font-medium cursor-pointer transition-colors duration-150"
              >Disable federation on this channel</button>
            </div>
            <div v-else class="flex items-center gap-3">
              <p class="text-sm text-red-400 flex-1">This will remove all links. Continue?</p>
              <button
                @click="showDisableConfirm = false"
                class="text-sm text-[var(--text-3)] hover:text-[var(--text-1)] px-3 py-1.5 rounded-lg cursor-pointer transition-colors duration-150"
              >Cancel</button>
              <button
                @click="disableFederation"
                class="text-sm font-semibold text-red-400 hover:text-red-300 px-3 py-1.5 rounded-lg bg-red-500/10 hover:bg-red-500/20 cursor-pointer transition-all duration-150"
              >Disable</button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>
