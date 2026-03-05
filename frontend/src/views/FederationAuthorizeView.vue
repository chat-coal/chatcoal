<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import api from '@/services/api.service'
import logoSvg from '@/assets/logo.svg'

const route = useRoute()
const auth = useAuthStore()

// Query params from the redirect URL built by the visiting instance.
const visiting = computed(() => route.query.visiting || '')
const nonce = computed(() => route.query.nonce || '')
const callback = computed(() => route.query.callback || '')

// UI state
const step = ref('loading') // 'loading' | 'login' | 'consent' | 'done' | 'error'
const error = ref('')
const authorizing = ref(false)

// Inline login form state (used when the user isn't authenticated yet).
const loginMode = ref('signin') // 'signin' | 'register'
const email = ref('')
const password = ref('')
const loginError = ref('')
const loginLoading = ref(false)
const showEmailForm = ref(false)

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

onMounted(async () => {
  // Validate required params.
  if (!visiting.value || !nonce.value || !callback.value) {
    error.value = 'Missing required federation parameters.'
    step.value = 'error'
    return
  }
  // Wait for auth to finish initialising.
  if (auth.loading) {
    await new Promise((resolve) => {
      const stop = auth.$subscribe(() => {
        if (!auth.loading) { stop(); resolve() }
      })
      // Guard against race: loading may have become false between
      // the if-check and the $subscribe registration.
      if (!auth.loading) { stop(); resolve() }
    })
  }
  step.value = auth.user ? 'consent' : 'login'
})

async function handleGoogle() {
  loginError.value = ''
  loginLoading.value = true
  try {
    await auth.loginWithGoogle()
    step.value = 'consent'
  } catch (e) {
    if (e.code !== 'auth/popup-closed-by-user' && e.code !== 'auth/cancelled-popup-request') {
      loginError.value = parseFirebaseError(e.code)
    }
  } finally {
    loginLoading.value = false
  }
}

async function handleEmailAuth() {
  loginError.value = ''
  loginLoading.value = true
  try {
    if (loginMode.value === 'signin') {
      await auth.loginWithEmail(email.value, password.value)
    } else {
      await auth.registerWithEmail(email.value, password.value)
    }
    step.value = 'consent'
  } catch (e) {
    loginError.value = parseFirebaseError(e.code)
  } finally {
    loginLoading.value = false
  }
}

async function handleAuthorize() {
  authorizing.value = true
  error.value = ''
  try {
    const { data } = await api.federationAuthorize(visiting.value, nonce.value, callback.value)
    step.value = 'done'
    // Redirect the browser to the visiting instance's callback URL.
    window.location.href = data.redirect_url
  } catch (e) {
    error.value = e?.response?.data?.error || 'Authorization failed. Please try again.'
    step.value = 'consent'
  } finally {
    authorizing.value = false
  }
}

function handleCancel() {
  // Close the tab or navigate back; the visiting instance will time out.
  if (window.history.length > 1) {
    window.history.back()
  } else {
    window.close()
  }
}

// Display the visiting domain clearly (strip the trailing dot if any).
const visitingDisplay = computed(() => visiting.value.replace(/\.$/, ''))
</script>

<template>
  <div class="min-h-screen flex bg-[var(--surface)]">
    <!-- Left branding panel -->
    <div class="hidden lg:flex lg:w-[45%] relative overflow-hidden bg-[var(--login-bg)] items-center justify-center noise-texture">
      <div class="absolute top-[-10%] right-[-8%] w-[500px] h-[500px] rounded-full bg-[#E8521A] animate-ember-pulse blur-[120px]" style="animation-delay: 0s;"></div>
      <div class="absolute bottom-[-15%] left-[-5%] w-[380px] h-[380px] rounded-full bg-[#D4782A] animate-ember-pulse blur-[100px]" style="animation-delay: 2s;"></div>
      <div class="relative z-10 text-center px-16 animate-fade-in-up">
        <img :src="logoSvg" alt="chatcoal" class="w-24 h-24 mx-auto mb-6 drop-shadow-[0_0_30px_rgba(232,82,26,0.4)]" />
        <h1 class="font-display text-7xl font-bold text-[var(--sb-text)] mb-5 tracking-tight">chatcoal</h1>
        <p class="text-[var(--sb-text-2)] text-lg max-w-sm mx-auto leading-relaxed">Federated identity — one account, every instance.</p>
      </div>
    </div>

    <!-- Right panel -->
    <div class="flex-1 bg-[var(--login-form-bg)] flex items-center justify-center px-6 py-12 overflow-y-auto">
      <div class="w-full max-w-sm animate-fade-in-up" style="animation-delay: 100ms;">

        <!-- Loading -->
        <div v-if="step === 'loading'" class="text-center">
          <svg class="w-8 h-8 animate-spin text-[#E8521A] mx-auto" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
          </svg>
        </div>

        <!-- Error (invalid params) -->
        <div v-else-if="step === 'error'" class="text-center">
          <div class="w-12 h-12 rounded-full bg-[#E8521A]/15 flex items-center justify-center mx-auto mb-5">
            <svg class="w-6 h-6 text-[#E8521A]" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <h2 class="text-[var(--text-1)] font-semibold text-lg mb-2">Invalid request</h2>
          <p class="text-[var(--text-3)] text-sm">{{ error }}</p>
        </div>

        <!-- Login form (user not yet authenticated) -->
        <div v-else-if="step === 'login'">
          <!-- Mobile branding -->
          <div class="lg:hidden text-center mb-8">
            <img :src="logoSvg" alt="chatcoal" class="w-14 h-14 mx-auto mb-3" />
            <h1 class="font-display text-5xl font-bold text-[var(--text-1)] mb-2 tracking-tight">chatcoal</h1>
          </div>

          <!-- Context banner -->
          <div class="bg-[var(--surface-2)] border border-[var(--surface-border)] rounded-xl px-4 py-3 mb-6 flex items-start gap-3">
            <svg class="w-4 h-4 text-[#E8521A] mt-0.5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <p class="text-[var(--text-3)] text-xs leading-relaxed">
              <span class="text-[var(--text-2)] font-semibold">{{ visitingDisplay }}</span> wants to verify your identity.
              Sign in to continue.
            </p>
          </div>

          <h2 class="text-2xl font-bold text-[var(--text-1)] mb-1.5">Sign in to continue</h2>
          <p class="text-[var(--text-3)] text-sm mb-8">Use your account on this instance</p>

          <div class="space-y-3">
            <!-- Google -->
            <button
              @click="handleGoogle"
              :disabled="loginLoading"
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

            <!-- Email toggle -->
            <button
              v-if="!showEmailForm"
              @click="showEmailForm = true"
              :disabled="loginLoading"
              class="w-full flex items-center justify-center gap-3 bg-[var(--surface-2)] text-[var(--text-1)] font-semibold py-3.5 px-4 rounded-xl border border-transparent hover:border-[var(--surface-border)] transition-all duration-200 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <svg class="w-5 h-5 shrink-0 text-[var(--text-3)]" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
              </svg>
              Continue with Email
            </button>

            <!-- Email/password form -->
            <div v-else class="animate-fade-in-up">
              <div class="flex bg-[var(--surface-2)] rounded-xl p-1 mb-3">
                <button type="button" @click="loginMode = 'signin'; loginError = ''"
                  class="flex-1 py-2 text-sm font-semibold rounded-lg transition-all duration-150 cursor-pointer"
                  :class="loginMode === 'signin' ? 'bg-[var(--surface-3)] text-[var(--text-1)] shadow-sm' : 'text-[var(--text-3)] hover:text-[var(--text-2)]'">
                  Sign in
                </button>
                <button type="button" @click="loginMode = 'register'; loginError = ''"
                  class="flex-1 py-2 text-sm font-semibold rounded-lg transition-all duration-150 cursor-pointer"
                  :class="loginMode === 'register' ? 'bg-[var(--surface-3)] text-[var(--text-1)] shadow-sm' : 'text-[var(--text-3)] hover:text-[var(--text-2)]'">
                  Create account
                </button>
              </div>
              <form @submit.prevent="handleEmailAuth" class="space-y-2.5">
                <input v-model="email" type="email" placeholder="Email address" required autocomplete="email"
                  class="w-full bg-[var(--surface-2)] text-[var(--text-1)] placeholder-[var(--text-4)] rounded-xl px-4 py-3.5 text-sm border border-transparent focus:outline-none focus:border-[#E8521A]/50 transition-all duration-150" />
                <input v-model="password" type="password" placeholder="Password" required
                  :autocomplete="loginMode === 'signin' ? 'current-password' : 'new-password'"
                  class="w-full bg-[var(--surface-2)] text-[var(--text-1)] placeholder-[var(--text-4)] rounded-xl px-4 py-3.5 text-sm border border-transparent focus:outline-none focus:border-[#E8521A]/50 transition-all duration-150" />
                <p v-if="loginError" class="text-[#E8521A] text-xs font-medium px-0.5 pt-0.5">{{ loginError }}</p>
                <button type="submit" :disabled="loginLoading"
                  class="w-full bg-[var(--surface-3)] text-[var(--text-1)] font-semibold py-3.5 px-4 rounded-xl hover:bg-[var(--surface-border)] transition-all duration-200 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed">
                  <span v-if="loginLoading" class="flex items-center justify-center gap-2">
                    <svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
                      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                      <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
                    </svg>
                  </span>
                  <span v-else>{{ loginMode === 'signin' ? 'Sign in' : 'Create account' }}</span>
                </button>
              </form>
            </div>
          </div>
        </div>

        <!-- Consent screen -->
        <div v-else-if="step === 'consent' || step === 'done'">
          <!-- Mobile branding -->
          <div class="lg:hidden text-center mb-8">
            <img :src="logoSvg" alt="chatcoal" class="w-12 h-12 mx-auto mb-3" />
            <h1 class="font-display text-4xl font-bold text-[var(--text-1)] tracking-tight">chatcoal</h1>
          </div>

          <!-- App icon -->
          <img :src="logoSvg" alt="chatcoal" class="w-14 h-14 mx-auto mb-6 hidden lg:block" />

          <h2 class="text-2xl font-bold text-[var(--text-1)] mb-2 text-center">Authorize access</h2>
          <p class="text-[var(--text-3)] text-sm text-center mb-8">
            <span class="text-[var(--text-1)] font-semibold">{{ visitingDisplay }}</span>
            wants to verify your identity
          </p>

          <!-- Permission summary -->
          <div class="bg-[var(--surface-2)] rounded-xl border border-[var(--surface-border)] divide-y divide-[var(--surface-border)] mb-6">
            <div class="flex items-center gap-3 px-4 py-3">
              <svg class="w-4 h-4 text-[var(--text-3)] shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
              </svg>
              <span class="text-[var(--text-2)] text-sm">Read your username and display name</span>
            </div>
            <div class="flex items-center gap-3 px-4 py-3">
              <svg class="w-4 h-4 text-[var(--text-3)] shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
              <span class="text-[var(--text-2)] text-sm">Read your avatar URL</span>
            </div>
          </div>

          <p v-if="error" class="text-[#E8521A] text-xs font-medium mb-4 text-center">{{ error }}</p>

          <!-- Authorize / Cancel -->
          <div class="space-y-2.5">
            <button
              @click="handleAuthorize"
              :disabled="authorizing || step === 'done'"
              class="w-full bg-[#E8521A] text-white font-semibold py-3.5 px-4 rounded-xl hover:bg-[#D44818] shadow-lg shadow-[#E8521A]/20 hover:shadow-[#E8521A]/30 transition-all duration-200 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <span v-if="authorizing || step === 'done'" class="flex items-center justify-center gap-2">
                <svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
                </svg>
              </span>
              <span v-else>Authorize</span>
            </button>

            <button
              @click="handleCancel"
              :disabled="authorizing || step === 'done'"
              class="w-full bg-[var(--surface-2)] text-[var(--text-2)] font-semibold py-3.5 px-4 rounded-xl border border-[var(--surface-border)] hover:border-[#E8521A]/30 transition-all duration-200 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Cancel
            </button>
          </div>

          <p class="text-[var(--text-4)] text-xs text-center mt-6">
            You'll be redirected to <span class="text-[var(--text-3)]">{{ visitingDisplay }}</span> after authorization.
          </p>
        </div>

      </div>
    </div>
  </div>
</template>
