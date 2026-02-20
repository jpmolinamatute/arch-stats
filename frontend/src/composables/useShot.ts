import type { components } from '@/types/types.generated'
import { ref } from 'vue'
import { api, ApiError } from '@/api/client'

type ShotCreate = components['schemas']['ShotCreate']
type ShotScore = components['schemas']['ShotScore']
type Stats = components['schemas']['Stats']
type LiveStat = components['schemas']['LiveStat']

// Manual definition since WS schema is not in OpenAPI client types fully
interface WebSocketMessage
{
    content: LiveStat
    content_type: 'shot.created' | string
}

const loading = ref(false)
const error = ref<string | null>(null)
const shots = ref<ShotScore[]>([])
const stats = ref<Stats | null>(null)

export function useShot() {
    /**
     * Create a new shot for the current authenticated archer.
     * Backend enforces that the shot's slot belongs to the caller.
     *
     * Returns the created shot_id as a string.
     */
    async function createShot(
        payload: ShotCreate | ShotCreate[],
    ): Promise<string | string[]> {
        loading.value = true
        error.value = null

        // Recursive helper to process shots in valid batches
        async function processBatch(shots: ShotCreate[]): Promise<string[]> {
            const results: string[] = []

            // Base case: empty
            if (shots.length === 0)
                return []

            // Case 1: Less than 3 shots
            // Backend requires min 3 for batch, so send these individually
            if (shots.length < 3) {
                for (const shot of shots) {
                    const result = await api.createShot(shot)
                    if (result && !Array.isArray(result)) {
                        results.push(result.shot_id)
                    }
                }
                return results
            }

            // Case 2: 3 to 10 shots (Valid batch size)
            if (shots.length <= 10) {
                const result = await api.createShot(shots)
                if (Array.isArray(result)) {
                    return result.map(s => s.shot_id)
                }
                // Fallback if single ID returned (shouldn't happen for array input but type safe)
                return result ? [result.shot_id] : []
            }

            // Case 3: More than 10 shots
            // Split into a batch of 10 and process the rest recursively
            const chunk = shots.slice(0, 10)
            const remainder = shots.slice(10)

            const chunkResult = await api.createShot(chunk)
            if (Array.isArray(chunkResult)) {
                results.push(...chunkResult.map(s => s.shot_id))
            }

            const remainderResult = await processBatch(remainder)
            results.push(...remainderResult)

            return results
        }

        try {
            // Handle single shot input
            if (!Array.isArray(payload)) {
                const data = await api.createShot(payload)
                if (!data)
                    throw new Error('No response from server')
                return Array.isArray(data) ? data.map(s => s.shot_id) : data.shot_id
            }

            // Handle array input with batching logic
            return await processBatch(payload)
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

    async function fetchShots(slotId: string) {
        loading.value = true
        error.value = null
        shots.value = []
        stats.value = null
        try {
            // Updated to use the new stats endpoint which returns both shots and stats
            const data = await api.get<LiveStat>(`/stats/${slotId}`)
            if (data) {
                shots.value = data.scores
                stats.value = data.stats
            }
        }
        catch (e) {
            error.value = e instanceof Error ? e.message : 'Failed to fetch shots'
            console.error(e)
        }
        finally {
            loading.value = false
        }
    }

    function subscribeToShots(slotId: string): WebSocket {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
        const host = window.location.host
        const wsUrl = `${protocol}//${host}/api/v0/stats/ws/${slotId}`

        const socket = new WebSocket(wsUrl)

        socket.onopen = () => {
            // Connected to WS
        }

        socket.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data) as WebSocketMessage

                if (data.content_type === 'shot.created' && data.content) {
                    // Update state directly from WS payload

                    const newShots = data.content.scores
                    // Avoid duplicates
                    const existingIds = new Set(shots.value.map(s => s.shot_id))
                    const uniqueNewShots = newShots.filter(s => !existingIds.has(s.shot_id))

                    shots.value = [...shots.value, ...uniqueNewShots]
                    stats.value = data.content.stats
                }
            }
            catch (e) {
                console.error('[useShot] Failed to parse WS message', e)
            }
        }

        socket.onclose = () => {
            // Disconnected from WS
        }

        socket.onerror = (err) => {
            console.error('[useShot] WS Error', err)
        }

        return socket
    }

    const shotCount = ref<number>(0)

    async function fetchShotCount(slotId: string) {
        loading.value = true
        error.value = null
        try {
            const data = await api.get<number>(`/shot/count-by-slot/${slotId}`)
            if (typeof data === 'number') {
                shotCount.value = data
            }
        }
        catch (e) {
            error.value = e instanceof Error ? e.message : 'Failed to fetch shot count'
            console.error(e)
        }
        finally {
            loading.value = false
        }
    }

    return {
        shots,
        shotCount,
        stats,
        loading,
        error,
        createShot,
        fetchShots,
        fetchShotCount,
        subscribeToShots,
    } as const
}
