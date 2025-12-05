import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import { createMemoryHistory, createRouter } from 'vue-router'
import Landing from '../../src/components/Landing.vue'

// Mock useAuth
const loginAsDummyMock = vi.fn()
vi.mock('@/composables/useAuth', () => ({
  useAuth: () => ({
    loginAsDummy: loginAsDummyMock,
  }),
}))

describe('landing Page', () => {
  it('renders header and handles Dev login', async () => {
    // Mock env var
    vi.stubEnv('ARCH_STATS_DEV_MODE', 'true')

    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        {
          path: '/',
          name: 'landing',
          component: Landing,
        },
      ],
    })

    router.push('/')
    await router.isReady()

    const wrapper = mount(Landing, {
      global: {
        plugins: [router],
      },
    })

    expect(wrapper.text()).toContain('Arch Stats')

    // Verify Dev Mode Button
    const button = wrapper.find('button')
    expect(button.exists()).toBe(true)
    expect(button.text()).toBe('Login as Dummy (Dev Only)')

    // Verify Click Action
    await button.trigger('click')
    expect(loginAsDummyMock).toHaveBeenCalled()
  })
})
