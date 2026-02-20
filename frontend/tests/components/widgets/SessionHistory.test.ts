import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import SessionHistory from '@/components/widgets/SessionHistory.vue'

describe('sessionHistory.vue', () => {
    const mockShots = [
        { shot_id: '1', score: 9, is_x: false, x: 10, y: 10, created_at: '2023-01-01T10:00:00Z', slot_id: 's1' },
        { shot_id: '2', score: 10, is_x: true, x: 5, y: 5, created_at: '2023-01-01T10:05:00Z', slot_id: 's1' },
        { shot_id: '3', score: null, is_x: false, x: null, y: null, created_at: '2023-01-01T10:10:00Z', slot_id: 's1' },
    ]

    it('renders "No confirmed shots yet." when shots array is empty', () => {
        const wrapper = mount(SessionHistory, {
            props: { shots: [] },
        })
        expect(wrapper.text()).toContain('No confirmed shots yet.')
    })

    it('renders shots sorted by newest first', () => {
        const wrapper = mount(SessionHistory, {
            props: { shots: mockShots },
        })

        const rows = wrapper.findAll('tbody tr')
        expect(rows.length).toBe(3)

        // Newest is shot 3
        expect(rows[0].text()).toContain('-') // Null text interpolation creates nothing or null string? Let's check classes specifically
        expect(rows[0].find('span').classes()).toContain('text-slate-500')

        // Second newest is shot 2 (10, X)
        expect(rows[1].text()).toContain('10')
        expect(rows[1].text()).toContain('X')

        // Oldest is shot 1 (9)
        expect(rows[2].text()).toContain('9')
        expect(rows[2].text()).toContain('-')
    })

    it('formats coordinates correctly when present', () => {
        const wrapper = mount(SessionHistory, {
            props: { shots: [mockShots[0]] }, // shot 1 has x:10.0, y:10.0
        })

        expect(wrapper.text()).toContain('(10.0, 10.0)')
    })

    it('renders "-" when coordinates are missing', () => {
        const wrapper = mount(SessionHistory, {
            props: { shots: [mockShots[2]] }, // shot 3 has null coords
        })

        const rows = wrapper.findAll('tbody tr')
        const cells = rows[0].findAll('td')
        // Coords are in the 4th column (index 3)
        expect(cells[3].text()).toBe('-')
    })

    it('returns correct color classes based on score', () => {
        const scores = [
            { score: 10, class: 'text-yellow-400' },
            { score: 9, class: 'text-yellow-400' },
            { score: 8, class: 'text-red-400' },
            { score: 7, class: 'text-red-400' },
            { score: 6, class: 'text-blue-400' },
            { score: 5, class: 'text-blue-400' },
            { score: 4, class: 'text-black' },
            { score: 3, class: 'text-black' },
            { score: 2, class: 'text-white' },
            { score: 1, class: 'text-white' },
            { score: 0, class: 'text-white' },
            { score: null, class: 'text-slate-500' },
        ]

        const shots = scores.map((s, i) => ({
            shot_id: String(i),
            score: s.score,
            is_x: false,
            x: 0,
            y: 0,
            created_at: `2023-01-01T10:${i.toString().padStart(2, '0')}:00Z`,
            slot_id: 's1',
        }))

        const wrapper = mount(SessionHistory, {
            props: { shots },
        })

        const rows = wrapper.findAll('tbody tr')
        // Rows are newest first, but let's just reverse scores to match
        const reversedScores = [...scores].reverse()

        rows.forEach((row, index) => {
            const scoreSpan = row.findAll('td')[1].find('span')
            expect(scoreSpan.classes()).toContain(reversedScores[index].class)
        })
    })
})
