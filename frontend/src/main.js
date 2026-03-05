import './assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import { useAuthStore } from '@/stores/auth'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)

// Initialize auth before installing the router so the initial navigation's
// beforeEach guards run with a fully resolved auth state.
const authStore = useAuthStore()
authStore.init().then(() => {
  app.use(router)
  app.mount('#app')
})
