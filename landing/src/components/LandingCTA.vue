<template>
  <section class="py-32 relative overflow-hidden">
    <div class="absolute inset-0 bg-coal-900" />

    <!-- Glow orb -->
    <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[700px] h-[400px] pointer-events-none"
         style="background: radial-gradient(ellipse, #E8521A1A 0%, transparent 65%)" />

    <!-- Grid -->
    <div class="absolute inset-0 opacity-[0.03]"
         style="background-image: linear-gradient(#E8E6E1 1px, transparent 1px), linear-gradient(90deg, #E8E6E1 1px, transparent 1px); background-size: 60px 60px;" />

    <div class="relative z-10 max-w-4xl mx-auto px-6 text-center">
      <!-- Ember icon -->
      <div class="flex justify-center mb-8">
        <div class="w-20 h-20 relative animate-float-up">
          <svg viewBox="0 0 1024 1024" fill="none" class="w-full h-full drop-shadow-[0_0_30px_#E8521A]">
            <path d="M416 87.4376C416 48.1063 466.688 27.9491 497.216 55.1681C591.2 139.015 645.824 323.113 581.216 462.784L577.376 470.562L577.952 470.695C607.952 476.64 635.696 451.476 688.496 373.573L695.216 363.562C699.11 357.724 704.399 352.802 710.669 349.177C716.939 345.552 724.022 343.322 731.368 342.66C738.714 341.999 746.125 342.923 753.025 345.362C759.925 347.802 766.129 351.689 771.152 356.724C835.184 420.861 896 552.353 896 636.96C896 827.583 723.632 981.333 512 981.333C459.866 981.333 410.116 972.003 364.737 955.089C323.716 970.987 278.893 981.333 237.565 981.333C215.984 981.333 205.821 963.398 219.66 945.112C230.873 931.204 244.349 912.826 256.19 893.798C177.574 830.801 128 739.167 128 636.915C128 536.263 177.056 426.134 254.336 355.293L283.376 328.968C294.944 318.42 304.208 309.75 313.04 301.079C381.68 233.501 416 170.391 416 87.4376Z" fill="url(#ctaGrad)"/>
            <path d="M374.519 824.835C374.519 824.835 372.693 823.579 369.778 821.433C326.092 788.898 298.667 741.752 298.667 689.262C298.667 591.407 394.193 512 512 512C629.807 512 725.333 591.407 725.333 689.262C725.333 787.163 629.807 863.831 512 863.831C501.95 863.831 485.452 863.191 462.507 861.913C432.593 880.635 388.93 896 350.72 896C338.892 896 333.322 886.639 340.907 877.096C352.427 863.488 368.308 841.684 374.471 824.812L374.519 824.835Z" fill="#0D0D0D"/>
            <defs>
              <linearGradient id="ctaGrad" x1="512" y1="42.66" x2="512" y2="981.333" gradientUnits="userSpaceOnUse">
                <stop stop-color="#FBB500"/><stop offset="1" stop-color="#FD4A13"/>
              </linearGradient>
            </defs>
          </svg>
        </div>
      </div>

      <h2 class="font-display font-800 text-5xl md:text-6xl lg:text-7xl text-white leading-tight mb-6">
        Ready to light<br/>
        <span class="text-gradient-ember">the forge?</span>
      </h2>

      <p class="text-white/45 text-lg max-w-xl mx-auto mb-12 font-light leading-relaxed">
        Free forever, open source, and yours to own. Be among the first to build on chatcoal.
      </p>

      <div class="flex flex-col sm:flex-row gap-4 justify-center">
        <a href="https://app.chatcoal.com"
           class="inline-flex items-center justify-center gap-2 bg-ember hover:bg-ember-glow text-white font-semibold px-10 py-4 rounded-xl text-lg transition-all duration-200 glow-ember hover:glow-ember-strong animate-glow-flicker">
          Open App
          <svg class="w-5 h-5" viewBox="0 0 20 20" fill="none">
            <path d="M10 3L17 10L10 17M17 10H3" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </a>
        <a href="https://github.com/chat-coal/chatcoal"
           class="inline-flex items-center justify-center gap-2 bg-transparent hover:bg-coal-800 border border-white/15 hover:border-white/25 text-white/70 hover:text-white font-semibold px-10 py-4 rounded-xl text-lg transition-all duration-200">
          <svg class="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
            <path d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z"/>
          </svg>
          Self-Host
        </a>
      </div>
    </div>
  </section>
</template>
