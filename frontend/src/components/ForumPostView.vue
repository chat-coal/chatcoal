<script setup>
import { ref, watch, nextTick } from 'vue'
import { useForumStore } from '@/stores/forum'
import { getAvatarColor, getDefaultAvatarStyle, resolveFileUrl, cssBackgroundUrl } from '@/utils/avatar'
import MessageItem from './MessageItem.vue'
import MessageInput from './MessageInput.vue'

const forumStore = useForumStore()
const messagesContainer = ref(null)

watch(
  () => forumStore.currentPost,
  async (post) => {
    if (post) {
      await forumStore.fetchMessages(post.id)
      scrollToBottom()
    }
  },
  { immediate: true },
)

watch(() => forumStore.messages.length, scrollToBottom)

function scrollToBottom() {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

function handleReply(message) {
  forumStore.setReplyingTo(message)
}

function handleCancelReply() {
  forumStore.clearReplyingTo()
}

function scrollToMessage(messageId) {
  const el = messagesContainer.value?.querySelector(`[data-message-id="${messageId}"]`)
  if (el) {
    el.scrollIntoView({ behavior: 'smooth', block: 'center' })
    el.classList.add('bg-[var(--surface-2)]')
    setTimeout(() => el.classList.remove('bg-[var(--surface-2)]'), 1500)
  }
}

async function loadMore() {
  if (!forumStore.hasMoreMessages || forumStore.messagesLoading) return
  const oldest = forumStore.messages[0]
  if (oldest) {
    await forumStore.fetchMessages(forumStore.currentPost.id, oldest.id)
  }
}

function formatTime(dateStr) {
  const d = new Date(dateStr)
  const today = new Date()
  if (d.toDateString() === today.toDateString()) {
    return 'Today at ' + d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }
  return d.toLocaleDateString() + ' ' + d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

</script>

<template>
  <div class="flex-1 flex flex-col min-w-0 overflow-hidden">
    <!-- Post header -->
    <div class="px-5 py-3 border-b border-[var(--surface-border)]">
      <button
        @click="forumStore.goBackToList()"
        class="flex items-center gap-1.5 text-[var(--text-4)] hover:text-[var(--text-1)] cursor-pointer text-sm mb-2 transition-colors duration-100"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="M10.5 19.5 3 12m0 0 7.5-7.5M3 12h18" />
        </svg>
        Back to posts
      </button>

      <h3 class="text-[var(--text-1)] font-bold text-lg">{{ forumStore.currentPost.title }}</h3>
      <div class="flex items-center gap-2 mt-1.5">
        <div
          v-if="forumStore.currentPost.author?.avatar_url"
          class="w-5 h-5 rounded-full bg-cover bg-center"
          :style="{ backgroundImage: cssBackgroundUrl(resolveFileUrl(forumStore.currentPost.author.avatar_url)) }"
        ></div>
        <div
          v-else
          class="w-5 h-5 rounded-full flex items-center justify-center text-white text-[9px] font-bold"
          :style="getDefaultAvatarStyle(forumStore.currentPost.author_id)"
        >
          {{ (forumStore.currentPost.author?.display_name || '?')[0].toUpperCase() }}
        </div>
        <span class="text-[var(--text-2)] text-xs font-medium">{{ forumStore.currentPost.author?.display_name || 'Unknown' }}</span>
        <span class="text-[var(--text-4)] text-[10px]">{{ formatTime(forumStore.currentPost.created_at) }}</span>
      </div>
      <p class="text-[var(--text-2)] text-sm mt-2 leading-relaxed">{{ forumStore.currentPost.content }}</p>
    </div>

    <!-- Messages -->
    <div ref="messagesContainer" class="flex-1 overflow-y-auto px-5 py-5 scrollbar-light">
      <!-- Loading skeleton for initial reply load -->
      <div v-if="forumStore.messagesLoading && forumStore.messages.length === 0" class="space-y-5">
        <div v-for="i in 5" :key="i" class="flex items-start gap-3 animate-pulse">
          <div class="w-10 h-10 rounded-full bg-[var(--surface-3)] shrink-0" />
          <div class="flex-1 min-w-0 pt-0.5">
            <div class="flex items-center gap-2 mb-1.5">
              <div class="h-3.5 rounded bg-[var(--surface-3)]" :style="{ width: [90, 70, 110, 80, 100][i - 1] + 'px' }" />
              <div class="h-2.5 w-10 rounded bg-[var(--surface-3)] opacity-50" />
            </div>
            <div class="space-y-1.5">
              <div class="h-3 rounded bg-[var(--surface-3)] opacity-70" :style="{ width: [85, 60, 95, 45, 75][i - 1] + '%' }" />
              <div v-if="i % 3 !== 0" class="h-3 rounded bg-[var(--surface-3)] opacity-50" :style="{ width: [50, 70, 0, 40, 60][i - 1] + '%' }" />
            </div>
          </div>
        </div>
      </div>

      <template v-else>
        <button
          v-if="forumStore.hasMoreMessages"
          @click="loadMore"
          class="text-[#E8521A] text-sm hover:text-[#D44818] mb-5 cursor-pointer font-medium"
        >
          Load older replies
        </button>

        <div
          v-for="message in forumStore.messages"
          :key="message.id"
          :data-message-id="message.id"
        >
          <MessageItem
            :message="message"
            mode="forum"
            @reply="handleReply"
            @scroll-to-message="scrollToMessage"
          />
        </div>

        <div v-if="forumStore.messages.length === 0 && !forumStore.messagesLoading" class="text-center mt-8">
          <p class="text-[var(--text-4)] text-sm">No replies yet. Be the first to respond!</p>
        </div>
      </template>
    </div>

    <!-- Input -->
    <MessageInput
      mode="forum"
      :forum-post-id="forumStore.currentPost.id"
      :replying-to="forumStore.replyingTo"
      @cancel-reply="handleCancelReply"
    />
  </div>
</template>
