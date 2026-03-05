<template>
  <div class="min-h-screen bg-coal-900 font-body">
    <LandingNav />
    <LandingHero />
    <LandingFeatures />
    <LandingBuiltDifferent />
    <LandingCTA />
    <LandingFooter />
  </div>
</template>

<script setup>
import LandingNav from '../components/LandingNav.vue'
import LandingHero from '../components/LandingHero.vue'
import LandingFeatures from '../components/LandingFeatures.vue'
import LandingBuiltDifferent from '../components/LandingBuiltDifferent.vue'
import LandingCTA from '../components/LandingCTA.vue'
import LandingFooter from '../components/LandingFooter.vue'
</script>
