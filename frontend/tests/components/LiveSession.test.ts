import type { components } from '@/types/types.generated'
import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import Face from '@/components/Face.vue'
import LiveSession from '@/components/LiveSession.vue'
import ConfirmModal from '@/components/widgets/ConfirmModal.vue'

import MiniTable from '@/components/widgets/MiniTable.vue'
import { useAuth } from '@/composables/useAuth'
import { useFaces } from '@/composables/useFaces'
import { useSession } from '@/composables/useSession'
import { useShot } from '@/composables/useShot'

import { useSlot } from '@/composables/useSlot'

type SessionRead = components['schemas']['SessionRead']
type SlotRead = components['schemas']['FullSlotInfo']
interface UserSession
{
    archer_id: string
    email: string
    first_name?: string | null
    last_name?: string | null
    picture_url?: string | null
    is_admin?: boolean
}

// Mock composables
vi.mock('@/composables/useAuth', () => ({
    useAuth: vi.fn(),
}))
vi.mock('@/composables/useSession', () => ({
    useSession: vi.fn(),
}))
vi.mock('@/composables/useSlot', () => ({
    useSlot: vi.fn(),
}))
vi.mock('@/composables/useShot', () => ({
    useShot: vi.fn().mockReturnValue({
        createShot: vi.fn(),
        fetchShots: vi.fn(),
        subscribeToShots: vi.fn(),
        shots: { value: [], __v_isRef: true },
        loading: { value: false, __v_isRef: true },
        error: { value: null, __v_isRef: true },
    }),
}))
vi.mock('@/composables/useFaces', () => ({
    useFaces: vi.fn().mockReturnValue({
        fetchFace: vi.fn(),
        loading: { value: false, __v_isRef: true },
        error: { value: null, __v_isRef: true },
        faces: { value: [], __v_isRef: true },
    }),
}))
vi.mock('vue-router', () => ({
    useRouter: vi.fn(),
}))

// Mock api
vi.mock('@/api/client', () => ({
    api: {
        createShot: vi.fn(),
    },
}))

// Mock child components
vi.mock('@/components/layout/AppHeader.vue', () => ({
    default: { template: '<div data-testid="app-header"></div>' },
}))
vi.mock('@/components/widgets/ConfirmModal.vue', () => ({
    default: {
        template: '<div data-testid="confirm-modal"></div>',
        props: ['show'],
        emits: ['confirm', 'cancel'],
    },
}))
vi.mock('@/components/forms/SlotJoinForm.vue', () => ({
    default: {
        template: '<div data-testid="slot-join-form"></div>',
        props: ['sessionId'],
        emits: ['slotAssigned'],
    },
}))
vi.mock('@/components/Face.vue', () => ({
    default: {
        template: '<div data-testid="face"></div>',
        props: ['face', 'shots'],
        emits: ['shot'],
    },
}))
vi.mock('@/components/widgets/MiniTable.vue', () => ({
    default: {
        template: '<div data-testid="mini-table"></div>',
        props: ['shots', 'face', 'maxShots'],
        emits: ['delete', 'clear', 'confirm'],
    },
}))
vi.mock('@/components/widgets/ShotsTable.vue', () => ({
    default: {
        template: '<div data-testid="shots-table"></div>',
        props: ['shots'],
    },
}))

describe('liveSession', () => {
    const mockRouterPush = vi.fn()
    const mockBootstrapAuth = vi.fn()
    const mockCheckForOpenSession = vi.fn()
    const mockCloseSession = vi.fn()
    const mockGetSlot = vi.fn()
    const mockGetSlotCached = vi.fn()

    beforeEach(() => {
        vi.clearAllMocks()
        localStorage.clear()
        vi.mocked(useRouter).mockReturnValue({ push: mockRouterPush } as unknown as ReturnType<
            typeof useRouter
        >)
    })

    it('redirects if not authenticated', async () => {
        vi.mocked(useAuth).mockReturnValue({
            bootstrapAuth: mockBootstrapAuth,
            isAuthenticated: ref(false),
            loading: ref(false),
            pendingRegistration: ref(null),
            user: ref(null),
            logout: vi.fn(),
            disableGoogleAutoSelect: vi.fn(),
            registerNewArcher: vi.fn(),
            initialized: ref(true),
            initError: ref(null),
            prompting: ref(false),
            initOneTap: vi.fn(),
            clientId: 'mock-client-id',
            beginGoogleLogin: vi.fn(),
            loginAsDummy: vi.fn(),
        })
        vi.mocked(useSession).mockReturnValue({
            checkForOpenSession: mockCheckForOpenSession,
            currentSession: ref(null),
            loading: ref(false),
            error: ref(null),
            createSession: vi.fn(),
            closeSession: vi.fn(),
            hasOpenSession: computed(() => false),
            clearSessionCache: vi.fn(),
        })
        vi.mocked(useSlot).mockReturnValue({
            currentSlot: ref(null),
            getSlot: mockGetSlot,
            getSlotCached: mockGetSlotCached,
            loading: ref(false),
            error: ref(null),
            joinSession: vi.fn(),
            leaveSession: vi.fn(),
            clearSlotCache: vi.fn(),
        })

        mount(LiveSession)
        await new Promise(resolve => setTimeout(resolve, 0))
        await new Promise(resolve => setTimeout(resolve, 0)) // Wait for async mount

        expect(mockRouterPush).toHaveBeenCalledWith('/app')
    })

    it('redirects if no session found', async () => {
        vi.mocked(useAuth).mockReturnValue({
            bootstrapAuth: mockBootstrapAuth,
            isAuthenticated: ref(true),
            loading: ref(false),
            pendingRegistration: ref(null),
            user: ref<UserSession>({
                first_name: 'John',
                last_name: 'Doe',
                archer_id: 'archer_1',
                email: 'john@example.com',
                is_admin: false,
            }),
            logout: vi.fn(),
            disableGoogleAutoSelect: vi.fn(),
            registerNewArcher: vi.fn(),
            initialized: ref(true),
            initError: ref(null),
            prompting: ref(false),
            initOneTap: vi.fn(),
            clientId: 'mock-client-id',
            beginGoogleLogin: vi.fn(),
            loginAsDummy: vi.fn(),
        })
        vi.mocked(useSession).mockReturnValue({
            checkForOpenSession: mockCheckForOpenSession,
            currentSession: ref(null),
            loading: ref(false),
            error: ref(null),
            createSession: vi.fn(),
            closeSession: vi.fn(),
            hasOpenSession: computed(() => false),
            clearSessionCache: vi.fn(),
        })
        vi.mocked(useSlot).mockReturnValue({
            currentSlot: ref(null),
            getSlot: mockGetSlot,
            getSlotCached: mockGetSlotCached,
            loading: ref(false),
            error: ref(null),
            joinSession: vi.fn(),
            leaveSession: vi.fn(),
            clearSlotCache: vi.fn(),
        })

        mount(LiveSession)
        await new Promise(resolve => setTimeout(resolve, 0))
        await new Promise(resolve => setTimeout(resolve, 0))

        expect(mockCheckForOpenSession).toHaveBeenCalledWith('archer_1')
        expect(mockRouterPush).toHaveBeenCalledWith('/app')
    })

    it('renders SlotJoinForm if session exists but no slot', async () => {
        vi.mocked(useAuth).mockReturnValue({
            bootstrapAuth: mockBootstrapAuth,
            isAuthenticated: ref(true),
            loading: ref(false),
            pendingRegistration: ref(null),
            user: ref<UserSession>({
                first_name: 'John',
                last_name: 'Doe',
                archer_id: 'archer_1',
                email: 'john@example.com',
                is_admin: false,
            }),
            logout: vi.fn(),
            disableGoogleAutoSelect: vi.fn(),
            registerNewArcher: vi.fn(),
            initialized: ref(true),
            initError: ref(null),
            prompting: ref(false),
            initOneTap: vi.fn(),
            clientId: 'mock-client-id',
            beginGoogleLogin: vi.fn(),
            loginAsDummy: vi.fn(),
        })
        vi.mocked(useSession).mockReturnValue({
            checkForOpenSession: mockCheckForOpenSession,
            currentSession: ref<SessionRead>({
                session_id: 'sess_1',
                session_location: 'Range A',
                owner_archer_id: 'archer_1',
                is_indoor: true,
                is_opened: true,
                shot_per_round: 3,
                created_at: '2023-01-01',
            }),
            loading: ref(false),
            error: ref(null),
            createSession: vi.fn(),
            closeSession: vi.fn(),
            hasOpenSession: computed(() => true),
            clearSessionCache: vi.fn(),
        })
        vi.mocked(useSlot).mockReturnValue({
            currentSlot: ref(null),
            getSlot: mockGetSlot.mockRejectedValue(new Error('404')),
            getSlotCached: mockGetSlotCached.mockReturnValue(null),
            loading: ref(false),
            error: ref(null),
            joinSession: vi.fn(),
            leaveSession: vi.fn(),
            clearSlotCache: vi.fn(),
        })

        const wrapper = mount(LiveSession)
        await new Promise(resolve => setTimeout(resolve, 0))
        await new Promise(resolve => setTimeout(resolve, 0))
        await new Promise(resolve => setTimeout(resolve, 0)) // Extra tick for slot check catch block

        expect(wrapper.find('[data-testid="slot-join-form"]').exists()).toBe(true)
        expect(wrapper.find('[data-testid="face"]').exists()).toBe(false)
    })

    it('renders Face if slot exists', async () => {
        vi.mocked(useAuth).mockReturnValue({
            bootstrapAuth: mockBootstrapAuth,
            isAuthenticated: ref(true),
            loading: ref(false),
            pendingRegistration: ref(null),
            user: ref<UserSession>({
                first_name: 'John',
                last_name: 'Doe',
                archer_id: 'archer_1',
                email: 'john@example.com',
                is_admin: false,
            }),
            logout: vi.fn(),
            disableGoogleAutoSelect: vi.fn(),
            registerNewArcher: vi.fn(),
            initialized: ref(true),
            initError: ref(null),
            prompting: ref(false),
            initOneTap: vi.fn(),
            clientId: 'mock-client-id',
            beginGoogleLogin: vi.fn(),
            loginAsDummy: vi.fn(),
        })
        vi.mocked(useSession).mockReturnValue({
            checkForOpenSession: mockCheckForOpenSession,
            currentSession: ref<SessionRead>({
                session_id: 'sess_1',
                session_location: 'Range A',
                owner_archer_id: 'archer_1',
                is_indoor: true,
                is_opened: true,
                shot_per_round: 3,
                created_at: '2023-01-01',
            }),
            loading: ref(false),
            error: ref(null),
            createSession: vi.fn(),
            closeSession: vi.fn(),
            hasOpenSession: computed(() => true),
            clearSessionCache: vi.fn(),
        })
        vi.mocked(useSlot).mockReturnValue({
            currentSlot: ref<SlotRead>({
                slot_id: 'slot_1',
                face_type: 'wa_40cm_full',
                session_id: 'sess_1',
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
            getSlot: mockGetSlot.mockResolvedValue({
                slot_id: 'slot_1',
                face_type: 'wa_40cm_full',
            }),
            getSlotCached: mockGetSlotCached.mockReturnValue(null),
            loading: ref(false),
            error: ref(null),
            joinSession: vi.fn(),
            leaveSession: vi.fn(),
            clearSlotCache: vi.fn(),
        })

        const wrapper = mount(LiveSession)
        await new Promise(resolve => setTimeout(resolve, 0))
        await new Promise(resolve => setTimeout(resolve, 0))

        expect(wrapper.find('[data-testid="face"]').exists()).toBe(true)
        expect(wrapper.find('[data-testid="slot-join-form"]').exists()).toBe(false)
    })

    it('closes session correctly', async () => {
        vi.mocked(useAuth).mockReturnValue({
            bootstrapAuth: mockBootstrapAuth,
            isAuthenticated: ref(true),
            loading: ref(false),
            pendingRegistration: ref(null),
            user: ref<UserSession>({
                first_name: 'John',
                last_name: 'Doe',
                archer_id: 'archer_1',
                email: 'john@example.com',
                is_admin: false,
            }),
            logout: vi.fn(),
            disableGoogleAutoSelect: vi.fn(),
            registerNewArcher: vi.fn(),
            initialized: ref(true),
            initError: ref(null),
            prompting: ref(false),
            initOneTap: vi.fn(),
            clientId: 'mock-client-id',
            beginGoogleLogin: vi.fn(),
            loginAsDummy: vi.fn(),
        })
        vi.mocked(useSession).mockReturnValue({
            checkForOpenSession: mockCheckForOpenSession,
            currentSession: ref<SessionRead>({
                session_id: 'sess_1',
                session_location: 'Range A',
                owner_archer_id: 'archer_1',
                is_indoor: true,
                is_opened: true,
                shot_per_round: 3,
                created_at: '2023-01-01',
            }),
            closeSession: mockCloseSession,
            loading: ref(false),
            error: ref(null),
            createSession: vi.fn(),
            hasOpenSession: computed(() => true),
            clearSessionCache: vi.fn(),
        })
        vi.mocked(useSlot).mockReturnValue({
            currentSlot: ref<SlotRead>({
                slot_id: 'slot_1',
                face_type: 'wa_40cm_full',
                session_id: 'sess_1',
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
            getSlot: mockGetSlot,
            getSlotCached: mockGetSlotCached,
            loading: ref(false),
            error: ref(null),
            joinSession: vi.fn(),
            leaveSession: vi.fn(),
            clearSlotCache: vi.fn(),
        })

        const wrapper = mount(LiveSession)
        await new Promise(resolve => setTimeout(resolve, 0))
        await new Promise(resolve => setTimeout(resolve, 0))

        // Click close button
        await wrapper.find('[data-testid="close-session-btn"]').trigger('click')

        // Confirm modal should be visible (prop passed to stub)
        expect((wrapper.findComponent(ConfirmModal) as any).props('show')).toBe(
            true,
        )

        // Simulate confirm event
        await (wrapper.findComponent(ConfirmModal) as any).vm.$emit('confirm')

        expect(mockCloseSession).toHaveBeenCalledWith('sess_1')
        expect(mockRouterPush).toHaveBeenCalledWith('/app')
    })

    it('adds shot to draft and confirms round', async () => {
        vi.mocked(useAuth).mockReturnValue({
            bootstrapAuth: mockBootstrapAuth,
            isAuthenticated: ref(true),
            loading: ref(false),
            pendingRegistration: ref(null),
            user: ref<UserSession>({
                first_name: 'John',
                last_name: 'Doe',
                archer_id: 'archer_1',
                email: 'john@example.com',
                is_admin: false,
            }),
            logout: vi.fn(),
            disableGoogleAutoSelect: vi.fn(),
            registerNewArcher: vi.fn(),
            initialized: ref(true),
            initError: ref(null),
            prompting: ref(false),
            initOneTap: vi.fn(),
            clientId: 'mock-client-id',
            beginGoogleLogin: vi.fn(),
            loginAsDummy: vi.fn(),
        })
        vi.mocked(useSession).mockReturnValue({
            checkForOpenSession: mockCheckForOpenSession,
            currentSession: ref<SessionRead>({
                session_id: 'sess_1',
                session_location: 'Range A',
                owner_archer_id: 'archer_1',
                is_indoor: true,
                is_opened: true,
                shot_per_round: 3,
                created_at: '2023-01-01',
            }),
            loading: ref(false),
            error: ref(null),
            createSession: vi.fn(),
            closeSession: vi.fn(),
            hasOpenSession: computed(() => true),
            clearSessionCache: vi.fn(),
        })
        vi.mocked(useSlot).mockReturnValue({
            currentSlot: ref<SlotRead>({
                slot_id: 'slot_1',
                face_type: 'wa_40cm_full',
                session_id: 'sess_1',
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
            getSlot: mockGetSlot,
            getSlotCached: mockGetSlotCached,
            loading: ref(false),
            error: ref(null),
            joinSession: vi.fn(),
            leaveSession: vi.fn(),
            clearSlotCache: vi.fn(),
        })

        // Mock useShot
        const mockCreateShot = vi.fn()
        vi.mocked(useShot).mockReturnValue({
            createShot: mockCreateShot,
            fetchShots: vi.fn(),
            subscribeToShots: vi.fn(),
            shots: ref([]),
            loading: ref(false),
            error: ref(null),
        })

        // Mock useFaces
        vi.mocked(useFaces).mockReturnValue({
            fetchFace: vi.fn().mockResolvedValue({
                face_type: 'wa_40cm_full',
                viewBox: 100,
                rings: [],
            }),
            loading: ref(false),
            error: ref(null),
            faces: ref([]),
            listFaces: vi.fn(),
            face: ref(null),
            createFace: vi.fn(),
            updateFace: vi.fn(),
            deleteFace: vi.fn(),
        })

        const wrapper = mount(LiveSession)
        await new Promise(resolve => setTimeout(resolve, 0))
        await new Promise(resolve => setTimeout(resolve, 0))

        const face = wrapper.findComponent(Face)
        expect(face.exists()).toBe(true)

        // Simulate shot event
        ; (face as any).vm.$emit('shot', { score: 9, x: 100, y: 100, is_x: true })

        // Should NOT call API yet
        expect(mockCreateShot).not.toHaveBeenCalled()

        // Find MiniTable and confirm
        const miniTable = wrapper.findComponent(MiniTable)
        expect(miniTable.exists()).toBe(true)

        ; (miniTable as any).vm.$emit('confirm')

        expect(mockCreateShot).toHaveBeenCalledWith([
            {
                slot_id: 'slot_1',
                score: 9,
                x: 100,
                y: 100,
                is_x: true,
            },
        ])
    })

    it('toggles between Target and Shots view', async () => {
        vi.mocked(useAuth).mockReturnValue({
            bootstrapAuth: mockBootstrapAuth,
            isAuthenticated: ref(true),
            loading: ref(false),
            pendingRegistration: ref(null),
            user: ref<UserSession>({
                first_name: 'John',
                last_name: 'Doe',
                archer_id: 'archer_1',
                email: 'john@example.com',
                is_admin: false,
            }),
            logout: vi.fn(),
            disableGoogleAutoSelect: vi.fn(),
            registerNewArcher: vi.fn(),
            initialized: ref(true),
            initError: ref(null),
            prompting: ref(false),
            initOneTap: vi.fn(),
            clientId: 'mock-client-id',
            beginGoogleLogin: vi.fn(),
            loginAsDummy: vi.fn(),
        })
        vi.mocked(useSession).mockReturnValue({
            checkForOpenSession: mockCheckForOpenSession,
            currentSession: ref<SessionRead>({
                session_id: 'sess_1',
                session_location: 'Range A',
                owner_archer_id: 'archer_1',
                is_indoor: true,
                is_opened: true,
                shot_per_round: 3,
                created_at: '2023-01-01',
            }),
            loading: ref(false),
            error: ref(null),
            createSession: vi.fn(),
            closeSession: vi.fn(),
            hasOpenSession: computed(() => true),
            clearSessionCache: vi.fn(),
        })
        vi.mocked(useSlot).mockReturnValue({
            currentSlot: ref<SlotRead>({
                slot_id: 'slot_1',
                face_type: 'wa_40cm_full',
                session_id: 'sess_1',
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
            getSlot: mockGetSlot,
            getSlotCached: mockGetSlotCached,
            loading: ref(false),
            error: ref(null),
            joinSession: vi.fn(),
            leaveSession: vi.fn(),
            clearSlotCache: vi.fn(),
        })

        // Mock useShot
        vi.mocked(useShot).mockReturnValue({
            createShot: vi.fn(),
            fetchShots: vi.fn(),
            subscribeToShots: vi.fn(),
            shots: ref([]),
            loading: ref(false),
            error: ref(null),
        })

        // Mock useFaces
        vi.mocked(useFaces).mockReturnValue({
            fetchFace: vi.fn().mockResolvedValue({
                face_type: 'wa_40cm_full',
                viewBox: 100,
                rings: [],
            }),
            loading: ref(false),
            error: ref(null),
            faces: ref([]),
            listFaces: vi.fn(),
            face: ref(null),
            createFace: vi.fn(),
            updateFace: vi.fn(),
            deleteFace: vi.fn(),
        })

        const wrapper = mount(LiveSession)
        await new Promise(resolve => setTimeout(resolve, 0))
        await new Promise(resolve => setTimeout(resolve, 0))

        // Default: Target view visible
        expect(wrapper.find('[data-testid="target-view"]').isVisible()).toBe(true)
        expect(wrapper.find('[data-testid="shots-view"]').isVisible()).toBe(false)

        // Find toggle button
        const shotsButton = wrapper.find('[data-testid="view-shots-btn"]')

        expect(shotsButton.exists()).toBe(true)
        expect(shotsButton.classes()).toContain('text-slate-400')

        // Click Shots button
        await shotsButton.trigger('click')

        // Force update
        await wrapper.vm.$nextTick()

        // Check if state updated by checking button class
        expect(shotsButton.classes()).toContain('bg-indigo-600')

        // Now Target should be hidden (v-show=false)
        // Using style check because isVisible can be flaky in some JSDOM setups with v-show
        expect(wrapper.find('[data-testid="target-view"]').attributes('style')).toContain('display: none')

        // ShotsTable should be visible
        const shotsViewStyle = wrapper.find('[data-testid="shots-view"]').attributes('style')
        expect(shotsViewStyle).not.toContain('display: none')
    })
})
