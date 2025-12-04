import type { components } from '@/types/types.generated'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { api, ApiError } from '../../src/api/client'
import { useAuth } from '../../src/composables/useAuth'

// Mock the api client
vi.mock('../../src/api/client', () => ({
  api: {
    get: vi.fn(),
    post: vi.fn(),
  },
  ApiError: class extends Error {
    status: number
    constructor(message: string, status: number) {
      super(message)
      this.status = status
    }
  },
}))

describe('useAuth', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    // Mock window.google to bypass script loading
    vi.stubGlobal('google', {
      accounts: {
        id: {
          initialize: vi.fn(),
          prompt: vi.fn(),
          renderButton: vi.fn(),
          disableAutoSelect: vi.fn(),
        },
      },
    })
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    vi.useRealTimers()
  })

  it('bootstrapAuth sets user when authenticated', async () => {
    const { bootstrapAuth, user, isAuthenticated } = useAuth()

    type MeResponse = components['schemas']['AuthAuthenticated']
    const mockUser: MeResponse = {
      archer: {
        archer_id: '123',
        email: 'test@example.com',
        first_name: 'Test',
        last_name: 'User',
        google_picture_url: 'http://example.com/pic.jpg',
      },
      access_token: 'token',
      token_type: 'bearer',
    } as unknown as MeResponse

    vi.mocked(api.get).mockResolvedValue(mockUser)

    await bootstrapAuth()
    vi.runAllTimers() // Run any timers (e.g. promptOnce timeout)

    expect(isAuthenticated.value).toBe(true)
    expect(user.value).toEqual({
      archer_id: '123',
      email: 'test@example.com',
      first_name: 'Test',
      last_name: 'User',
      picture_url: 'http://example.com/pic.jpg',
    })
  })

  it('bootstrapAuth handles 401 correctly', async () => {
    const { bootstrapAuth, user, isAuthenticated } = useAuth()

    vi.mocked(api.get).mockRejectedValue(new ApiError('Unauthorized', 401))

    await bootstrapAuth()
    vi.runAllTimers()

    expect(isAuthenticated.value).toBe(false)
    expect(user.value).toBeNull()
  })

  it('logout clears session', async () => {
    const { logout, user, isAuthenticated } = useAuth()

    // Set initial state
    isAuthenticated.value = true
    user.value = { archer_id: '1', email: 'a@b.c' }

    await logout()

    expect(api.post).toHaveBeenCalledWith('/auth/logout')
    expect(isAuthenticated.value).toBe(false)
    expect(user.value).toBeNull()
  })
})
