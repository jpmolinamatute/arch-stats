import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import NoFaceSession from '@/components/widgets/NoFaceSession.vue'

describe('noFaceSession.vue', () => {
    it('renders total shots correctly', () => {
        const wrapper = mount(NoFaceSession, {
            props: {
                totalShots: 42,
                loading: false,
            },
        })

        expect(wrapper.text()).toContain('42')
        expect(wrapper.text()).toContain('Shots')
        expect(wrapper.text()).toContain('Control')
    })

    it('emits add-shots event with valid input', async () => {
        const wrapper = mount(NoFaceSession, {
            props: {
                totalShots: 0,
                loading: false,
            },
        })

        const input = wrapper.find('input[type="number"]')
        await input.setValue(10)

        await wrapper.find('button').trigger('click')

        expect(wrapper.emitted()).toHaveProperty('addShots')
        expect(wrapper.emitted('addShots')![0]).toEqual([10])
    })

    it('disables button when input is invalid or loading', async () => {
        const wrapper = mount(NoFaceSession, {
            props: {
                totalShots: 0,
                loading: true,
            },
        })

        const button = wrapper.find('button')
        expect(button.attributes('disabled')).toBeDefined()

        await wrapper.setProps({ loading: false })
        const input = wrapper.find('input[type="number"]')
        await input.setValue(-1)
        expect(button.attributes('disabled')).toBeDefined()

        await input.setValue(5)
        expect(button.attributes('disabled')).toBeUndefined()
    })
})
