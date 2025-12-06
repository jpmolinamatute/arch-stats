import type { components } from '@/types/types.generated'
import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import SlotJoinForm from '@/components/forms/SlotJoinForm.vue'
import { useArcher } from '@/composables/useArcher'
import { useAuth } from '@/composables/useAuth'
import { useFaces } from '@/composables/useFaces'

import { useSlot } from '@/composables/useSlot'

type ArcherRead = components['schemas']['ArcherRead']
type FaceMinimal = components['schemas']['FaceMinimal']
interface UserSession {
    archer_id: string
    email: string
    first_name?: string | null
    last_name?: string | null
    picture_url?: string | null
    is_admin?: boolean
}

// Mock composables
vi.mock('@/composables/useSlot', () => ({
    useSlot: vi.fn(),
}))
vi.mock('@/composables/useArcher', () => ({
    useArcher: vi.fn(),
}))
vi.mock('@/composables/useFaces', () => ({
    useFaces: vi.fn(),
}))
vi.mock('@/composables/useAuth', () => ({
    useAuth: vi.fn(),
}))
vi.mock('vue-router', () => ({
    useRouter: vi.fn(),
}))

describe('slotJoinForm', () => {
    const mockJoinSession = vi.fn()
    const mockGetSlot = vi.fn()
    const mockGetSlotCached = vi.fn()
    const mockGetArcher = vi.fn()
    const mockListFaces = vi.fn()
    const mockRouterPush = vi.fn()

    beforeEach(() => {
        vi.clearAllMocks()
        vi.mocked(useRouter).mockReturnValue({ push: mockRouterPush } as unknown as ReturnType<
            typeof useRouter
        >)
    })

    it('renders form and fetches initial data', async () => {
        vi.mocked(useSlot).mockReturnValue({
            joinSession: mockJoinSession,
            getSlot: mockGetSlot,
            getSlotCached: mockGetSlotCached,
            currentSlot: ref(null),
            loading: ref(false),
            error: ref(null),
            leaveSession: vi.fn(),
            clearSlotCache: vi.fn(),
        })
        vi.mocked(useArcher).mockReturnValue({
            getArcher: mockGetArcher.mockResolvedValue({
                bowstyle: 'recurve',
                draw_weight: 30,
                first_name: 'John',
                last_name: 'Doe',
                date_of_birth: '1990-01-01',
                gender: 'male',
                created_at: '2023-01-01',
            } as ArcherRead),
            loading: ref(false),
            error: ref(null),
        })
        vi.mocked(useFaces).mockReturnValue({
            listFaces: mockListFaces,
            faces: ref<FaceMinimal[]>([{ face_type: 'wa_40cm_full', face_name: 'WA 40cm' }]),
            loading: ref(false),
            error: ref(null),
            fetchFace: vi.fn(),
            face: ref(null),
            createFace: vi.fn(),
            updateFace: vi.fn(),
            deleteFace: vi.fn(),
        })
        vi.mocked(useAuth).mockReturnValue({
            user: ref<UserSession>({
                first_name: 'John',
                last_name: 'Doe',
                archer_id: 'archer_1',
                email: 'john@example.com',
                is_admin: false,
            }),
            isAuthenticated: ref(true),
            loading: ref(false),
            pendingRegistration: ref(null),
            bootstrapAuth: vi.fn(),
            logout: vi.fn(),
            disableGoogleAutoSelect: vi.fn(),
            registerNewArcher: vi.fn(),
            initialized: ref(false),
            initError: ref(null),
            prompting: ref(false),
            initOneTap: vi.fn(),
            beginGoogleLogin: vi.fn(),
            clientId: 'mock-client-id',
            loginAsDummy: vi.fn(),
        })

        const wrapper = mount(SlotJoinForm, {
            props: { sessionId: 'sess_1' },
        })

        // Wait for onMounted
        await new Promise(resolve => setTimeout(resolve, 0))
        await new Promise(resolve => setTimeout(resolve, 0))

        expect(mockListFaces).toHaveBeenCalled()
        expect(mockGetArcher).toHaveBeenCalledWith('archer_1')
        expect(wrapper.find('select').exists()).toBe(true) // Face type select
    })

    it('shows validation error for invalid distance', async () => {
        vi.mocked(useSlot).mockReturnValue({
            joinSession: mockJoinSession,
            getSlot: mockGetSlot,
            getSlotCached: mockGetSlotCached,
            currentSlot: ref(null),
            loading: ref(false),
            error: ref(null),
            leaveSession: vi.fn(),
            clearSlotCache: vi.fn(),
        })
        vi.mocked(useArcher).mockReturnValue({
            getArcher: mockGetArcher.mockResolvedValue({
                bowstyle: 'recurve',
                draw_weight: 30,
                first_name: 'John',
                last_name: 'Doe',
                date_of_birth: '1990-01-01',
                gender: 'male',
                created_at: '2023-01-01',
            } as ArcherRead),
            loading: ref(false),
            error: ref(null),
        })
        vi.mocked(useFaces).mockReturnValue({
            listFaces: mockListFaces,
            faces: ref<FaceMinimal[]>([{ face_type: 'wa_40cm_full', face_name: 'WA 40cm' }]),
            loading: ref(false),
            error: ref(null),
            fetchFace: vi.fn(),
            face: ref(null),
            createFace: vi.fn(),
            updateFace: vi.fn(),
            deleteFace: vi.fn(),
        })
        vi.mocked(useAuth).mockReturnValue({
            user: ref<UserSession>({
                first_name: 'John',
                last_name: 'Doe',
                archer_id: 'archer_1',
                email: 'john@example.com',
                is_admin: false,
            }),
            isAuthenticated: ref(true),
            loading: ref(false),
            pendingRegistration: ref(null),
            bootstrapAuth: vi.fn(),
            logout: vi.fn(),
            disableGoogleAutoSelect: vi.fn(),
            registerNewArcher: vi.fn(),
            initialized: ref(false),
            initError: ref(null),
            prompting: ref(false),
            initOneTap: vi.fn(),
            beginGoogleLogin: vi.fn(),
            clientId: 'mock-client-id',
            loginAsDummy: vi.fn(),
        })

        const wrapper = mount(SlotJoinForm, {
            props: { sessionId: 'sess_1' },
        })

        // Wait for onMounted
        await new Promise(resolve => setTimeout(resolve, 0))

        // Set invalid distance
        await wrapper.findAll('input[type="number"]')[1].setValue(150)

        await wrapper.find('form').trigger('submit')

        expect(mockJoinSession).not.toHaveBeenCalled()
        expect(wrapper.text()).toContain('Distance must be between 1 and 100 meters')
    })

    it('submits form and redirects on success', async () => {
        vi.mocked(useSlot).mockReturnValue({
            joinSession: mockJoinSession.mockResolvedValue({ slot_id: 'slot_1' }),
            getSlot: mockGetSlot.mockResolvedValue({ slot_id: 'slot_1' }),
            getSlotCached: mockGetSlotCached.mockReturnValue(null),
            currentSlot: ref(null),
            loading: ref(false),
            error: ref(null),
            leaveSession: vi.fn(),
            clearSlotCache: vi.fn(),
        })
        vi.mocked(useArcher).mockReturnValue({
            getArcher: mockGetArcher.mockResolvedValue({
                bowstyle: 'recurve',
                draw_weight: 30,
                first_name: 'John',
                last_name: 'Doe',
                date_of_birth: '1990-01-01',
                gender: 'male',
                created_at: '2023-01-01',
            } as ArcherRead),
            loading: ref(false),
            error: ref(null),
        })
        vi.mocked(useFaces).mockReturnValue({
            listFaces: mockListFaces,
            faces: ref<FaceMinimal[]>([{ face_type: 'wa_40cm_full', face_name: 'WA 40cm' }]),
            loading: ref(false),
            error: ref(null),
            fetchFace: vi.fn(),
            face: ref(null),
            createFace: vi.fn(),
            updateFace: vi.fn(),
            deleteFace: vi.fn(),
        })
        vi.mocked(useAuth).mockReturnValue({
            user: ref<UserSession>({
                first_name: 'John',
                last_name: 'Doe',
                archer_id: 'archer_1',
                email: 'john@example.com',
                is_admin: false,
            }),
            isAuthenticated: ref(true),
            loading: ref(false),
            pendingRegistration: ref(null),
            bootstrapAuth: vi.fn(),
            logout: vi.fn(),
            disableGoogleAutoSelect: vi.fn(),
            registerNewArcher: vi.fn(),
            initialized: ref(false),
            initError: ref(null),
            prompting: ref(false),
            initOneTap: vi.fn(),
            beginGoogleLogin: vi.fn(),
            clientId: 'mock-client-id',
            loginAsDummy: vi.fn(),
        })

        const wrapper = mount(SlotJoinForm, {
            props: { sessionId: 'sess_1' },
        })

        // Wait for onMounted
        await new Promise(resolve => setTimeout(resolve, 0))
        await new Promise(resolve => setTimeout(resolve, 0))

        // Fill form (defaults are populated, but let's be explicit)
        await wrapper.findAll('select')[0].setValue('wa_40cm_full') // Face
        await wrapper.findAll('select')[1].setValue('recurve') // Bowstyle
        await wrapper.findAll('input[type="number"]')[0].setValue(30) // Draw weight
        await wrapper.findAll('input[type="number"]')[1].setValue(18) // Distance

        await wrapper.find('form').trigger('submit')

        expect(mockJoinSession).toHaveBeenCalledWith({
            archer_id: 'archer_1',
            session_id: 'sess_1',
            face_type: 'wa_40cm_full',
            is_shooting: true,
            bowstyle: 'recurve',
            draw_weight: 30,
            distance: 18,
            club_id: null,
        })
        expect(mockGetSlot).toHaveBeenCalled()
        expect(wrapper.emitted('slotAssigned')).toBeTruthy()
        expect(mockRouterPush).toHaveBeenCalledWith('/app/live-session')
    })
})
