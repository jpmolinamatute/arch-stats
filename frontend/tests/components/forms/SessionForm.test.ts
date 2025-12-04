import type { ComputedRef } from 'vue'
import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { computed, ref } from 'vue'
import SessionForm from '@/components/forms/SessionForm.vue'
import { useAuth } from '@/composables/useAuth'
import { useSession } from '@/composables/useSession'

interface UserSession {
  archer_id: string
  email: string
  first_name?: string | null
  last_name?: string | null
  picture_url?: string | null
  is_admin?: boolean
}

// Mock composables
vi.mock('@/composables/useSession', () => ({
  useSession: vi.fn(),
}))
vi.mock('@/composables/useAuth', () => ({
  useAuth: vi.fn(),
}))

describe('sessionForm', () => {
  const mockCreateSession = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders form correctly', () => {
    vi.mocked(useSession).mockReturnValue({
      createSession: mockCreateSession,
      loading: ref(false),
      error: ref(null),
      currentSession: ref(null),
      checkForOpenSession: vi.fn(),
      closeSession: vi.fn(),
      hasOpenSession: computed(() => false) as unknown as ComputedRef<boolean>,
      clearSessionCache: vi.fn(),
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
      initialized: ref(true),
      initError: ref(null),
      prompting: ref(false),
      initOneTap: vi.fn(),
      beginGoogleLogin: vi.fn(),
      clientId: 'test-client-id',
    })

    const wrapper = mount(SessionForm)
    expect(wrapper.find('input[type="text"]').exists()).toBe(true) // Location
    expect(wrapper.find('input[type="checkbox"]').exists()).toBe(true) // Indoor
    expect(wrapper.find('input[type="number"]').exists()).toBe(true) // Shots per round
  })

  it('shows validation error if location is empty', async () => {
    vi.mocked(useSession).mockReturnValue({
      createSession: mockCreateSession,
      loading: ref(false),
      error: ref(null),
      currentSession: ref(null),
      checkForOpenSession: vi.fn(),
      closeSession: vi.fn(),
      hasOpenSession: computed(() => false),
      clearSessionCache: vi.fn(),
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
      initialized: ref(true),
      initError: ref(null),
      prompting: ref(false),
      initOneTap: vi.fn(),
      beginGoogleLogin: vi.fn(),
      clientId: 'test-client-id',
    })

    const wrapper = mount(SessionForm)

    // Submit without filling location
    await wrapper.find('form').trigger('submit')

    expect(mockCreateSession).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('Session location is required')
  })

  it('submits form with correct data', async () => {
    vi.mocked(useSession).mockReturnValue({
      createSession: mockCreateSession.mockResolvedValue('sess_123'),
      loading: ref(false),
      error: ref(null),
      currentSession: ref(null),
      checkForOpenSession: vi.fn(),
      closeSession: vi.fn(),
      hasOpenSession: computed(() => false),
      clearSessionCache: vi.fn(),
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
      initialized: ref(true),
      initError: ref(null),
      prompting: ref(false),
      initOneTap: vi.fn(),
      beginGoogleLogin: vi.fn(),
      clientId: 'test-client-id',
    })

    const wrapper = mount(SessionForm)

    await wrapper.find('input[type="text"]').setValue('Range A')
    await wrapper.find('input[type="checkbox"]').setValue(true) // Indoor
    await wrapper.find('input[type="number"]').setValue(3) // Shots per round

    await wrapper.find('form').trigger('submit')

    expect(mockCreateSession).toHaveBeenCalledWith({
      owner_archer_id: 'archer_1',
      session_location: 'Range A',
      is_indoor: true,
      is_opened: true,
      shot_per_round: 3,
    })
    expect(wrapper.emitted('sessionCreated')).toBeTruthy()
    expect(wrapper.emitted('sessionCreated')?.[0]).toEqual(['sess_123'])
  })

  it('displays error from createSession', async () => {
    vi.mocked(useSession).mockReturnValue({
      createSession: mockCreateSession.mockRejectedValue(new Error('Network error')),
      loading: ref(false),
      error: ref(null),
      currentSession: ref(null),
      checkForOpenSession: vi.fn(),
      closeSession: vi.fn(),
      hasOpenSession: computed(() => false) as unknown as ComputedRef<boolean>,
      clearSessionCache: vi.fn(),
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
      initialized: ref(true),
      initError: ref(null),
      prompting: ref(false),
      initOneTap: vi.fn(),
      beginGoogleLogin: vi.fn(),
      clientId: 'test-client-id',
    })

    const wrapper = mount(SessionForm)

    const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

    await wrapper.find('input[type="text"]').setValue('Range A')
    await wrapper.find('form').trigger('submit')

    // Wait for async operation
    await new Promise(resolve => setTimeout(resolve, 0))

    expect(wrapper.text()).toContain('Network error')
    expect(consoleErrorSpy).toHaveBeenCalledWith('Error creating session:', expect.any(Error))

    consoleErrorSpy.mockRestore()
  })
})
