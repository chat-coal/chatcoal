<script setup>
import { ref, computed, watch, onUnmounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useVoiceStore } from '@/stores/voice'
import { useToastStore } from '@/stores/toast'
import { useRouter } from 'vue-router'
import { API_URL } from '@/services/api.service'
import { getAvatarColor, getDefaultAvatarStyle, gifFirstFrame } from '@/utils/avatar'
import api from '@/services/api.service'
import { useEscapeClose } from '@/composables/useEscapeClose'
import { auth as firebaseAuth, EmailAuthProvider, reauthenticateWithCredential, updatePassword } from '@/services/firebase'

const emit = defineEmits(['close'])
const authStore = useAuthStore()
const voiceStore = useVoiceStore()
const toastStore = useToastStore()
const router = useRouter()

const isAnon = computed(() => authStore.dbUser?.is_anonymous)
const isRestricted = computed(() => authStore.dbUser?.is_anonymous || authStore.dbUser?.email_verified === false)

const displayName = ref(authStore.dbUser?.display_name || '')
const username = ref(authStore.dbUser?.username || '')
const status = ref(authStore.dbUser?.status === 'invisible' ? 'invisible' : 'online')
const avatarFile = ref(null)
const clearAvatar = ref(false)
const existingUrl = authStore.dbUser?.avatar_url || ''
const avatarPreview = ref(
  existingUrl ? (existingUrl.startsWith('http') ? existingUrl : `${API_URL}${existingUrl}`) : ''
)
const saving = ref(false)
const usernameError = ref('')
const showDeleteConfirm = ref(false)
const deleteConfirmText = ref('')
const deleting = ref(false)
const DELETE_PHRASE = 'delete my account'
const usernameAvailable = ref(null) // null=unchecked, true=available, false=taken
let checkTimer = null

// PTT key recording
const recordingPttKey = ref(false)
const isElectron = !!window.electronAPI?.isElectron
const isMac = window.electronAPI?.platform === 'darwin'
const requestingAccess = ref(false)

async function requestAccessibility() {
  requestingAccess.value = true
  try {
    const available = await window.electronAPI.requestAccessibility()
    voiceStore.globalPttAvailable = available
  } finally {
    requestingAccess.value = false
  }
}

// Poll accessibility while the banner is visible + re-check on window focus
let accessPollTimer = null
function checkAccessibility() {
  if (voiceStore.globalPttAvailable) {
    clearInterval(accessPollTimer)
    window.removeEventListener('focus', checkAccessibility)
    return
  }
  window.electronAPI.checkAccessibility().then((available) => {
    voiceStore.globalPttAvailable = available
    if (available) {
      clearInterval(accessPollTimer)
      window.removeEventListener('focus', checkAccessibility)
    }
  })
}
function startAccessPoll() {
  if (accessPollTimer) return
  checkAccessibility()
  accessPollTimer = setInterval(checkAccessibility, 5000)
  window.addEventListener('focus', checkAccessibility)
}
function stopAccessPoll() {
  clearInterval(accessPollTimer)
  accessPollTimer = null
  window.removeEventListener('focus', checkAccessibility)
}
if (isMac && !voiceStore.globalPttAvailable && voiceStore.inputMode === 'push_to_talk') {
  startAccessPoll()
}
if (isMac) {
  watch(() => [voiceStore.inputMode, voiceStore.globalPttAvailable], ([mode, available]) => {
    if (mode === 'push_to_talk' && !available) startAccessPoll()
    else stopAccessPoll()
  })
}
onUnmounted(() => {
  stopAccessPoll()
})

// Change password (email/password users only)
const isPasswordUser = computed(() => {
  const user = firebaseAuth.currentUser
  return user?.providerData?.some(p => p.providerId === 'password') ?? false
})
const currentPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const passwordSaving = ref(false)
const passwordError = ref('')
const passwordSuccess = ref(false)

async function handleChangePassword() {
  passwordError.value = ''
  passwordSuccess.value = false
  if (!newPassword.value || !confirmPassword.value) {
    passwordError.value = 'Please fill in all fields.'
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    passwordError.value = 'New passwords do not match.'
    return
  }
  if (newPassword.value.length < 6) {
    passwordError.value = 'Password must be at least 6 characters.'
    return
  }
  passwordSaving.value = true
  try {
    const user = firebaseAuth.currentUser
    const credential = EmailAuthProvider.credential(user.email, currentPassword.value)
    await reauthenticateWithCredential(user, credential)
    await updatePassword(user, newPassword.value)
    passwordSuccess.value = true
    currentPassword.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
  } catch (e) {
    const map = {
      'auth/wrong-password': 'Current password is incorrect.',
      'auth/invalid-credential': 'Current password is incorrect.',
      'auth/weak-password': 'Password must be at least 6 characters.',
      'auth/too-many-requests': 'Too many attempts. Try again later.',
      'auth/requires-recent-login': 'Please sign out and sign back in, then try again.',
    }
    passwordError.value = map[e.code] || 'Failed to change password. Please try again.'
  } finally {
    passwordSaving.value = false
  }
}

useEscapeClose(() => {
  if (recordingPttKey.value) return // PTT handler catches ESC itself
  if (showDeleteConfirm.value) { showDeleteConfirm.value = false; return }
  emit('close')
})

function pttKeyLabel(code) {
  if (code === 'Space') return 'Space'
  if (code.startsWith('Key')) return code.slice(3)
  if (code.startsWith('Digit')) return code.slice(5)
  return code.replace(/Left|Right/, '')
}

function startRecordingPttKey() {
  recordingPttKey.value = true
  function onKey(e) {
    e.preventDefault()
    e.stopPropagation()
    window.removeEventListener('keydown', onKey, true)
    recordingPttKey.value = false
    if (e.code === 'Escape') return
    voiceStore.setPttKey(e.code)
  }
  window.addEventListener('keydown', onKey, true)
}

const usernameChanged = computed(() => username.value !== (authStore.dbUser?.username || ''))
const usernameFormatValid = computed(() => {
  const t = username.value.trim()
  return /^[a-zA-Z0-9_]{2,32}$/.test(t) && !t.startsWith('_') && !t.endsWith('_')
})

watch(username, (val) => {
  usernameAvailable.value = null
  usernameError.value = ''
  clearTimeout(checkTimer)
  const trimmed = val.trim()
  if (!trimmed || trimmed === authStore.dbUser?.username) return
  if (!/^[a-zA-Z0-9_]{2,32}$/.test(trimmed) || trimmed.startsWith('_') || trimmed.endsWith('_')) {
    usernameError.value = 'Username must be 2–32 chars: letters, numbers, underscores'
    return
  }
  checkTimer = setTimeout(async () => {
    try {
      const { available } = await api.checkUsername(trimmed)
      usernameAvailable.value = available
      if (!available) usernameError.value = 'Username is already taken'
    } catch {
      usernameAvailable.value = null
    }
  }, 400)
})

onUnmounted(() => {
  clearTimeout(checkTimer)
  if (avatarPreview.value && avatarPreview.value.startsWith('blob:')) {
    URL.revokeObjectURL(avatarPreview.value)
  }
})

async function onFileSelect(e) {
  const raw = e.target.files[0]
  if (!raw) return
  const file = await gifFirstFrame(raw)
  avatarFile.value = file
  if (avatarPreview.value && avatarPreview.value.startsWith('blob:')) {
    URL.revokeObjectURL(avatarPreview.value)
  }
  avatarPreview.value = URL.createObjectURL(file)
  clearAvatar.value = false
}

function removeAvatar() {
  avatarFile.value = null
  avatarPreview.value = ''
  clearAvatar.value = true
}

function openDeleteConfirm() {
  deleteConfirmText.value = ''
  showDeleteConfirm.value = true
}

function closeDeleteConfirm() {
  showDeleteConfirm.value = false
  deleteConfirmText.value = ''
}

async function confirmDeleteAccount() {
  if (deleting.value || deleteConfirmText.value.toLowerCase() !== DELETE_PHRASE) return
  deleting.value = true
  try {
    await authStore.deleteAccount()
    router.push('/login')
  } catch (e) {
    const msg = e?.response?.data?.error || 'Failed to delete account'
    toastStore.add(msg)
  } finally {
    deleting.value = false
    showDeleteConfirm.value = false
  }
}

async function save() {
  if (saving.value) return
  const trimmedUsername = username.value.trim()
  if (!isAnon.value && usernameChanged.value) {
    if (!/^[a-zA-Z0-9_]{2,32}$/.test(trimmedUsername) || trimmedUsername.startsWith('_') || trimmedUsername.endsWith('_')) {
      usernameError.value = 'Username must be 2–32 chars: letters, numbers, underscores'
      return
    }
    if (usernameAvailable.value === false) {
      usernameError.value = 'Username is already taken'
      return
    }
  }
  saving.value = true
  try {
    await authStore.updateProfile({
      displayName: !isAnon.value ? (displayName.value.trim() || undefined) : undefined,
      username: (!isAnon.value && usernameChanged.value) ? trimmedUsername : undefined,
      avatarFile: avatarFile.value || undefined,
      clearAvatar: clearAvatar.value || undefined,
      status: status.value || undefined,
    })
    emit('close')
  } catch {
    toastStore.add('Failed to save profile')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <Teleport to="body">
    <div class="fixed inset-0 bg-[var(--backdrop)] backdrop-blur-sm flex items-center justify-center z-50" @click.self="emit('close')">
      <div class="bg-[var(--modal-bg)] rounded-2xl p-7 w-full max-w-md max-h-[90vh] overflow-y-auto shadow-2xl shadow-black/10 animate-scale-in border border-[var(--modal-border)]">
        <h2 class="font-display text-2xl font-bold text-[var(--text-1)] mb-6">Profile Settings</h2>

        <!-- Avatar upload -->
        <div class="flex flex-col items-center mb-6 gap-2">
          <label
            class="w-20 h-20 rounded-full bg-[var(--surface)] border-2 border-dashed border-[var(--surface-border)] flex items-center justify-center cursor-pointer hover:border-[#E8521A]/40 transition-colors duration-200 overflow-hidden relative group"
          >
            <img v-if="avatarPreview" :src="avatarPreview" class="w-full h-full object-cover" />
            <span
              v-else
              class="text-white text-2xl font-bold w-full h-full flex items-center justify-center rounded-full"
              :style="getDefaultAvatarStyle(authStore.dbUser?.id)"
            >
              {{ (authStore.dbUser?.display_name || '?')[0].toUpperCase() }}
            </span>
            <div class="absolute inset-0 bg-black/40 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity duration-200 rounded-full">
              <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z" />
                <circle cx="12" cy="13" r="3" />
              </svg>
            </div>
            <input type="file" accept="image/png,image/jpeg,image/gif,image/webp" class="hidden" @change="onFileSelect" />
          </label>
          <button
            v-if="avatarPreview && !isRestricted"
            type="button"
            @click="removeAvatar"
            class="text-xs text-[var(--text-4)] hover:text-red-400 transition-colors duration-150 cursor-pointer"
          >
            Remove avatar
          </button>
        </div>

        <!-- Username (non-anon only) -->
        <template v-if="!isAnon">
          <label class="text-[11px] font-bold text-[var(--text-4)] uppercase tracking-[0.1em]">Username</label>
          <div class="relative mt-2 mb-1">
            <span class="absolute left-3.5 top-1/2 -translate-y-1/2 text-[var(--text-3)] text-sm font-semibold select-none pointer-events-none">@</span>
            <input
              v-model="username"
              type="text"
              maxlength="32"
              autocomplete="off"
              autocorrect="off"
              autocapitalize="off"
              spellcheck="false"
              class="w-full bg-[var(--surface)] text-[var(--text-1)] pl-8 pr-3.5 py-2.5 rounded-xl placeholder-[var(--text-4)] text-sm border border-[var(--surface-border)]"
            />
            <div v-if="usernameChanged && username.trim()" class="absolute right-3.5 top-1/2 -translate-y-1/2">
              <svg v-if="usernameFormatValid && usernameAvailable === null" class="w-3.5 h-3.5 text-[var(--text-4)] animate-spin" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
              </svg>
              <svg v-else-if="usernameAvailable === true" class="w-3.5 h-3.5 text-[#D4782A]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
                <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
              </svg>
              <svg v-else class="w-3.5 h-3.5 text-[#E8521A]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
                <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </div>
          </div>
          <p v-if="usernameError" class="text-[#E8521A] text-xs font-medium px-0.5 mb-4">{{ usernameError }}</p>
          <div v-else class="mb-4"></div>
        </template>

        <!-- Display name -->
        <label class="text-[11px] font-bold text-[var(--text-4)] uppercase tracking-[0.1em]">Display Name</label>
        <input
          v-model="displayName"
          @keyup.enter="save"
          :disabled="isAnon"
          :class="[
            'w-full bg-[var(--surface)] text-[var(--text-1)] px-3.5 py-2.5 rounded-xl mt-2 mb-6 placeholder-[var(--text-4)] text-sm border border-[var(--surface-border)]',
            isAnon ? 'opacity-50 cursor-not-allowed' : ''
          ]"
        />

        <!-- Online status -->
        <label class="text-[11px] font-bold text-[var(--text-4)] uppercase tracking-[0.1em] mb-2 block">Online Status</label>
        <div class="flex gap-3 mb-6">
          <button
            @click="status = 'online'"
            :class="[
              'flex items-center gap-2 px-4 py-2.5 rounded-xl border text-sm font-medium cursor-pointer transition-all duration-200',
              status === 'online'
                ? 'border-green-500 bg-green-500/10 text-[var(--text-1)]'
                : 'border-[var(--surface-border)] text-[var(--text-3)] hover:border-[var(--text-4)]'
            ]"
          >
            <span class="w-2.5 h-2.5 rounded-full bg-green-500"></span>
            Online
          </button>
          <button
            @click="status = 'invisible'"
            :class="[
              'flex items-center gap-2 px-4 py-2.5 rounded-xl border text-sm font-medium cursor-pointer transition-all duration-200',
              status === 'invisible'
                ? 'border-[var(--offline)] bg-[var(--offline)]/10 text-[var(--text-1)]'
                : 'border-[var(--surface-border)] text-[var(--text-3)] hover:border-[var(--text-4)]'
            ]"
          >
            <span class="w-2.5 h-2.5 rounded-full bg-[var(--offline)]"></span>
            Invisible
          </button>
        </div>

        <!-- Voice settings (desktop only — PTT requires a physical keyboard) -->
        <div class="hidden md:block">
        <label class="text-[11px] font-bold text-[var(--text-4)] uppercase tracking-[0.1em] mb-2 block">Input Mode</label>
        <div class="flex gap-3 mb-4">
          <button
            @click="voiceStore.setInputMode('voice_activity')"
            :class="[
              'flex items-center gap-2 px-4 py-2.5 rounded-xl border text-sm font-medium cursor-pointer transition-all duration-200',
              voiceStore.inputMode === 'voice_activity'
                ? 'border-[#E8521A] bg-[#E8521A]/10 text-[var(--text-1)]'
                : 'border-[var(--surface-border)] text-[var(--text-3)] hover:border-[var(--text-4)]'
            ]"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 18.75a6 6 0 0 0 6-6v-1.5m-6 7.5a6 6 0 0 1-6-6v-1.5m6 7.5v3.75m-3.75 0h7.5M12 15.75a3 3 0 0 1-3-3V4.5a3 3 0 1 1 6 0v8.25a3 3 0 0 1-3 3Z" />
            </svg>
            Voice Activity
          </button>
          <button
            @click="voiceStore.setInputMode('push_to_talk')"
            :class="[
              'flex items-center gap-2 px-4 py-2.5 rounded-xl border text-sm font-medium cursor-pointer transition-all duration-200',
              voiceStore.inputMode === 'push_to_talk'
                ? 'border-[#E8521A] bg-[#E8521A]/10 text-[var(--text-1)]'
                : 'border-[var(--surface-border)] text-[var(--text-3)] hover:border-[var(--text-4)]'
            ]"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M15.75 5.25a3 3 0 0 1 3 3m3 0a6 6 0 0 1-7.029 5.912c-.563-.097-1.159.026-1.563.43L10.5 17.25H8.25v2.25H6v2.25H2.25v-2.818c0-.597.237-1.17.659-1.591l6.499-6.499c.404-.404.527-1 .43-1.563A6 6 0 1 1 21.75 8.25Z" />
            </svg>
            Push to Talk
          </button>
        </div>

        <!-- PTT key selector -->
        <div v-if="voiceStore.inputMode === 'push_to_talk'" class="mb-4">
          <label class="text-[11px] font-bold text-[var(--text-4)] uppercase tracking-[0.1em] mb-2 block">PTT Key</label>
          <button
            @click="startRecordingPttKey"
            class="px-4 py-2.5 rounded-xl border text-sm font-medium cursor-pointer transition-all duration-200"
            :class="recordingPttKey
              ? 'border-[#E8521A] bg-[#E8521A]/10 text-[#E8521A] animate-pulse'
              : 'border-[var(--surface-border)] text-[var(--text-1)] hover:border-[var(--text-4)]'"
          >
            {{ recordingPttKey ? 'Press any key…' : pttKeyLabel(voiceStore.pttKey) }}
          </button>
        </div>

        <!-- Accessibility permission banner (Electron + PTT mode + no global hook) -->
        <div v-if="voiceStore.inputMode === 'push_to_talk' && isMac && !voiceStore.globalPttAvailable" class="mb-6 rounded-xl border border-amber-500/20 bg-amber-500/5 px-4 py-3.5">
          <div class="flex items-start gap-3">
            <svg class="w-5 h-5 text-amber-400 shrink-0 mt-0.5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" />
            </svg>
            <div class="flex-1 min-w-0">
              <p class="text-sm font-semibold text-amber-300 mb-1">System-wide PTT unavailable</p>
              <p class="text-xs text-[var(--text-3)] leading-relaxed mb-2.5">
                Accessibility permission is required for the PTT key to work when chatcoal is in the background. Without it, PTT only works while the window is focused.
              </p>
              <p class="text-xs text-[var(--text-4)] leading-relaxed mb-3">
                Go to <span class="text-[var(--text-2)] font-medium">System Settings &rsaquo; Privacy &amp; Security &rsaquo; Accessibility</span> and enable chatcoal.
              </p>
              <button
                @click="requestAccessibility"
                :disabled="requestingAccess"
                class="text-xs font-semibold px-3 py-1.5 rounded-lg bg-amber-500/15 text-amber-300 hover:bg-amber-500/25 transition-colors duration-150 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {{ requestingAccess ? 'Requesting…' : 'Grant Accessibility Permission' }}
              </button>
            </div>
          </div>
        </div>
        <div v-else-if="voiceStore.inputMode !== 'push_to_talk'" class="mb-2"></div>
        <div v-else class="mb-6"></div>
        </div>

        <!-- Change Password (email/password users only) -->
        <template v-if="isPasswordUser">
          <label class="text-[11px] font-bold text-[var(--text-4)] uppercase tracking-[0.1em] mb-2 block">Change Password</label>
          <div class="space-y-2.5 mb-6">
            <input
              v-model="currentPassword"
              type="password"
              placeholder="Current password"
              autocomplete="current-password"
              class="w-full bg-[var(--surface)] text-[var(--text-1)] placeholder-[var(--text-4)] rounded-xl px-3.5 py-2.5 text-sm border border-[var(--surface-border)]"
            />
            <input
              v-model="newPassword"
              type="password"
              placeholder="New password"
              autocomplete="new-password"
              class="w-full bg-[var(--surface)] text-[var(--text-1)] placeholder-[var(--text-4)] rounded-xl px-3.5 py-2.5 text-sm border border-[var(--surface-border)]"
            />
            <input
              v-model="confirmPassword"
              type="password"
              placeholder="Confirm new password"
              autocomplete="new-password"
              class="w-full bg-[var(--surface)] text-[var(--text-1)] placeholder-[var(--text-4)] rounded-xl px-3.5 py-2.5 text-sm border border-[var(--surface-border)]"
            />
            <p v-if="passwordError" class="text-[#E8521A] text-xs font-medium px-0.5">{{ passwordError }}</p>
            <div v-if="passwordSuccess" class="bg-green-500/10 border border-green-500/20 rounded-xl px-3.5 py-2.5 text-xs text-green-400 font-medium">
              Password changed successfully.
            </div>
            <button
              type="button"
              @click="handleChangePassword"
              :disabled="passwordSaving || !currentPassword || !newPassword || !confirmPassword"
              class="bg-[var(--surface-3)] text-[var(--text-1)] text-sm font-semibold px-4 py-2.5 rounded-xl hover:bg-[var(--surface-border)] transition-all duration-150 cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
            >
              {{ passwordSaving ? 'Updating…' : 'Update password' }}
            </button>
          </div>
        </template>

        <!-- Actions -->
        <div class="flex justify-end gap-3">
          <button @click="emit('close')" class="text-[var(--text-3)] hover:text-[var(--text-1)] px-4 py-2.5 rounded-xl cursor-pointer font-medium transition-colors duration-150">
            Cancel
          </button>
          <button
            @click="save"
            :disabled="saving"
            class="bg-[#E8521A] text-white px-5 py-2.5 rounded-xl hover:bg-[#D44818] disabled:opacity-40 cursor-pointer font-semibold shadow-lg shadow-[#E8521A]/15 transition-all duration-200"
          >
            Save
          </button>
        </div>

        <!-- Danger Zone -->
        <div class="mt-8 pt-6 border-t border-[var(--modal-border)]">
          <p class="text-[11px] font-bold text-[var(--text-4)] uppercase tracking-[0.1em] mb-3">Danger Zone</p>
          <button
            @click="openDeleteConfirm"
            class="w-full flex items-center gap-3 px-4 py-3 rounded-xl border border-red-500/20 text-red-400 hover:bg-red-500/10 hover:border-red-500/40 cursor-pointer transition-all duration-150 text-sm font-medium"
          >
            <svg class="w-4 h-4 shrink-0" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
            </svg>
            Delete Account
          </button>
        </div>
      </div>
    </div>
  </Teleport>

  <!-- Delete account confirmation modal -->
  <Teleport to="body">
    <div
      v-if="showDeleteConfirm"
      class="fixed inset-0 bg-[var(--backdrop)] backdrop-blur-sm flex items-center justify-center z-[60]"
      @click.self="closeDeleteConfirm"
    >
      <div class="bg-[var(--modal-bg)] rounded-2xl p-7 w-full max-w-sm shadow-2xl shadow-black/10 animate-scale-in border border-[var(--modal-border)]">
        <div class="flex items-center gap-3 mb-2">
          <div class="w-9 h-9 rounded-xl bg-red-500/10 flex items-center justify-center shrink-0">
            <svg class="w-5 h-5 text-red-400" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
            </svg>
          </div>
          <h3 class="text-[var(--text-1)] font-bold text-lg">Delete account?</h3>
        </div>
        <p class="text-[var(--text-3)] text-sm mb-4 leading-relaxed">
          This is permanent and cannot be undone. Your messages will remain visible as "Deleted User".
        </p>
        <p class="text-[var(--text-3)] text-sm mb-2">
          Type <span class="font-mono font-semibold text-[var(--text-1)]">{{ DELETE_PHRASE }}</span> to confirm:
        </p>
        <input
          v-model="deleteConfirmText"
          @keyup.enter="confirmDeleteAccount"
          type="text"
          autocomplete="off"
          spellcheck="false"
          :placeholder="DELETE_PHRASE"
          class="w-full bg-[var(--surface)] text-[var(--text-1)] px-3.5 py-2.5 rounded-xl mb-5 placeholder-[var(--text-4)] text-sm border border-[var(--surface-border)] focus:outline-none focus:border-red-500/50"
        />
        <div class="flex gap-2.5">
          <button
            @click="closeDeleteConfirm"
            class="flex-1 py-2.5 px-4 rounded-xl text-sm font-semibold text-[var(--text-2)] bg-[var(--surface-2)] hover:bg-[var(--surface-3)] transition-colors duration-150 cursor-pointer"
          >
            Cancel
          </button>
          <button
            @click="confirmDeleteAccount"
            :disabled="deleting || deleteConfirmText.toLowerCase() !== DELETE_PHRASE"
            class="flex-1 py-2.5 px-4 rounded-xl text-sm font-semibold text-white bg-red-500 hover:bg-red-600 disabled:opacity-40 disabled:cursor-not-allowed shadow-lg shadow-red-500/15 transition-all duration-150 cursor-pointer"
          >
            {{ deleting ? 'Deleting…' : 'Delete account' }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>
