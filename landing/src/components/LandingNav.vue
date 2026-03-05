<template>
  <nav
    class="fixed top-0 left-0 right-0 z-50 transition-all duration-300"
    :class="scrolled ? 'bg-coal-900/95 backdrop-blur-md border-b border-white/5' : 'bg-transparent'"
  >
    <div class="max-w-7xl mx-auto px-6 h-16 flex items-center justify-between">
      <!-- Logo -->
      <a href="#" class="flex items-center gap-2.5 group">
        <div class="w-8 h-8 relative">
          <svg viewBox="0 0 1024 1024" fill="none" class="w-full h-full">
            <path d="M416 87.4376C416 48.1063 466.688 27.9491 497.216 55.1681C591.2 139.015 645.824 323.113 581.216 462.784L577.376 470.562L577.952 470.695C607.952 476.64 635.696 451.476 688.496 373.573L695.216 363.562C699.11 357.724 704.399 352.802 710.669 349.177C716.939 345.552 724.022 343.322 731.368 342.66C738.714 341.999 746.125 342.923 753.025 345.362C759.925 347.802 766.129 351.689 771.152 356.724C835.184 420.861 896 552.353 896 636.96C896 827.583 723.632 981.333 512 981.333C459.866 981.333 410.116 972.003 364.737 955.089C323.716 970.987 278.893 981.333 237.565 981.333C215.984 981.333 205.821 963.398 219.66 945.112C230.873 931.204 244.349 912.826 256.19 893.798C177.574 830.801 128 739.167 128 636.915C128 536.263 177.056 426.134 254.336 355.293L283.376 328.968C294.944 318.42 304.208 309.75 313.04 301.079C381.68 233.501 416 170.391 416 87.4376Z" fill="url(#navGrad)"/>
            <path d="M374.519 824.835C374.519 824.835 372.693 823.579 369.778 821.433C326.092 788.898 298.667 741.752 298.667 689.262C298.667 591.407 394.193 512 512 512C629.807 512 725.333 591.407 725.333 689.262C725.333 787.163 629.807 863.831 512 863.831C501.95 863.831 485.452 863.191 462.507 861.913C432.593 880.635 388.93 896 350.72 896C338.892 896 333.322 886.639 340.907 877.096C352.427 863.488 368.308 841.684 374.471 824.812L374.519 824.835Z" fill="#0D0D0D"/>
            <defs>
              <linearGradient id="navGrad" x1="512" y1="42.66" x2="512" y2="981.333" gradientUnits="userSpaceOnUse">
                <stop stop-color="#FBB500"/><stop offset="1" stop-color="#FD4A13"/>
              </linearGradient>
            </defs>
          </svg>
        </div>
        <span class="font-display font-700 text-lg tracking-tight text-white group-hover:text-ember-glow transition-colors">
          chatcoal
        </span>
      </a>

      <!-- Nav links -->
      <div class="hidden md:flex items-center gap-8">
        <a v-for="link in navLinks" :key="link.label"
           :href="link.href"
           class="text-sm text-white/50 hover:text-white/90 transition-colors font-medium tracking-wide">
          {{ link.label }}
        </a>
      </div>

      <!-- CTA -->
      <a href="https://app.chatcoal.com"
         class="text-sm font-semibold bg-ember hover:bg-ember-glow text-white px-5 py-2 rounded-lg transition-all duration-200 hover:shadow-[0_0_20px_#E8521A55]">
        Open App
      </a>
    </div>
  </nav>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'

const scrolled = ref(false)

const navLinks = [
  { label: 'Features', href: '#features' },
  { label: 'Open Source', href: '#open-source' },
]

function handleScroll() {
  scrolled.value = window.scrollY > 20
}

onMounted(() => window.addEventListener('scroll', handleScroll))
onUnmounted(() => window.removeEventListener('scroll', handleScroll))
</script>
