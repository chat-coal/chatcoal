import { API_URL } from '@/services/api.service'
import logoSvg from '@/assets/logo.svg'

const colors = ['#E8521A', '#D4782A', '#E8893A', '#C85A1A', '#B84818']

export function getAvatarColor(id) {
  const str = String(id || '0')
  let hash = 0
  for (let i = 0; i < str.length; i++) {
    hash = str.charCodeAt(i) + ((hash << 5) - hash)
  }
  return colors[Math.abs(hash) % colors.length]
}

export function getDefaultAvatarStyle(id) {
  return {
    backgroundColor: getAvatarColor(id),
    backgroundImage: `url(${logoSvg})`,
    backgroundSize: '70%',
    backgroundPosition: 'center',
    backgroundRepeat: 'no-repeat',
  }
}

// If the file is an animated GIF, return a new File containing only the first
// frame (as PNG). For all other image types the original file is returned as-is.
export function gifFirstFrame(file) {
  if (file.type !== 'image/gif') return Promise.resolve(file)
  return new Promise((resolve) => {
    const img = new Image()
    const url = URL.createObjectURL(file)
    img.onload = () => {
      const canvas = document.createElement('canvas')
      canvas.width = img.naturalWidth
      canvas.height = img.naturalHeight
      canvas.getContext('2d').drawImage(img, 0, 0)
      URL.revokeObjectURL(url)
      canvas.toBlob((blob) => {
        resolve(new File([blob], file.name.replace(/\.gif$/i, '.png'), { type: 'image/png' }))
      }, 'image/png')
    }
    img.onerror = () => { URL.revokeObjectURL(url); resolve(file) }
    img.src = url
  })
}

// Resolve a stored avatar/file URL (which may be a relative /api/files/... path)
// to an absolute URL pointing at the backend.
export function resolveFileUrl(url) {
  if (!url) return ''
  if (url.startsWith('http')) return url
  return `${API_URL}${url}`
}

// Return a CSS-safe url() value for use in inline backgroundImage styles.
// Escapes characters that could break out of the CSS url() context.
export function cssBackgroundUrl(url) {
  if (!url) return 'none'
  const safe = url.replace(/[()'"\\]/g, (ch) => '\\' + ch)
  return `url(${safe})`
}
