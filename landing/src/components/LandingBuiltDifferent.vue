<template>
  <section id="open-source" class="py-32 relative overflow-hidden">
    <!-- Background -->
    <div class="absolute inset-0 bg-coal-950" />
    <!-- Ember stripe -->
    <div class="absolute left-0 top-0 bottom-0 w-px"
         style="background: linear-gradient(to bottom, transparent 0%, #E8521A44 30%, #E8521A66 50%, #E8521A44 70%, transparent 100%)" />

    <div class="relative z-10 max-w-7xl mx-auto px-6">
      <!-- Header -->
      <div class="max-w-2xl mb-24">
        <div class="inline-flex items-center gap-2 bg-coal-800 border border-ember/20 rounded-full px-4 py-1.5 mb-6">
          <span class="w-1.5 h-1.5 rounded-full bg-ember" />
          <span class="text-xs font-semibold text-ember tracking-widest uppercase">Open Source</span>
        </div>
        <h2 class="font-display font-800 text-5xl md:text-6xl text-white leading-tight">
          Forged in the open.<br/>
          <span class="text-gradient-ember">Owned by no one.</span>
        </h2>
      </div>

      <!-- Steps -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-px bg-white/5 rounded-2xl overflow-hidden">
        <div v-for="(step, i) in steps" :key="step.title"
             class="bg-coal-900 p-10 relative group hover:bg-coal-800 transition-colors duration-300">
          <!-- Step number -->
          <div class="text-[80px] font-display font-800 leading-none mb-6 select-none"
               style="color: #E8521A0A">{{ String(i + 1).padStart(2, '0') }}</div>

          <div class="absolute top-10 left-10">
            <div class="w-10 h-10 rounded-xl flex items-center justify-center bg-ember/10 text-ember mb-4"
                 v-html="step.icon" />
          </div>

          <div class="mt-6">
            <h3 class="font-display font-700 text-xl text-white mb-3">{{ step.title }}</h3>
            <p class="text-white/45 text-sm leading-relaxed">{{ step.desc }}</p>
          </div>

          <!-- Bottom accent line -->
          <div class="absolute bottom-0 left-0 right-0 h-px bg-ember/0 group-hover:bg-ember/30 transition-colors duration-300" />
        </div>
      </div>

      <!-- Bottom callout -->
      <div class="mt-16 flex flex-col md:flex-row items-center gap-8 p-8 rounded-2xl border border-ember/15 bg-coal-800"
           style="background: linear-gradient(135deg, #14100E 0%, #111 100%)">
        <div class="flex-1">
          <h3 class="font-display font-700 text-2xl text-white mb-2">100% open source</h3>
          <p class="text-white/45 text-sm leading-relaxed max-w-lg">
            Every line of code is public. No hidden telemetry, no dark patterns, no investor-driven enshittification. Run it yourself or trust us to do it — your choice.
          </p>
        </div>
        <div class="flex gap-3 flex-shrink-0">
          <a href="https://github.com"
             class="inline-flex items-center gap-2 bg-coal-700 hover:bg-coal-600 border border-white/10 text-white/70 hover:text-white font-semibold px-5 py-3 rounded-xl text-sm transition-all duration-200">
            <svg class="w-4 h-4" viewBox="0 0 24 24" fill="currentColor">
              <path d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z"/>
            </svg>
            View on GitHub
          </a>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup>
const steps = [
  {
    title: 'Create Your Server',
    desc: 'Set up a server in seconds. Add channels, configure roles, and invite your community. No verification, no waitlist.',
    icon: `<svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5" class="w-5 h-5"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15"/></svg>`,
  },
  {
    title: 'Invite & Connect',
    desc: 'Share an invite link. Your community joins instantly, no account required for a look-around. Seamless onboarding.',
    icon: `<svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5" class="w-5 h-5"><path stroke-linecap="round" stroke-linejoin="round" d="M13.19 8.688a4.5 4.5 0 011.242 7.244l-4.5 4.5a4.5 4.5 0 01-6.364-6.364l1.757-1.757m13.35-.622l1.757-1.757a4.5 4.5 0 00-6.364-6.364l-4.5 4.5a4.5 4.5 0 001.242 7.244"/></svg>`,
  },
  {
    title: 'Own Your Space',
    desc: 'Self-host on your own infrastructure or use our managed service. Your data, your rules.',
    icon: `<svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5" class="w-5 h-5"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z"/></svg>`,
  },
]
</script>

<style scoped>
div :deep(svg) {
  width: 20px;
  height: 20px;
}
</style>
