import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import MiniTable from '@/components/widgets/MiniTable.vue'

describe('miniTable', () => {
    const mockShots = [
        { score: 10, x: 0, y: 0, is_x: false, color: 'yellow' },
        { score: 9, x: 0, y: 0, is_x: false, color: 'yellow' },
        { score: 10, x: 0, y: 0, is_x: true, color: 'yellow' }, // X
    ]
    const maxShots = 6

    it('renders correct number of slots based on maxShots', () => {
        const wrapper = mount(MiniTable, {
            props: {
                shots: [],
                maxShots,
            },
        })

        // Should have maxShots header slots + 1 empty
        // Header row
        const headerRow = wrapper.findAll('.grid').at(0)
        expect(headerRow?.findAll('div.text-lg').length).toBe(maxShots)

        // Slots row
        const slotsRow = wrapper.findAll('.grid').at(1)
        // Excluding the action buttons container, we count children divs minus the actions div?
        // Actually the structure is:
        // div.grid (row 2)
        //   div (slot 1) ... div (slot 6)
        //   div (actions)
        // So direct children of the second grid div should be maxShots + 1
        expect(slotsRow?.element.children.length).toBe(maxShots + 1)
    })

    it('renders shots with correct values and styles', () => {
        const wrapper = mount(MiniTable, {
            props: {
                shots: mockShots,
                maxShots,
            },
        })

        const buttons = wrapper.findAll('button.shadow-sm') // Shot buttons have shadow-sm class
        expect(buttons.length).toBe(3)

        // Check content
        expect(buttons[0].text()).toBe('10')
        expect(buttons[1].text()).toBe('9')
        expect(buttons[2].text()).toBe('X')

        // Check styles (colors)
        // Note: styles are applied via :style binding, text color is computed
        expect(buttons[0].attributes('style')).toContain('background-color: yellow')
    })

    it('emits delete event when clicking a shot', async () => {
        const wrapper = mount(MiniTable, {
            props: {
                shots: mockShots,
                maxShots,
            },
        })

        const buttons = wrapper.findAll('button.shadow-sm')
        await buttons[1].trigger('click')

        expect(wrapper.emitted('delete')).toBeTruthy()
        expect(wrapper.emitted('delete')![0]).toEqual([1]) // index 1
    })

    it('handles clear button state and emit', async () => {
        const wrapper = mount(MiniTable, {
            props: {
                shots: mockShots,
                maxShots,
            },
        })

        const clearBtn = wrapper.findAll('button').filter(b => b.text() === 'CLR')[0]
        expect(clearBtn.attributes('disabled')).toBeUndefined()

        await clearBtn.trigger('click')
        expect(wrapper.emitted('clear')).toBeTruthy()

        // Test disabled state
        await wrapper.setProps({ shots: [] })
        expect(clearBtn.attributes('disabled')).toBeDefined()
    })

    it('handles confirm button state and emit', async () => {
        const wrapper = mount(MiniTable, {
            props: {
                shots: mockShots, // 3 shots
                maxShots, // 6
            },
        })

        const confirmBtn = wrapper.findAll('button').filter(b => b.text() === 'OK')[0]

        // Should be disabled because shots (3) < maxShots (6)
        expect(confirmBtn.attributes('disabled')).toBeDefined()

        // Fill up shots
        const fullShots: Array<{ score: number, x: number, y: number, is_x: boolean, color: string }> = Array.from({ length: 6 }, () => ({
            score: 10,
            x: 0,
            y: 0,
            is_x: false,
            color: 'yellow',
        }))
        await wrapper.setProps({ shots: fullShots })

        expect(confirmBtn.attributes('disabled')).toBeUndefined()

        await confirmBtn.trigger('click')
        expect(wrapper.emitted('confirm')).toBeTruthy()
    })
})
