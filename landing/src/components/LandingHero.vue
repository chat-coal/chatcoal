<template>
  <section class="relative overflow-hidden pt-16">
    <!-- Background layers -->
    <div class="absolute inset-0 bg-coal-950" />

    <!-- Ember glow at bottom center -->
    <div class="absolute bottom-0 left-1/2 -translate-x-1/2 w-[800px] h-[400px] rounded-full"
         style="background: radial-gradient(ellipse at 50% 100%, #E8521A1A 0%, transparent 70%)" />

    <!-- Grid lines -->
    <div class="absolute inset-0 opacity-[0.04]"
         style="background-image: linear-gradient(#E8E6E1 1px, transparent 1px), linear-gradient(90deg, #E8E6E1 1px, transparent 1px); background-size: 80px 80px;" />

    <!-- Floating ember particles -->
    <div class="absolute inset-0 pointer-events-none overflow-hidden">
      <div v-for="p in particles" :key="p.id"
           class="absolute rounded-full"
           :style="{
             left: p.x + '%',
             bottom: p.startY + '%',
             width: p.size + 'px',
             height: p.size + 'px',
             backgroundColor: p.color,
             animationDuration: p.duration + 's',
             animationDelay: p.delay + 's',
             filter: 'blur(' + p.blur + 'px)',
           }"
           style="animation-name: particleRise; animation-timing-function: ease-out; animation-iteration-count: infinite;" />
    </div>

    <!-- Diagonal accent line -->
    <div class="absolute top-0 right-0 w-px h-full opacity-10"
         style="background: linear-gradient(to bottom, transparent, #E8521A, transparent)" />

    <!-- Content -->
    <div class="relative z-10 max-w-5xl mx-auto px-6 text-center pt-24 pb-16">
      <!-- Headline -->
      <h1 class="font-display font-800 text-6xl sm:text-7xl md:text-8xl leading-[0.9] tracking-tight mb-6"
          style="animation: slideUp 0.6s ease-out 0.1s both;">
        <span class="block text-white">Where</span>
        <span class="block text-gradient-ember">Communities</span>
        <span class="block text-white">Ignite.</span>
      </h1>

      <!-- Sub -->
      <p class="text-white/50 text-lg sm:text-xl max-w-2xl mx-auto mb-10 font-light leading-relaxed"
         style="animation: slideUp 0.6s ease-out 0.25s both;">
        Real-time messaging built from the ground up. Servers, channels, DMs, and voice — fast, private, and fully open source. No bloat. No tracking. Yours to own.
      </p>

      <!-- CTAs -->
      <div class="flex flex-col sm:flex-row gap-4 justify-center items-center"
           style="animation: slideUp 0.6s ease-out 0.4s both;">
        <a href="https://app.chatcoal.com"
           class="group relative inline-flex items-center gap-2 bg-ember hover:bg-ember-glow text-white font-semibold px-8 py-4 rounded-xl text-base transition-all duration-200 glow-ember hover:glow-ember-strong">
          <svg class="w-4 h-4" viewBox="0 0 16 16" fill="none">
            <path d="M8 1L15 8L8 15M15 8H1" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
          Open App
        </a>
      </div>

      <!-- Social proof -->
      <p class="text-white/25 text-sm mt-8" style="animation: fadeIn 0.8s ease-out 0.6s both;">
        Free forever · Self-hostable · No tracking
      </p>
    </div>

    <!-- App preview mockup -->
    <div class="relative z-10 w-full max-w-5xl mx-auto px-6 pointer-events-none"
         style="animation: slideUp 0.8s ease-out 0.5s both;">
      <AppMockup />
    </div>
  </section>
</template>

<script setup>
import { ref } from 'vue'
import AppMockup from './AppMockup.vue'

const particles = ref(
  Array.from({ length: 20 }, (_, i) => ({
    id: i,
    x: Math.random() * 100,
    startY: Math.random() * 20,
    size: Math.random() * 4 + 2,
    color: ['#E8521A', '#D4782A', '#FF6B35', '#FF8C5A'][Math.floor(Math.random() * 4)],
    duration: Math.random() * 4 + 4,
    delay: Math.random() * 6,
    blur: Math.random() * 2,
  }))
)
</script>

<style scoped>
h1 span {
  display: block;
  opacity: 0;
  animation: slideUp 0.6s ease-out forwards;
}
h1 span:nth-child(1) { animation-delay: 0.1s; }
h1 span:nth-child(2) { animation-delay: 0.2s; }
h1 span:nth-child(3) { animation-delay: 0.3s; }

@keyframes slideUp {
  from { opacity: 0; transform: translateY(30px); }
  to { opacity: 1; transform: translateY(0); }
}
@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}
</style>
