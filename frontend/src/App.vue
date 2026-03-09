<script setup>
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const isElectron = !!window.electronAPI?.isElectron
const isWindows = window.electronAPI?.platform === 'win32'
const isLinux = window.electronAPI?.platform === 'linux'
const linuxDesktop = window.electronAPI?.linuxDesktop || ''
const isGnome = /gnome|unity|pantheon|budgie|cinnamon|cosmic/.test(linuxDesktop)
const isKde = /kde|plasma/.test(linuxDesktop)
const electronAPI = window.electronAPI
</script>

<template>
  <!-- Electron: transparent drag region for window dragging -->
  <div v-if="isElectron" class="fixed top-0 left-0 right-0 h-7 z-[9999]" style="-webkit-app-region: drag">
    <!-- Windows: custom window controls (frameless window) -->
    <div v-if="isWindows" class="flex items-center h-full float-right" style="-webkit-app-region: no-drag">
      <button class="w-11 h-full flex items-center justify-center hover:bg-[var(--surface-2)] transition-colors" @click="electronAPI.windowMinimize()">
        <svg width="10" height="1" viewBox="0 0 10 1"><rect fill="currentColor" width="10" height="1" /></svg>
      </button>
      <button class="w-11 h-full flex items-center justify-center hover:bg-[var(--surface-2)] transition-colors" @click="electronAPI.windowMaximize()">
        <svg width="10" height="10" viewBox="0 0 10 10" fill="none"><rect x="0.5" y="0.5" width="9" height="9" stroke="currentColor" stroke-width="1" /></svg>
      </button>
      <button class="w-11 h-full flex items-center justify-center hover:bg-red-500 hover:text-white transition-colors" @click="electronAPI.windowClose()">
        <svg width="10" height="10" viewBox="0 0 10 10"><line x1="0" y1="0" x2="10" y2="10" stroke="currentColor" stroke-width="1.2" /><line x1="10" y1="0" x2="0" y2="10" stroke="currentColor" stroke-width="1.2" /></svg>
      </button>
    </div>

    <!-- Linux GNOME/GTK: Adwaita-style circular window controls -->
    <div v-else-if="isLinux && isGnome" class="flex items-center h-full gap-2 px-2 float-right" style="-webkit-app-region: no-drag">
      <button class="gnome-btn group" @click="electronAPI.windowMinimize()">
        <svg class="opacity-0 group-hover:opacity-100 transition-opacity" width="8" height="1" viewBox="0 0 8 1"><rect fill="#2e3436" width="8" height="1" /></svg>
      </button>
      <button class="gnome-btn group" @click="electronAPI.windowMaximize()">
        <svg class="opacity-0 group-hover:opacity-100 transition-opacity" width="8" height="8" viewBox="0 0 8 8" fill="none"><rect x="0.5" y="0.5" width="7" height="7" stroke="#2e3436" stroke-width="1" rx="0.5" /></svg>
      </button>
      <button class="gnome-btn-close group" @click="electronAPI.windowClose()">
        <svg class="opacity-0 group-hover:opacity-100 transition-opacity" width="8" height="8" viewBox="0 0 8 8"><line x1="1" y1="1" x2="7" y2="7" stroke="white" stroke-width="1.2" /><line x1="7" y1="1" x2="1" y2="7" stroke="white" stroke-width="1.2" /></svg>
      </button>
    </div>

    <!-- Linux KDE/Breeze-style window controls -->
    <div v-else-if="isLinux && isKde" class="flex items-center h-full float-right" style="-webkit-app-region: no-drag">
      <button class="kde-btn group" @click="electronAPI.windowMinimize()">
        <svg width="10" height="1" viewBox="0 0 10 1"><rect fill="currentColor" width="10" height="1" /></svg>
      </button>
      <button class="kde-btn group" @click="electronAPI.windowMaximize()">
        <svg width="10" height="10" viewBox="0 0 10 10" fill="none"><rect x="0.5" y="0.5" width="9" height="9" stroke="currentColor" stroke-width="1" rx="1" /></svg>
      </button>
      <button class="kde-btn kde-btn-close group" @click="electronAPI.windowClose()">
        <svg width="10" height="10" viewBox="0 0 10 10"><line x1="1" y1="1" x2="9" y2="9" stroke="currentColor" stroke-width="1.2" /><line x1="9" y1="1" x2="1" y2="9" stroke="currentColor" stroke-width="1.2" /></svg>
      </button>
    </div>

    <!-- Linux fallback: generic window controls -->
    <div v-else-if="isLinux" class="flex items-center h-full float-right" style="-webkit-app-region: no-drag">
      <button class="w-11 h-full flex items-center justify-center hover:bg-[var(--surface-2)] transition-colors" @click="electronAPI.windowMinimize()">
        <svg width="10" height="1" viewBox="0 0 10 1"><rect fill="currentColor" width="10" height="1" /></svg>
      </button>
      <button class="w-11 h-full flex items-center justify-center hover:bg-[var(--surface-2)] transition-colors" @click="electronAPI.windowMaximize()">
        <svg width="10" height="10" viewBox="0 0 10 10" fill="none"><rect x="0.5" y="0.5" width="9" height="9" stroke="currentColor" stroke-width="1" /></svg>
      </button>
      <button class="w-11 h-full flex items-center justify-center hover:bg-red-500 hover:text-white transition-colors" @click="electronAPI.windowClose()">
        <svg width="10" height="10" viewBox="0 0 10 10"><line x1="0" y1="0" x2="10" y2="10" stroke="currentColor" stroke-width="1.2" /><line x1="10" y1="0" x2="0" y2="10" stroke="currentColor" stroke-width="1.2" /></svg>
      </button>
    </div>
  </div>
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

<style scoped>
/* GNOME/Adwaita-style circular window buttons */
.gnome-btn {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--surface-2, #3d3846);
  transition: background 150ms;
}
.gnome-btn:hover {
  background: var(--surface-3, #504a5a);
}
.gnome-btn-close {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--surface-2, #3d3846);
  transition: background 150ms;
}
.gnome-btn-close:hover {
  background: #e01b24;
}

/* KDE/Breeze-style window buttons */
.kde-btn {
  width: 28px;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 150ms;
}
.kde-btn:hover {
  background: var(--surface-2, #3d3846);
}
.kde-btn-close:hover {
  background: #da4453;
  color: white;
}
</style>
