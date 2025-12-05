import type { components } from '@/types/types.generated'
import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import SessionManager from '@/components/SessionManager.vue'
import { useAuth } from '@/composables/useAuth'

import { useSession } from '@/composables/useSession'

type SessionRead = components['schemas']['SessionRead']
interface UserSession {
  archer_id: string
  email: string
  first_name?: string | null
  last_name?: string | null
  picture_url?: string | null
  is_admin?: boolean // Added based on usage in tests
}

// Mock composables
vi.mock('@/composables/useAuth', () => ({
  useAuth: vi.fn(),
}))
vi.mock('@/composables/useSession', () => ({
  useSession: vi.fn(),
}))
vi.mock('vue-router', () => ({
  useRouter: vi.fn(),
}))

// Mock child components
vi.mock('@/components/forms/SessionForm.vue', () => ({
  default: {
    template: '<div data-testid="session-form"></div>',
    emits: ['sessionCreated'],
  },
}))
vi.mock('@/components/forms/SlotJoinForm.vue', () => ({
  default: {
    template: '<div data-testid="slot-join-form"></div>',
    props: ['sessionId'],
    emits: ['slotAssigned'],
  },
}))

describe('sessionManager', () => {
  const mockCheckForOpenSession = vi.fn()
  const mockRouterPush = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(useRouter).mockReturnValue({ push: mockRouterPush } as unknown as ReturnType<
            typeof useRouter
    >)
  })

  it('checks for open session on mount', async () => {
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

    mount(SessionManager)

    // Wait for onMounted
    await new Promise(resolve => setTimeout(resolve, 0))

    expect(mockCheckForOpenSession).toHaveBeenCalledWith('archer_1')
  })

  it('redirects to live session if session exists', async () => {
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
      initialized: ref(true),
      initError: ref(null),
      prompting: ref(false),
      initOneTap: vi.fn(),
      clientId: 'mock-client-id',
      beginGoogleLogin: vi.fn(),
      loginAsDummy: vi.fn(),
    })
    vi.mocked(useSession).mockReturnValue({
      checkForOpenSession: mockCheckForOpenSession.mockResolvedValue({
        session_id: 'sess_1',
        session_location: 'Range A',
        owner_archer_id: 'archer_1',
        is_indoor: true,
        is_opened: true,
        shot_per_round: 3,
        created_at: '2023-01-01',
      }),
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

    mount(SessionManager)
    await new Promise(resolve => setTimeout(resolve, 0))

    expect(mockRouterPush).toHaveBeenCalledWith('/app/live-session')
  })

  it('renders open session button if no session exists', async () => {
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

    const wrapper = mount(SessionManager)
    await new Promise(resolve => setTimeout(resolve, 0))

    expect(wrapper.find('button').text()).toBe('Open Session')
  })

  it('shows session form when open session button is clicked', async () => {
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

    const wrapper = mount(SessionManager)
    await new Promise(resolve => setTimeout(resolve, 0))

    await wrapper.find('button').trigger('click')

    expect(wrapper.find('[data-testid="session-form"]').exists()).toBe(true)
  })

  it('shows slot join form when session is created', async () => {
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

    const wrapper = mount(SessionManager)
    await new Promise(resolve => setTimeout(resolve, 0))

    // Open session form
    await wrapper.find('button').trigger('click')

    // Emit sessionCreated from SessionForm
    await (
      wrapper.findComponent('[data-testid="session-form"]') as unknown as {
        vm: { $emit: (e: string, v: string) => void }
      }
    ).vm.$emit('sessionCreated', 'sess_new')

    expect(wrapper.find('[data-testid="slot-join-form"]').exists()).toBe(true)
    // Check prop

    expect(
      (
        wrapper.findComponent('[data-testid="slot-join-form"]') as unknown as {
          props: (name: string) => unknown
        }
      ).props('sessionId'),
    ).toBe('sess_new')
  })

  it('redirects when slot is assigned', async () => {
    vi.mocked(useAuth).mockReturnValue({
      user: ref<UserSession>({
        // Actually, UserSession defined in test does NOT have status.
        // But the previous mock had it.
        // Let's check if 'status' is used in tests.
        // The test checks 'isAuthenticated' ref.
        // So I will just use the flat structure.
        first_name: 'John',
        last_name: 'Doe',
        archer_id: 'archer_1',
        email: 'john@example.com',
        is_admin: false,
        picture_url: null,
      }),
      isAuthenticated: ref(true),
      loading: ref(false),
      pendingRegistration: ref(null),
      bootstrapAuth: vi.fn(),
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

    const wrapper = mount(SessionManager)
    await new Promise(resolve => setTimeout(resolve, 0))

    // Simulate flow to get to slot form
    await wrapper.find('button').trigger('click')

    await (
      wrapper.findComponent('[data-testid="session-form"]') as unknown as {
        vm: { $emit: (e: string, v: string) => void }
      }
    ).vm.$emit('sessionCreated', 'sess_new')

    // Emit slotAssigned
    await (
      wrapper.findComponent('[data-testid="slot-join-form"]') as unknown as {
        vm: { $emit: (e: string, v: string) => void }
      }
    ).vm.$emit('slotAssigned', 'slot_1')

    expect(mockRouterPush).toHaveBeenCalledWith('/app/live-session')
  })
})
