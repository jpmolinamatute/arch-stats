import type { components } from '@/types/types.generated'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { api } from '../../src/api/client'
import { useSession } from '../../src/composables/useSession'

// Mock the api client
vi.mock('../../src/api/client', () => ({
    api: {
        get: vi.fn(),
        post: vi.fn(),
        patch: vi.fn(),
    },
    ApiError: class extends Error {
        status: number
        constructor(message: string, status: number) {
            super(message)
            this.status = status
        }
    },
}))

// Mock useSlot
vi.mock('../../src/composables/useSlot', () => ({
    useSlot: () => ({
        leaveSession: vi.fn(),
        getSlot: vi.fn().mockResolvedValue(null),
        clearSlotCache: vi.fn(),
    }),
}))

describe('useSession', () => {
    beforeEach(() => {
        vi.clearAllMocks()
    })

    it('checkForOpenSession returns session when active slot found (Step 1)', async () => {
        const { checkForOpenSession, currentSession } = useSession()

        const mockSession: components['schemas']['SessionRead'] = {
            session_id: 'sess_123',
            owner_archer_id: 'archer_1',
            session_location: 'Range',
            is_indoor: true,
            is_opened: true,
            shot_per_round: 3,
            created_at: '2023-01-01',
            closed_at: null,
        }

        // Mock Step 1: Slot found
        vi.mocked(api.get).mockResolvedValueOnce({ session_id: 'sess_123' })
        // Mock Step 3: Fetch session
        vi.mocked(api.get).mockResolvedValueOnce(mockSession)

        const result = await checkForOpenSession('archer_1')

        expect(api.get).toHaveBeenCalledWith('/session/slot/archer/archer_1', { ignoreStatus: [404] })
        expect(result).toEqual(mockSession)
        expect(currentSession.value).toEqual(mockSession)
    })

    it('checkForOpenSession returns session when owned session found (Step 2 - Recovery)', async () => {
        const { checkForOpenSession, currentSession } = useSession()

        const mockSession: components['schemas']['SessionRead'] = {
            session_id: 'sess_123',
            owner_archer_id: 'archer_1',
            session_location: 'Range',
            is_indoor: true,
            is_opened: true,
            shot_per_round: 3,
            created_at: '2023-01-01',
            closed_at: null,
        }

        // Mock Step 1: Slot NOT found (null due to ignoreStatus)
        vi.mocked(api.get).mockResolvedValueOnce(null)
        // Mock Step 2: Owned session found
        vi.mocked(api.get).mockResolvedValueOnce({ session_id: 'sess_123' })
        // Mock Step 3: Fetch session
        vi.mocked(api.get).mockResolvedValueOnce(mockSession)

        const result = await checkForOpenSession('archer_1')

        expect(api.get).toHaveBeenCalledWith('/session/slot/archer/archer_1', { ignoreStatus: [404] })
        expect(api.get).toHaveBeenCalledWith('/session/archer/archer_1/open-session', { ignoreStatus: [404] })
        expect(result).toEqual(mockSession)
        expect(currentSession.value).toEqual(mockSession)
    })

    it('checkForOpenSession returns null when neither found', async () => {
        const { checkForOpenSession, currentSession } = useSession()

        // Mock Step 1: Slot NOT found (null)
        vi.mocked(api.get).mockResolvedValueOnce(null)
        // Mock Step 2: Owned session NOT found (null)
        vi.mocked(api.get).mockResolvedValueOnce(null)

        const result = await checkForOpenSession('archer_1')

        expect(result).toBeNull()
        expect(currentSession.value).toBeNull()
    })

    it('checkForOpenSession handles 404 error during session details fetch gracefully', async () => {
        const { checkForOpenSession, currentSession, error } = useSession()
        const { ApiError } = await import('../../src/api/client')

        // Mock Step 1: Slot found (successful)
        vi.mocked(api.get).mockResolvedValueOnce({ session_id: 'sess_123' })
        // Mock Step 3: Fetch session fails with 404
        vi.mocked(api.get).mockRejectedValueOnce(new ApiError('Not Found', 404))

        const result = await checkForOpenSession('archer_1')

        expect(result).toBeNull()
        expect(currentSession.value).toBeNull()
        expect(error.value).toBeNull()
    })

    it('createSession sets current session', async () => {
        const { createSession, currentSession } = useSession()

        type SessionCreate = components['schemas']['SessionCreate']
        const payload: SessionCreate = {
            owner_archer_id: 'archer_1',
            session_location: 'Range',
            is_indoor: true,
            is_opened: true,
            shot_per_round: 3,
        }

        vi.mocked(api.post).mockResolvedValue({ session_id: 'sess_new' })

        await createSession(payload)

        expect(currentSession.value).toMatchObject({
            session_id: 'sess_new',
            ...payload,
        })
    })

    it('closeSession clears current session', async () => {
        const { closeSession, currentSession } = useSession()

        // Set initial state
        currentSession.value = {
            session_id: 'sess_1',
        } as unknown as components['schemas']['SessionRead']

        // Mock successful close response
        vi.mocked(api.patch).mockResolvedValue({ status: 'closed' })

        await closeSession('sess_1')

        expect(api.patch).toHaveBeenCalledWith('/session/close', { session_id: 'sess_1' })
        expect(currentSession.value).toBeNull()
    })
})
