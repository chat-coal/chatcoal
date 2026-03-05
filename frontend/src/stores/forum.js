import { ref } from 'vue'
import { defineStore } from 'pinia'
import api from '@/services/api.service'
import { useAuthStore } from '@/stores/auth'

export const useForumStore = defineStore('forum', () => {
  const posts = ref([])
  const currentPost = ref(null)
  const messages = ref([])
  const loading = ref(false)
  const hasMorePosts = ref(true)
  const hasMoreMessages = ref(true)
  const messagesLoading = ref(false)
  const replyingTo = ref(null)
  let tempIdCounter = 0

  async function fetchPosts(channelId, before = null) {
    loading.value = true
    try {
      const data = await api.getForumPosts(channelId, before)
      if (before) {
        posts.value = [...posts.value, ...data]
      } else {
        posts.value = data
      }
      hasMorePosts.value = data.length === 25
    } catch {
      posts.value = []
    }
    loading.value = false
  }

  async function createPost(channelId, title, content) {
    const post = await api.createForumPost(channelId, title, content)
    addPost(post)
    return post
  }

  async function deletePost(postId) {
    await api.deleteForumPost(postId)
    posts.value = posts.value.filter((p) => p.id !== postId)
    if (currentPost.value?.id === postId) {
      currentPost.value = null
      messages.value = []
    }
  }

  function selectPost(post) {
    currentPost.value = post
    messages.value = []
    hasMoreMessages.value = true
    replyingTo.value = null
  }

  function goBackToList() {
    currentPost.value = null
    messages.value = []
    replyingTo.value = null
  }

  async function fetchMessages(postId, before = null) {
    messagesLoading.value = true
    try {
      const data = await api.getForumPostMessages(postId, before)
      if (before) {
        messages.value = [...data.reverse(), ...messages.value]
      } else {
        messages.value = data.reverse()
      }
      hasMoreMessages.value = data.length === 50
    } catch {
      messages.value = []
    }
    messagesLoading.value = false
  }

  async function sendMessage(postId, content, replyToId = null) {
    const authStore = useAuthStore()
    const tempId = `_temp_${Date.now()}_${++tempIdCounter}`
    const optimistic = {
      id: tempId,
      content: content || '',
      author_id: authStore.dbUser?.id,
      author: authStore.dbUser ? { ...authStore.dbUser } : null,
      forum_post_id: postId,
      created_at: new Date().toISOString(),
      reply_to_id: replyToId || null,
      reply_to: replyToId && replyingTo.value ? {
        id: replyingTo.value.id,
        content: replyingTo.value.content,
        author: replyingTo.value.author,
      } : null,
      reactions: [],
      _sending: true,
    }
    messages.value.push(optimistic)

    try {
      const msg = await api.sendForumPostMessage(postId, content, replyToId)
      if (messages.value.find(m => m.id === tempId)) {
        messages.value = messages.value.filter(m => m.id !== tempId)
        addMessage(msg)
      }
      return msg
    } catch (err) {
      const idx = messages.value.findIndex(m => m.id === tempId)
      if (idx !== -1) {
        messages.value[idx] = { ...messages.value[idx], _sending: false, _failed: true, _error: err?.response?.data?.error || 'Failed to send' }
      }
    }
  }

  async function editPost(postId, title, content) {
    const post = await api.editForumPost(postId, title, content)
    updatePost(post)
    return post
  }

  function updatePost(post) {
    const idx = posts.value.findIndex((p) => p.id === post.id)
    if (idx !== -1) posts.value[idx] = post
    if (currentPost.value?.id === post.id) currentPost.value = post
  }

  function addMessage(msg) {
    if (!messages.value.find((m) => m.id === msg.id)) {
      messages.value.push(msg)
    }
  }

  function updateMessage(msg) {
    const idx = messages.value.findIndex((m) => m.id === msg.id)
    if (idx !== -1) {
      messages.value[idx] = msg
    }
  }

  function removeMessage(id, forumPostId) {
    messages.value = messages.value.filter((m) => m.id !== id)
    // Decrement reply count
    const postId = forumPostId || currentPost.value?.id
    if (postId) {
      if (currentPost.value && currentPost.value.id === postId) {
        currentPost.value = { ...currentPost.value, reply_count: Math.max(0, (currentPost.value.reply_count || 1) - 1) }
      }
      const idx = posts.value.findIndex((p) => p.id === postId)
      if (idx !== -1) {
        posts.value[idx] = { ...posts.value[idx], reply_count: Math.max(0, (posts.value[idx].reply_count || 1) - 1) }
      }
    }
  }

  function incrementPostReplyCount(postId, replyAt) {
    if (currentPost.value && currentPost.value.id === postId) {
      currentPost.value = { ...currentPost.value, reply_count: (currentPost.value.reply_count || 0) + 1 }
    }
    const idx = posts.value.findIndex((p) => p.id === postId)
    if (idx !== -1) {
      posts.value[idx] = {
        ...posts.value[idx],
        reply_count: (posts.value[idx].reply_count || 0) + 1,
        last_reply_at: replyAt,
      }
    }
  }

  function addPost(post) {
    if (!posts.value.find((p) => p.id === post.id)) {
      posts.value = [post, ...posts.value]
    }
  }

  function removePost(id) {
    posts.value = posts.value.filter((p) => p.id !== id)
    if (currentPost.value?.id === id) {
      currentPost.value = null
      messages.value = []
    }
  }

  function setReplyingTo(message) {
    replyingTo.value = message
  }

  function clearReplyingTo() {
    replyingTo.value = null
  }

  function clear() {
    posts.value = []
    currentPost.value = null
    messages.value = []
    hasMorePosts.value = true
    hasMoreMessages.value = true
    replyingTo.value = null
  }

  return {
    posts, currentPost, messages, loading, hasMorePosts, hasMoreMessages, messagesLoading, replyingTo,
    fetchPosts, createPost, editPost, deletePost, selectPost, goBackToList,
    fetchMessages, sendMessage, addMessage, updateMessage, removeMessage,
    incrementPostReplyCount, addPost, updatePost, removePost, setReplyingTo, clearReplyingTo, clear,
  }
})
