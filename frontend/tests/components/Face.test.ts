import type { components } from '@/types/types.generated'
import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import Face from '@/components/Face.vue'
import { useFaces } from '@/composables/useFaces'

type FaceSchema = components['schemas']['Face']

// Mock composables and api
vi.mock('@/composables/useFaces', () => ({
  useFaces: vi.fn(),
}))

describe('face', () => {
  const mockFetchFace = vi.fn()

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

  it('fetches face data and renders SVG', async () => {
    vi.mocked(useFaces).mockReturnValue({
      fetchFace: mockFetchFace.mockResolvedValue({
        face_type: 'wa_40cm_full',
        face_name: 'WA 40cm',
        viewBox: 100,
        rings: [
          { r: 50, data_score: 1, fill: 'white', stroke: 'black', stroke_width: 1 },
          { r: 25, data_score: 10, fill: 'yellow', stroke: 'black', stroke_width: 1 },
        ],
        render_cross: true,
        spots: [],
      } as FaceSchema),
      loading: ref(false),
      error: ref(null),
      faces: ref([]),
      listFaces: vi.fn(),
      face: ref(null),
      createFace: vi.fn(),
      updateFace: vi.fn(),
      deleteFace: vi.fn(),
    })

    const wrapper = mount(Face, {
      props: {
        faceId: 'wa_40cm_full',
      },
    })

    // Wait for onMounted
    await new Promise(resolve => setTimeout(resolve, 0))
    await new Promise(resolve => setTimeout(resolve, 0))

    expect(mockFetchFace).toHaveBeenCalledWith('wa_40cm_full')
    expect(wrapper.find('svg').exists()).toBe(true)
    // 2 rings + 0 shots
    expect(wrapper.findAll('circle').length).toBe(2)
  })

  it('emits shot event on click and shows visual feedback', async () => {
    vi.mocked(useFaces).mockReturnValue({
      fetchFace: mockFetchFace.mockResolvedValue({
        face_type: 'wa_40cm_full',
        face_name: 'WA 40cm',
        viewBox: 100,
        rings: [{ r: 50, data_score: 1, fill: 'white', stroke: 'black', stroke_width: 1 }],
        render_cross: false,
        spots: [],
      } as FaceSchema),
      loading: ref(false),
      error: ref(null),
      faces: ref([]),
      listFaces: vi.fn(),
      face: ref(null),
      createFace: vi.fn(),
      updateFace: vi.fn(),
      deleteFace: vi.fn(),
    })

    const wrapper = mount(Face, {
      props: {
        faceId: 'wa_40cm_full',
        maxShots: 3,
      },
    })

    // Wait for onMounted
    await new Promise(resolve => setTimeout(resolve, 0))
    await new Promise(resolve => setTimeout(resolve, 0))

    // Click on the ring (g element)
    await wrapper.find('g').trigger('click', { clientX: 10, clientY: 10 })

    expect(wrapper.emitted('shot')).toBeTruthy()
    expect(wrapper.emitted('shot')![0][0]).toEqual(expect.objectContaining({
      score: 1,
      is_x: false,
    }))

    // Check if a new circle was added (1 ring + 1 shot = 2 circles)
    expect(wrapper.findAll('circle').length).toBe(2)

    // Click 3 more times to exceed limit (limit is 3)
    await wrapper.find('g').trigger('click', { clientX: 20, clientY: 20 })
    await wrapper.find('g').trigger('click', { clientX: 30, clientY: 30 })
    await wrapper.find('g').trigger('click', { clientX: 40, clientY: 40 })

    // Should have 1 ring + 3 shots = 4 circles
    expect(wrapper.findAll('circle').length).toBe(4)
  })

  it('does not render if faceId is missing', async () => {
    vi.mocked(useFaces).mockReturnValue({
      fetchFace: mockFetchFace,
      loading: ref(false),
      error: ref(null),
      faces: ref([]),
      listFaces: vi.fn(),
      face: ref(null),
      createFace: vi.fn(),
      updateFace: vi.fn(),
      deleteFace: vi.fn(),
    })

    const wrapper = mount(Face, {
      props: {
        faceId: 'none',
      },
    })
    await new Promise(resolve => setTimeout(resolve, 0))

    expect(mockFetchFace).not.toHaveBeenCalled()
    expect(wrapper.find('svg').exists()).toBe(false)
  })
})
