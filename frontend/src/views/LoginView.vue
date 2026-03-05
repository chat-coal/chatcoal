<script setup>
import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useRoute, useRouter } from 'vue-router'
import { auth as firebaseAuth, sendPasswordResetEmail } from '@/services/firebase'
import logoSvg from '@/assets/logo.svg'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()

const mode = ref('signin') // 'signin' | 'register'
const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)
const showEmailForm = ref(false)
const showVerificationNotice = ref(false)

// Forgot password
const showForgotPassword = ref(false)
const resetLoading = ref(false)
const resetMessage = ref('')
const resetError = ref('')

async function handleForgotPassword() {
  resetError.value = ''
  resetMessage.value = ''
  const target = email.value.trim()
  if (!target) {
    resetError.value = 'Please enter your email in the field above.'
    return
  }
  resetLoading.value = true
  try {
    await sendPasswordResetEmail(firebaseAuth, target)
    resetMessage.value = 'Password reset email sent! Check your inbox.'
  } catch (e) {
    const map = {
      'auth/user-not-found': 'No account found with that email.',
      'auth/invalid-email': 'Please enter a valid email address.',
      'auth/too-many-requests': 'Too many attempts. Try again later.',
    }
    resetError.value = map[e.code] || 'Failed to send reset email. Please try again.'
  } finally {
    resetLoading.value = false
  }
}

// Federated login
const federatedId = ref('')
const showFederatedForm = ref(false)
const federatedError = ref('')
const federatedLoading = ref(false)

function redirectAfterLogin() {
  if (!auth.dbUser?.display_name) {
    router.push('/onboarding')
  } else {
    router.push(route.query.redirect || '/channels/@me')
  }
}

function parseFirebaseError(code) {
  const map = {
    'auth/user-not-found': 'No account found with that email.',
    'auth/wrong-password': 'Incorrect password.',
    'auth/invalid-credential': 'Invalid email or password.',
    'auth/email-already-in-use': 'An account with that email already exists.',
    'auth/weak-password': 'Password must be at least 6 characters.',
    'auth/invalid-email': 'Please enter a valid email address.',
    'auth/too-many-requests': 'Too many attempts. Try again later.',
  }
  return map[code] || 'Something went wrong. Please try again.'
}

async function handleGoogle() {
  error.value = ''
  loading.value = true
  try {
    await auth.loginWithGoogle()
    redirectAfterLogin()
  } catch (e) {
    if (e.code !== 'auth/popup-closed-by-user' && e.code !== 'auth/cancelled-popup-request') {
      error.value = parseFirebaseError(e.code)
    }
  } finally {
    loading.value = false
  }
}

async function handleEmailAuth() {
  error.value = ''
  loading.value = true
  try {
    if (mode.value === 'signin') {
      await auth.loginWithEmail(email.value, password.value)
      redirectAfterLogin()
    } else {
      await auth.registerWithEmail(email.value, password.value)
      showVerificationNotice.value = true
      setTimeout(() => redirectAfterLogin(), 3000)
    }
  } catch (e) {
    error.value = parseFirebaseError(e.code)
  } finally {
    loading.value = false
  }
}

async function handleAnonymous() {
  error.value = ''
  loading.value = true
  try {
    await auth.loginAnonymously()
    redirectAfterLogin()
  } catch {
    error.value = 'Failed to continue as guest.'
  } finally {
    loading.value = false
  }
}

async function handleFederatedLogin() {
  federatedError.value = ''
  const id = federatedId.value.trim()
  // Validate format: username@domain.tld
  const parts = id.split('@')
  if (parts.length !== 2 || !parts[0] || !parts[1].includes('.')) {
    federatedError.value = 'Enter your account as username@instance.com'
    return
  }
  federatedLoading.value = true
  try {
    await auth.loginWithFederation(id)
    // loginWithFederation redirects the browser — no return expected.
  } catch (e) {
    federatedError.value = e?.response?.data?.error || 'Could not reach that instance. Please check the address.'
    federatedLoading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex bg-[var(--surface)]">
    <!-- Left: Branding panel -->
    <div class="hidden lg:flex lg:w-[45%] relative overflow-hidden bg-[var(--login-bg)] items-center justify-center noise-texture">
      <!-- Ember glow blobs -->
      <div class="absolute top-[-10%] right-[-8%] w-[500px] h-[500px] rounded-full bg-[#E8521A] animate-ember-pulse blur-[120px]" style="animation-delay: 0s;"></div>
      <div class="absolute bottom-[-15%] left-[-5%] w-[380px] h-[380px] rounded-full bg-[#D4782A] animate-ember-pulse blur-[100px]" style="animation-delay: 2s;"></div>
      <div class="absolute top-[40%] left-[10%] w-[200px] h-[200px] rounded-full bg-[#E8521A] animate-ember-pulse blur-[80px]" style="animation-delay: 1s; opacity: 0.07;"></div>

      <!-- Ember spark dots -->
      <div class="absolute inset-0 overflow-hidden pointer-events-none">
        <div class="absolute w-px h-px rounded-full bg-[#E8521A] opacity-60" style="top: 20%; left: 35%; box-shadow: 0 0 6px 2px rgba(232,82,26,0.8);"></div>
        <div class="absolute w-px h-px rounded-full bg-[#D4782A] opacity-40" style="top: 65%; left: 70%; box-shadow: 0 0 8px 3px rgba(212,120,42,0.6);"></div>
        <div class="absolute w-px h-px rounded-full bg-[#E8521A] opacity-50" style="top: 80%; left: 25%; box-shadow: 0 0 5px 2px rgba(232,82,26,0.7);"></div>
        <div class="absolute w-px h-px rounded-full bg-[#E8893A] opacity-35" style="top: 35%; left: 75%; box-shadow: 0 0 7px 3px rgba(232,137,58,0.5);"></div>
      </div>

      <div class="relative z-10 text-center px-16 animate-fade-in-up">
        <img :src="logoSvg" alt="chatcoal" class="w-24 h-24 mx-auto mb-6 drop-shadow-[0_0_30px_rgba(232,82,26,0.4)]" />
        <h1 class="font-display text-7xl font-bold text-[var(--sb-text)] mb-5 tracking-tight">
          chatcoal
        </h1>
        <p class="text-[var(--sb-text-2)] text-lg max-w-sm mx-auto leading-relaxed">
          Dark by design. Conversations that glow.
        </p>
        <div class="flex items-center justify-center gap-2 mt-10">
          <div class="w-2 h-2 rounded-full bg-[#E8521A] shadow-[0_0_8px_2px_rgba(232,82,26,0.6)]"></div>
          <div class="w-1.5 h-1.5 rounded-full bg-[#D4782A] opacity-70"></div>
          <div class="w-1 h-1 rounded-full bg-[#E8893A] opacity-50"></div>
        </div>
      </div>
    </div>

    <!-- Right: Auth form -->
    <div class="flex-1 bg-[var(--login-form-bg)] flex items-center justify-center px-6 py-12 overflow-y-auto">
      <div class="w-full max-w-sm animate-fade-in-up" style="animation-delay: 100ms;">
        <!-- Mobile branding -->
        <div class="lg:hidden text-center mb-10">
          <img :src="logoSvg" alt="chatcoal" class="w-16 h-16 mx-auto mb-4" />
          <h1 class="font-display text-5xl font-bold text-[var(--text-1)] mb-2 tracking-tight">
            chatcoal
          </h1>
          <p class="text-[var(--text-3)]">Dark by design. Conversations that glow.</p>
        </div>

        <h2 class="text-2xl font-bold text-[var(--text-1)] mb-1.5">Welcome back</h2>
        <p class="text-[var(--text-3)] text-sm mb-8">Sign in or create an account to get started</p>

        <div class="space-y-3">
          <!-- Google -->
          <button
            @click="handleGoogle"
            :disabled="loading"
            class="w-full flex items-center justify-center gap-3 bg-[var(--surface-2)] text-[var(--text-1)] font-semibold py-3.5 px-4 rounded-xl border border-[var(--surface-border)] hover:border-[#E8521A]/40 hover:shadow-lg hover:shadow-[#E8521A]/5 transition-all duration-200 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <svg class="w-5 h-5 shrink-0" viewBox="0 0 24 24">
              <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z" />
              <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" />
              <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" />
              <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" />
            </svg>
            Continue with Google
          </button>

          <!-- Email toggle button (collapsed state) -->
          <button
            v-if="!showEmailForm"
            @click="showEmailForm = true"
            :disabled="loading"
            class="w-full flex items-center justify-center gap-3 bg-[var(--surface-2)] text-[var(--text-1)] font-semibold py-3.5 px-4 rounded-xl border border-transparent hover:border-[var(--surface-border)] transition-all duration-200 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <svg class="w-5 h-5 shrink-0 text-[var(--text-3)]" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
            </svg>
            Continue with Email
          </button>

          <!-- Email/Password form (expanded state) -->
          <div v-else class="animate-fade-in-up">
            <!-- Mode tabs -->
            <div class="flex bg-[var(--surface-2)] rounded-xl p-1 mb-3">
              <button
                type="button"
                @click="mode = 'signin'; error = ''"
                class="flex-1 py-2 text-sm font-semibold rounded-lg transition-all duration-150 cursor-pointer"
                :class="mode === 'signin' ? 'bg-[var(--surface-3)] text-[var(--text-1)] shadow-sm' : 'text-[var(--text-3)] hover:text-[var(--text-2)]'"
              >
                Sign in
              </button>
              <button
                type="button"
                @click="mode = 'register'; error = ''"
                class="flex-1 py-2 text-sm font-semibold rounded-lg transition-all duration-150 cursor-pointer"
                :class="mode === 'register' ? 'bg-[var(--surface-3)] text-[var(--text-1)] shadow-sm' : 'text-[var(--text-3)] hover:text-[var(--text-2)]'"
              >
                Create account
              </button>
            </div>

            <form @submit.prevent="handleEmailAuth" class="space-y-2.5">
              <input
                v-model="email"
                type="email"
                placeholder="Email address"
                required
                autocomplete="email"
                class="w-full bg-[var(--surface-2)] text-[var(--text-1)] placeholder-[var(--text-4)] rounded-xl px-4 py-3.5 text-sm border border-transparent focus:outline-none focus:border-[#E8521A]/50 transition-all duration-150"
              />
              <input
                v-if="!showForgotPassword"
                v-model="password"
                type="password"
                placeholder="Password"
                required
                :autocomplete="mode === 'signin' ? 'current-password' : 'new-password'"
                class="w-full bg-[var(--surface-2)] text-[var(--text-1)] placeholder-[var(--text-4)] rounded-xl px-4 py-3.5 text-sm border border-transparent focus:outline-none focus:border-[#E8521A]/50 transition-all duration-150"
              />

              <!-- Forgot password link (sign-in mode only) -->
              <div v-if="mode === 'signin' && !showForgotPassword" class="flex justify-end">
                <button
                  type="button"
                  @click="showForgotPassword = true; resetError = ''; resetMessage = ''"
                  class="text-[var(--text-4)] hover:text-[#E8521A] text-xs font-medium transition-colors duration-150 cursor-pointer"
                >
                  Forgot password?
                </button>
              </div>

              <!-- Forgot password inline form -->
              <div v-if="showForgotPassword" class="bg-[var(--surface-2)] rounded-xl px-4 py-3 space-y-2.5 animate-fade-in-up">
                <p class="text-[var(--text-2)] text-xs font-medium">We'll send a reset link to <strong class="text-[var(--text-1)]">{{ email || 'the email above' }}</strong></p>
                <p v-if="resetError" class="text-[#E8521A] text-xs font-medium">{{ resetError }}</p>
                <div v-if="resetMessage" class="bg-green-500/10 border border-green-500/20 rounded-lg px-3 py-2 text-xs text-green-400 font-medium">
                  {{ resetMessage }}
                </div>
                <div class="flex gap-2">
                  <button
                    type="button"
                    @click="showForgotPassword = false"
                    class="text-[var(--text-4)] hover:text-[var(--text-2)] text-xs font-medium cursor-pointer transition-colors duration-150"
                  >
                    Back
                  </button>
                  <button
                    type="button"
                    @click="handleForgotPassword"
                    :disabled="resetLoading"
                    class="bg-[var(--surface-3)] text-[var(--text-1)] text-xs font-semibold px-3.5 py-1.5 rounded-lg hover:bg-[var(--surface-border)] transition-all duration-150 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {{ resetLoading ? 'Sending…' : 'Send reset email' }}
                  </button>
                </div>
              </div>

              <p v-if="error" class="text-[#E8521A] text-xs font-medium px-0.5 pt-0.5">{{ error }}</p>

              <!-- Verification email sent notice -->
              <div v-if="showVerificationNotice" class="bg-green-500/10 border border-green-500/20 rounded-xl px-4 py-3 text-sm text-green-400 font-medium flex items-center gap-2.5">
                <svg class="w-4 h-4 animate-spin shrink-0" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
                </svg>
                Verification email sent! Check your inbox.
              </div>

              <button
                type="submit"
                :disabled="loading || showVerificationNotice"
                class="w-full bg-[var(--surface-3)] text-[var(--text-1)] font-semibold py-3.5 px-4 rounded-xl hover:bg-[var(--surface-border)] transition-all duration-200 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <span v-if="loading" class="flex items-center justify-center gap-2">
                  <svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
                  </svg>
                </span>
                <span v-else>{{ mode === 'signin' ? 'Sign in' : 'Create account' }}</span>
              </button>
            </form>
          </div>

          <!-- Divider -->
          <div class="relative py-1">
            <div class="absolute inset-0 flex items-center">
              <div class="w-full border-t border-[var(--surface-border)]"></div>
            </div>
            <div class="relative flex justify-center">
              <span class="px-3 bg-[var(--login-form-bg)] text-[var(--text-4)] text-xs font-semibold uppercase tracking-[0.15em]">or</span>
            </div>
          </div>

          <!-- Continue as Guest -->
          <button
            @click="handleAnonymous"
            :disabled="loading"
            class="w-full bg-[#E8521A] text-white font-semibold py-3.5 px-4 rounded-xl hover:bg-[#D44818] shadow-lg shadow-[#E8521A]/20 hover:shadow-[#E8521A]/30 transition-all duration-200 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Continue as Guest
          </button>

          <!-- Divider -->
          <div class="relative py-1">
            <div class="absolute inset-0 flex items-center">
              <div class="w-full border-t border-[var(--surface-border)]"></div>
            </div>
            <div class="relative flex justify-center">
              <span class="px-3 bg-[var(--login-form-bg)] text-[var(--text-4)] text-xs font-semibold uppercase tracking-[0.15em]">federated</span>
            </div>
          </div>

          <!-- Sign in from another instance -->
          <button
            v-if="!showFederatedForm"
            @click="showFederatedForm = true"
            :disabled="loading"
            class="w-full flex items-center justify-center gap-3 bg-[var(--surface-2)] text-[var(--text-1)] font-semibold py-3.5 px-4 rounded-xl border border-transparent hover:border-[var(--surface-border)] transition-all duration-200 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <svg class="w-5 h-5 shrink-0 text-[var(--text-3)]" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
            </svg>
            Sign in from another instance
          </button>

          <div v-else class="animate-fade-in-up">
            <form @submit.prevent="handleFederatedLogin" class="space-y-2.5">
              <input
                v-model="federatedId"
                type="text"
                placeholder="username@instance.com"
                required
                autocomplete="off"
                class="w-full bg-[var(--surface-2)] text-[var(--text-1)] placeholder-[var(--text-4)] rounded-xl px-4 py-3.5 text-sm border border-transparent focus:outline-none focus:border-[#E8521A]/50 transition-all duration-150"
              />
              <p v-if="federatedError" class="text-[#E8521A] text-xs font-medium px-0.5 pt-0.5">{{ federatedError }}</p>
              <button
                type="submit"
                :disabled="federatedLoading"
                class="w-full bg-[var(--surface-3)] text-[var(--text-1)] font-semibold py-3.5 px-4 rounded-xl hover:bg-[var(--surface-border)] transition-all duration-200 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <span v-if="federatedLoading" class="flex items-center justify-center gap-2">
                  <svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
                  </svg>
                </span>
                <span v-else>Continue</span>
              </button>
            </form>
          </div>
        </div>

        <p class="text-[var(--text-4)] text-xs text-center mt-8">
          By continuing, you agree to keep things friendly.
        </p>
      </div>
    </div>
  </div>
</template>
