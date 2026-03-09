<template>
  <section id="roadmap" class="py-32 relative overflow-hidden">
    <div class="absolute inset-0 bg-coal-900" />

    <!-- Subtle glow -->
    <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[500px] h-[500px] rounded-full pointer-events-none"
         style="background: radial-gradient(circle, #E8521A08 0%, transparent 70%)" />

    <div class="relative z-10 max-w-4xl mx-auto px-6">
      <!-- Header -->
      <div class="text-center mb-20">
        <div class="inline-flex items-center gap-2 bg-coal-800 border border-ember/20 rounded-full px-4 py-1.5 mb-6">
          <span class="w-1.5 h-1.5 rounded-full bg-ember" />
          <span class="text-xs font-semibold text-ember tracking-widest uppercase">Roadmap</span>
        </div>
        <h2 class="font-display font-800 text-5xl md:text-6xl text-white leading-tight mb-4">
          What's coming<br/>
          <span class="text-gradient-ember">next.</span>
        </h2>
        <p class="text-white/40 text-lg max-w-xl mx-auto">
          chatcoal is growing fast. Here's what we're building.
        </p>
      </div>

      <!-- Timeline -->
      <div class="relative">
        <!-- Vertical line -->
        <div class="absolute left-6 md:left-1/2 md:-translate-x-px top-0 bottom-0 w-px"
             style="background: linear-gradient(to bottom, transparent 0%, #E8521A33 10%, #E8521A33 90%, transparent 100%)" />

        <div class="space-y-12">
          <div v-for="(item, i) in items" :key="item.title"
               class="relative flex items-start gap-8"
               :class="i % 2 === 0 ? 'md:flex-row' : 'md:flex-row-reverse'">

            <!-- Dot on the line -->
            <div class="absolute left-6 md:left-1/2 -translate-x-1/2 w-3 h-3 rounded-full bg-coal-900 border-2 border-ember/50 z-10 mt-7" />

            <!-- Spacer for the other side (desktop) -->
            <div class="hidden md:block md:w-1/2" />

            <!-- Card -->
            <div class="ml-14 md:ml-0 md:w-1/2 group">
              <div class="bg-coal-800 border border-white/5 rounded-2xl p-8 hover:border-ember/20 transition-colors duration-300">
                <div class="flex items-center gap-3 mb-4">
                  <div class="w-10 h-10 rounded-xl flex items-center justify-center bg-ember/10 text-ember"
                       v-html="item.icon" />
                  <span v-if="item.tag"
                        class="text-[10px] font-semibold tracking-widest uppercase px-2 py-0.5 rounded-full"
                        :class="item.tagClass">
                    {{ item.tag }}
                  </span>
                </div>
                <h3 class="font-display font-700 text-xl text-white mb-2">{{ item.title }}</h3>
                <p class="text-white/40 text-sm leading-relaxed">{{ item.desc }}</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup>
const items = [
  {
    title: 'Desktop Apps',
    desc: 'Apps for macOS, Windows, and Linux. Fast startup and system-wide push to talk.',
    tag: 'In Progress',
    tagClass: 'bg-ember/15 text-ember',
    icon: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="w-5 h-5"><path stroke-linecap="round" stroke-linejoin="round" d="M9 17.25v1.007a3 3 0 01-.879 2.122L7.5 21h9l-.621-.621A3 3 0 0115 18.257V17.25m6-12V15a2.25 2.25 0 01-2.25 2.25H5.25A2.25 2.25 0 013 15V5.25A2.25 2.25 0 015.25 3h13.5A2.25 2.25 0 0121 5.25z"/></svg>`,
  },
  {
    title: 'iOS App',
    desc: 'A full-featured native iOS app. Stay connected on the go with push notifications and a smooth mobile experience.',
    tag: 'Planned',
    tagClass: 'bg-white/5 text-white/40',
    icon: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="w-5 h-5"><path stroke-linecap="round" stroke-linejoin="round" d="M10.5 1.5H8.25A2.25 2.25 0 006 3.75v16.5a2.25 2.25 0 002.25 2.25h7.5A2.25 2.25 0 0018 20.25V3.75a2.25 2.25 0 00-2.25-2.25H13.5m-3 0V3h3V1.5m-3 0h3m-3 18.75h3"/></svg>`,
  },
  {
    title: 'Android App',
    desc: 'chatcoal in your pocket. Native Android app with Material design and notifications.',
    tag: 'Planned',
    tagClass: 'bg-white/5 text-white/40',
    icon: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="w-5 h-5"><path stroke-linecap="round" stroke-linejoin="round" d="M10.5 1.5H8.25A2.25 2.25 0 006 3.75v16.5a2.25 2.25 0 002.25 2.25h7.5A2.25 2.25 0 0018 20.25V3.75a2.25 2.25 0 00-2.25-2.25H13.5m-3 0V3h3V1.5m-3 0h3m-3 18.75h3"/></svg>`,
  },
]
</script>
