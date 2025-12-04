import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import { createMemoryHistory, createRouter } from 'vue-router'
import Landing from '../../src/components/Landing.vue'

describe('landing Page', () => {
  it('renders header', async () => {
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
  })
})
