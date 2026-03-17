import { createRouter, createWebHistory } from 'vue-router'
import Home from './views/Home.vue'
import PrivacyPolicy from './views/PrivacyPolicy.vue'
import TermsOfService from './views/TermsOfService.vue'
import Download from './views/Download.vue'
import NotFound from './views/NotFound.vue'

export default createRouter({
  history: createWebHistory(),
  scrollBehavior(to, from, savedPosition) {
    if (savedPosition) return savedPosition
    if (to.hash) return { el: to.hash, behavior: 'smooth' }
    return new Promise((resolve) => {
      setTimeout(() => resolve({ top: 0 }), 0)
    })
  },
  routes: [
    { path: '/', component: Home },
    { path: '/download', component: Download },
    { path: '/privacy', component: PrivacyPolicy },
    { path: '/terms', component: TermsOfService },
    { path: '/:pathMatch(.*)*', component: NotFound },
  ],
})
