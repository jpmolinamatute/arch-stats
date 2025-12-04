import type { components } from '@/types/types.generated'
import { ref } from 'vue'

type ShotCreate = components['schemas']['ShotCreate']
type ShotId = components['schemas']['ShotId']
// type ShotRead = components['schemas']['ShotRead']; // For future expansion (e.g., list by slot)

const loading = ref(false)
const error = ref<string | null>(null)

export function useShot() {
  /**
   * Create a new shot for the current authenticated archer.
   * Backend enforces that the shot's slot belongs to the caller.
   *
   * Returns the created shot_id as a string.
   */
  async function createShot(
    payload: ShotCreate,
    opts?: { signal?: AbortSignal },
  ): Promise<string> {
    loading.value = true
    error.value = null
    try {
      const response = await fetch('/api/v0/shot', {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
        signal: opts?.signal,
      })

      if (!response.ok) {
        // Try to parse backend-provided detail for a better error
        let errorMessage = `Failed to create shot: ${response.status}`
        try {
          const errorData = (await response.json()) as { detail?: string }
          if (errorData?.detail)
            errorMessage = errorData.detail
        }
        catch {
          /* ignore parse errors */
        }

        if (response.status === 401) {
          throw new Error('Not authenticated. Please sign in again.')
        }
        if (response.status === 403) {
          throw new Error('Forbidden: You are not allowed to add a shot to this slot')
        }
        if (response.status === 404) {
          throw new Error('Slot not found')
        }
        if (response.status === 422) {
          // Validation error
          throw new Error(errorMessage)
        }
        throw new Error(errorMessage)
      }

      const data = (await response.json()) as ShotId
      if (!data.shot_id)
        throw new Error('No shot ID returned from server')
      return data.shot_id
    }
    catch (e) {
      if (e instanceof DOMException && e.name === 'AbortError') {
        // Surface a consistent message for aborted requests
        error.value = 'Request aborted'
        throw e
      }
      error.value = e instanceof Error ? e.message : 'Failed to create shot'
      throw e
    }
    finally {
      loading.value = false
    }
  }

  return {
    loading,
    error,
    createShot,
  } as const
}
