import { Room, RoomEvent } from 'livekit-client'
import { send, on } from './websocket.service'
import { useVoiceStore } from '@/stores/voice'
import { useAuthStore } from '@/stores/auth'
import api from './api.service'

const LIVEKIT_URL = import.meta.env.VITE_LIVEKIT_URL
const ICE_SERVERS = [{ urls: 'stun:stun.l.google.com:19302' }]

// Current transport: 'livekit' | 'p2p' | null
let mode = null
let currentServerId = null

// --- LiveKit state ---
let room = null

// --- P2P state ---
let localStream = null
const peerConnections = new Map()
const remoteAudios = new Map()
let unsubOffer = null
let unsubAnswer = null
let unsubIce = null
let unsubVoiceState = null

export async function joinChannel(channelId, serverId) {
  const voiceStore = useVoiceStore()
  voiceStore.isJoining = true

  // Try LiveKit first if URL is configured
  if (LIVEKIT_URL) {
    try {
      const tokenData = await api.getVoiceToken(serverId, channelId)

      room = new Room()

      room.on(RoomEvent.TrackSubscribed, (track, publication, participant) => {
        if (track.kind === 'audio') {
          const el = track.attach()
          el.id = `lk-audio-${participant.identity}`
          document.body.appendChild(el)
          if (voiceStore.isDeafened) el.muted = true
        }
      })

      room.on(RoomEvent.TrackUnsubscribed, (track) => {
        track.detach().forEach((el) => el.remove())
      })

      room.on(RoomEvent.Disconnected, () => {
        cleanupLiveKit()
      })

      await room.connect(LIVEKIT_URL, tokenData.token)
      await room.localParticipant.setMicrophoneEnabled(true)

      mode = 'livekit'
      currentServerId = serverId
      voiceStore.connectionMode = 'livekit'
      voiceStore.isJoining = false
      voiceStore.setCurrentVoiceChannel(channelId)
      send({ type: 'voice_join', channel_id: channelId, server_id: serverId })
      return
    } catch {
      console.warn('LiveKit unavailable, falling back to P2P')
      room = null
    }
  }

  // P2P fallback
  try {
    localStream = await navigator.mediaDevices.getUserMedia({ audio: true, video: false })
  } catch {
    console.error('Failed to get microphone access')
    voiceStore.isJoining = false
    return
  }

  mode = 'p2p'
  currentServerId = serverId
  voiceStore.connectionMode = 'p2p'
  voiceStore.isJoining = false
  voiceStore.setCurrentVoiceChannel(channelId)
  send({ type: 'voice_join', channel_id: channelId, server_id: serverId })

  unsubOffer = on('webrtc_offer', handleOffer)
  unsubAnswer = on('webrtc_answer', handleAnswer)
  unsubIce = on('webrtc_ice', handleIce)
  unsubVoiceState = on('voice_state_update', handleVoiceStateForPeers)
}

export async function leaveChannel() {
  const voiceStore = useVoiceStore()
  send({ type: 'voice_leave' })

  if (mode === 'livekit') {
    if (room) {
      // Stop local tracks explicitly before disconnect (iOS Safari won't release mic otherwise)
      room.localParticipant.audioTrackPublications.forEach((pub) => {
        pub.track?.stop()
      })
      await room.disconnect()
      cleanupLiveKit()
    }
  } else if (mode === 'p2p') {
    cleanupP2P()
  }

  mode = null
  currentServerId = null
  voiceStore.setCurrentVoiceChannel(null)
  voiceStore.isMuted = false
  voiceStore.isDeafened = false
  voiceStore.connectionMode = null
}

// --- LiveKit helpers ---

function cleanupLiveKit() {
  if (!room) return
  document.querySelectorAll('[id^="lk-audio-"]').forEach((el) => {
    el.pause()
    el.srcObject = null
    el.remove()
  })
  room.removeAllListeners()
  room = null
}

// --- P2P helpers ---

function cleanupP2P() {
  for (const [, pc] of peerConnections) {
    pc.close()
  }
  peerConnections.clear()

  for (const [, audio] of remoteAudios) {
    audio.pause()
    audio.srcObject = null
    audio.remove()
  }
  remoteAudios.clear()

  if (localStream) {
    localStream.getTracks().forEach((t) => t.stop())
    localStream = null
  }

  unsubOffer?.()
  unsubAnswer?.()
  unsubIce?.()
  unsubVoiceState?.()
  unsubOffer = unsubAnswer = unsubIce = unsubVoiceState = null
}

function createPeerConnection(remoteUserId, isOfferer) {
  const pc = new RTCPeerConnection({ iceServers: ICE_SERVERS })
  peerConnections.set(remoteUserId, pc)

  if (localStream) {
    localStream.getTracks().forEach((track) => pc.addTrack(track, localStream))
  }

  pc.ontrack = (event) => {
    let audio = remoteAudios.get(remoteUserId)
    if (!audio) {
      audio = document.createElement('audio')
      audio.autoplay = true
      remoteAudios.set(remoteUserId, audio)
    }
    audio.srcObject = event.streams[0]
    const voiceStore = useVoiceStore()
    audio.muted = voiceStore.isDeafened
  }

  pc.onicecandidate = (event) => {
    if (event.candidate) {
      send({
        type: 'webrtc_ice',
        target_user_id: remoteUserId,
        candidate: event.candidate.toJSON(),
      })
    }
  }

  if (isOfferer) {
    pc.createOffer()
      .then((offer) => pc.setLocalDescription(offer))
      .then(() => {
        send({
          type: 'webrtc_offer',
          target_user_id: remoteUserId,
          sdp: pc.localDescription.toJSON(),
        })
      })
  }

  return pc
}

function handleOffer(data) {
  const { user_id, sdp } = data
  if (peerConnections.has(user_id)) {
    peerConnections.get(user_id).close()
  }
  const pc = createPeerConnection(user_id, false)
  pc.setRemoteDescription(new RTCSessionDescription(sdp))
    .then(() => pc.createAnswer())
    .then((answer) => pc.setLocalDescription(answer))
    .then(() => {
      send({
        type: 'webrtc_answer',
        target_user_id: user_id,
        sdp: pc.localDescription.toJSON(),
      })
    })
}

function handleAnswer(data) {
  const { user_id, sdp } = data
  const pc = peerConnections.get(user_id)
  if (pc && pc.signalingState === 'have-local-offer') {
    pc.setRemoteDescription(new RTCSessionDescription(sdp))
  }
}

function handleIce(data) {
  const { user_id, candidate } = data
  const pc = peerConnections.get(user_id)
  if (pc && candidate && pc.remoteDescription) {
    pc.addIceCandidate(new RTCIceCandidate(candidate)).catch(() => {})
  }
}

function handleVoiceStateForPeers(data) {
  const voiceStore = useVoiceStore()
  const authStore = useAuthStore()
  const { channel_id, user_id, action } = data

  if (String(channel_id) !== String(voiceStore.currentVoiceChannelId)) return
  if (String(user_id) === String(authStore.dbUser?.id)) return

  if (action === 'join') {
    if (!peerConnections.has(user_id)) {
      createPeerConnection(user_id, true)
    }
  } else if (action === 'leave') {
    const pc = peerConnections.get(user_id)
    if (pc) {
      pc.close()
      peerConnections.delete(user_id)
    }
    const audio = remoteAudios.get(user_id)
    if (audio) {
      audio.pause()
      audio.srcObject = null
      audio.remove()
      remoteAudios.delete(user_id)
    }
  }
}

// --- Mode-agnostic exports ---

export function setMuted(muted) {
  if (mode === 'livekit' && room) {
    room.localParticipant.setMicrophoneEnabled(!muted)
  } else if (mode === 'p2p' && localStream) {
    localStream.getAudioTracks().forEach((track) => {
      track.enabled = !muted
    })
  }
}

export function setDeafened(deafened) {
  if (mode === 'livekit') {
    document.querySelectorAll('[id^="lk-audio-"]').forEach((el) => {
      el.muted = deafened
    })
  } else if (mode === 'p2p') {
    for (const audio of remoteAudios.values()) {
      audio.muted = deafened
    }
  }
}

export function cleanup() {
  const voiceStore = useVoiceStore()
  if (voiceStore.currentVoiceChannelId) {
    leaveChannel()
  }
}
