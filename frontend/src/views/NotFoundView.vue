<template>
  <div class="not-found">
    <div class="not-found-content">
      <!-- Decorative background embers -->
      <div class="ember ember-1"></div>
      <div class="ember ember-2"></div>
      <div class="ember ember-3"></div>

      <!-- Charcoal illustration -->
      <div class="illustration">
        <svg width="120" height="120" viewBox="0 0 120 120" fill="none" xmlns="http://www.w3.org/2000/svg">
          <!-- Charcoal chunk -->
          <path d="M30 85 L20 60 L35 35 L60 20 L90 30 L100 55 L95 80 L70 95 Z" fill="#2a2a2a" stroke="#444" stroke-width="1.5"/>
          <path d="M35 80 L28 58 L40 38 L62 26 L86 34 L94 56 L90 76 L68 88 Z" fill="#333"/>
          <!-- Cracks with ember glow -->
          <path d="M45 75 L55 50 L70 40" stroke="#E8521A" stroke-width="2" stroke-linecap="round" opacity="0.7"/>
          <path d="M50 80 L60 60 L75 55" stroke="#F97316" stroke-width="1.5" stroke-linecap="round" opacity="0.5"/>
          <path d="M40 65 L50 45" stroke="#E8521A" stroke-width="1" stroke-linecap="round" opacity="0.4"/>
          <!-- Smoke wisps -->
          <path d="M55 18 Q52 10 55 2" stroke="#666" stroke-width="1.5" stroke-linecap="round" fill="none" class="smoke smoke-1"/>
          <path d="M65 15 Q68 8 65 0" stroke="#555" stroke-width="1" stroke-linecap="round" fill="none" class="smoke smoke-2"/>
          <path d="M48 20 Q44 12 48 4" stroke="#555" stroke-width="1" stroke-linecap="round" fill="none" class="smoke smoke-3"/>
        </svg>
      </div>

      <h1 class="error-code">404</h1>
      <p class="error-message">This page burned to ash</p>
      <p class="error-detail">Nothing but embers here. The page you're looking for doesn't exist.</p>

      <div class="actions">
        <button @click="goHome" class="btn-primary">
          <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
            <path d="M8.354 1.146a.5.5 0 0 0-.708 0l-6 6A.5.5 0 0 0 1.5 7.5v7a.5.5 0 0 0 .5.5h4.5a.5.5 0 0 0 .5-.5v-4h2v4a.5.5 0 0 0 .5.5H14a.5.5 0 0 0 .5-.5v-7a.5.5 0 0 0-.146-.354L8.354 1.146z"/>
          </svg>
          Back to Home
        </button>
        <button @click="goBack" class="btn-secondary">Go Back</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'

const router = useRouter()

function goHome() {
  router.push('/channels/@me')
}

function goBack() {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.push('/channels/@me')
  }
}
</script>

<style scoped>
.not-found {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--background, #111);
  position: relative;
  overflow: hidden;
}

.not-found-content {
  text-align: center;
  position: relative;
  z-index: 1;
  animation: fade-in 0.5s ease-out;
}

/* Decorative embers */
.ember {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  opacity: 0.15;
  pointer-events: none;
}

.ember-1 {
  width: 300px;
  height: 300px;
  background: #E8521A;
  top: -100px;
  right: -50px;
  animation: ember-drift 8s ease-in-out infinite alternate;
}

.ember-2 {
  width: 200px;
  height: 200px;
  background: #F97316;
  bottom: -80px;
  left: -40px;
  animation: ember-drift 6s ease-in-out infinite alternate-reverse;
}

.ember-3 {
  width: 150px;
  height: 150px;
  background: #B91C1C;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  animation: ember-drift 10s ease-in-out infinite alternate;
}

.illustration {
  margin-bottom: 1.5rem;
  opacity: 0;
  animation: fade-in 0.6s ease-out 0.1s forwards;
}

.error-code {
  font-size: 6rem;
  font-weight: 800;
  line-height: 1;
  margin: 0;
  background: linear-gradient(135deg, #E8521A 0%, #F97316 50%, #E8521A 100%);
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
  opacity: 0;
  animation: fade-in 0.6s ease-out 0.2s forwards;
}

.error-message {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--text-1, #e5e5e5);
  margin: 0.75rem 0 0.5rem;
  opacity: 0;
  animation: fade-in 0.6s ease-out 0.3s forwards;
}

.error-detail {
  font-size: 0.875rem;
  color: var(--text-3, #888);
  margin: 0 0 2rem;
  max-width: 320px;
  line-height: 1.5;
  opacity: 0;
  animation: fade-in 0.6s ease-out 0.4s forwards;
}

.actions {
  display: flex;
  gap: 0.75rem;
  justify-content: center;
  opacity: 0;
  animation: fade-in 0.6s ease-out 0.5s forwards;
}

.btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.625rem 1.25rem;
  background: #E8521A;
  color: white;
  border: none;
  border-radius: 0.5rem;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-primary:hover {
  background: #d4471a;
  box-shadow: 0 0 20px rgba(232, 82, 26, 0.3);
}

.btn-secondary {
  padding: 0.625rem 1.25rem;
  background: transparent;
  color: var(--text-2, #aaa);
  border: 1px solid var(--surface-border, #333);
  border-radius: 0.5rem;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-secondary:hover {
  background: var(--surface, #1a1a1a);
  color: var(--text-1, #e5e5e5);
  border-color: var(--text-3, #888);
}

/* SVG smoke animation */
.smoke {
  opacity: 0;
  animation: smoke-rise 3s ease-in-out infinite;
}

.smoke-1 { animation-delay: 0s; }
.smoke-2 { animation-delay: 1s; }
.smoke-3 { animation-delay: 2s; }

@keyframes fade-in {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes ember-drift {
  from { transform: translate(0, 0); }
  to { transform: translate(20px, -20px); }
}

@keyframes smoke-rise {
  0% { opacity: 0; transform: translateY(0); }
  30% { opacity: 0.5; }
  100% { opacity: 0; transform: translateY(-10px); }
}
</style>
