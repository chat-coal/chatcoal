<template>
  <footer class="border-t border-white/6 bg-coal-950">
    <div class="max-w-7xl mx-auto px-6 py-16">
      <div class="grid grid-cols-1 md:grid-cols-4 gap-12 mb-16">
        <!-- Brand -->
        <div class="md:col-span-1">
          <div class="flex items-center gap-2.5 mb-4">
            <div class="w-7 h-7">
              <svg viewBox="0 0 1024 1024" fill="none" class="w-full h-full">
                <path d="M416 87.4376C416 48.1063 466.688 27.9491 497.216 55.1681C591.2 139.015 645.824 323.113 581.216 462.784L577.376 470.562L577.952 470.695C607.952 476.64 635.696 451.476 688.496 373.573L695.216 363.562C699.11 357.724 704.399 352.802 710.669 349.177C716.939 345.552 724.022 343.322 731.368 342.66C738.714 341.999 746.125 342.923 753.025 345.362C759.925 347.802 766.129 351.689 771.152 356.724C835.184 420.861 896 552.353 896 636.96C896 827.583 723.632 981.333 512 981.333C459.866 981.333 410.116 972.003 364.737 955.089C323.716 970.987 278.893 981.333 237.565 981.333C215.984 981.333 205.821 963.398 219.66 945.112C230.873 931.204 244.349 912.826 256.19 893.798C177.574 830.801 128 739.167 128 636.915C128 536.263 177.056 426.134 254.336 355.293L283.376 328.968C294.944 318.42 304.208 309.75 313.04 301.079C381.68 233.501 416 170.391 416 87.4376Z" fill="url(#footerGrad)"/>
                <path d="M374.519 824.835C374.519 824.835 372.693 823.579 369.778 821.433C326.092 788.898 298.667 741.752 298.667 689.262C298.667 591.407 394.193 512 512 512C629.807 512 725.333 591.407 725.333 689.262C725.333 787.163 629.807 863.831 512 863.831C501.95 863.831 485.452 863.191 462.507 861.913C432.593 880.635 388.93 896 350.72 896C338.892 896 333.322 886.639 340.907 877.096C352.427 863.488 368.308 841.684 374.471 824.812L374.519 824.835Z" fill="#0A0A0F"/>
                <defs>
                  <linearGradient id="footerGrad" x1="512" y1="42.66" x2="512" y2="981.333" gradientUnits="userSpaceOnUse">
                    <stop stop-color="#FBB500"/><stop offset="1" stop-color="#FD4A13"/>
                  </linearGradient>
                </defs>
              </svg>
            </div>
            <span class="font-display font-700 text-white">chatcoal</span>
          </div>
          <p class="text-white/35 text-sm leading-relaxed mb-5">
            Real-time messaging built from the ground up.
          </p>
          <!-- Social links -->
          <div class="flex gap-3">
            <a v-for="social in socials" :key="social.name"
               :href="social.href"
               class="w-8 h-8 rounded-lg bg-coal-700 hover:bg-coal-600 border border-white/8 hover:border-ember/30 flex items-center justify-center text-white/40 hover:text-white transition-all duration-200"
               :title="social.name"
               v-html="social.icon" />
          </div>
        </div>

        <!-- Links -->
        <div v-for="group in linkGroups" :key="group.label">
          <h4 class="text-xs font-semibold text-white/40 uppercase tracking-widest mb-5">{{ group.label }}</h4>
          <ul class="space-y-3">
            <li v-for="link in group.links" :key="link.label">
              <component :is="link.href.startsWith('/') ? 'RouterLink' : 'a'"
                 v-bind="link.href.startsWith('/') ? { to: link.href } : { href: link.href }"
                 class="text-sm text-white/45 hover:text-white/80 transition-colors">
                {{ link.label }}
              </component>
            </li>
          </ul>
        </div>
      </div>

      <!-- Bottom bar -->
      <div class="flex flex-col sm:flex-row items-center justify-between gap-4 pt-8 border-t border-white/5">
        <p class="text-white/25 text-xs">
          © 2026 chatcoal. Open source under the AGPL License.
        </p>
      </div>
    </div>
  </footer>
</template>

<script setup>
const socials = [
  {
    name: 'GitHub',
    href: 'https://github.com',
    icon: `<svg class="w-4 h-4" viewBox="0 0 24 24" fill="currentColor"><path d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z"/></svg>`,
  },
  {
    name: 'Twitter/X',
    href: 'https://x.com',
    icon: `<svg class="w-4 h-4" viewBox="0 0 24 24" fill="currentColor"><path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/></svg>`,
  },
]

const linkGroups = [
  {
    label: 'Product',
    links: [
      { label: 'Features', href: '/#features' },
      { label: 'Open App', href: 'https://app.chatcoal.com' },
    ],
  },
  {
    label: 'Company',
    links: [
      { label: 'Privacy Policy', href: '/privacy' },
      { label: 'Terms of Service', href: '/terms' },
    ],
  },
]
</script>
