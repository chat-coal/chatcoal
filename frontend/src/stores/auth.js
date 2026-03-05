import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { onAuthStateChanged } from 'firebase/auth'
import {
  auth,
  googleProvider,
  appleProvider,
  signInAnonymously,
  signInWithPopup,
  signInWithEmailAndPassword,
  createUserWithEmailAndPassword,
  sendEmailVerification,
  linkWithPopup,
  linkWithCredential,
  EmailAuthProvider,
} from '@/services/firebase'
import api from '@/services/api.service'
import { isElectron } from '@/utils/platform'

export const useAuthStore = defineStore('auth', () => {
  const user = ref(null)
  const dbUser = ref(null)
  const loading = ref(true)
  const federationToken = ref(localStorage.getItem('fed_token') || null)
  const isFederated = computed(() => !!federationToken.value)

  function init() {
    return new Promise((resolve) => {
      // If we have a federation token, hydrate dbUser from /api/auth/me.
      if (federationToken.value) {
        api.getMe()
          .then((u) => { dbUser.value = u })
          .catch(() => {
            // Expired or invalid — clear it.
            localStorage.removeItem('fed_token')
            federationToken.value = null
          })
          .finally(() => {
            loading.value = false
            resolve()
          })
        return
      }

      onAuthStateChanged(auth, async (firebaseUser) => {
        if (!firebaseUser) {
          loading.value = false
          resolve()
          return
        }

        user.value = firebaseUser
        try {
          dbUser.value = await api.login()
        } catch {
          dbUser.value = null
        }
        loading.value = false
        resolve()
      })
    })
  }

  async function _fetchDbUser(firebaseUser) {
    user.value = firebaseUser
    try {
      dbUser.value = await api.login()
    } catch {
      dbUser.value = null
    }
  }

  async function loginWithGoogle() {
    const result = await signInWithPopup(auth, googleProvider)
    await _fetchDbUser(result.user)
  }

  async function loginWithApple() {
    const result = await signInWithPopup(auth, appleProvider)
    await _fetchDbUser(result.user)
  }

  async function loginWithEmail(email, password) {
    const result = await signInWithEmailAndPassword(auth, email, password)
    await _fetchDbUser(result.user)
  }

  async function registerWithEmail(email, password) {
    const result = await createUserWithEmailAndPassword(auth, email, password)
    await sendEmailVerification(result.user)
    await _fetchDbUser(result.user)
  }

  async function loginAnonymously() {
    if (auth.currentUser?.isAnonymous) {
      if (!dbUser.value) {
        await _fetchDbUser(auth.currentUser)
      }
      return
    }
    const result = await signInAnonymously(auth)
    await _fetchDbUser(result.user)
  }

  async function resendVerificationEmail() {
    if (auth.currentUser) {
      await sendEmailVerification(auth.currentUser)
    }
  }

  async function refreshEmailVerification() {
    if (!auth.currentUser) return
    await auth.currentUser.reload()
    await auth.currentUser.getIdToken(true)
    dbUser.value = await api.login()
  }

  // Initiates federated login by calling /api/federation/begin and
  // redirecting the browser to the home instance's authorize page.
  async function loginWithFederation(federatedId) {
    const { data } = await api.beginFederation(federatedId)

    if (isElectron) {
      // In Electron, open auth in a managed child window and get the token via IPC
      const token = await window.electronAPI.federationAuth(
        data.auth_url,
        window.location.origin,
      )
      await loginWithFederationCallback(token)
      return
    }

    window.location.href = data.auth_url
  }

  // Called by FederationCallbackView after a successful /api/federation/verify.
  async function loginWithFederationCallback(token) {
    const { data } = await api.verifyFederation(token)
    localStorage.setItem('fed_token', data.session_token)
    federationToken.value = data.session_token
    dbUser.value = data.user
  }

  async function updateProfile(data) {
    const updated = await api.updateProfile(data)
    dbUser.value = updated
    return updated
  }

  async function logout() {
    // Clear federation session if active.
    if (federationToken.value) {
      localStorage.removeItem('fed_token')
      federationToken.value = null
      dbUser.value = null
      return
    }
    user.value = null
    dbUser.value = null
    await auth.signOut()
  }

  async function linkWithGoogle() {
    await linkWithPopup(auth.currentUser, googleProvider)
    user.value = auth.currentUser
    dbUser.value = await api.login()
  }

  async function linkWithEmail(emailAddr, password) {
    const credential = EmailAuthProvider.credential(emailAddr, password)
    await linkWithCredential(auth.currentUser, credential)
    user.value = auth.currentUser
    dbUser.value = await api.login()
  }

  async function deleteAccount() {
    if (federationToken.value) {
      // Federated user — no Firebase session to sign out of.
      await api.deleteAccount()
      localStorage.removeItem('fed_token')
      federationToken.value = null
      user.value = null
      dbUser.value = null
      return
    }
    // Anonymous users can't re-authenticate, so force-refresh the token so
    // the backend's freshness check passes.
    if (auth.currentUser?.isAnonymous) {
      await auth.currentUser.getIdToken(true)
    }
    await api.deleteAccount()
    user.value = null
    dbUser.value = null
    await auth.signOut()
  }

  return {
    user,
    dbUser,
    loading,
    federationToken,
    isFederated,
    init,
    loginWithGoogle,
    loginWithApple,
    loginWithEmail,
    registerWithEmail,
    resendVerificationEmail,
    refreshEmailVerification,
    loginAnonymously,
    loginWithFederation,
    loginWithFederationCallback,
    updateProfile,
    logout,
    deleteAccount,
    linkWithGoogle,
    linkWithEmail,
  }
})
