const urlRegex = /https?:\/\/[^\s<>"'`)\]]+/g
const mentionRegex = /<@(\d+)>/g

/**
 * Splits message content into an array of parts with types: 'text', 'link', or 'mention'.
 * Mention parts include a userId field.
 * @param {string} content
 * @param {Map<string, object>} [memberMap] - optional Map of userId → user object for display name resolution
 */
export function linkify(content, memberMap) {
  if (!content) return [{ type: 'text', value: '' }]

  // Collect all matches (URLs and mentions) sorted by position
  const tokens = []

  urlRegex.lastIndex = 0
  let match
  while ((match = urlRegex.exec(content)) !== null) {
    let url = match[0].replace(/[.,;:!?]+$/, '')
    tokens.push({ type: 'link', value: url, index: match.index, length: url.length })
    urlRegex.lastIndex = match.index + url.length
  }

  mentionRegex.lastIndex = 0
  while ((match = mentionRegex.exec(content)) !== null) {
    const userId = match[1]
    const user = memberMap?.get(userId)
    const displayName = user?.display_name || user?.user?.display_name || userId
    tokens.push({
      type: 'mention',
      value: `@${displayName}`,
      userId,
      index: match.index,
      length: match[0].length,
    })
  }

  // Sort by position
  tokens.sort((a, b) => a.index - b.index)

  if (tokens.length === 0) {
    return [{ type: 'text', value: content }]
  }

  const parts = []
  let lastIndex = 0

  for (const token of tokens) {
    // Skip overlapping tokens
    if (token.index < lastIndex) continue

    if (token.index > lastIndex) {
      parts.push({ type: 'text', value: content.slice(lastIndex, token.index) })
    }

    if (token.type === 'mention') {
      parts.push({ type: 'mention', value: token.value, userId: token.userId })
    } else {
      parts.push({ type: 'link', value: token.value })
    }

    lastIndex = token.index + token.length
  }

  if (lastIndex < content.length) {
    parts.push({ type: 'text', value: content.slice(lastIndex) })
  }

  return parts.length ? parts : [{ type: 'text', value: content }]
}
