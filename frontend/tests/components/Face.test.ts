import type { components } from '@/types/types.generated'
import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import Face from '@/components/Face.vue'

type FaceSchema = components['schemas']['Face']

describe('face', () => {
    const mockFace: FaceSchema = {
        face_type: 'wa_40cm_full',
        face_name: 'WA 40cm',
        viewBox: 100,
        rings: [
            { r: 50, data_score: 1, fill: 'white', stroke: 'black', stroke_width: 1 },
            { r: 25, data_score: 10, fill: 'yellow', stroke: 'black', stroke_width: 1 },
        ],
        render_cross: true,
        spots: [],
    }

    beforeEach(() => {
        vi.clearAllMocks()

        // Mock window.matchMedia
        Object.defineProperty(window, 'matchMedia', {
            writable: true,
            value: vi.fn().mockImplementation(query => ({
                matches: false,
                media: query,
                onchange: null,
                addListener: vi.fn(), // deprecated
                removeListener: vi.fn(), // deprecated
                addEventListener: vi.fn(),
                removeEventListener: vi.fn(),
                dispatchEvent: vi.fn(),
            })),
        })

        // Mock ResizeObserver
        vi.stubGlobal(
            'ResizeObserver',
            class ResizeObserver {
                observe = vi.fn()
                unobserve = vi.fn()
                disconnect = vi.fn()
            },
        )

        // Mock requestAnimationFrame
        vi.stubGlobal('requestAnimationFrame', (cb: FrameRequestCallback) => {
            return setTimeout(cb, 0)
        })
        vi.stubGlobal('cancelAnimationFrame', (id: number) => {
            clearTimeout(id)
        })

        // Mock SVG methods
        window.SVGSVGElement.prototype.createSVGPoint = function () {
            return {
                x: 0,
                y: 0,
                matrixTransform() {
                    return { x: this.x, y: this.y }
                },
            } as DOMPoint
        }
        window.SVGSVGElement.prototype.getScreenCTM = function () {
            return {
                inverse() {
                    return {} as DOMMatrix
                },
            } as DOMMatrix
        }
    })

    it('renders SVG based on face prop', async () => {
        const wrapper = mount(Face, {
            props: {
                face: mockFace,
                shots: [],
            },
        })

        expect(wrapper.find('svg').exists()).toBe(true)
        // 2 rings + 0 shots
        expect(wrapper.findAll('circle').length).toBe(2)
    })

    it('emits shot event on click', async () => {
        const wrapper = mount(Face, {
            props: {
                face: mockFace,
                shots: [],
            },
        })

        // Click on the ring (g element)
        await wrapper.find('g').trigger('click', { clientX: 10, clientY: 10 })

        expect(wrapper.emitted('shot')).toBeTruthy()
        expect(wrapper.emitted('shot')![0][0]).toEqual(expect.objectContaining({
            score: 1,
            is_x: false,
            color: 'white',
        }))
    })

    it('renders shots passed via props', async () => {
        const wrapper = mount(Face, {
            props: {
                face: mockFace,
                shots: [{ x: 50, y: 50 }, { x: 60, y: 60 }],
            },
        })

        // 2 rings + 2 shots = 4 circles
        expect(wrapper.findAll('circle').length).toBe(4)

        // Update props
        await wrapper.setProps({
            shots: [{ x: 50, y: 50 }],
        })

        // 2 rings + 1 shot = 3 circles
        expect(wrapper.findAll('circle').length).toBe(3)
    })

    it('does not render if face is null', async () => {
        const wrapper = mount(Face, {
            props: {
                face: null,
                shots: [],
            },
        })

        expect(wrapper.find('svg').exists()).toBe(false)
    })
    it('updates SVG width when width prop changes', async () => {
        const wrapper = mount(Face, {
            props: {
                face: mockFace,
                shots: [],
                width: 300,
            },
        })

        const svg = wrapper.find('svg')
        expect(svg.attributes('width')).toBe('300')

        await wrapper.setProps({ width: 500 })
        expect(svg.attributes('width')).toBe('500')
    })

    it('emits miss (score 0) when clicking background', async () => {
        const wrapper = mount(Face, {
            props: {
                face: mockFace,
                shots: [],
            },
        })

        // Find the background rect
        const rect = wrapper.find('rect')
        await rect.trigger('click', { clientX: 5, clientY: 5 })

        expect(wrapper.emitted('shot')).toBeTruthy()
        expect(wrapper.emitted('shot')![0][0]).toEqual(expect.objectContaining({
            score: 0,
            color: 'white',
        }))
    })
})
