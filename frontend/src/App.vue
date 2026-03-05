<script setup>
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const isElectron = !!window.electronAPI?.isElectron
</script>

<template>
  <!-- Electron: transparent drag region for window dragging -->
  <div v-if="isElectron" class="fixed top-0 left-0 right-0 h-7 z-[9999]" style="-webkit-app-region: drag" />
  <div :class="{ 'pt-7': isElectron }">
    <div v-if="auth.loading" class="h-screen bg-[var(--surface)] flex items-center justify-center">
      <div class="flex flex-col items-center gap-4 animate-fade-in">
        <div class="w-8 h-8 border-2 border-[#E8521A] border-t-transparent rounded-full animate-spin"></div>
        <span class="text-[var(--text-3)] text-sm font-medium tracking-wide">Loading</span>
      </div>
    </div>
    <RouterView v-else />
  </div>
</template>
