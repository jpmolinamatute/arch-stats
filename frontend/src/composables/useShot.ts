import type { components } from '@/types/types.generated'
import { ref } from 'vue'
import { api, ApiError } from '@/api/client'

type ShotCreate = components['schemas']['ShotCreate']
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
        payload: ShotCreate | ShotCreate[],
    // opts?: { signal?: AbortSignal }, // Signal not supported in api.createShot yet, removing for now
    ): Promise<string | string[]> {
        loading.value = true
        error.value = null
        try {
            const data = await api.createShot(payload)

            if (!data)
                throw new Error('No response from server')

            if (Array.isArray(data)) {
                return data.map(s => s.shot_id)
            }
            return data.shot_id
        }
        catch (e) {
            if (e instanceof ApiError) {
                if (e.status === 401) {
                    throw new Error('Not authenticated. Please sign in again.')
                }
                if (e.status === 403) {
                    throw new Error('Forbidden: You are not allowed to add a shot to this slot')
                }
                if (e.status === 404) {
                    throw new Error('Slot not found')
                }
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
