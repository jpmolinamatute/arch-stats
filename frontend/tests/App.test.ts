import { mount } from '@vue/test-utils';
import { describe, it, expect } from 'vitest';
import { createMemoryHistory, createRouter } from 'vue-router';
import App from '../src/App.vue';
import Landing from '../src/components/Landing.vue';

describe('Landing Page', () => {
    it('renders header', async () => {
        // Provide a minimal in-memory router so <RouterView> can render
        const router = createRouter({
            history: createMemoryHistory(),
            routes: [
                {
                    path: '/',
                    name: 'landing',
                    component: Landing,
                },
            ],
        });

        // Navigate to root route and wait for router to be ready
        router.push('/');
        await router.isReady();

        const wrapper = mount(App, {
            global: {
                plugins: [router],
            },
        });

        expect(wrapper.text()).toContain('Arch Stats');
    });
});
