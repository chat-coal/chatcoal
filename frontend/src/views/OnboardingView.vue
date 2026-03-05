<script setup>
import { ref, computed, watch } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'
import api from '@/services/api.service'
import logoSvg from '@/assets/logo.svg'

const auth = useAuthStore()
const router = useRouter()

const isAnon = computed(() => auth.user?.isAnonymous === true)

// Anonymous users keep their random display name; Google users start blank
const anonSuggestedName = isAnon.value ? (auth.dbUser?.display_name || '') : ''
const username = ref('')
const displayName = ref(anonSuggestedName)
const loading = ref(false)
const error = ref('')
const usernameAvailable = ref(null) // null = unchecked, true = available, false = taken
let checkTimer = null

const trimmedUsername = computed(() => username.value.trim())
const trimmedDisplayName = computed(() => displayName.value.trim())

const usernameValidationMessage = computed(() => {
  if (!trimmedUsername.value) return ''
  if (trimmedUsername.value.length < 2) return 'Must be at least 2 characters'
  if (trimmedUsername.value.length > 32) return 'Must be 32 characters or fewer'
  if (!/^[a-zA-Z0-9_]+$/.test(trimmedUsername.value)) return 'Only letters, numbers, and underscores'
  if (trimmedUsername.value.startsWith('_') || trimmedUsername.value.endsWith('_')) return 'Cannot start or end with underscore'
  return ''
})

const displayNameValidationMessage = computed(() => {
  if (!trimmedDisplayName.value) return ''
  if (trimmedDisplayName.value.length < 1) return 'Cannot be empty'
  if (trimmedDisplayName.value.length > 50) return 'Must be 50 characters or fewer'
  return ''
})

const isUsernameFormatValid = computed(() => trimmedUsername.value.length >= 2 && !usernameValidationMessage.value)
const isDisplayNameValid = computed(() => trimmedDisplayName.value.length >= 1 && !displayNameValidationMessage.value)

const isValid = computed(() => {
  if (isAnon.value) return isDisplayNameValid.value
  return isUsernameFormatValid.value && isDisplayNameValid.value && usernameAvailable.value === true
})

// Check username availability with debounce (non-anon only)
watch(trimmedUsername, (val) => {
  if (isAnon.value) return
  usernameAvailable.value = null
  clearTimeout(checkTimer)
  if (!val || usernameValidationMessage.value) return
  checkTimer = setTimeout(async () => {
    try {
      const { available } = await api.checkUsername(val)
      usernameAvailable.value = available
    } catch {
      usernameAvailable.value = null
    }
  }, 400)
})

async function handleContinue() {
  if (!isValid.value || loading.value) return
  error.value = ''
  loading.value = true
  try {
    if (isAnon.value) {
      await auth.updateProfile({ displayName: trimmedDisplayName.value })
    } else {
      await auth.updateProfile({ displayName: trimmedDisplayName.value, username: trimmedUsername.value })
    }
    router.push('/channels/@me')
  } catch {
    error.value = 'Could not save. Please try again.'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex bg-[var(--surface)]">
    <!-- Left: Branding panel -->
    <div class="hidden lg:flex lg:w-[45%] relative overflow-hidden bg-[var(--login-bg)] items-center justify-center noise-texture">
      <div class="absolute top-[-15%] right-[-5%] w-[450px] h-[450px] rounded-full bg-[#E8521A] animate-ember-pulse blur-[100px]"></div>
      <div class="absolute bottom-[-15%] left-[-8%] w-[350px] h-[350px] rounded-full bg-[#D4782A] animate-ember-pulse blur-[90px]" style="animation-delay: 2s;"></div>
      <div class="absolute top-[50%] left-[20%] w-[250px] h-[250px] rounded-full bg-[#E8521A] animate-ember-pulse blur-[80px]" style="animation-delay: 1s; opacity: 0.07;"></div>

      <div class="relative z-10 text-center px-16 animate-fade-in-up">
        <img :src="logoSvg" alt="chatcoal" class="w-24 h-24 mx-auto mb-6 drop-shadow-[0_0_30px_rgba(232,82,26,0.4)]" />
        <h1 class="font-display text-7xl font-bold text-[var(--sb-text)] mb-5 tracking-tight">
          chatcoal
        </h1>
        <p class="text-[var(--sb-text-2)] text-lg max-w-sm mx-auto leading-relaxed">
          One last step. Make it yours.
        </p>

        <!-- Step indicator: step 2 of 2 -->
        <div class="flex items-center justify-center gap-2 mt-10">
          <div class="w-6 h-1.5 rounded-full bg-[#E8521A] opacity-40"></div>
          <div class="w-6 h-1.5 rounded-full bg-[#E8521A]"></div>
        </div>
      </div>
    </div>

    <!-- Right: Form -->
    <div class="flex-1 bg-[var(--login-form-bg)] flex items-center justify-center px-6">
      <div class="w-full max-w-sm animate-fade-in-up" style="animation-delay: 100ms;">
        <!-- Mobile branding -->
        <div class="lg:hidden text-center mb-10">
          <img :src="logoSvg" alt="chatcoal" class="w-16 h-16 mx-auto mb-4" />
          <h1 class="font-display text-5xl font-bold text-[var(--text-1)] mb-2 tracking-tight">
            chatcoal
          </h1>
        </div>

        <!-- Step indicator (mobile) -->
        <div class="lg:hidden flex items-center gap-2 mb-8">
          <div class="w-6 h-1.5 rounded-full bg-[#E8521A] opacity-40"></div>
          <div class="w-6 h-1.5 rounded-full bg-[#E8521A]"></div>
          <span class="ml-1 text-[var(--text-4)] text-xs font-semibold uppercase tracking-wider">Step 2 of 2</span>
        </div>

        <h2 class="text-2xl font-bold text-[var(--text-1)] mb-1.5">
          {{ isAnon ? 'Choose a display name' : 'Set up your profile' }}
        </h2>
        <p class="text-[var(--text-3)] text-sm mb-8">
          {{ isAnon ? 'This is how others will see you.' : 'Choose your username and display name.' }}
        </p>

        <form @submit.prevent="handleContinue" class="space-y-5">
          <!-- Anonymous: single display name input -->
          <template v-if="isAnon">
            <div>
              <input
                v-model="displayName"
                type="text"
                placeholder="Display name"
                maxlength="50"
                autocomplete="off"
                autocorrect="off"
                autocapitalize="off"
                spellcheck="false"
                autofocus
                class="w-full bg-[var(--surface-2)] text-[var(--text-1)] placeholder-[var(--text-4)] rounded-xl py-3.5 px-4 text-sm border border-transparent focus:outline-none focus:border-[#E8521A]/50 transition-all duration-150"
              />
              <div class="flex items-center justify-between px-0.5 mt-2">
                <p class="text-xs" :class="displayNameValidationMessage ? 'text-[#E8521A] font-medium' : 'text-[var(--text-4)]'">
                  <template v-if="displayNameValidationMessage">{{ displayNameValidationMessage }}</template>
                  <template v-else>Up to 50 characters</template>
                </p>
              </div>
            </div>
          </template>

          <!-- Non-anonymous: username + display name -->
          <template v-else>
            <!-- Username -->
            <div>
              <label class="block text-xs font-semibold text-[var(--text-3)] uppercase tracking-wider mb-2">Username</label>
              <div class="relative">
                <span class="absolute left-4 top-1/2 -translate-y-1/2 text-[var(--text-3)] text-sm font-semibold select-none pointer-events-none">@</span>
                <input
                  v-model="username"
                  type="text"
                  placeholder="username"
                  maxlength="32"
                  autocomplete="off"
                  autocorrect="off"
                  autocapitalize="off"
                  spellcheck="false"
                  autofocus
                  class="w-full bg-[var(--surface-2)] text-[var(--text-1)] placeholder-[var(--text-4)] rounded-xl py-3.5 pl-8 pr-10 text-sm border border-transparent focus:outline-none focus:border-[#E8521A]/50 transition-all duration-150"
                />
                <!-- Validity icon -->
                <div v-if="trimmedUsername" class="absolute right-4 top-1/2 -translate-y-1/2">
                  <svg v-if="isUsernameFormatValid && usernameAvailable === null" class="w-4 h-4 text-[var(--text-4)] animate-spin" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
                  </svg>
                  <svg v-else-if="isUsernameFormatValid && usernameAvailable === true" class="w-4 h-4 text-[#D4782A]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                  </svg>
                  <svg v-else class="w-4 h-4 text-[#E8521A]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </div>
              </div>
              <div class="flex items-center justify-between px-0.5 mt-2">
                <p class="text-xs" :class="usernameValidationMessage || usernameAvailable === false ? 'text-[#E8521A] font-medium' : 'text-[var(--text-4)]'">
                  <template v-if="usernameValidationMessage">{{ usernameValidationMessage }}</template>
                  <template v-else-if="usernameAvailable === false">Username is already taken</template>
                  <template v-else>2–32 chars · letters, numbers, underscores</template>
                </p>
                <p class="text-xs text-[var(--text-4)] tabular-nums">{{ trimmedUsername.length }}/32</p>
              </div>
            </div>

            <!-- Display Name -->
            <div>
              <label class="block text-xs font-semibold text-[var(--text-3)] uppercase tracking-wider mb-2">Display Name</label>
              <input
                v-model="displayName"
                type="text"
                placeholder="How others will see you"
                maxlength="50"
                autocomplete="off"
                autocorrect="off"
                autocapitalize="off"
                spellcheck="false"
                class="w-full bg-[var(--surface-2)] text-[var(--text-1)] placeholder-[var(--text-4)] rounded-xl py-3.5 px-4 text-sm border border-transparent focus:outline-none focus:border-[#E8521A]/50 transition-all duration-150"
              />
              <div class="flex items-center justify-between px-0.5 mt-2">
                <p class="text-xs" :class="displayNameValidationMessage ? 'text-[#E8521A] font-medium' : 'text-[var(--text-4)]'">
                  <template v-if="displayNameValidationMessage">{{ displayNameValidationMessage }}</template>
                  <template v-else>Up to 50 characters</template>
                </p>
              </div>
            </div>
          </template>

          <p v-if="error" class="text-[#E8521A] text-xs font-medium px-0.5">{{ error }}</p>

          <button
            type="submit"
            :disabled="!isValid || loading"
            class="w-full mt-2 bg-[#E8521A] text-white font-semibold py-3.5 px-4 rounded-xl hover:bg-[#D44818] shadow-lg shadow-[#E8521A]/15 hover:shadow-[#E8521A]/25 transition-all duration-200 cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed disabled:shadow-none"
          >
            <span v-if="loading" class="flex items-center justify-center gap-2">
              <svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
              </svg>
              Setting up...
            </span>
            <span v-else>Let's go &rarr;</span>
          </button>
        </form>

        <p class="text-[var(--text-4)] text-xs text-center mt-8">
          You can always change this in your profile settings.
        </p>
      </div>
    </div>
  </div>
</template>
