import { createRouter, createWebHistory } from 'vue-router'
import AboutPlaceholder from '@/components/AboutPlaceholder.vue'
import AppContainer from '@/components/AppContainer.vue'
import DocsPlaceholder from '@/components/DocsPlaceholder.vue'
import FeedbackPage from '@/components/FeedbackPage.vue'
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
            path: '/docs',
            name: 'docs',
            component: DocsPlaceholder,
        },
        {
            path: '/about',
            name: 'about',
            component: AboutPlaceholder,
        },
        {
            path: '/feedback',
            name: 'feedback',
            component: FeedbackPage,
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
