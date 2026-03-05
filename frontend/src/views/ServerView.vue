<script setup>
import { onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useServersStore } from '@/stores/servers'
import { useChannelsStore } from '@/stores/channels'

const route = useRoute()
const router = useRouter()
const serversStore = useServersStore()
const channelsStore = useChannelsStore()

async function loadFromRoute() {
  const serverId = route.params.serverId
  const channelId = route.params.channelId || null

  if (serverId && (!serversStore.currentServer || serversStore.currentServer.id !== serverId)) {
    const server = serversStore.servers.find((s) => s.id === serverId)
    if (server) {
      serversStore.selectServer(server)
    }
  }

  if (channelId && (!channelsStore.currentChannel || channelsStore.currentChannel.id !== channelId)) {
    await new Promise((r) => setTimeout(r, 100))
    const channel = channelsStore.channels.find((c) => c.id === channelId)
    if (channel) {
      channelsStore.selectChannel(channel)
    }
  }
}

onMounted(loadFromRoute)

watch(() => route.params, loadFromRoute)
</script>

<template>
  <!-- ServerView is a route handler; rendering happens in AppLayout -->
  <div></div>
</template>
