import { createRouter, createWebHistory } from 'vue-router'
import AppContainer from '@/components/AppContainer.vue'
import Landing from '@/components/Landing.vue'
import LiveSession from '@/components/LiveSession.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'landing',
      component: Landing,
    },
    {
      path: '/app',
      name: 'app',
      component: AppContainer,
    },
    {
      path: '/app/live-session',
      name: 'live-session',
      component: LiveSession,
    },
  ],
})

export default router
