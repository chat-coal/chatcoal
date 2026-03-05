<template>
  <section id="features" class="py-32 relative overflow-hidden">
    <!-- Section glow -->
    <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[600px] rounded-full pointer-events-none"
         style="background: radial-gradient(circle, #E8521A0A 0%, transparent 70%)" />

    <div class="max-w-7xl mx-auto px-6">
      <!-- Header -->
      <div class="text-center mb-20">
        <div class="inline-flex items-center gap-2 bg-coal-800 border border-ember/20 rounded-full px-4 py-1.5 mb-6">
          <span class="w-1.5 h-1.5 rounded-full bg-ember" />
          <span class="text-xs font-semibold text-ember tracking-widest uppercase">Features</span>
        </div>
        <h2 class="font-display font-800 text-5xl md:text-6xl text-white leading-tight mb-4">
          Everything you need.<br/>
          <span class="text-gradient-ember">Nothing you don't.</span>
        </h2>
        <p class="text-white/40 text-lg max-w-xl mx-auto">
          Built from scratch with performance and privacy at the core.
        </p>
      </div>

      <!-- Feature grid -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <FeatureCard
          v-for="(feature, i) in features"
          :key="feature.title"
          :feature="feature"
          :large="i === 0 || i === 3"
        />
      </div>
    </div>
  </section>
</template>

<script setup>
import FeatureCard from './FeatureCard.vue'

const features = [
  {
    icon: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M2.25 12.76c0 1.6 1.123 2.994 2.707 3.227 1.087.16 2.185.283 3.293.369V21l4.076-4.076a1.526 1.526 0 011.037-.443 48.282 48.282 0 005.68-.494c1.584-.233 2.707-1.626 2.707-3.228V6.741c0-1.602-1.123-2.995-2.707-3.228A48.394 48.394 0 0012 3c-2.392 0-4.744.175-7.043.513C3.373 3.746 2.25 5.14 2.25 6.741v6.018z"/></svg>`,
    title: 'Servers & Channels',
    desc: 'Organize your community with servers, text channels, voice channels, and forum channels. Keep conversations focused and discoverable.',
    accent: true,
  },
  {
    icon: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M19.114 5.636a9 9 0 010 12.728M16.463 8.288a5.25 5.25 0 010 7.424M6.75 8.25l4.72-4.72a.75.75 0 011.28.53v15.88a.75.75 0 01-1.28.53l-4.72-4.72H4.51c-.88 0-1.704-.507-1.938-1.354A9.01 9.01 0 012.25 12c0-.83.112-1.633.322-2.396C2.806 8.756 3.63 8.25 4.51 8.25H6.75z"/></svg>`,
    title: 'Crystal Voice',
    desc: 'Join voice channels and talk with your community in real time. Simple, low-latency, and built right into your server.',
  },
  {
    icon: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M15.75 6a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0zM4.501 20.118a7.5 7.5 0 0114.998 0A17.933 17.933 0 0112 21.75c-2.676 0-5.216-.584-7.499-1.632z"/></svg>`,
    title: 'Direct Messages',
    desc: 'Send direct messages to anyone on the platform. Fast, private, and separate from your servers.',
  },
  {
    icon: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M15.59 14.37a6 6 0 01-5.84 7.38v-4.8m5.84-2.58a14.98 14.98 0 006.16-12.12A14.98 14.98 0 009.631 8.41m5.96 5.96a14.926 14.926 0 01-5.841 2.58m-.119-8.54a6 6 0 00-7.381 5.84h4.8m2.581-5.84a14.927 14.927 0 00-2.58 5.84m2.699 2.7c-.103.021-.207.041-.311.06a15.09 15.09 0 01-2.448-2.448 14.9 14.9 0 01.06-.312m-2.24 2.39a4.493 4.493 0 00-1.757 4.306 4.493 4.493 0 004.306-1.758M16.5 9a1.5 1.5 0 11-3 0 1.5 1.5 0 013 0z"/></svg>`,
    title: 'Open Source',
    desc: 'Fully open source and self-hostable. Audit the code, run your own instance, or contribute upstream.',
    accent: true,
  },
  {
    icon: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M14.857 17.082a23.848 23.848 0 005.454-1.31A8.967 8.967 0 0118 9.75v-.7V9A6 6 0 006 9v.75a8.967 8.967 0 01-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 01-5.714 0m5.714 0a3 3 0 11-5.714 0"/></svg>`,
    title: 'Smart Notifications',
    desc: 'Per-channel notification rules so you only hear what matters.',
  },
  {
    icon: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z"/></svg>`,
    title: 'Instant Search',
    desc: 'Full-text search across all channels and messages. Find anything fast.',
  },
]
</script>
