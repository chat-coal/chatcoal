<script setup>
import { ref, onUnmounted } from 'vue'
import { useEscapeClose } from '@/composables/useEscapeClose'

const emit = defineEmits(['close'])
useEscapeClose(() => emit('close'))

const isTesting = ref(false)
const isLoopback = ref(false)
const volumeLevel = ref(0)
const error = ref('')

let audioContext = null
let mediaStream = null
let analyser = null
let gainNode = null
let animFrameId = null

async function startTest() {
  error.value = ''
  try {
    mediaStream = await navigator.mediaDevices.getUserMedia({ audio: true })
  } catch (e) {
    if (e.name === 'NotAllowedError') {
      error.value = 'Microphone permission was denied. Please allow access in your browser settings.'
    } else if (e.name === 'NotFoundError') {
      error.value = 'No microphone found. Please connect a microphone and try again.'
    } else {
      error.value = 'Could not access microphone. Please check your device settings.'
    }
    return
  }

  audioContext = new AudioContext()
  const source = audioContext.createMediaStreamSource(mediaStream)
  analyser = audioContext.createAnalyser()
  analyser.fftSize = 256
  analyser.smoothingTimeConstant = 0.5

  gainNode = audioContext.createGain()
  gainNode.gain.value = 0

  source.connect(analyser)
  analyser.connect(gainNode)
  gainNode.connect(audioContext.destination)

  isTesting.value = true
  isLoopback.value = false
  pollVolume()
}

function pollVolume() {
  if (!analyser) return
  const data = new Uint8Array(analyser.frequencyBinCount)
  analyser.getByteFrequencyData(data)
  let sum = 0
  for (let i = 0; i < data.length; i++) sum += data[i]
  const avg = sum / data.length
  volumeLevel.value = Math.min(100, Math.round((avg / 128) * 100))
  animFrameId = requestAnimationFrame(pollVolume)
}

function toggleLoopback() {
  if (!gainNode || !audioContext) return
  isLoopback.value = !isLoopback.value
  const target = isLoopback.value ? 1 : 0
  gainNode.gain.setTargetAtTime(target, audioContext.currentTime, 0.05)
}

function stopTest() {
  if (animFrameId) {
    cancelAnimationFrame(animFrameId)
    animFrameId = null
  }
  if (mediaStream) {
    mediaStream.getTracks().forEach(t => t.stop())
    mediaStream = null
  }
  if (audioContext) {
    audioContext.close()
    audioContext = null
  }
  analyser = null
  gainNode = null
  isTesting.value = false
  isLoopback.value = false
  volumeLevel.value = 0
}

function close() {
  stopTest()
  emit('close')
}

onUnmounted(() => {
  stopTest()
})

function barColor(index) {
  if (index < 12) return 'bg-green-500'
  if (index < 17) return 'bg-amber-400'
  return 'bg-orange-500'
}
</script>

<template>
  <Teleport to="body">
    <div
      class="fixed inset-0 bg-[var(--backdrop)] backdrop-blur-sm flex items-center justify-center z-50"
      @click.self="close"
    >
      <div class="bg-[var(--modal-bg)] rounded-2xl p-7 w-full max-w-sm shadow-2xl shadow-black/10 animate-scale-in border border-[var(--modal-border)]">
        <!-- Header -->
        <div class="flex items-center gap-3 mb-2">
          <div class="w-9 h-9 rounded-xl bg-[#E8521A]/10 flex items-center justify-center shrink-0">
            <svg class="w-5 h-5 text-[#E8521A]" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 18.75a6 6 0 0 0 6-6v-1.5m-6 7.5a6 6 0 0 1-6-6v-1.5m6 7.5v3.75m-3.75 0h7.5M12 15.75a3 3 0 0 1-3-3V4.5a3 3 0 1 1 6 0v8.25a3 3 0 0 1-3 3Z" />
            </svg>
          </div>
          <div>
            <h3 class="text-[var(--text-1)] font-bold text-lg">Microphone Test</h3>
          </div>
        </div>
        <p class="text-[var(--text-3)] text-sm mb-5 leading-relaxed">
          Test your microphone before joining a voice channel.
        </p>

        <!-- Error -->
        <div v-if="error" class="mb-4 p-3 rounded-xl bg-red-500/10 border border-red-500/20">
          <p class="text-red-400 text-sm">{{ error }}</p>
        </div>

        <!-- Volume meter -->
        <div class="mb-5">
          <div class="flex items-center gap-1 h-6">
            <div
              v-for="i in 20"
              :key="i"
              class="flex-1 h-full rounded-sm transition-opacity duration-75"
              :class="[
                barColor(i - 1),
                (i - 1) * 5 < volumeLevel ? 'opacity-100' : 'opacity-15'
              ]"
            ></div>
          </div>
          <p class="text-[var(--text-4)] text-xs mt-1.5 text-center">
            {{ isTesting ? 'Speak into your microphone' : 'Click Start Test to begin' }}
          </p>
        </div>

        <!-- Loopback toggle -->
        <button
          v-if="isTesting"
          @click="toggleLoopback"
          class="w-full mb-5 flex items-center gap-3 py-2.5 px-4 rounded-xl text-sm font-medium border transition-all duration-150 cursor-pointer"
          :class="isLoopback
            ? 'bg-[#E8521A]/10 border-[#E8521A]/30 text-[#E8521A]'
            : 'bg-[var(--surface-2)] border-[var(--surface-border)] text-[var(--text-2)] hover:border-[var(--text-4)]'"
        >
          <svg class="w-4 h-4 shrink-0" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" d="M19.114 5.636a9 9 0 0 1 0 12.728M16.463 8.288a5.25 5.25 0 0 1 0 7.424M6.75 8.25l4.72-4.72a.75.75 0 0 1 1.28.53v15.88a.75.75 0 0 1-1.28.53l-4.72-4.72H4.51c-.88 0-1.704-.507-1.938-1.354A9.009 9.009 0 0 1 2.25 12c0-.83.112-1.633.322-2.396C2.806 8.756 3.63 8.25 4.51 8.25H6.75Z" />
          </svg>
          {{ isLoopback ? 'Loopback On' : 'Enable Loopback' }}
          <span v-if="isLoopback" class="ml-auto text-xs opacity-70">Use headphones to avoid feedback</span>
        </button>

        <!-- Actions -->
        <div class="flex gap-2.5">
          <button
            @click="close"
            class="flex-1 py-2.5 px-4 rounded-xl text-sm font-semibold text-[var(--text-2)] bg-[var(--surface-2)] hover:bg-[var(--surface-3)] transition-colors duration-150 cursor-pointer"
          >
            Close
          </button>
          <button
            v-if="!isTesting"
            @click="startTest"
            class="flex-1 py-2.5 px-4 rounded-xl text-sm font-semibold text-white bg-[#E8521A] hover:bg-[#D44818] shadow-lg shadow-[#E8521A]/15 transition-all duration-150 cursor-pointer"
          >
            Start Test
          </button>
          <button
            v-else
            @click="stopTest"
            class="flex-1 py-2.5 px-4 rounded-xl text-sm font-semibold text-red-400 bg-red-500/10 hover:bg-red-500/20 border border-red-500/20 hover:border-red-500/40 transition-all duration-150 cursor-pointer"
          >
            Stop Test
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>
