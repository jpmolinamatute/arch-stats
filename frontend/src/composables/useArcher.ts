import type { components } from '@/types/types.generated'
import { ref } from 'vue'

type ArcherRead = components['schemas']['ArcherRead']

const loading = ref(false)
const error = ref<string | null>(null)

export function useArcher() {
  /**
   * Get archer details by ID.
   */
  async function getArcher(archerId: string): Promise<ArcherRead> {
    loading.value = true
    error.value = null
    try {
      const response = await fetch(`/api/v0/archer/${archerId}`, {
        method: 'GET',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
      })

      if (!response.ok) {
        let errorMessage = `Failed to fetch archer: ${response.status}`
        try {
          const errorData = await response.json()
          if (errorData.detail) {
            errorMessage = errorData.detail
          }
        }
        catch (parseError) {
          console.error('Could not parse error response:', parseError)
        }

        if (response.status === 404) {
          throw new Error('Archer not found')
        }
        if (response.status === 401) {
          throw new Error('Not authenticated. Please sign in again.')
        }
        throw new Error(errorMessage)
      }

      return (await response.json()) as ArcherRead
    }
    catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to fetch archer'
      throw e
    }
    finally {
      loading.value = false
    }
  }

  return {
    loading,
    error,
    getArcher,
  } as const
}
