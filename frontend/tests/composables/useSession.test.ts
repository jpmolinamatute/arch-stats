import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useSession } from '../../src/composables/useSession';
import { api, ApiError } from '../../src/api/client';
import type { components } from '@/types/types.generated';

// Mock the api client
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

// Mock useSlot
vi.mock('../../src/composables/useSlot', () => ({
    useSlot: () => ({
        leaveSession: vi.fn(),
        getSlot: vi.fn().mockResolvedValue(null),
        clearSlotCache: vi.fn(),
    }),
}));

describe('useSession', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('checkForOpenSession returns session when found', async () => {
        const { checkForOpenSession, currentSession } = useSession();

        type SessionId = components['schemas']['SessionId'];
        type SessionRead = components['schemas']['SessionRead'];

        const mockSessionId: SessionId = { session_id: 'sess_123' };
        const mockSession: SessionRead = {
            session_id: 'sess_123',
            owner_archer_id: 'archer_1',
            session_location: 'Range',
            is_indoor: true,
            is_opened: true,
            shot_per_round: 3,
            created_at: '2023-01-01',
            closed_at: null,
        };

        vi.mocked(api.get).mockResolvedValueOnce(mockSessionId).mockResolvedValueOnce(mockSession);

        const result = await checkForOpenSession('archer_1');

        expect(result).toEqual(mockSession);
        expect(currentSession.value).toEqual(mockSession);
    });

    it('checkForOpenSession returns null when no open session', async () => {
        const { checkForOpenSession, currentSession } = useSession();

        vi.mocked(api.get).mockRejectedValue(new ApiError('Not Found', 404));

        const result = await checkForOpenSession('archer_1');

        expect(result).toBeNull();
        expect(currentSession.value).toBeNull();
    });

    it('createSession sets current session', async () => {
        const { createSession, currentSession } = useSession();

        type SessionCreate = components['schemas']['SessionCreate'];
        const payload: SessionCreate = {
            owner_archer_id: 'archer_1',
            session_location: 'Range',
            is_indoor: true,
            is_opened: true,
            shot_per_round: 3,
        };

        vi.mocked(api.post).mockResolvedValue({ session_id: 'sess_new' });

        await createSession(payload);

        expect(currentSession.value).toMatchObject({
            session_id: 'sess_new',
            ...payload,
        });
    });

    it('closeSession clears current session', async () => {
        const { closeSession, currentSession } = useSession();

        // Set initial state
        currentSession.value = {
            session_id: 'sess_1',
        } as unknown as components['schemas']['SessionRead'];

        await closeSession('sess_1');

        expect(api.patch).toHaveBeenCalledWith('/session/close', { session_id: 'sess_1' });
        expect(currentSession.value).toBeNull();
    });
});
