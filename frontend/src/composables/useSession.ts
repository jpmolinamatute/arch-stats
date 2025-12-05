import type { components } from '@/types/types.generated'
import { computed, ref } from 'vue'
import { api, ApiError } from '@/api/client'
import { useSlot } from './useSlot'

type SessionRead = components['schemas']['SessionRead']
type SessionCreate = components['schemas']['SessionCreate']
type SessionId = components['schemas']['SessionId']

const currentSession = ref<SessionRead | null>(null)
const loading = ref(false)
const error = ref<string | null>(null)

export function useSession() {
  const hasOpenSession = computed(() => currentSession.value !== null)

  // Local cache (localStorage) helpers
  const SESSION_CACHE_PREFIX = 'arch-stats:session:'
  function getCacheKey(archerId: string): string {
    return `${SESSION_CACHE_PREFIX}${archerId}`
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

  function setSessionCache(archerId: string, session: SessionRead): void {
    try {
      const ls = getLS()
      ls?.setItem(getCacheKey(archerId), JSON.stringify(session))
    }
    catch {
      /* ignore storage errors */
    }
  }

  function readSessionCache(archerId: string): SessionRead | null {
    try {
      const ls = getLS()
      const raw = ls?.getItem(getCacheKey(archerId))
      if (!raw)
        return null
      return JSON.parse(raw) as SessionRead
    }
    catch {
      return null
    }
  }

  function clearSessionCache(archerId?: string): void {
    try {
      // If no specific archerId provided, we can't easily clear specific session
      // unless we store current archerId somewhere else or pass it in.
      // For now, we'll rely on the caller passing it or just clearing if we have a user context.
      // But useSession doesn't have direct access to user context inside these helpers unless we pass it.
      // Let's assume the caller handles it or we use the one passed to functions.
      if (archerId) {
        const ls = getLS()
        ls?.removeItem(getCacheKey(archerId))
      }
    }
    catch {
      /* ignore */
    }
  }

  /**
   * Check if the authenticated archer has an open session.
   * Returns the session if found, null otherwise.
   */
  async function checkForOpenSession(archerId: string): Promise<SessionRead | null> {
    loading.value = true
    error.value = null
    try {
      // Try cache first
      const cached = readSessionCache(archerId)
      if (cached) {
        // We have a cached session, but we must verify it's still valid with the backend.
        // If we return it immediately, we risk showing a stale session that was closed on another device
        // or if the backend state changed.
        // So we just set it to currentSession for immediate UI feedback, but continue to fetch.
        currentSession.value = cached
      }

      // Step 1: Check if the archer has an active slot (is participating)
      // This is the happy path for active users.
      let sessionId: string | null = null
      const slot = await api.get<{ session_id: string }>(
        `/session/slot/archer/${archerId}`,
        { ignoreStatus: [404] },
      )

      if (slot) {
        sessionId = slot.session_id
      }

      if (!sessionId) {
        const idData = await api.get<SessionId>(
          `/session/archer/${archerId}/open-session`,
          { ignoreStatus: [404] },
        )
        if (idData) {
          sessionId = idData.session_id || null
        }
      }

      if (!sessionId) {
        currentSession.value = null
        clearSessionCache(archerId)
        return null
      }

      // Step 3: Fetch full session details
      const session = await api.get<SessionRead>(`/session/${sessionId}`)

      // Handle potential null if session fetch fails with ignored status (though we didn't ignore any here)
      if (!session) {
        throw new Error('Failed to fetch session details')
      }

      currentSession.value = session
      setSessionCache(archerId, session)
      return session
    }
    catch (e) {
      // If 404, it means no open session found - this is a valid state
      if (e instanceof ApiError && e.status === 404) {
        currentSession.value = null
        clearSessionCache(archerId)
        return null
      }

      // Only set error for unexpected failures, not for "not participating"
      error.value = e instanceof Error ? e.message : 'Failed to check for open session'
      console.error('[useSession] Error checking for open session:', e)
      return null
    }
    finally {
      loading.value = false
    }
  }

  /**
   * Create a new shooting session.
   * Returns the created session ID.
   */
  async function createSession(payload: SessionCreate): Promise<string> {
    loading.value = true
    error.value = null
    try {
      const data = await api.post<SessionId>('/session', payload)

      if (!data || !data.session_id) {
        throw new Error('No session ID returned from server')
      }

      // Update current session
      const session: SessionRead = {
        session_id: data.session_id,
        owner_archer_id: payload.owner_archer_id,
        session_location: payload.session_location,
        is_indoor: payload.is_indoor,
        is_opened: payload.is_opened,
        shot_per_round: payload.shot_per_round,
        created_at: new Date().toISOString(),
        closed_at: null,
      }
      currentSession.value = session
      setSessionCache(payload.owner_archer_id, session)

      return data.session_id
    }
    catch (e) {
      if (e instanceof ApiError) {
        if (e.status === 409) {
          error.value = 'You already have an open session'
        }
        else if (e.status === 403) {
          error.value = 'Not authorized to create session for this archer'
        }
        else if (e.status === 401) {
          error.value = 'Not authenticated. Please sign in again.'
        }
        else {
          error.value = e.message
        }
      }
      else {
        error.value = e instanceof Error ? e.message : 'Failed to create session'
      }
      throw e
    }
    finally {
      loading.value = false
    }
  }

  /**
   * Close the current session.
   * This will automatically leave the slot before closing.
   * @param sessionId - The session to close
   */
  async function closeSession(sessionId: string): Promise<void> {
    loading.value = true
    error.value = null
    try {
      // First, leave the slot if the archer is in one (required before closing)
      const { leaveSession, getSlot, clearSlotCache } = useSlot()
      try {
        // Try to get the slot for this archer in this session
        const slot = await getSlot()
        if (slot && slot.slot_id) {
          await leaveSession(slot.slot_id)
        }
      }
      catch (leaveError) {
        // Silently ignore 404/409 errors - archer not in slot is acceptable when closing
        if (
          leaveError instanceof Error
          && (leaveError.message.includes('not participating')
            || leaveError.message.includes('not found'))
        ) {
          // Not in slot, proceed
        }
        else {
          // Re-throw any other errors
          throw leaveError
        }
      }

      // Then close the session
      try {
        await api.patch('/session/close', { session_id: sessionId })
      }
      catch (e) {
        if (e instanceof ApiError) {
          if (e.status === 422) {
            throw new Error('Cannot close session with active participants')
          }
          if (e.status === 404) {
            throw new Error('Session not found')
          }
        }
        throw e
      }

      // Clear current session
      const archerId = currentSession.value?.owner_archer_id
      currentSession.value = null

      if (archerId) {
        clearSessionCache(archerId)
      }

      // Also clear any cached slot data to avoid stale state on next load
      try {
        clearSlotCache()
      }
      catch {
        /* ignore */
      }
    }
    catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to close session'
      throw e
    }
    finally {
      loading.value = false
    }
  }
  return {
    currentSession,
    hasOpenSession,
    loading,
    error,
    checkForOpenSession,
    createSession,
    closeSession,
    // Expose cache helpers if needed, or just use internally
    clearSessionCache,
  } as const
}
