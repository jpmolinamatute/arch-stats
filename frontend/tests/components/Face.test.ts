import { describe, it, expect, vi, beforeEach } from 'vitest';
import { mount } from '@vue/test-utils';
import { ref } from 'vue';
import Face from '@/components/Face.vue';
import { useFaces } from '@/composables/useFaces';
import { useSlot } from '@/composables/useSlot';
import { api } from '@/api/client';

import type { components } from '@/types/types.generated';

type FaceSchema = components['schemas']['Face'];
type SlotRead = components['schemas']['FullSlotInfo'];

// Mock composables and api
vi.mock('@/composables/useFaces', () => ({
    useFaces: vi.fn(),
}));
vi.mock('@/composables/useSlot', () => ({
    useSlot: vi.fn(),
}));
vi.mock('@/api/client', () => ({
    api: {
        createShot: vi.fn(),
    },
}));

describe('Face', () => {
    const mockFetchFace = vi.fn();

    beforeEach(() => {
        vi.clearAllMocks();

        // Mock window.matchMedia
        Object.defineProperty(window, 'matchMedia', {
            writable: true,
            value: vi.fn().mockImplementation((query) => ({
                matches: false,
                media: query,
                onchange: null,
                addListener: vi.fn(), // deprecated
                removeListener: vi.fn(), // deprecated
                addEventListener: vi.fn(),
                removeEventListener: vi.fn(),
                dispatchEvent: vi.fn(),
            })),
        });

        // Mock ResizeObserver
        vi.stubGlobal(
            'ResizeObserver',
            class ResizeObserver {
                observe = vi.fn();
                unobserve = vi.fn();
                disconnect = vi.fn();
            },
        );

        // Mock requestAnimationFrame
        vi.stubGlobal('requestAnimationFrame', (cb: FrameRequestCallback) => {
            return setTimeout(cb, 0);
        });
        vi.stubGlobal('cancelAnimationFrame', (id: number) => {
            clearTimeout(id);
        });
    });

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
        });
        vi.mocked(useSlot).mockReturnValue({
            currentSlot: ref<SlotRead>({
                face_type: 'wa_40cm_full',
                session_id: 'sess_1',
                slot_id: 'slot_1',
                target_id: 'target_1',
                archer_id: 'archer_1',
                is_shooting: true,
                bowstyle: 'recurve',
                draw_weight: 40,
                distance: 18,
                created_at: '2023-01-01',
                slot_letter: 'A',
                lane: 1,
                slot: '1A',
            }),
            joinSession: vi.fn(),
            getSlot: vi.fn(),
            getSlotCached: vi.fn(),
            loading: ref(false),
            error: ref(null),
            leaveSession: vi.fn(),
            clearSlotCache: vi.fn(),
        });

        const wrapper = mount(Face);

        // Wait for onMounted
        await new Promise((resolve) => setTimeout(resolve, 0));
        await new Promise((resolve) => setTimeout(resolve, 0));

        expect(mockFetchFace).toHaveBeenCalledWith('wa_40cm_full');
        expect(wrapper.find('svg').exists()).toBe(true);
        expect(wrapper.findAll('circle').length).toBe(2);
    });

    it('records shot on click', async () => {
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
        });
        vi.mocked(useSlot).mockReturnValue({
            currentSlot: ref<SlotRead>({
                face_type: 'wa_40cm_full',
                session_id: 'sess_1',
                slot_id: 'slot_1',
                target_id: 'target_1',
                archer_id: 'archer_1',
                is_shooting: true,
                bowstyle: 'recurve',
                draw_weight: 40,
                distance: 18,
                created_at: '2023-01-01',
                slot_letter: 'A',
                lane: 1,
                slot: '1A',
            }),
            joinSession: vi.fn(),
            getSlot: vi.fn(),
            getSlotCached: vi.fn(),
            loading: ref(false),
            error: ref(null),
            leaveSession: vi.fn(),
            clearSlotCache: vi.fn(),
        });

        const wrapper = mount(Face);

        // Wait for onMounted
        await new Promise((resolve) => setTimeout(resolve, 0));
        await new Promise((resolve) => setTimeout(resolve, 0));

        // Click on the ring (g element)
        await wrapper.find('g').trigger('click');

        expect(api.createShot).toHaveBeenCalledWith(
            expect.objectContaining({
                slot_id: 'slot_1',
                score: 1,
                is_x: false,
            }),
        );
    });

    it('handles missing face type in slot', async () => {
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
        });
        vi.mocked(useSlot).mockReturnValue({
            currentSlot: ref({
                face_type: null,
            } as unknown as SlotRead), // Intentionally invalid for test case
            joinSession: vi.fn(),
            getSlot: vi.fn(),
            getSlotCached: vi.fn(),
            loading: ref(false),
            error: ref(null),
            leaveSession: vi.fn(),
            clearSlotCache: vi.fn(),
        });

        const consoleWarnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
        const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

        const errorHandler = vi.fn();
        mount(Face, {
            global: {
                config: {
                    errorHandler,
                },
            },
        });
        await new Promise((resolve) => setTimeout(resolve, 0));

        expect(errorHandler).toHaveBeenCalled();
        const error = errorHandler.mock.calls[0][0];
        expect(error).toBeInstanceOf(Error);
        expect(error.message).toBe('No valid face_type found in current slot');

        expect(mockFetchFace).not.toHaveBeenCalled();
        consoleWarnSpy.mockRestore();
        consoleErrorSpy.mockRestore();
    });
});
