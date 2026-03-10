import { ref, onMounted, onUnmounted } from 'vue'
import { isElectron } from '@/utils/platform'

const WEB_POLL_INTERVAL = 5 * 60 * 1000 // 5 minutes
const ELECTRON_POLL_INTERVAL = 60 * 60 * 1000 // 1 hour

export function useVersionCheck() {
  const updateAvailable = ref(false)
  const electronUpdate = ref(null) // { version, status, percent }
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

  // Electron: auto-updater events are pushed from the main process.
  // A manual check can also be triggered via the IPC call.
  function handleUpdateStatus(data) {
    if (data.status === 'available') {
      updateAvailable.value = true
      electronUpdate.value = { version: data.version, status: 'downloading', percent: 0 }
    } else if (data.status === 'downloading') {
      if (electronUpdate.value) electronUpdate.value.percent = data.percent
    } else if (data.status === 'ready') {
      if (electronUpdate.value) {
        electronUpdate.value.status = 'ready'
        electronUpdate.value.percent = 100
      } else {
        updateAvailable.value = true
        electronUpdate.value = { version: data.version, status: 'ready', percent: 100 }
      }
    } else if (data.status === 'error') {
      // Silently ignore update errors — don't disrupt the user
    }
  }

  function checkElectronVersion() {
    window.electronAPI.checkForUpdates()
  }

  function handleVisibilityChange() {
    if (document.visibilityState === 'visible') {
      isElectron ? checkElectronVersion() : checkWebVersion()
    }
  }

  onMounted(() => {
    if (import.meta.env.DEV) return

    if (isElectron) {
      window.electronAPI.onUpdateStatus(handleUpdateStatus)
      // Main process auto-checks 3s after launch; also poll hourly
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
    if (isElectron) window.electronAPI?.removeUpdateListener()
  })

  const reloading = ref(false)

  function reload() {
    reloading.value = true
    window.location.reload()
  }

  function installUpdate() {
    window.electronAPI.installUpdate()
  }

  function dismissUpdate() {
    updateAvailable.value = false
    electronUpdate.value = null
  }

  return { updateAvailable, electronUpdate, reloading, reload, installUpdate, dismissUpdate }
}
