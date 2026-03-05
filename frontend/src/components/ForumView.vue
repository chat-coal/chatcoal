<script setup>
import { ref, watch, computed } from 'vue'
import { useChannelsStore } from '@/stores/channels'
import { useForumStore } from '@/stores/forum'
import { useAuthStore } from '@/stores/auth'
import { useServersStore } from '@/stores/servers'
import { getAvatarColor, getDefaultAvatarStyle, resolveFileUrl, cssBackgroundUrl } from '@/utils/avatar'
import ForumPostView from './ForumPostView.vue'

const channelsStore = useChannelsStore()
const forumStore = useForumStore()
const authStore = useAuthStore()
const serversStore = useServersStore()

const showNewPost = ref(false)
const newTitle = ref('')
const newContent = ref('')
const showDeleteConfirm = ref(null)
const editingPost = ref(null)
const editTitle = ref('')
const editContent = ref('')

watch(
  () => channelsStore.currentChannel,
  (channel) => {
    if (channel?.type === 'forum') {
      forumStore.clear()
      forumStore.fetchPosts(channel.id)
    }
  },
  { immediate: true },
)

const canManage = computed(() => serversStore.canManageMessages)

async function createPost() {
  if (!newTitle.value.trim() || !newContent.value.trim()) return
  await forumStore.createPost(channelsStore.currentChannel.id, newTitle.value.trim(), newContent.value.trim())
  newTitle.value = ''
  newContent.value = ''
  showNewPost.value = false
}

function openPost(post) {
  forumStore.selectPost(post)
}

function openEditPost(post) {
  editingPost.value = post
  editTitle.value = post.title
  editContent.value = post.content
}

async function confirmEditPost() {
  if (!editTitle.value.trim() || !editContent.value.trim()) return
  await forumStore.editPost(editingPost.value.id, editTitle.value.trim(), editContent.value.trim())
  editingPost.value = null
}

async function confirmDeletePost(postId) {
  await forumStore.deletePost(postId)
  showDeleteConfirm.value = null
}

function formatTime(dateStr) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  const today = new Date()
  if (d.toDateString() === today.toDateString()) {
    return 'Today at ' + d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }
  return d.toLocaleDateString() + ' ' + d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

</script>

<template>
  <!-- If viewing a post, show post view -->
  <ForumPostView v-if="forumStore.currentPost" />

  <!-- Otherwise show post list -->
  <div v-else class="flex-1 flex flex-col min-w-0 overflow-hidden">
    <!-- Action bar -->
    <div class="px-5 py-3 border-b border-[var(--surface-border)]">
      <button
        @click="showNewPost = !showNewPost"
        class="bg-[#E8521A] text-white px-4 py-2 rounded-xl text-sm font-semibold hover:bg-[#D44818] cursor-pointer transition-colors duration-150 shadow-lg shadow-[#E8521A]/15"
      >
        New Post
      </button>
    </div>

    <!-- New post form -->
    <div v-if="showNewPost" class="px-5 py-4 border-b border-[var(--surface-border)] animate-fade-in-up">
      <input
        v-model="newTitle"
        placeholder="Post title"
        class="w-full bg-[var(--card)] text-[var(--text-1)] px-4 py-2.5 rounded-xl border border-[var(--surface-border)] text-sm mb-2 outline-none"
      />
      <textarea
        v-model="newContent"
        placeholder="What do you want to discuss?"
        rows="3"
        class="w-full bg-[var(--card)] text-[var(--text-1)] px-4 py-2.5 rounded-xl border border-[var(--surface-border)] text-sm resize-none outline-none"
      ></textarea>
      <div class="flex justify-end gap-2 mt-2">
        <button @click="showNewPost = false; newTitle = ''; newContent = ''" class="text-[var(--text-3)] hover:text-[var(--text-1)] px-3 py-1.5 rounded-lg cursor-pointer text-sm font-medium transition-colors duration-150">
          Cancel
        </button>
        <button
          @click="createPost"
          :disabled="!newTitle.trim() || !newContent.trim()"
          class="bg-[#E8521A] text-white px-4 py-1.5 rounded-lg text-sm font-semibold hover:bg-[#D44818] cursor-pointer transition-colors duration-150 disabled:opacity-40 disabled:cursor-not-allowed"
        >
          Create Post
        </button>
      </div>
    </div>

    <!-- Post list -->
    <div class="flex-1 overflow-y-auto px-5 py-4 scrollbar-light">
      <div v-if="forumStore.posts.length > 0" class="space-y-2">
        <div
          v-for="post in forumStore.posts"
          :key="post.id"
          class="group"
        >
          <button
            @click="openPost(post)"
            class="w-full text-left p-4 rounded-xl bg-[var(--card)] hover:bg-[var(--surface-2)] border border-[var(--surface-border)] cursor-pointer transition-colors duration-100"
          >
            <div class="flex items-start justify-between gap-3">
              <div class="min-w-0 flex-1">
                <h4 class="text-[var(--text-1)] font-semibold text-sm mb-1 truncate">{{ post.title }}</h4>
                <p class="text-[var(--text-3)] text-xs leading-relaxed line-clamp-2 mb-2">{{ post.content }}</p>
                <div class="flex items-center gap-3 text-[10px] text-[var(--text-4)]">
                  <div class="flex items-center gap-1.5">
                    <div
                      v-if="post.author?.avatar_url"
                      class="w-4 h-4 rounded-full bg-cover bg-center"
                      :style="{ backgroundImage: cssBackgroundUrl(resolveFileUrl(post.author.avatar_url)) }"
                    ></div>
                    <div
                      v-else
                      class="w-4 h-4 rounded-full flex items-center justify-center text-white text-[8px] font-bold"
                      :style="getDefaultAvatarStyle(post.author_id)"
                    >
                      {{ (post.author?.display_name || '?')[0].toUpperCase() }}
                    </div>
                    <span>{{ post.author?.display_name || 'Unknown' }}</span>
                  </div>
                  <span>{{ post.reply_count }} {{ post.reply_count === 1 ? 'reply' : 'replies' }}</span>
                  <span v-if="post.last_reply_at">Last reply {{ formatTime(post.last_reply_at) }}</span>
                  <span v-else>{{ formatTime(post.created_at) }}</span>
                </div>
              </div>
              <!-- Edit + Delete buttons -->
              <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-all duration-100">
                <button
                  v-if="post.author_id === authStore.dbUser?.id"
                  @click.stop="openEditPost(post)"
                  class="p-1.5 text-[var(--text-4)] hover:text-[var(--text-1)] hover:bg-[var(--surface-2)] rounded-lg cursor-pointer"
                  title="Edit post"
                >
                  <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L10.582 16.07a4.5 4.5 0 0 1-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 0 1 1.13-1.897l8.932-8.931Zm0 0L19.5 7.125" />
                  </svg>
                </button>
                <button
                  v-if="post.author_id === authStore.dbUser?.id || canManage"
                  @click.stop="showDeleteConfirm = post.id"
                  class="p-1.5 text-[var(--text-4)] hover:text-[#E8521A] hover:bg-[#E8521A]/10 rounded-lg cursor-pointer"
                  title="Delete post"
                >
                  <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0" />
                  </svg>
                </button>
              </div>
            </div>
          </button>
        </div>

        <button
          v-if="forumStore.hasMorePosts"
          @click="forumStore.fetchPosts(channelsStore.currentChannel.id, forumStore.posts[forumStore.posts.length - 1]?.id)"
          class="text-[#E8521A] text-sm hover:text-[#D44818] cursor-pointer font-medium"
        >
          Load more posts
        </button>
      </div>

      <div v-else-if="!forumStore.loading" class="text-center mt-16">
        <div class="w-16 h-16 rounded-2xl bg-[#E8521A]/10 flex items-center justify-center mx-auto mb-4">
          <svg class="w-7 h-7 text-[#E8521A]" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" d="M20 13V6a2 2 0 0 0-2-2H6a2 2 0 0 0-2 2v7m16 0v1a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2v-1m16 0h-2.586a1 1 0 0 0-.707.293l-2.414 2.414a1 1 0 0 1-.707.293h-3.172a1 1 0 0 1-.707-.293l-2.414-2.414A1 1 0 0 0 6.586 13H4" />
          </svg>
        </div>
        <p class="text-[var(--text-3)] font-medium">No posts yet</p>
        <p class="text-[var(--text-4)] text-sm mt-1">Create a post to start a discussion!</p>
      </div>
    </div>
  </div>

  <!-- Edit post modal -->
  <Teleport to="body">
    <div v-if="editingPost" class="fixed inset-0 bg-[var(--backdrop)] backdrop-blur-sm flex items-center justify-center z-[60]" @click.self="editingPost = null">
      <div class="bg-[var(--modal-bg)] rounded-2xl p-7 w-full max-w-md shadow-2xl shadow-black/10 animate-scale-in border border-[var(--modal-border)]">
        <h2 class="font-display text-xl font-bold text-[var(--text-1)] mb-4">Edit Post</h2>
        <input
          v-model="editTitle"
          placeholder="Post title"
          maxlength="200"
          class="w-full bg-[var(--card)] text-[var(--text-1)] px-4 py-2.5 rounded-xl border border-[var(--surface-border)] text-sm mb-3 outline-none"
        />
        <textarea
          v-model="editContent"
          placeholder="Post content"
          rows="4"
          class="w-full bg-[var(--card)] text-[var(--text-1)] px-4 py-2.5 rounded-xl border border-[var(--surface-border)] text-sm resize-none outline-none"
        ></textarea>
        <div class="flex justify-end gap-3 mt-4">
          <button @click="editingPost = null" class="text-[var(--text-3)] hover:text-[var(--text-1)] px-4 py-2.5 rounded-xl cursor-pointer font-medium transition-colors duration-150">
            Cancel
          </button>
          <button
            @click="confirmEditPost"
            :disabled="!editTitle.trim() || !editContent.trim()"
            class="bg-[#E8521A] text-white px-5 py-2.5 rounded-xl hover:bg-[#D44818] cursor-pointer font-semibold shadow-lg shadow-[#E8521A]/15 transition-all duration-200 disabled:opacity-40 disabled:cursor-not-allowed"
          >
            Save
          </button>
        </div>
      </div>
    </div>
  </Teleport>

  <!-- Delete confirmation modal -->
  <Teleport to="body">
    <div v-if="showDeleteConfirm" class="fixed inset-0 bg-[var(--backdrop)] backdrop-blur-sm flex items-center justify-center z-[60]" @click.self="showDeleteConfirm = null">
      <div class="bg-[var(--modal-bg)] rounded-2xl p-7 w-full max-w-sm shadow-2xl shadow-black/10 animate-scale-in border border-[var(--modal-border)]">
        <h2 class="font-display text-xl font-bold text-[var(--text-1)] mb-2">Delete Post</h2>
        <p class="text-[var(--text-3)] text-sm mb-5">
          Are you sure you want to delete this post and all its replies? This cannot be undone.
        </p>
        <div class="flex justify-end gap-3">
          <button @click="showDeleteConfirm = null" class="text-[var(--text-3)] hover:text-[var(--text-1)] px-4 py-2.5 rounded-xl cursor-pointer font-medium transition-colors duration-150">
            Cancel
          </button>
          <button
            @click="confirmDeletePost(showDeleteConfirm)"
            class="bg-[#E8521A] text-white px-5 py-2.5 rounded-xl hover:bg-[#D44818] cursor-pointer font-semibold shadow-lg shadow-[#E8521A]/15 transition-all duration-200"
          >
            Delete
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>
