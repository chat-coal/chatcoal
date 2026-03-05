import { createRouter, createWebHistory } from 'vue-router'
import Home from './views/Home.vue'
import PrivacyPolicy from './views/PrivacyPolicy.vue'
import TermsOfService from './views/TermsOfService.vue'
import NotFound from './views/NotFound.vue'

export default createRouter({
  history: createWebHistory(),
  scrollBehavior(to) {
    if (to.hash) return { el: to.hash, behavior: 'smooth' }
    return { top: 0 }
  },
  routes: [
    { path: '/', component: Home },
    { path: '/privacy', component: PrivacyPolicy },
    { path: '/terms', component: TermsOfService },
    { path: '/:pathMatch(.*)*', component: NotFound },
  ],
})
