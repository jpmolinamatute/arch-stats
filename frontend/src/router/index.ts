import { createRouter, createWebHistory } from 'vue-router';
import Landing from '@/components/Landing.vue';
import AppContainer from '@/components/AppContainer.vue';
import LiveSession from '@/components/LiveSession.vue';

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
});

export default router;
