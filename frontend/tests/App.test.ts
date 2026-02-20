import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import { createMemoryHistory, createRouter } from 'vue-router'
import App from '@/App.vue'

describe('app.vue', () => {
    it('renders router-view', async () => {
        const router = createRouter({
            history: createMemoryHistory(),
            routes: [{ path: '/', component: { template: '<div data-testid="mock-route">Mock Route</div>' } }],
        })

        router.push('/')
        await router.isReady()

        const wrapper = mount(App, {
            global: {
                plugins: [router],
            },
        })

        expect(wrapper.find('[data-testid="mock-route"]').exists()).toBe(true)
    })
})
