<template>
  <div
    class="group relative rounded-2xl border transition-all duration-300 overflow-hidden cursor-default"
    :class="[
      feature.accent
        ? 'border-ember/25 hover:border-ember/50 bg-gradient-to-br from-coal-800 to-coal-900'
        : 'border-white/6 hover:border-white/12 bg-coal-800',
      large ? 'md:col-span-2 lg:col-span-1' : ''
    ]"
  >
    <!-- Accent glow on hover -->
    <div v-if="feature.accent"
         class="absolute inset-0 opacity-0 group-hover:opacity-100 transition-opacity duration-500 pointer-events-none"
         style="background: radial-gradient(ellipse at top left, #E8521A0D 0%, transparent 60%)" />

    <div class="p-7">
      <!-- Icon -->
      <div class="w-11 h-11 rounded-xl mb-5 flex items-center justify-center transition-all duration-300"
           :class="feature.accent ? 'bg-ember/15 text-ember group-hover:bg-ember/25' : 'bg-white/6 text-white/50 group-hover:text-white/80 group-hover:bg-white/10'"
           v-html="feature.icon" />

      <!-- Title -->
      <h3 class="font-display font-700 text-lg text-white mb-2 leading-tight">
        {{ feature.title }}
      </h3>

      <!-- Desc -->
      <p class="text-white/45 text-sm leading-relaxed">{{ feature.desc }}</p>
    </div>

    <!-- Ember corner accent -->
    <div v-if="feature.accent"
         class="absolute top-0 right-0 w-24 h-24 pointer-events-none opacity-20"
         style="background: radial-gradient(circle at top right, #E8521A, transparent 70%)" />
  </div>
</template>

<script setup>
defineProps({
  feature: Object,
  large: Boolean,
})
</script>

<style scoped>
div :deep(svg) {
  width: 22px;
  height: 22px;
}
</style>
