<script setup>
import { watch } from 'vue'
import { useRoute } from 'vue-router'
import { useDMsStore } from '@/stores/dms'
import { useServersStore } from '@/stores/servers'

const route = useRoute()
const dmsStore = useDMsStore()
const serversStore = useServersStore()

function selectFromRoute() {
  serversStore.selectServer(null)
  const dmChannelId = route.params.dmChannelId
  const channel = dmsStore.dmChannels.find((c) => c.id === dmChannelId)
  if (channel) {
    dmsStore.selectDM(channel)
  }
}

selectFromRoute()

watch(() => route.params.dmChannelId, selectFromRoute)
</script>

<template>
  <div />
</template>
