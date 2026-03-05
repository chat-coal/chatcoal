<script setup>
import { ref, computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { API_URL } from '@/services/api.service'
import { getAvatarColor, getDefaultAvatarStyle, cssBackgroundUrl } from '@/utils/avatar'
import { useRouter } from 'vue-router'
import ProfileSettingsModal from './ProfileSettingsModal.vue'
import MicTestModal from './MicTestModal.vue'

const authStore = useAuthStore()
const router = useRouter()
const showSettings = ref(false)
const showMicTest = ref(false)
const showLogoutConfirm = ref(false)

const isGuest = computed(() => authStore.user?.isAnonymous === true)
const isRestricted = computed(() => authStore.dbUser?.is_anonymous || authStore.dbUser?.email_verified === false)

const tooltip = ref({ visible: false, text: '', top: 0, left: 0 })
let hideTimeout = null

function showTooltip(event, text) {
  if (window.matchMedia('(hover: none)').matches) return
  clearTimeout(hideTimeout)
  const rect = event.currentTarget.getBoundingClientRect()
  tooltip.value = {
    visible: true,
    text,
    top: rect.top - 8,
    left: rect.left + rect.width / 2,
  }
}

function hideTooltip() {
  hideTimeout = setTimeout(() => {
    tooltip.value.visible = false
  }, 50)
}

// Guest account-linking state
const showLinkEmailForm = ref(false)
const linkEmail = ref('')
const linkPassword = ref('')
const linkError = ref('')
const linking = ref(false)

const avatarUrl = computed(() => {
  const url = authStore.dbUser?.avatar_url
  if (!url) return ''
  if (url.startsWith('http')) return url
  return `${API_URL}${url}`
})

function openLogoutConfirm() {
  showLinkEmailForm.value = false
  linkEmail.value = ''
  linkPassword.value = ''
  linkError.value = ''
  showLogoutConfirm.value = true
}

function closeLogoutConfirm() {
  showLogoutConfirm.value = false
  showLinkEmailForm.value = false
  linkError.value = ''
}

async function confirmLogout() {
  showLogoutConfirm.value = false
  await authStore.logout()
  router.push('/login')
}

async function confirmGuestLogout() {
  showLogoutConfirm.value = false
  try {
    await authStore.deleteAccount()
  } catch {
    await authStore.logout()
  }
  router.push('/login')
}

function parseFirebaseError(code) {
  const map = {
    'auth/email-already-in-use': 'An account with that email already exists.',
    'auth/weak-password': 'Password must be at least 6 characters.',
    'auth/invalid-email': 'Please enter a valid email address.',
    'auth/credential-already-in-use': 'That Google account is already linked to another user.',
    'auth/popup-closed-by-user': '',
    'auth/cancelled-popup-request': '',
  }
  return map[code] ?? 'Something went wrong. Please try again.'
}

async function handleLinkGoogle() {
  linkError.value = ''
  linking.value = true
  try {
    await authStore.linkWithGoogle()
    showLogoutConfirm.value = false
    router.push('/onboarding')
  } catch (e) {
    const msg = parseFirebaseError(e.code)
    if (msg) linkError.value = msg
  } finally {
    linking.value = false
  }
}

async function handleLinkEmail() {
  if (!linkEmail.value || !linkPassword.value) return
  linkError.value = ''
  linking.value = true
  try {
    await authStore.linkWithEmail(linkEmail.value, linkPassword.value)
    showLogoutConfirm.value = false
    router.push('/onboarding')
  } catch (e) {
    linkError.value = parseFirebaseError(e.code) || 'Something went wrong.'
  } finally {
    linking.value = false
  }
}
</script>

<template>
  <div class="relative z-10 bg-[var(--sb-bg)] px-3 py-2.5 flex items-center gap-2.5 border-t border-[var(--sb-border)]">
    <div class="relative">
      <div
        v-if="authStore.dbUser?.avatar_url"
        class="w-8 h-8 rounded-full bg-cover bg-center"
        :style="{ backgroundImage: cssBackgroundUrl(avatarUrl) }"
      ></div>
      <div
        v-else
        class="w-8 h-8 rounded-full flex items-center justify-center text-white text-xs font-bold"
        :style="getDefaultAvatarStyle(authStore.dbUser?.id)"
      >
        {{ (authStore.dbUser?.display_name || '?')[0].toUpperCase() }}
      </div>
      <div
        class="absolute -bottom-0.5 -right-0.5 w-3 h-3 rounded-full border-2 border-[var(--sb-bg)]"
        :class="authStore.dbUser?.status === 'online' ? 'bg-green-500' : 'bg-[var(--offline)]'"
      ></div>
    </div>
    <div class="flex-1 min-w-0">
      <p
        class="text-sm font-semibold truncate"
        :class="isRestricted ? 'text-[var(--sb-text-3)]' : 'text-[var(--sb-text)]'"
      >{{ authStore.dbUser?.display_name || 'User' }}</p>
      <p v-if="authStore.dbUser?.username" class="text-[var(--sb-text-3)] text-[11px] truncate">@{{ authStore.dbUser.username }}</p>
      <p v-else class="text-[var(--sb-text-3)] text-[11px]">{{ authStore.dbUser?.status === 'online' ? 'Online' : 'Invisible' }}</p>
    </div>
    <button
      v-if="authStore.dbUser?.is_site_admin"
      @click="router.push('/admin')"
      @mouseenter="showTooltip($event, 'Admin')"
      @mouseleave="hideTooltip"
      class="text-[var(--sb-text-3)] hover:text-[var(--sb-text)] p-1.5 rounded-lg hover:bg-[var(--sb-hover)] cursor-pointer transition-colors duration-150"
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75m-3-7.036A11.959 11.959 0 0 1 3.598 6 11.99 11.99 0 0 0 3 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285Z" />
      </svg>
    </button>
    <button
      @click="showSettings = true"
      @mouseenter="showTooltip($event, 'Profile Settings')"
      @mouseleave="hideTooltip"
      class="text-[var(--sb-text-3)] hover:text-[var(--sb-text)] p-1.5 rounded-lg hover:bg-[var(--sb-hover)] cursor-pointer transition-colors duration-150"
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" d="M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.325.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 0 1 1.37.49l1.296 2.247a1.125 1.125 0 0 1-.26 1.431l-1.003.827c-.293.241-.438.613-.43.992a7.723 7.723 0 0 1 0 .255c-.008.378.137.75.43.991l1.004.827c.424.35.534.955.26 1.43l-1.298 2.247a1.125 1.125 0 0 1-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.47 6.47 0 0 1-.22.128c-.331.183-.581.495-.644.869l-.213 1.281c-.09.543-.56.94-1.11.94h-2.594c-.55 0-1.019-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 0 1-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 0 1-1.369-.49l-1.297-2.247a1.125 1.125 0 0 1 .26-1.431l1.004-.827c.292-.24.437-.613.43-.991a6.932 6.932 0 0 1 0-.255c.007-.38-.138-.751-.43-.992l-1.004-.827a1.125 1.125 0 0 1-.26-1.43l1.297-2.247a1.125 1.125 0 0 1 1.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.086.22-.128.332-.183.582-.495.644-.869l.214-1.28Z" />
        <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" />
      </svg>
    </button>
    <button
      @click="showMicTest = true"
      @mouseenter="showTooltip($event, 'Microphone Test')"
      @mouseleave="hideTooltip"
      class="text-[var(--sb-text-3)] hover:text-[var(--sb-text)] p-1.5 rounded-lg hover:bg-[var(--sb-hover)] cursor-pointer transition-colors duration-150"
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" viewBox="0 0 24 24">
        <path d="M9 5a3 3 0 0 1 3 -3a3 3 0 0 1 3 3v5a3 3 0 0 1 -3 3a3 3 0 0 1 -3 -3l0 -5" /><path d="M5 10a7 7 0 0 0 14 0" /><path d="M8 21l8 0" /><path d="M12 17l0 4" />
      </svg>
    </button>
    <button
      @click="openLogoutConfirm"
      @mouseenter="showTooltip($event, 'Sign Out')"
      @mouseleave="hideTooltip"
      class="text-[var(--sb-text-3)] hover:text-[#E8521A] p-1.5 rounded-lg hover:bg-[var(--sb-hover)] cursor-pointer transition-colors duration-150"
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" d="M15.75 9V5.25A2.25 2.25 0 0 0 13.5 3h-6a2.25 2.25 0 0 0-2.25 2.25v13.5A2.25 2.25 0 0 0 7.5 21h6a2.25 2.25 0 0 0 2.25-2.25V15m3 0 3-3m0 0-3-3m3 3H9" />
      </svg>
    </button>
  </div>
  <!-- Tooltip (teleported above buttons) -->
  <Teleport to="body">
    <Transition
      enter-active-class="transition duration-150 ease-out"
      enter-from-class="opacity-0 translate-y-1"
      enter-to-class="opacity-100 -translate-y-0"
      leave-active-class="transition duration-100 ease-in"
      leave-from-class="opacity-100 -translate-y-0"
      leave-to-class="opacity-0 translate-y-1"
    >
      <div
        v-if="tooltip.visible"
        class="fixed z-[9999] pointer-events-none flex flex-col items-center -translate-x-1/2"
        :style="{ top: tooltip.top + 'px', left: tooltip.left + 'px' }"
      >
        <div class="bg-[#111214] text-white text-sm font-semibold px-3 py-1.5 rounded-md shadow-lg shadow-black/30 whitespace-nowrap -translate-y-full">{{ tooltip.text }}</div>
        <div class="w-2 h-2 bg-[#111214] rotate-45 rounded-[2px] shrink-0 -mt-1.5"></div>
      </div>
    </Transition>
  </Teleport>

  <ProfileSettingsModal v-if="showSettings" @close="showSettings = false" />
  <MicTestModal v-if="showMicTest" @close="showMicTest = false" />

  <!-- Guest sign-out modal -->
  <Teleport to="body">
    <div
      v-if="showLogoutConfirm && isGuest"
      class="fixed inset-0 bg-[var(--backdrop)] backdrop-blur-sm flex items-center justify-center z-50"
      @click.self="closeLogoutConfirm"
    >
      <div class="bg-[var(--modal-bg)] rounded-2xl p-7 w-full max-w-sm shadow-2xl shadow-black/10 animate-scale-in border border-[var(--modal-border)]">

        <!-- Email form step -->
        <template v-if="showLinkEmailForm">
          <button
            @click="showLinkEmailForm = false; linkError = ''"
            class="flex items-center gap-1.5 text-[var(--text-3)] hover:text-[var(--text-2)] text-sm mb-5 transition-colors duration-150 cursor-pointer"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M15.75 19.5 8.25 12l7.5-7.5" />
            </svg>
            Back
          </button>
          <h3 class="text-[var(--text-1)] font-bold text-lg mb-1">Create an account</h3>
          <p class="text-[var(--text-3)] text-sm mb-5 leading-relaxed">Your guest data will be kept.</p>
          <form @submit.prevent="handleLinkEmail" class="space-y-2.5">
            <input
              v-model="linkEmail"
              type="email"
              placeholder="Email address"
              required
              autocomplete="email"
              class="w-full bg-[var(--surface)] text-[var(--text-1)] placeholder-[var(--text-4)] rounded-xl px-3.5 py-2.5 text-sm border border-[var(--surface-border)] focus:outline-none focus:border-[#E8521A]/50 transition-all duration-150"
            />
            <input
              v-model="linkPassword"
              type="password"
              placeholder="Password (min. 6 characters)"
              required
              autocomplete="new-password"
              class="w-full bg-[var(--surface)] text-[var(--text-1)] placeholder-[var(--text-4)] rounded-xl px-3.5 py-2.5 text-sm border border-[var(--surface-border)] focus:outline-none focus:border-[#E8521A]/50 transition-all duration-150"
            />
            <p v-if="linkError" class="text-[#E8521A] text-xs font-medium px-0.5">{{ linkError }}</p>
            <button
              type="submit"
              :disabled="linking"
              class="w-full py-2.5 px-4 rounded-xl text-sm font-semibold text-white bg-[#E8521A] hover:bg-[#D44818] shadow-lg shadow-[#E8521A]/15 transition-all duration-150 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {{ linking ? 'Creating…' : 'Create account' }}
            </button>
          </form>
        </template>

        <!-- Main step -->
        <template v-else>
          <div class="flex items-center gap-3 mb-2">
            <div class="w-9 h-9 rounded-xl bg-amber-500/10 flex items-center justify-center shrink-0">
              <svg class="w-5 h-5 text-amber-500" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" />
              </svg>
            </div>
            <h3 class="text-[var(--text-1)] font-bold text-lg">Sign out?</h3>
          </div>
          <p class="text-[var(--text-3)] text-sm mb-5 leading-relaxed">
            You're signed in as a guest. Signing out will permanently delete your account and all your data.
          </p>

          <div class="space-y-2 mb-4">
            <!-- Google link -->
            <button
              @click="handleLinkGoogle"
              :disabled="linking"
              class="w-full flex items-center gap-3 bg-[var(--surface-2)] text-[var(--text-1)] font-semibold py-2.5 px-4 rounded-xl border border-[var(--surface-border)] hover:border-[#E8521A]/40 transition-all duration-150 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <svg class="w-4 h-4 shrink-0" viewBox="0 0 24 24">
                <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z" />
                <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" />
                <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" />
                <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" />
              </svg>
              Create account with Google
            </button>

            <!-- Email link -->
            <button
              @click="showLinkEmailForm = true"
              :disabled="linking"
              class="w-full flex items-center gap-3 bg-[var(--surface-2)] text-[var(--text-1)] font-semibold py-2.5 px-4 rounded-xl border border-[var(--surface-border)] hover:border-[var(--surface-border)] transition-all duration-150 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <svg class="w-4 h-4 shrink-0 text-[var(--text-3)]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
              </svg>
              Create account with email
            </button>
          </div>

          <p v-if="linkError" class="text-[#E8521A] text-xs font-medium mb-3 px-0.5">{{ linkError }}</p>

          <div class="relative mb-4">
            <div class="absolute inset-0 flex items-center">
              <div class="w-full border-t border-[var(--modal-border)]"></div>
            </div>
            <div class="relative flex justify-center">
              <span class="px-3 bg-[var(--modal-bg)] text-[var(--text-4)] text-xs font-semibold uppercase tracking-[0.12em]">or</span>
            </div>
          </div>

          <div class="flex gap-2.5">
            <button
              @click="closeLogoutConfirm"
              class="flex-1 py-2.5 px-4 rounded-xl text-sm font-semibold text-[var(--text-2)] bg-[var(--surface-2)] hover:bg-[var(--surface-3)] transition-colors duration-150 cursor-pointer"
            >
              Cancel
            </button>
            <button
              @click="confirmGuestLogout"
              :disabled="linking"
              class="flex-1 py-2.5 px-4 rounded-xl text-sm font-semibold text-red-400 bg-red-500/10 hover:bg-red-500/20 border border-red-500/20 hover:border-red-500/40 transition-all duration-150 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Sign out
            </button>
          </div>
        </template>

      </div>
    </div>
  </Teleport>

  <!-- Regular sign-out modal (non-guests) -->
  <Teleport to="body">
    <div
      v-if="showLogoutConfirm && !isGuest"
      class="fixed inset-0 bg-[var(--backdrop)] backdrop-blur-sm flex items-center justify-center z-50"
      @click.self="closeLogoutConfirm"
    >
      <div class="bg-[var(--modal-bg)] rounded-2xl p-7 w-full max-w-sm shadow-2xl shadow-black/10 animate-scale-in border border-[var(--modal-border)]">
        <div class="flex items-center gap-3 mb-2">
          <div class="w-9 h-9 rounded-xl bg-[#E8521A]/10 flex items-center justify-center shrink-0">
            <svg class="w-5 h-5 text-[#E8521A]" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M15.75 9V5.25A2.25 2.25 0 0 0 13.5 3h-6a2.25 2.25 0 0 0-2.25 2.25v13.5A2.25 2.25 0 0 0 7.5 21h6a2.25 2.25 0 0 0 2.25-2.25V15m3 0 3-3m0 0-3-3m3 3H9" />
            </svg>
          </div>
          <h3 class="text-[var(--text-1)] font-bold text-lg">Sign out?</h3>
        </div>
        <p class="text-[var(--text-3)] text-sm mb-6 leading-relaxed">
          You'll be returned to the login screen.
        </p>
        <div class="flex gap-2.5">
          <button
            @click="closeLogoutConfirm"
            class="flex-1 py-2.5 px-4 rounded-xl text-sm font-semibold text-[var(--text-2)] bg-[var(--surface-2)] hover:bg-[var(--surface-3)] transition-colors duration-150 cursor-pointer"
          >
            Cancel
          </button>
          <button
            @click="confirmLogout"
            class="flex-1 py-2.5 px-4 rounded-xl text-sm font-semibold text-white bg-[#E8521A] hover:bg-[#D44818] shadow-lg shadow-[#E8521A]/15 transition-all duration-150 cursor-pointer"
          >
            Sign out
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>
