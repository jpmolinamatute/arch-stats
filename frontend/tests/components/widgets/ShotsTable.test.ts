import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import ShotsTable from '@/components/widgets/ShotsTable.vue'

describe('shotsTable', () => {
    const createShot = (id: string, score: number, createdAt: string, isX = false) => ({
        shot_id: id,
        slot_id: 'slot_1',
        score,
        is_x: isX,
        x: 0,
        y: 0,
        confidence: 1,
        created_at: createdAt,
    })

    const mockShots = [
        // Round 1 (Complete, 3 shots)
        createShot('1', 9, '2023-01-01T10:00:00Z'),
        createShot('2', 9, '2023-01-01T10:01:00Z'),
        createShot('3', 10, '2023-01-01T10:02:00Z'), // Total: 28

        // Round 2 (Complete, 3 shots)
        createShot('4', 8, '2023-01-01T10:05:00Z'),
        createShot('5', 8, '2023-01-01T10:06:00Z'),
        createShot('6', 8, '2023-01-01T10:07:00Z'), // Total: 24 (Cum: 52)

        // Round 3 (Partial, 1 shot)
        createShot('7', 10, '2023-01-01T10:10:00Z'), // Total: 10 (Cum: 62)
    ] // Note: Sorted by time here for clarity, component sorts internally

    it('renders correct columns based on shotPerRound', () => {
        const wrapper = mount(ShotsTable, {
            props: {
                shots: [],
                shotPerRound: 3,
            },
        })

        const headers = wrapper.findAll('th')
        // Expected: # + 3 score + Total + Sum = 6 columns
        expect(headers.length).toBe(6)
        expect(headers[1].text()).toBe('1')
        expect(headers[2].text()).toBe('2')
        expect(headers[3].text()).toBe('3')
    })

    it('groups shots into rounds correctly', () => {
        const wrapper = mount(ShotsTable, {
            props: {
                shots: mockShots,
                shotPerRound: 3,
            },
        })

        const rows = wrapper.findAll('tbody tr')
        // Should have 3 rows (Round 3, Round 2, Round 1 in descending order)
        expect(rows.length).toBe(3)

        // Row 1 (Top) = Round 1 (Oldest)
        const row1Cells = rows[0].findAll('td')
        expect(row1Cells[0].text()).toBe('1') // Round ID
        expect(row1Cells[4].text()).toBe('28') // Total (9+9+10)
        expect(row1Cells[5].text()).toBe('28') // Cumulative

        // Row 2 = Round 2
        const row2Cells = rows[1].findAll('td')
        expect(row2Cells[0].text()).toBe('2')
        expect(row2Cells[4].text()).toBe('24') // Total (8+8+8)
        expect(row2Cells[5].text()).toBe('52') // Cumulative (28+24)

        // Row 3 = Round 3 (Newest)
        const row3Cells = rows[2].findAll('td')
        expect(row3Cells[0].text()).toBe('3')
        expect(row3Cells[1].text()).toBe('10') // Shot 1
        expect(row3Cells[2].text()).toBe('-') // Shot 2 (Empty)
        expect(row3Cells[3].text()).toBe('-') // Shot 3 (Empty)
        expect(row3Cells[4].text()).toBe('10') // Total
        expect(row3Cells[5].text()).toBe('62') // Cumulative
    })

    it('handles empty shots list', () => {
        const wrapper = mount(ShotsTable, {
            props: {
                shots: [],
                shotPerRound: 6,
            },
        })
        expect(wrapper.text()).toContain('No confirmed shots yet')
    })

    it('renders "M" for score 0', () => {
        const wrapper = mount(ShotsTable, {
            props: {
                shots: [createShot('1', 0, '2023-01-01T10:00:00Z')],
                shotPerRound: 6,
            },
        })
        const cell = wrapper.find('td div.w-10')
        expect(cell.text()).toBe('M')
    })

    it('renders "X" for is_x shot', () => {
        const wrapper = mount(ShotsTable, {
            props: {
                shots: [createShot('1', 10, '2023-01-01T10:00:00Z', true)],
                shotPerRound: 6,
            },
        })
        const cell = wrapper.find('td div.w-10')
        expect(cell.text()).toBe('X')
    })

    it('applies correct background and text colors based on score', () => {
        const shots = [
            createShot('1', 10, '2023-01-01T10:00:00Z'), // Yellow (9-10)
            createShot('2', 9, '2023-01-01T10:01:00Z'), // Yellow
            createShot('3', 8, '2023-01-01T10:02:00Z'), // Red (7-8)
            createShot('4', 7, '2023-01-01T10:03:00Z'), // Red
            createShot('5', 6, '2023-01-01T10:04:00Z'), // Blue (5-6)
            createShot('6', 5, '2023-01-01T10:05:00Z'), // Blue
            createShot('7', 4, '2023-01-01T10:06:00Z'), // Black (3-4)
            createShot('8', 3, '2023-01-01T10:07:00Z'), // Black
            createShot('9', 2, '2023-01-01T10:08:00Z'), // White (1-2)
            createShot('10', 1, '2023-01-01T10:09:00Z'), // White
            createShot('11', 0, '2023-01-01T10:10:00Z'), // White (Miss)
        ]

        const wrapper = mount(ShotsTable, {
            props: {
                shots,
                shotPerRound: 12, // Ensure all in one round for easy indexing
            },
        })

        const cells = wrapper.findAll('td div.w-10')

        const checkStyle = (index: number, bg: string, color: string) => {
            const style = cells[index].attributes('style')
            expect(style).toContain(`background-color: ${bg}`)
            expect(style).toContain(`color: ${color}`)
        }

        checkStyle(0, 'rgb(255, 229, 82)', 'rgb(0, 0, 0)') // 10 (Yellow)
        checkStyle(1, 'rgb(255, 229, 82)', 'rgb(0, 0, 0)') // 9 (Yellow)
        checkStyle(2, 'rgb(246, 80, 88)', 'rgb(0, 0, 0)') // 8 (Red)
        checkStyle(3, 'rgb(246, 80, 88)', 'rgb(0, 0, 0)') // 7 (Red)
        checkStyle(4, 'rgb(0, 180, 228)', 'rgb(0, 0, 0)') // 6 (Blue)
        checkStyle(5, 'rgb(0, 180, 228)', 'rgb(0, 0, 0)') // 5 (Blue)
        checkStyle(6, 'rgb(0, 0, 0)', 'rgb(255, 255, 255)') // 4 (Black)
        checkStyle(7, 'rgb(0, 0, 0)', 'rgb(255, 255, 255)') // 3 (Black)
        checkStyle(8, 'rgb(255, 255, 255)', 'rgb(0, 0, 0)') // 2 (White)
        checkStyle(9, 'rgb(255, 255, 255)', 'rgb(0, 0, 0)') // 1 (White)
        checkStyle(10, 'rgb(255, 255, 255)', 'rgb(0, 0, 0)') // 0 (White/Miss)
    })
})
