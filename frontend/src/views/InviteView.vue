<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useServersStore } from '@/stores/servers'
import api from '@/services/api.service'

const route = useRoute()
const router = useRouter()
const serversStore = useServersStore()

const server = ref(null)
const error = ref('')
const joining = ref(false)

onMounted(async () => {
  try {
    const data = await api.resolveInvite(route.params.code)
    server.value = data.server
  } catch {
    error.value = 'This invite is invalid or has expired.'
  }
})

async function accept() {
  joining.value = true
  try {
    const joined = await serversStore.joinServer(route.params.code)
    serversStore.selectServer(joined)
    router.push(`/channels/${joined.id}`)
  } catch {
    error.value = 'Failed to join server.'
  } finally {
    joining.value = false
  }
}
</script>

<template>
  <div class="min-h-screen bg-[var(--surface)] flex items-center justify-center px-6">
    <!-- Decorative background -->
    <div class="fixed top-[-20%] right-[-10%] w-[500px] h-[500px] rounded-full bg-[#E8521A] opacity-[0.04] blur-[120px]"></div>
    <div class="fixed bottom-[-20%] left-[-10%] w-[400px] h-[400px] rounded-full bg-[#D4782A] opacity-[0.04] blur-[100px]"></div>

    <div class="relative bg-[var(--modal-bg)] rounded-2xl p-8 w-full max-w-sm shadow-xl shadow-black/5 text-center animate-scale-in border border-[var(--modal-border)]">
      <template v-if="error">
        <div class="w-14 h-14 rounded-full bg-[#E8521A]/10 flex items-center justify-center mx-auto mb-4">
          <svg class="w-6 h-6 text-[#E8521A]" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </div>
        <h1 class="font-display text-xl font-bold text-[var(--text-1)] mb-2">Invalid Invite</h1>
        <p class="text-[var(--text-3)] text-sm mb-8">{{ error }}</p>
        <button
          @click="router.push('/channels/@me')"
          class="bg-[var(--text-1)] text-[var(--surface)] px-6 py-2.5 rounded-xl hover:opacity-90 cursor-pointer text-sm font-semibold transition-all duration-200"
        >
          Go Home
        </button>
      </template>

      <template v-else-if="server">
        <div
          class="w-16 h-16 rounded-2xl bg-[#E8521A] flex items-center justify-center text-white text-2xl font-bold mx-auto mb-5 shadow-lg shadow-[#E8521A]/20"
        >
          {{ server.name?.charAt(0)?.toUpperCase() }}
        </div>
        <p class="text-[var(--text-3)] text-sm mb-1">You've been invited to join</p>
        <h1 class="font-display text-2xl font-bold text-[var(--text-1)] mb-8">{{ server.name }}</h1>
        <button
          @click="accept"
          :disabled="joining"
          class="w-full bg-[#E8521A] text-white py-3 rounded-xl hover:bg-[#D44818] disabled:opacity-50 cursor-pointer text-sm font-semibold shadow-lg shadow-[#E8521A]/15 transition-all duration-200"
        >
          {{ joining ? 'Joining...' : 'Accept Invite' }}
        </button>
      </template>

      <template v-else>
        <div class="py-10 flex flex-col items-center gap-3">
          <div class="w-8 h-8 border-2 border-[#E8521A] border-t-transparent rounded-full animate-spin"></div>
          <span class="text-[var(--text-3)] text-sm">Loading invite...</span>
        </div>
      </template>
    </div>
  </div>
</template>
