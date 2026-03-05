import { createRouter, createWebHistory, createWebHashHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import AppLayout from '@/views/AppLayout.vue'

const router = createRouter({
  history: window.electronAPI?.isElectron
    ? createWebHashHistory()
    : createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
    },
    {
      path: '/onboarding',
      name: 'onboarding',
      component: () => import('@/views/OnboardingView.vue'),
    },
    {
      path: '/invite/:code',
      name: 'invite',
      component: () => import('@/views/InviteView.vue'),
    },
    {
      path: '/federation/callback',
      name: 'federation-callback',
      component: () => import('@/views/FederationCallbackView.vue'),
    },
    {
      path: '/federation/authorize',
      name: 'federation-authorize',
      component: () => import('@/views/FederationAuthorizeView.vue'),
    },
    {
      path: '/admin',
      name: 'admin',
      component: () => import('@/views/AdminView.vue'),
    },
    {
      path: '/channels',
      component: AppLayout,
      children: [
        {
          path: '@me',
          name: 'home',
          component: () => import('@/views/ServerView.vue'),
        },
        {
          path: '@me/:dmChannelId',
          name: 'dm',
          component: () => import('@/views/DMView.vue'),
        },
        {
          path: ':serverId/:channelId?',
          name: 'server',
          component: () => import('@/views/ServerView.vue'),
        },
      ],
    },
    {
      path: '/',
      redirect: () => {
        const saved = localStorage.getItem('lastChannel')
        return saved && saved.startsWith('/channels/') ? saved : '/channels/@me'
      },
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('@/views/NotFoundView.vue'),
    },
  ],
})

const PUBLIC_ROUTES = ['login', 'invite', 'federation-callback', 'federation-authorize', 'not-found']

router.beforeEach((to) => {
  const auth = useAuthStore()

  // Still initializing — let it through; guards will re-run after load
  if (auth.loading) return

  // A federated session counts as authenticated.
  const isAuthenticated = auth.isFederated ? !!auth.dbUser : !!auth.user

  // Not authenticated: send to login
  if (!isAuthenticated && !PUBLIC_ROUTES.includes(to.name)) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }

  // Federated users: skip onboarding (they set up their profile on their home instance).
  if (auth.isFederated) {
    if (auth.dbUser?.display_name && to.name === 'login') {
      return { name: 'home' }
    }
    return
  }

  // Authenticated but no username yet: send to onboarding
  // (allow login/onboarding/invite through so they can complete auth)
  // Anonymous users always get a display name assigned by the backend — never redirect them.
  if (auth.user && !auth.user.isAnonymous && !auth.dbUser?.username && to.name !== 'login' && to.name !== 'onboarding' && to.name !== 'invite' && to.name !== 'admin') {
    return { name: 'onboarding' }
  }

  // Already set up: skip onboarding
  // Anon users are set up once they have a display_name; non-anon users need a username.
  const isSetUp = auth.user?.isAnonymous ? !!auth.dbUser?.display_name : !!auth.dbUser?.username
  if (auth.user && isSetUp && to.name === 'onboarding') {
    return { name: 'home' }
  }

  // Authenticated and set up: skip login
  if (auth.user && isSetUp && to.name === 'login') {
    return { name: 'home' }
  }
})

router.afterEach((to) => {
  if (to.path.startsWith('/channels/')) {
    localStorage.setItem('lastChannel', to.fullPath)
  }
})

export default router
