<template>
  <div class="min-h-screen bg-coal-900 font-body relative overflow-hidden">
    <LandingNav />

    <!-- Background embers -->
    <div class="absolute inset-0 pointer-events-none">
      <div class="absolute top-1/4 left-1/4 w-[500px] h-[500px] bg-ember/5 rounded-full blur-[120px] animate-ember-pulse" />
      <div class="absolute bottom-1/4 right-1/4 w-[400px] h-[400px] bg-ember-deep/8 rounded-full blur-[100px] animate-ember-pulse" style="animation-delay: 1.5s" />
    </div>

    <main class="relative z-10 flex flex-col items-center justify-center min-h-screen px-6 text-center">
      <!-- Charcoal illustration -->
      <div class="mb-8 opacity-0 animate-fade-in" style="animation-delay: 0.1s">
        <div class="relative w-32 h-32 mx-auto">
          <svg viewBox="0 0 120 120" fill="none" class="w-full h-full">
            <!-- Charcoal chunk -->
            <path d="M30 85 L20 60 L35 35 L60 20 L90 30 L100 55 L95 80 L70 95 Z" fill="#1C1C1C" stroke="#333" stroke-width="1"/>
            <path d="M35 80 L28 58 L40 38 L62 26 L86 34 L94 56 L90 76 L68 88 Z" fill="#252525"/>
            <!-- Ember cracks -->
            <path d="M45 75 L55 50 L70 40" stroke="#E8521A" stroke-width="2.5" stroke-linecap="round" opacity="0.8">
              <animate attributeName="opacity" values="0.8;0.4;0.8" dur="3s" repeatCount="indefinite"/>
            </path>
            <path d="M50 80 L60 60 L75 55" stroke="#FF6B35" stroke-width="2" stroke-linecap="round" opacity="0.5">
              <animate attributeName="opacity" values="0.5;0.25;0.5" dur="4s" repeatCount="indefinite"/>
            </path>
            <path d="M40 65 L50 45" stroke="#E8521A" stroke-width="1.5" stroke-linecap="round" opacity="0.4">
              <animate attributeName="opacity" values="0.4;0.2;0.4" dur="2.5s" repeatCount="indefinite"/>
            </path>
            <!-- Smoke wisps -->
            <g opacity="0.4">
              <path d="M55 18 Q52 10 55 2" stroke="#555" stroke-width="1.5" stroke-linecap="round" fill="none">
                <animate attributeName="d" values="M55 18 Q52 10 55 2;M55 18 Q58 10 55 2;M55 18 Q52 10 55 2" dur="3s" repeatCount="indefinite"/>
                <animate attributeName="opacity" values="0.4;0.15;0.4" dur="3s" repeatCount="indefinite"/>
              </path>
              <path d="M65 15 Q68 8 65 0" stroke="#444" stroke-width="1" stroke-linecap="round" fill="none">
                <animate attributeName="d" values="M65 15 Q68 8 65 0;M65 15 Q62 8 65 0;M65 15 Q68 8 65 0" dur="4s" repeatCount="indefinite"/>
                <animate attributeName="opacity" values="0.3;0.1;0.3" dur="4s" repeatCount="indefinite"/>
              </path>
            </g>
          </svg>
        </div>
      </div>

      <!-- 404 -->
      <h1 class="font-display font-800 text-[7rem] sm:text-[9rem] leading-none tracking-tight text-gradient-ember opacity-0 animate-fade-in" style="animation-delay: 0.2s">
        404
      </h1>

      <!-- Message -->
      <p class="font-display font-600 text-xl sm:text-2xl text-white/90 mt-4 mb-3 opacity-0 animate-fade-in" style="animation-delay: 0.3s">
        This page burned to ash
      </p>
      <p class="text-white/40 text-sm sm:text-base max-w-md leading-relaxed mb-10 opacity-0 animate-fade-in" style="animation-delay: 0.4s">
        Nothing but embers here. The page you're looking for doesn't exist or has been moved.
      </p>

      <!-- Actions -->
      <div class="flex flex-col sm:flex-row items-center gap-3 opacity-0 animate-fade-in" style="animation-delay: 0.5s">
        <RouterLink
          to="/"
          class="inline-flex items-center gap-2.5 bg-ember hover:bg-ember-glow text-white font-semibold text-sm px-6 py-3 rounded-lg transition-all duration-200 hover:shadow-[0_0_24px_#E8521A55]"
        >
          <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
            <path d="M8.354 1.146a.5.5 0 0 0-.708 0l-6 6A.5.5 0 0 0 1.5 7.5v7a.5.5 0 0 0 .5.5h4.5a.5.5 0 0 0 .5-.5v-4h2v4a.5.5 0 0 0 .5.5H14a.5.5 0 0 0 .5-.5v-7a.5.5 0 0 0-.146-.354L8.354 1.146z"/>
          </svg>
          Back to Home
        </RouterLink>
        <a
          href="https://app.chatcoal.com"
          class="inline-flex items-center gap-2 text-sm font-medium text-white/50 hover:text-white/80 px-5 py-3 rounded-lg border border-white/8 hover:border-white/15 transition-all duration-200"
        >
          Open App
          <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor" class="opacity-50">
            <path fill-rule="evenodd" d="M8.636 3.5a.5.5 0 0 0-.5-.5H1.5A1.5 1.5 0 0 0 0 4.5v10A1.5 1.5 0 0 0 1.5 16h10a1.5 1.5 0 0 0 1.5-1.5V7.864a.5.5 0 0 0-1 0V14.5a.5.5 0 0 1-.5.5h-10a.5.5 0 0 1-.5-.5v-10a.5.5 0 0 1 .5-.5h6.636a.5.5 0 0 0 .5-.5z"/>
            <path fill-rule="evenodd" d="M16 .5a.5.5 0 0 0-.5-.5h-5a.5.5 0 0 0 0 1h3.793L6.146 9.146a.5.5 0 1 0 .708.708L15 1.707V5.5a.5.5 0 0 0 1 0v-5z"/>
          </svg>
        </a>
      </div>
    </main>

    <LandingFooter />
  </div>
</template>

<script setup>
import LandingNav from '../components/LandingNav.vue'
import LandingFooter from '../components/LandingFooter.vue'
</script>
