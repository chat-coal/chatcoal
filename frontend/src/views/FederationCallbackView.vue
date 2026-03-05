<script setup>
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const error = ref('')

onMounted(async () => {
  const token = route.query.token
  if (!token) {
    error.value = 'Missing token in callback URL.'
    return
  }

  try {
    await auth.loginWithFederationCallback(token)
    // Clean the token from the URL to prevent leakage via history/Referer.
    router.replace({ query: {} })
    const lastChannel = localStorage.getItem('lastChannel')
    const dest = lastChannel && lastChannel.startsWith('/channels/') ? lastChannel : '/channels/@me'
    router.replace(dest)
  } catch (e) {
    error.value = e?.response?.data?.error || e?.message || 'Federation login failed.'
  }
})
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-[var(--surface)]">
    <div class="w-full max-w-sm text-center px-6 animate-fade-in-up">
      <template v-if="!error">
        <!-- Loading state -->
        <svg class="w-10 h-10 animate-spin text-[#E8521A] mx-auto mb-6" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
        </svg>
        <p class="text-[var(--text-2)] text-sm">Completing sign-in…</p>
      </template>

      <template v-else>
        <!-- Error state -->
        <div class="w-12 h-12 rounded-full bg-[#E8521A]/15 flex items-center justify-center mx-auto mb-5">
          <svg class="w-6 h-6 text-[#E8521A]" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
        <h2 class="text-[var(--text-1)] font-semibold text-lg mb-2">Sign-in failed</h2>
        <p class="text-[var(--text-3)] text-sm mb-6">{{ error }}</p>
        <a
          href="/login"
          class="inline-block bg-[var(--surface-2)] text-[var(--text-1)] text-sm font-medium px-5 py-2.5 rounded-xl hover:bg-[var(--surface-3)] transition-colors"
        >
          Back to login
        </a>
      </template>
    </div>
  </div>
</template>
