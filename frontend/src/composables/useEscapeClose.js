import { onMounted, onUnmounted } from 'vue'

export function useEscapeClose(callback) {
  function onKeydown(e) {
    if (e.key === 'Escape') {
      e.stopPropagation()
      callback()
    }
  }
  onMounted(() => window.addEventListener('keydown', onKeydown))
  onUnmounted(() => window.removeEventListener('keydown', onKeydown))
}
