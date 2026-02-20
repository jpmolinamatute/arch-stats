import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import ConfirmModal from '@/components/widgets/ConfirmModal.vue'

describe('confirmModal', () => {
    it('renders when show is true', () => {
        const wrapper = mount(ConfirmModal, {
            props: {
                show: true,
                title: 'Test Title',
                message: 'Test Message',
            },
            global: {
                stubs: {
                    Teleport: true,
                    Transition: true,
                },
            },
        })

        expect(wrapper.text()).toContain('Test Title')
        expect(wrapper.text()).toContain('Test Message')
    })

    it('has standard accessibility attributes when shown', () => {
        const wrapper = mount(ConfirmModal, {
            props: { show: true },
            global: { stubs: { Teleport: true, Transition: true } },
        })

        const dialog = wrapper.find('div[role="dialog"]')
        expect(dialog.exists()).toBe(true)
        expect(dialog.attributes('aria-modal')).toBe('true')
    })

    it('does not render when show is false', () => {
        const wrapper = mount(ConfirmModal, {
            props: {
                show: false,
            },
            global: {
                stubs: {
                    Teleport: true,
                    Transition: true,
                },
            },
        })

        expect(wrapper.find('div[role="dialog"]').exists()).toBe(false)
    })

    it('emits confirm event when confirm button is clicked', async () => {
        const wrapper = mount(ConfirmModal, {
            props: {
                show: true,
            },
            global: {
                stubs: {
                    Teleport: true,
                    Transition: true,
                },
            },
        })

        const confirmButton = wrapper.findAll('button').find(b => b.text() === 'Confirm')
        await confirmButton?.trigger('click')

        expect(wrapper.emitted('confirm')).toBeTruthy()
    })

    it('emits cancel event when cancel button is clicked', async () => {
        const wrapper = mount(ConfirmModal, {
            props: {
                show: true,
            },
            global: {
                stubs: {
                    Teleport: true,
                    Transition: true,
                },
            },
        })

        const cancelButton = wrapper.findAll('button').find(b => b.text() === 'Cancel')
        await cancelButton?.trigger('click')

        expect(wrapper.emitted('cancel')).toBeTruthy()
    })
})
