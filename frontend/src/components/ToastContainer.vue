<script setup>
import { useToastStore } from '@/stores/toast'
const toastStore = useToastStore()
</script>

<template>
  <Teleport to="body">
    <div class="fixed bottom-5 right-5 z-[100] flex flex-col gap-2 pointer-events-none">
      <TransitionGroup name="toast">
        <div
          v-for="toast in toastStore.toasts"
          :key="toast.id"
          class="pointer-events-auto flex items-start gap-3 bg-[var(--card)] border border-[var(--surface-border)] rounded-xl px-4 py-3 shadow-xl max-w-sm"
        >
          <!-- Error icon -->
          <svg
            v-if="toast.type === 'error'"
            class="w-4 h-4 text-red-400 shrink-0 mt-0.5"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            viewBox="0 0 24 24"
          >
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z" />
          </svg>
          <!-- Success icon -->
          <svg
            v-else-if="toast.type === 'success'"
            class="w-4 h-4 text-green-400 shrink-0 mt-0.5"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            viewBox="0 0 24 24"
          >
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span class="text-sm text-[var(--text-2)] flex-1">{{ toast.message }}</span>
          <button
            @click="toastStore.remove(toast.id)"
            class="text-[var(--text-4)] hover:text-[var(--text-2)] transition-colors duration-150 shrink-0"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M18 6L6 18M6 6l12 12" />
            </svg>
          </button>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<style scoped>
.toast-enter-active,
.toast-leave-active {
  transition: all 0.25s ease;
}
.toast-enter-from {
  opacity: 0;
  transform: translateX(1rem);
}
.toast-leave-to {
  opacity: 0;
  transform: translateX(1rem);
}
</style>
