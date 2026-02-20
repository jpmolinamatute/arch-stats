import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import LiveStatsTable from '@/components/widgets/LiveStatsTable.vue'

describe('liveStatsTable.vue', () => {
    it('renders "No stats available." when stats is null', () => {
        const wrapper = mount(LiveStatsTable, {
            props: {
                stats: null,
                xCount: 0,
            },
        })
        expect(wrapper.text()).toContain('No stats available.')
    })

    it('renders stats correctly when provided', () => {
        const wrapper = mount(LiveStatsTable, {
            props: {
                stats: {
                    slot_id: 'dummy_slot_123',
                    number_of_shots: 10,
                    total_score: 95,
                    max_score: 10,
                    mean: 9.5,
                },
                xCount: 3,
            },
        })

        const tds = wrapper.findAll('td')
        expect(tds.length).toBe(5)
        expect(tds[0].text()).toBe('10') // number of shots
        expect(tds[1].text()).toBe('95') // total score
        expect(tds[2].text()).toBe('10') // max score
        expect(tds[3].text()).toBe('9.50') // mean (.toFixed(2))
        expect(tds[4].text()).toBe('3') // X count
    })
})
