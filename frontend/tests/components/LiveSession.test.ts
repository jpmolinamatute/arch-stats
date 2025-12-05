import type { components } from '@/types/types.generated'
import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '@/api/client'
import LiveSession from '@/components/LiveSession.vue'
import { useAuth } from '@/composables/useAuth'
import { useSession } from '@/composables/useSession'

import { useSlot } from '@/composables/useSlot'

type SessionRead = components['schemas']['SessionRead']
type SlotRead = components['schemas']['FullSlotInfo']
interface UserSession {
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
    props: ['faceId', 'maxShots'],
    emits: ['shot'],
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
    await wrapper.find('button').trigger('click')

    // Confirm modal should be visible (prop passed to stub)
    expect((wrapper.findComponent('[data-testid="confirm-modal"]') as any).props('show')).toBe(
      true,
    )

    // Simulate confirm event
    await (wrapper.findComponent('[data-testid="confirm-modal"]') as any).vm.$emit('confirm')

    expect(mockCloseSession).toHaveBeenCalledWith('sess_1')
    expect(mockRouterPush).toHaveBeenCalledWith('/app')
  })

  it('calls api.createShot when Face emits shot', async () => {
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

    const wrapper = mount(LiveSession)
    await new Promise(resolve => setTimeout(resolve, 0))
    await new Promise(resolve => setTimeout(resolve, 0))

    const face = wrapper.findComponent('[data-testid="face"]')
    expect(face.exists()).toBe(true)

    // Simulate shot event
    ;(face as any).vm.$emit('shot', { score: 9, x: 100, y: 100 })

    expect(api.createShot).toHaveBeenCalledWith({
      slot_id: 'slot_1',
      score: 9,
      x: 100,
      y: 100,
      is_x: false,
    })
  })
})
