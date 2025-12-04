import type { components } from '@/types/types.generated'
import { ref } from 'vue'
import { api, ApiError } from '@/api/client'
import { useAuth } from '@/composables/useAuth'

type SlotJoinRequest = components['schemas']['SlotJoinRequest']
type SlotJoinResponse = components['schemas']['SlotJoinResponse']
type FullSlotInfo = components['schemas']['FullSlotInfo']

const currentSlot = ref<FullSlotInfo | null>(null)
const loading = ref(false)
const error = ref<string | null>(null)
// Use a single auth composable instance at module scope to preserve reactive state
const { user } = useAuth()

export function useSlot() {
  // Local cache (localStorage) helpers
  const SLOT_CACHE_PREFIX = 'arch-stats:slot:'
  function getCacheKey(archerId: string): string {
    return `${SLOT_CACHE_PREFIX}${archerId}`
  }

  function getLS(): Storage | null {
    try {
      return typeof window !== 'undefined' && window.localStorage
        ? window.localStorage
        : null
    }
    catch {
      return null
    }
  }

  function setSlotCache(archerId: string, slot: FullSlotInfo): void {
    try {
      // Store only the slot object.
      const ls = getLS()
      ls?.setItem(getCacheKey(archerId), JSON.stringify(slot))
    }
    catch {
      /* ignore storage errors */
    }
  }

  function readSlotCache(archerId: string): FullSlotInfo | null {
    try {
      const ls = getLS()
      const raw = ls?.getItem(getCacheKey(archerId))
      if (!raw)
        return null
      return JSON.parse(raw) as FullSlotInfo
    }
    catch {
      return null
    }
  }

  function clearSlotCache(archerId?: string): void {
    try {
      const aid = archerId ?? user.value?.archer_id ?? null
      if (aid) {
        const ls = getLS()
        ls?.removeItem(getCacheKey(aid))
      }
    }
    catch {
      /* ignore */
    }
  }

  /**
   * Join a session by requesting a slot assignment.
   * Returns the slot ID.
   */
  async function joinSession(payload: SlotJoinRequest): Promise<SlotJoinResponse> {
    loading.value = true
    error.value = null
    try {
      const data = await api.post<SlotJoinResponse>('/session/slot', payload)

      // After joining, try to populate state and cache quickly
      try {
        if (user.value?.archer_id) {
          const full = await getSlot(true)
          currentSlot.value = full
          setSlotCache(user.value.archer_id, full)
        }
      }
      catch {
        // non-fatal
      }
      return data
    }
    catch (e) {
      if (e instanceof ApiError) {
        if (e.status === 409) {
          error.value = 'You are already participating in an open session'
        }
        else if (e.status === 422) {
          error.value = 'Session not found or is closed'
        }
        else if (e.status === 401) {
          error.value = 'Not authenticated. Please sign in again.'
        }
        else {
          error.value = e.message
        }
      }
      else {
        error.value = e instanceof Error ? e.message : 'Failed to join session'
      }
      throw e
    }
    finally {
      loading.value = false
    }
  }

  /**
   * Get current slot details for the authenticated archer.
   * Resolves the archer_id from the session (/api/v0/auth/me) to avoid mismatches.
   * Endpoint: GET /api/v0/session/slot/archer/{archer_id}
   */
  async function getSlot(forceRefresh = false): Promise<FullSlotInfo> {
    loading.value = true
    error.value = null
    try {
      const archerId = user.value?.archer_id
      if (!archerId) {
        throw new Error('User context missing; auth must be initialized before slot fetch')
      }

      // Try cache first unless forced refresh
      if (!forceRefresh) {
        const cached = readSlotCache(archerId)
        if (cached) {
          currentSlot.value = cached
          return cached
        }
      }

      const full = await api.get<FullSlotInfo>(`/session/slot/archer/${archerId}`)

      currentSlot.value = full
      setSlotCache(archerId, full)
      return full
    }
    catch (e) {
      if (e instanceof ApiError) {
        if (e.status === 404) {
          throw new Error('Slot not found')
        }
        if (e.status === 401) {
          throw new Error('Not authenticated. Please sign in again.')
        }
        // Propagate original message for other errors
        error.value = e.message
      }
      else {
        error.value = e instanceof Error ? e.message : 'Failed to fetch slot'
      }
      throw e
    }
    finally {
      loading.value = false
    }
  }

  /**
   * Leave a session (remove archer from slot).
   */
  async function leaveSession(slotId: string): Promise<void> {
    loading.value = true
    error.value = null
    try {
      await api.patch(`/session/slot/leave/${slotId}`)

      // Clear current slot after leaving
      currentSlot.value = null
      clearSlotCache()
    }
    catch (e) {
      if (e instanceof ApiError) {
        if (e.status === 409) {
          throw new Error('Archer is not participating in this session')
        }
        if (e.status === 422) {
          throw new Error('Session not found or is closed')
        }
        if (e.status === 403) {
          throw new Error('Not authorized to leave this session')
        }
        if (e.status === 401) {
          throw new Error('Not authenticated. Please sign in again.')
        }
        error.value = e.message
      }
      else {
        error.value = e instanceof Error ? e.message : 'Failed to leave session'
      }
      throw e
    }
    finally {
      loading.value = false
    }
  }

  return {
    currentSlot,
    loading,
    error,
    joinSession,
    getSlot,
    leaveSession,
    // cache utilities (optional exports)
    getSlotCached: (): FullSlotInfo | null => {
      const archerId = user.value?.archer_id
      return archerId ? readSlotCache(archerId) : null
    },
    clearSlotCache,
  } as const
}
