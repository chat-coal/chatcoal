<template>
  <!-- App UI mockup — dark chrome preview -->
  <div class="relative mx-auto" style="max-width: 880px;">
    <!-- Fade to bottom -->
    <div class="absolute inset-x-0 bottom-0 h-32 z-10 pointer-events-none"
         style="background: linear-gradient(to bottom, transparent, #0D0D0D)" />

    <!-- App frame -->
    <div class="mx-8 rounded-t-2xl overflow-hidden border border-white/8 shadow-[0_0_80px_#E8521A1A]"
         style="background: #111;">
      <!-- Window chrome -->
      <div class="flex items-center gap-1.5 px-4 py-3 border-b border-white/5"
           style="background: #0D0D0D;">
        <div class="w-3 h-3 rounded-full bg-white/10" />
        <div class="w-3 h-3 rounded-full bg-white/10" />
        <div class="w-3 h-3 rounded-full bg-white/10" />
        <div class="flex-1 mx-4 h-5 rounded bg-white/5 flex items-center px-3">
          <span class="text-white/20 text-xs font-mono">app.chatcoal.com</span>
        </div>
      </div>

      <!-- App layout -->
      <div class="flex" style="height: 340px;">
        <!-- Server sidebar (68px) -->
        <div class="w-14 flex flex-col items-center py-3 gap-2 border-r border-white/5" style="background: #0A0A0A;">
          <!-- Logo icon -->
          <div class="w-9 h-9 rounded-xl flex items-center justify-center mb-1" style="background: #1E1F22;">
            <svg viewBox="0 0 1024 1024" fill="none" class="w-5 h-5">
              <path d="M416 87.4376C416 48.1063 466.688 27.9491 497.216 55.1681C591.2 139.015 645.824 323.113 581.216 462.784L577.376 470.562L577.952 470.695C607.952 476.64 635.696 451.476 688.496 373.573L695.216 363.562C699.11 357.724 704.399 352.802 710.669 349.177C716.939 345.552 724.022 343.322 731.368 342.66C738.714 341.999 746.125 342.923 753.025 345.362C759.925 347.802 766.129 351.689 771.152 356.724C835.184 420.861 896 552.353 896 636.96C896 827.583 723.632 981.333 512 981.333C459.866 981.333 410.116 972.003 364.737 955.089C323.716 970.987 278.893 981.333 237.565 981.333C215.984 981.333 205.821 963.398 219.66 945.112C230.873 931.204 244.349 912.826 256.19 893.798C177.574 830.801 128 739.167 128 636.915C128 536.263 177.056 426.134 254.336 355.293L283.376 328.968C294.944 318.42 304.208 309.75 313.04 301.079C381.68 233.501 416 170.391 416 87.4376Z" fill="url(#mockupGrad)"/>
              <path d="M374.519 824.835C374.519 824.835 372.693 823.579 369.778 821.433C326.092 788.898 298.667 741.752 298.667 689.262C298.667 591.407 394.193 512 512 512C629.807 512 725.333 591.407 725.333 689.262C725.333 787.163 629.807 863.831 512 863.831C501.95 863.831 485.452 863.191 462.507 861.913C432.593 880.635 388.93 896 350.72 896C338.892 896 333.322 886.639 340.907 877.096C352.427 863.488 368.308 841.684 374.471 824.812L374.519 824.835Z" fill="#0A0A0A"/>
              <defs>
                <linearGradient id="mockupGrad" x1="512" y1="42.66" x2="512" y2="981.333" gradientUnits="userSpaceOnUse">
                  <stop stop-color="#FBB500"/><stop offset="1" stop-color="#FD4A13"/>
                </linearGradient>
              </defs>
            </svg>
          </div>
          <div class="w-px h-4 bg-white/10 rounded" />
          <div v-for="n in 4" :key="n"
               class="w-9 h-9 rounded-2xl hover:rounded-xl transition-all"
               :style="{ background: n === 1 ? '#E8521A22' : '#1C1C1C', border: n === 1 ? '1px solid #E8521A44' : '1px solid transparent' }" />
          <div class="mt-auto w-9 h-9 rounded-full bg-coal-600 border border-white/10" />
        </div>

        <!-- Channel sidebar (180px) -->
        <div class="w-44 flex flex-col border-r border-white/5 py-3" style="background: #111;">
          <div class="px-3 mb-3">
            <div class="text-xs font-semibold text-white/70 font-display px-1 mb-1">General</div>
            <div class="text-xs text-white/30 px-1 text-[10px]">Community server</div>
          </div>
          <!-- Channels -->
          <div class="px-2 space-y-0.5">
            <div class="text-[10px] text-white/30 px-2 pt-2 pb-1 uppercase tracking-wider font-semibold">Text Channels</div>
            <div v-for="ch in channels" :key="ch.name"
                 class="flex items-center gap-1.5 px-2 py-1 rounded-md cursor-pointer transition-colors"
                 :class="ch.active ? 'bg-white/10 text-white/90' : 'text-white/40 hover:text-white/60'">
              <span class="text-white/30 text-xs">#</span>
              <span class="text-xs">{{ ch.name }}</span>
            </div>
            <div class="text-[10px] text-white/30 px-2 pt-3 pb-1 uppercase tracking-wider font-semibold">Voice</div>
            <div v-for="vc in voiceChannels" :key="vc.name"
                 class="flex items-center gap-1.5 px-2 py-1 rounded-md text-white/40">
              <span class="text-white/30 text-xs">🔊</span>
              <span class="text-xs">{{ vc.name }}</span>
            </div>
          </div>
          <!-- User bar -->
          <div class="mt-auto mx-2 p-2 rounded-lg flex items-center gap-2" style="background: #0D0D0D;">
            <div class="w-6 h-6 rounded-full bg-ember/30 border border-ember/40 flex items-center justify-center">
              <span class="text-[8px] font-bold text-ember">L</span>
            </div>
            <div class="flex-1 min-w-0">
              <div class="text-[10px] font-semibold text-white/80 leading-none mb-0.5">luis</div>
              <div class="text-[8px] text-white/30 leading-none">Online</div>
            </div>
          </div>
        </div>

        <!-- Chat area -->
        <div class="flex-1 flex flex-col">
          <!-- Channel header -->
          <div class="flex items-center gap-2 px-4 py-2.5 border-b border-white/5">
            <span class="text-white/50 font-bold">#</span>
            <span class="text-sm font-semibold text-white/80">general</span>
            <div class="w-px h-4 bg-white/10 mx-1" />
            <span class="text-xs text-white/30">The main hub for announcements</span>
          </div>

          <!-- Messages -->
          <div class="flex-1 overflow-hidden px-4 pt-3 space-y-3">
            <div v-for="msg in messages" :key="msg.id" class="flex gap-2.5">
              <div class="w-7 h-7 rounded-full flex-shrink-0 flex items-center justify-center text-[10px] font-bold"
                   :style="{ background: msg.color + '33', color: msg.color }">
                {{ msg.avatar }}
              </div>
              <div>
                <div class="flex items-baseline gap-2 mb-0.5">
                  <span class="text-xs font-semibold" :style="{ color: msg.color }">{{ msg.name }}</span>
                  <span class="text-[9px] text-white/20">{{ msg.time }}</span>
                </div>
                <p class="text-xs text-white/60 leading-relaxed">{{ msg.content }}</p>
                <div v-if="msg.reaction" class="mt-1 inline-flex items-center gap-1 bg-white/5 border border-white/10 rounded px-1.5 py-0.5">
                  <span class="text-[10px]">{{ msg.reaction }}</span>
                  <span class="text-[9px] text-white/40">{{ msg.reactionCount }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Input -->
          <div class="px-4 pb-3">
            <div class="bg-coal-700 rounded-xl px-4 py-2.5 flex items-center gap-2 border border-white/5">
              <span class="text-xs text-white/25 flex-1">Message #general</span>
              <svg class="w-3.5 h-3.5 text-white/20" viewBox="0 0 16 16" fill="none">
                <path d="M14 8H2M8 2L2 8L8 14" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
            </div>
          </div>
        </div>

        <!-- Member list -->
        <div class="w-36 border-l border-white/5 py-3 px-2 hidden lg:block" style="background: #111;">
          <div class="text-[10px] text-white/30 px-2 pb-2 uppercase tracking-wider font-semibold">Online — 4</div>
          <div v-for="m in members" :key="m.name" class="flex items-center gap-2 px-2 py-1">
            <div class="relative">
              <div class="w-6 h-6 rounded-full flex items-center justify-center text-[9px] font-bold"
                   :style="{ background: m.color + '33', color: m.color }">
                {{ m.avatar }}
              </div>
              <div class="absolute -bottom-0.5 -right-0.5 w-2.5 h-2.5 rounded-full border border-coal-900"
                   :class="m.status === 'online' ? 'bg-green-500' : 'bg-yellow-500'" />
            </div>
            <span class="text-[10px] text-white/50">{{ m.name }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
const channels = [
  { name: 'general', active: true },
  { name: 'announcements', active: false },
  { name: 'off-topic', active: false },
  { name: 'dev-chat', active: false },
]

const voiceChannels = [
  { name: 'Lounge' },
  { name: 'Gaming' },
]

const messages = [
  { id: 1, name: 'kai', avatar: 'K', color: '#E8521A', time: 'Today at 2:14 PM', content: 'just shipped the new voice feature, works flawlessly 🔥', reaction: '🔥', reactionCount: 5 },
  { id: 2, name: 'mira', avatar: 'M', color: '#7B9EFF', time: 'Today at 2:16 PM', content: 'tested it — latency is incredible compared to what we had before' },
  { id: 3, name: 'theo', avatar: 'T', color: '#4AC99B', time: 'Today at 2:18 PM', content: 'open source AND this fast? chatcoal is the real deal' },
]

const members = [
  { name: 'kai', avatar: 'K', color: '#E8521A', status: 'online' },
  { name: 'mira', avatar: 'M', color: '#7B9EFF', status: 'online' },
  { name: 'theo', avatar: 'T', color: '#4AC99B', status: 'online' },
  { name: 'alex', avatar: 'A', color: '#C97BFF', status: 'idle' },
]
</script>
