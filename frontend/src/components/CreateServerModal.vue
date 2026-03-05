<script setup>
import { ref, onMounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useServersStore } from '@/stores/servers'
import { useToastStore } from '@/stores/toast'
import { useEscapeClose } from '@/composables/useEscapeClose'

const emit = defineEmits(['close'])
useEscapeClose(() => emit('close'))
const router = useRouter()
const serversStore = useServersStore()
const toastStore = useToastStore()
const name = ref('')
const isPublic = ref(true)
const nameInput = ref(null)

onMounted(() => nextTick(() => nameInput.value?.focus()))

async function create() {
  if (!name.value.trim()) return
  try {
    const server = await serversStore.createServer(name.value.trim(), isPublic.value)
    serversStore.selectServer(server)
    router.push(`/channels/${server.id}`)
    emit('close')
  } catch {
    toastStore.add('Failed to create server')
  }
}
</script>

<template>
  <Teleport to="body">
    <div class="fixed inset-0 bg-[var(--backdrop)] backdrop-blur-sm flex items-center justify-center z-50" @click.self="emit('close')">
      <div class="bg-[var(--modal-bg)] rounded-2xl p-7 w-full max-w-md shadow-2xl shadow-black/10 animate-scale-in border border-[var(--modal-border)]">
        <h2 class="font-display text-2xl font-bold text-[var(--text-1)] mb-1">Create a Server</h2>
        <p class="text-[var(--text-3)] text-sm mb-6">Give your new server a name to get started</p>

        <label class="text-[11px] font-bold text-[var(--text-4)] uppercase tracking-[0.1em]">Server Name</label>
        <input
          ref="nameInput"
          v-model="name"
          @keyup.enter="create"
          placeholder="My Awesome Server"
          class="w-full bg-[var(--surface)] text-[var(--text-1)] px-3.5 py-2.5 rounded-xl mt-2 mb-5 placeholder-[var(--text-4)] text-sm border border-[var(--surface-border)]"
        />

        <!-- Visibility toggle -->
        <div class="flex items-center justify-between mb-6 p-3.5 rounded-xl bg-[var(--surface)] border border-[var(--surface-border)]">
          <div>
            <p class="text-sm font-medium text-[var(--text-1)]">Public Server</p>
            <p class="text-xs text-[var(--text-4)] mt-0.5">
              {{ isPublic ? 'Anyone can discover and join in Explore' : 'Only accessible via invite link' }}
            </p>
          </div>
          <button
            type="button"
            @click="isPublic = !isPublic"
            class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors duration-200 cursor-pointer shrink-0 ml-4"
            :class="isPublic ? 'bg-[#E8521A]' : 'bg-[var(--surface-border)]'"
          >
            <span
              class="inline-block h-4 w-4 transform rounded-full bg-white shadow transition-transform duration-200"
              :class="isPublic ? 'translate-x-6' : 'translate-x-1'"
            />
          </button>
        </div>

        <div class="flex justify-end gap-3">
          <button @click="emit('close')" class="text-[var(--text-3)] hover:text-[var(--text-1)] px-4 py-2.5 rounded-xl cursor-pointer font-medium transition-colors duration-150">
            Cancel
          </button>
          <button
            @click="create"
            :disabled="!name.trim()"
            class="bg-[#E8521A] text-white px-5 py-2.5 rounded-xl hover:bg-[#D44818] disabled:opacity-40 cursor-pointer font-semibold shadow-lg shadow-[#E8521A]/15 transition-all duration-200"
          >
            Create
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>
