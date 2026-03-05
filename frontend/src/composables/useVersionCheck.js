import { ref, onMounted, onUnmounted } from 'vue'
import { isElectron } from '@/utils/platform'

const WEB_POLL_INTERVAL = 5 * 60 * 1000 // 5 minutes
const ELECTRON_POLL_INTERVAL = 60 * 60 * 1000 // 1 hour

export function useVersionCheck() {
  const updateAvailable = ref(false)
  const electronUpdate = ref(null) // { latestVersion, downloadUrl }
  let currentVersion = null
  let timer = null

  // Web: poll /version.json for hash-based version changes
  async function checkWebVersion() {
    try {
      const res = await fetch('/version.json?t=' + Date.now())
      if (!res.ok) return
      const { version } = await res.json()
      if (currentVersion && version !== currentVersion) {
        updateAvailable.value = true
      }
      if (!currentVersion) currentVersion = version
    } catch {
      // ignore fetch errors
    }
  }

  // Electron: check GitHub Releases for a newer version
  async function checkElectronVersion() {
    try {
      const result = await window.electronAPI.checkForUpdates()
      if (result) {
        updateAvailable.value = true
        electronUpdate.value = result
      }
    } catch {
      // ignore check errors
    }
  }

  function handleVisibilityChange() {
    if (document.visibilityState === 'visible') {
      isElectron ? checkElectronVersion() : checkWebVersion()
    }
  }

  onMounted(() => {
    if (import.meta.env.DEV) return

    if (isElectron) {
      checkElectronVersion()
      timer = setInterval(checkElectronVersion, ELECTRON_POLL_INTERVAL)
    } else {
      checkWebVersion()
      timer = setInterval(checkWebVersion, WEB_POLL_INTERVAL)
    }
    document.addEventListener('visibilitychange', handleVisibilityChange)
  })

  onUnmounted(() => {
    clearInterval(timer)
    document.removeEventListener('visibilitychange', handleVisibilityChange)
  })

  const reloading = ref(false)

  function reload() {
    reloading.value = true
    window.location.reload()
  }

  function openDownload() {
    if (electronUpdate.value?.downloadUrl) {
      window.electronAPI.openExternal(electronUpdate.value.downloadUrl)
    }
  }

  function dismissUpdate() {
    updateAvailable.value = false
    electronUpdate.value = null
  }

  return { updateAvailable, electronUpdate, reloading, reload, openDownload, dismissUpdate }
}
