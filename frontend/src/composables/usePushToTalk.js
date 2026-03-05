import { onUnmounted, watch } from 'vue'
import { useVoiceStore } from '@/stores/voice'
import { setMuted } from '@/services/voice.service'

// Web KeyboardEvent.code → uiohook keycode mapping
const CODE_TO_UIOHOOK = {
  'KeyA': 0x1E, 'KeyB': 0x30, 'KeyC': 0x2E, 'KeyD': 0x20,
  'KeyE': 0x12, 'KeyF': 0x21, 'KeyG': 0x22, 'KeyH': 0x23,
  'KeyI': 0x17, 'KeyJ': 0x24, 'KeyK': 0x25, 'KeyL': 0x26,
  'KeyM': 0x32, 'KeyN': 0x31, 'KeyO': 0x18, 'KeyP': 0x19,
  'KeyQ': 0x10, 'KeyR': 0x13, 'KeyS': 0x1F, 'KeyT': 0x14,
  'KeyU': 0x16, 'KeyV': 0x2F, 'KeyW': 0x11, 'KeyX': 0x2D,
  'KeyY': 0x15, 'KeyZ': 0x2C,
  'Digit0': 0x0B, 'Digit1': 0x02, 'Digit2': 0x03, 'Digit3': 0x04,
  'Digit4': 0x05, 'Digit5': 0x06, 'Digit6': 0x07, 'Digit7': 0x08,
  'Digit8': 0x09, 'Digit9': 0x0A,
  'Space': 0x39, 'Enter': 0x1C, 'Tab': 0x0F,
  'Backspace': 0x0E, 'Escape': 0x01, 'CapsLock': 0x3A,
  'ShiftLeft': 0x2A, 'ShiftRight': 0x36,
  'ControlLeft': 0x1D, 'ControlRight': 0x0E1D,
  'AltLeft': 0x38, 'AltRight': 0x0E38,
  'MetaLeft': 0x0E5B, 'MetaRight': 0x0E5C,
  'F1': 0x3B, 'F2': 0x3C, 'F3': 0x3D, 'F4': 0x3E,
  'F5': 0x3F, 'F6': 0x40, 'F7': 0x41, 'F8': 0x42,
  'F9': 0x43, 'F10': 0x44, 'F11': 0x57, 'F12': 0x58,
  'Semicolon': 0x27, 'Equal': 0x0D, 'Comma': 0x33,
  'Minus': 0x0C, 'Period': 0x34, 'Slash': 0x35,
  'Backquote': 0x29, 'BracketLeft': 0x1A, 'Backslash': 0x2B,
  'BracketRight': 0x1B, 'Quote': 0x28,
}

export function usePushToTalk() {
  const voiceStore = useVoiceStore()

  // When true, the OS-level global hook handles PTT and browser
  // listeners are skipped (including the blur→mute fallback).
  let globalPttActive = false

  // --- Browser PTT (fallback, works only when window is focused) ---

  function isTyping() {
    const tag = document.activeElement?.tagName
    if (tag === 'INPUT' || tag === 'TEXTAREA') return true
    if (document.activeElement?.isContentEditable) return true
    return false
  }

  function onKeyDown(e) {
    if (globalPttActive) return
    if (voiceStore.inputMode !== 'push_to_talk') return
    if (!voiceStore.currentVoiceChannelId) return
    if (e.repeat) return
    if (e.code !== voiceStore.pttKey) return
    if (isTyping()) return

    if (e.code === 'Space') e.preventDefault()

    voiceStore.pttActive = true
    voiceStore.isMuted = false
    setMuted(false)
  }

  function onKeyUp(e) {
    if (globalPttActive) return
    if (voiceStore.inputMode !== 'push_to_talk') return
    if (e.code !== voiceStore.pttKey) return

    voiceStore.pttActive = false
    voiceStore.isMuted = true
    setMuted(true)
  }

  function onBlur() {
    if (globalPttActive) return
    if (voiceStore.inputMode !== 'push_to_talk') return
    if (!voiceStore.pttActive) return

    voiceStore.pttActive = false
    voiceStore.isMuted = true
    setMuted(true)
  }

  window.addEventListener('keydown', onKeyDown)
  window.addEventListener('keyup', onKeyUp)
  window.addEventListener('blur', onBlur)

  // --- Electron global PTT (system-wide, works across all apps) ---

  if (window.electronAPI?.isElectron) {
    // Sends config to main process; the first call with enabled=true
    // triggers the accessibility permission prompt and starts the hook.
    // Returns whether the global hook is available.
    async function syncPttConfig() {
      const enabled = voiceStore.inputMode === 'push_to_talk' && !!voiceStore.currentVoiceChannelId
      const keyCode = CODE_TO_UIOHOOK[voiceStore.pttKey] ?? null
      const available = await window.electronAPI.setPttConfig({ keyCode, enabled })
      globalPttActive = available
      voiceStore.globalPttAvailable = available
    }

    watch(
      () => [voiceStore.pttKey, voiceStore.inputMode, voiceStore.currentVoiceChannelId],
      syncPttConfig,
      { immediate: true },
    )

    window.electronAPI.onPttState((active) => {
      if (!globalPttActive) return
      if (voiceStore.inputMode !== 'push_to_talk') return
      if (!voiceStore.currentVoiceChannelId) return
      voiceStore.pttActive = active
      voiceStore.isMuted = !active
      setMuted(!active)
    })

    // If macOS accessibility permission is revoked while the hook is
    // running, the main process stops the hook and notifies us so we
    // can fall back to browser-level PTT seamlessly.
    window.electronAPI.onGlobalPttRevoked(() => {
      globalPttActive = false
      voiceStore.globalPttAvailable = false
      // Release PTT if it was held at the moment of revocation
      if (voiceStore.pttActive) {
        voiceStore.pttActive = false
        voiceStore.isMuted = true
        setMuted(true)
      }
    })

    // If accessibility permission is granted mid-call, the main process
    // auto-starts the global hook and notifies us to switch over.
    window.electronAPI.onGlobalPttAvailable(() => {
      globalPttActive = true
      voiceStore.globalPttAvailable = true
    })
  }

  // --- Cleanup (registered synchronously) ---

  onUnmounted(() => {
    window.removeEventListener('keydown', onKeyDown)
    window.removeEventListener('keyup', onKeyUp)
    window.removeEventListener('blur', onBlur)
    if (window.electronAPI?.isElectron) {
      window.electronAPI.removePttListener()
      window.electronAPI.setPttConfig({ keyCode: null, enabled: false })
    }
  })
}
