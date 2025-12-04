import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import AppHeader from '@/components/layout/AppHeader.vue'
import { useAuth } from '@/composables/useAuth'

interface UserSession {
  archer_id: string
  email: string
  first_name?: string | null
  last_name?: string | null
  picture_url?: string | null
  is_admin?: boolean
}

// Mock useAuth
vi.mock('@/composables/useAuth', () => ({
  useAuth: vi.fn(),
}))

describe('appHeader', () => {
  const mockLogout = vi.fn()
  const mockDisableGoogleAutoSelect = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders title correctly', () => {
    vi.mocked(useAuth).mockReturnValue({
      isAuthenticated: ref(false),
      user: ref(null),
      logout: mockLogout,
      disableGoogleAutoSelect: mockDisableGoogleAutoSelect,
      bootstrapAuth: vi.fn(),
      loading: ref(false),
      pendingRegistration: ref(null),
      registerNewArcher: vi.fn(),
      initialized: ref(true),
      initError: ref(null),
      prompting: ref(false),
      initOneTap: vi.fn(),
      beginGoogleLogin: vi.fn(),
      clientId: 'test-client-id',
    })

    const wrapper = mount(AppHeader)
    expect(wrapper.text()).toContain('Arch Stats')
  })

  it('renders user info and logout button when authenticated', async () => {
    vi.mocked(useAuth).mockReturnValue({
      isAuthenticated: ref(true),
      user: ref<UserSession>({
        first_name: 'John',
        last_name: 'Doe',
        archer_id: '123',
        email: 'john@example.com',
        is_admin: false,
      }),
      logout: mockLogout,
      disableGoogleAutoSelect: mockDisableGoogleAutoSelect,
      bootstrapAuth: vi.fn(),
      loading: ref(false),
      pendingRegistration: ref(null),
      registerNewArcher: vi.fn(),
      initialized: ref(true),
      initError: ref(null),
      prompting: ref(false),
      initOneTap: vi.fn(),
      beginGoogleLogin: vi.fn(),
      clientId: 'test-client-id',
    })

    const wrapper = mount(AppHeader)
    expect(wrapper.text()).toContain('John Doe')
    expect(wrapper.find('button').text()).toBe('Logout')

    await wrapper.find('button').trigger('click')
    expect(mockLogout).toHaveBeenCalled()
  })

  it('does not render user info when not authenticated', () => {
    vi.mocked(useAuth).mockReturnValue({
      isAuthenticated: ref(false),
      user: ref(null),
      logout: mockLogout,
      disableGoogleAutoSelect: mockDisableGoogleAutoSelect,
      bootstrapAuth: vi.fn(),
      loading: ref(false),
      pendingRegistration: ref(null),
      registerNewArcher: vi.fn(),
      initialized: ref(true),
      initError: ref(null),
      prompting: ref(false),
      initOneTap: vi.fn(),
      beginGoogleLogin: vi.fn(),
      clientId: 'test-client-id',
    })

    const wrapper = mount(AppHeader)
    expect(wrapper.find('button').exists()).toBe(false)
  })

  it('calls disableGoogleAutoSelect on mount if authenticated', () => {
    vi.mocked(useAuth).mockReturnValue({
      isAuthenticated: ref(true),
      user: ref<UserSession>({
        first_name: 'John',
        last_name: 'Doe',
        archer_id: '123',
        email: 'john@example.com',
        is_admin: false,
      }),
      logout: mockLogout,
      disableGoogleAutoSelect: mockDisableGoogleAutoSelect,
      bootstrapAuth: vi.fn(),
      loading: ref(false),
      pendingRegistration: ref(null),
      registerNewArcher: vi.fn(),
      initialized: ref(true),
      initError: ref(null),
      prompting: ref(false),
      initOneTap: vi.fn(),
      beginGoogleLogin: vi.fn(),
      clientId: 'test-client-id',
    })

    mount(AppHeader)
    expect(mockDisableGoogleAutoSelect).toHaveBeenCalled()
  })
})
