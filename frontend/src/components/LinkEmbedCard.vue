<script setup>
defineProps({
  embed: { type: Object, required: true },
})
</script>

<template>
  <a
    :href="embed.url"
    target="_blank"
    rel="noopener noreferrer"
    class="block mt-2 max-w-[400px] rounded-xl border border-[var(--surface-border)] bg-[var(--surface-2)] hover:bg-[var(--surface-3)] transition-colors duration-100 overflow-hidden"
  >
    <img
      v-if="embed.image"
      :src="embed.image"
      :alt="embed.title || ''"
      class="w-full h-[160px] object-cover"
      loading="lazy"
      @error="$event.target.style.display = 'none'"
    />
    <div class="px-3.5 py-2.5">
      <div v-if="embed.site_name" class="text-[10px] font-medium text-[var(--text-4)] uppercase tracking-wide mb-0.5">
        {{ embed.site_name }}
      </div>
      <div v-if="embed.title" class="text-sm font-semibold text-[#E8521A] leading-snug line-clamp-2">
        {{ embed.title }}
      </div>
      <p v-if="embed.description" class="text-xs text-[var(--text-3)] leading-relaxed mt-1 line-clamp-3">
        {{ embed.description }}
      </p>
    </div>
  </a>
</template>
