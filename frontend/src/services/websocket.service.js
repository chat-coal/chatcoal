import { auth } from './firebase'

const WS_URL = import.meta.env.VITE_WS_URL || 'ws://localhost:3000'

let socket = null
let reconnectTimer = null
let currentUserID = null
let reconnectAttempts = 0
const MAX_RECONNECT_DELAY = 30000
const listeners = new Map()

export async function connect(userID) {
  disconnect()
  currentUserID = userID

  const user = auth.currentUser
  if (!user) return

  const token = await user.getIdToken()
  socket = new WebSocket(`${WS_URL}/ws`)

  socket.onopen = () => {
    reconnectAttempts = 0
    socket.send(JSON.stringify({ type: 'auth', token }))
    emit('connected', {})
  }

  socket.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data)
      emit(msg.type, msg.data, msg.server_id)
    } catch {
      // ignore malformed messages
    }
  }

  socket.onclose = () => {
    emit('disconnected', {})
    // Only reconnect if user is still logged in
    if (!currentUserID) return
    const user = auth.currentUser
    const fedToken = sessionStorage.getItem('fed_token')
    if (!user && !fedToken) return
    const delay = Math.min(1000 * 2 ** reconnectAttempts, MAX_RECONNECT_DELAY)
    reconnectAttempts++
    reconnectTimer = setTimeout(() => {
      if (currentUserID) {
        connect(currentUserID)
      }
    }, delay)
  }

  socket.onerror = () => {
    socket?.close()
  }
}

export function disconnect() {
  clearTimeout(reconnectTimer)
  currentUserID = null
  if (socket) {
    socket.onclose = null
    socket.close()
    socket = null
  }
}

export function send(data) {
  if (socket?.readyState === WebSocket.OPEN) {
    socket.send(JSON.stringify(data))
  }
}

export function sendTyping(channelId, serverId) {
  send({ type: 'typing', channel_id: channelId, server_id: serverId })
}

export function sendDMTyping(dmChannelId, targetUserId) {
  send({ type: 'dm_typing', dm_channel_id: dmChannelId, target_user_id: targetUserId })
}

export function subscribe(serverId) {
  send({ type: 'subscribe', server_id: serverId })
}

export function unsubscribe(serverId) {
  send({ type: 'unsubscribe', server_id: serverId })
}

export function on(event, callback) {
  if (!listeners.has(event)) {
    listeners.set(event, new Set())
  }
  listeners.get(event).add(callback)
  return () => listeners.get(event)?.delete(callback)
}

export function off(event, callback) {
  listeners.get(event)?.delete(callback)
}

function emit(event, data, serverId) {
  listeners.get(event)?.forEach((cb) => cb(data, serverId))
}
