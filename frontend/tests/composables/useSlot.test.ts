import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { useSlot } from '../../src/composables/useSlot';
import { api } from '../../src/api/client';
import type { components } from '@/types/types.generated';

// Mock api
vi.mock('../../src/api/client', () => ({
    api: {
        get: vi.fn(),
        post: vi.fn(),
        patch: vi.fn(),
    },
    ApiError: class extends Error {
        status: number;
        constructor(message: string, status: number) {
            super(message);
            this.status = status;
        }
    },
}));

// Mock useAuth
const { mockUserValue } = vi.hoisted(() => ({
    mockUserValue: { value: { archer_id: 'archer_1' } as { archer_id: string } | null },
}));

vi.mock('../../src/composables/useAuth', () => ({
    useAuth: () => ({ user: mockUserValue }),
}));

describe('useSlot', () => {
    beforeEach(() => {
        vi.clearAllMocks();
        mockUserValue.value = { archer_id: 'archer_1' };

        // Mock window.localStorage
        vi.stubGlobal('window', {
            localStorage: {
                getItem: vi.fn(),
                setItem: vi.fn(),
                removeItem: vi.fn(),
            },
        });
    });

    afterEach(() => {
        vi.unstubAllGlobals();
    });

    it('joinSession calls api and updates state', async () => {
        const { joinSession, currentSlot } = useSlot();

        type SlotJoinRequest = components['schemas']['SlotJoinRequest'];
        type SlotJoinResponse = components['schemas']['SlotJoinResponse'];
        type FullSlotInfo = components['schemas']['FullSlotInfo'];

        const payload: SlotJoinRequest = {
            session_id: 'sess_1',
            face_type: 'wa_40cm_full',
            distance: 18,
            archer_id: 'archer_1',
            is_shooting: true,
            bowstyle: 'recurve',
            draw_weight: 40,
        };
        const response: SlotJoinResponse = { slot_id: 'slot_1', slot: 'A' };
        const fullSlot: FullSlotInfo = {
            slot_id: 'slot_1',
            slot: 'A',
            archer_id: 'archer_1',
            session_id: 'sess_1',
            face_type: 'wa_40cm_full',
            distance: 18,
            is_shooting: true,
            bowstyle: 'recurve',
            draw_weight: 40,
            target_id: 'target_1',
            slot_letter: 'A',
            created_at: '2023-01-01',
            lane: 1,
        };

        vi.mocked(api.post).mockResolvedValue(response);
        vi.mocked(api.get).mockResolvedValue(fullSlot);

        const result = await joinSession(payload);

        expect(vi.mocked(api.post)).toHaveBeenCalledWith('/session/slot', payload);
        expect(result).toEqual(response);
        // It tries to fetch full slot after joining
        expect(vi.mocked(api.get)).toHaveBeenCalled();
        expect(currentSlot.value).toEqual(fullSlot);
    });

    it('getSlot fetches slot', async () => {
        const { getSlot, currentSlot } = useSlot();

        type FullSlotInfo = components['schemas']['FullSlotInfo'];
        const fullSlot: FullSlotInfo = {
            slot_id: 'slot_1',
            slot: 'A',
            archer_id: 'archer_1',
            session_id: 'sess_1',
            face_type: 'wa_40cm_full',
            distance: 18,
            is_shooting: true,
            bowstyle: 'recurve',
            draw_weight: 40,
            target_id: 'target_1',
            slot_letter: 'A',
            created_at: '2023-01-01',
            lane: 1,
        };
        vi.mocked(api.get).mockResolvedValue(fullSlot);

        const result = await getSlot(true); // force refresh to bypass cache logic for now

        expect(vi.mocked(api.get)).toHaveBeenCalledWith('/session/slot/archer/archer_1');
        expect(result).toEqual(fullSlot);
        expect(currentSlot.value).toEqual(fullSlot);
    });

    it('leaveSession calls api and clears state', async () => {
        const { leaveSession, currentSlot } = useSlot();

        type FullSlotInfo = components['schemas']['FullSlotInfo'];
        currentSlot.value = { slot_id: 'slot_1' } as unknown as FullSlotInfo;

        await leaveSession('slot_1');

        expect(vi.mocked(api.patch)).toHaveBeenCalledWith('/session/slot/leave/slot_1');
        expect(currentSlot.value).toBeNull();
    });
});
