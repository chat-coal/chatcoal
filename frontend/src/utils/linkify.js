const urlRegex = /https?:\/\/[^\s<>"'`)\]]+/g

/**
 * Splits message content into an array of { type: 'text' | 'link', value: string } parts.
 */
export function linkify(content) {
  if (!content) return [{ type: 'text', value: '' }]

  const parts = []
  let lastIndex = 0
  let match

  urlRegex.lastIndex = 0
  while ((match = urlRegex.exec(content)) !== null) {
    // Trim trailing punctuation that's likely not part of the URL
    let url = match[0].replace(/[.,;:!?]+$/, '')

    if (match.index > lastIndex) {
      parts.push({ type: 'text', value: content.slice(lastIndex, match.index) })
    }
    parts.push({ type: 'link', value: url })
    lastIndex = match.index + url.length
    // Reset regex index since we may have trimmed characters
    urlRegex.lastIndex = lastIndex
  }

  if (lastIndex < content.length) {
    parts.push({ type: 'text', value: content.slice(lastIndex) })
  }

  return parts.length ? parts : [{ type: 'text', value: content }]
}
