import type { components } from '@/types/types.generated'
import { ref } from 'vue'
import { api, ApiError } from '@/api/client'

type ShotCreate = components['schemas']['ShotCreate']
type ShotRead = components['schemas']['ShotRead']
// Manual definition since WS schema is not in OpenAPI
interface WebSocketMessage
{
    content: any
    content_type: 'shot_created' | string
}

const loading = ref(false)
const error = ref<string | null>(null)
const shots = ref<ShotRead[]>([])

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

    async function fetchShots(slotId: string) {
        loading.value = true
        error.value = null
        shots.value = [] // Clear previous shots to avoid stale data flash
        try {
            const data = await api.get<ShotRead[]>(`/shot/by-slot/${slotId}`)
            if (data) {
                shots.value = data
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
        const host = window.location.host // includes port if present
        // If running in dev mode with separate servers, this might need adjustment,
        // but typically vite proxies /api calls. The websocket endpoint is /api/v0/ws/{slot_id}
        // However, standard fetch proxying doesn't always apply to WS.
        // Let's assume the /api prefix is proxied or we use the backend URL directly.
        // For local dev typically backend is 8000, frontend 5173.

        // We can use a relative URL if the dev server proxies WS, which Vite does if configured.
        // Let's try relative path first, if that fails we might need an env var.
        const wsUrl = `${protocol}//${host}/api/v0/ws/${slotId}`

        const socket = new WebSocket(wsUrl)

        socket.onopen = () => {
            // Connected to WS
        }

        socket.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data) as WebSocketMessage
                // We expect content_type == 'shot_created' and content to be Stats object
                // The backend sends:
                // message = WebSocketMessage(content=payload, content_type=WSContentType.SHOT_CREATED)
                // payload is 'Stats' object which has 'shots': list[ShotScore], 'stats': LiveStat

                // Wait, checking backend ShotModel.listen_for_shots:
                // It yields 'Stats' object.
                // WebSocket sends WebSocketMessage where content is logic-less "object".

                // We need to fetch the full shot details?
                // The notification payload only has shot_id and score (ShotScore).
                // It does NOT have x, y, is_x, etc.
                //
                // Strategy: When we get a notification, we can either:
                // 1. Trust the score and append a "partial" shot (missing x/y).
                // 2. Re-fetch the list of shots (safe, easy).
                // 3. Update the implementation to send full shot data in notification.

                // Given the instructions to "come up with a plan" and I didn't spot this earlier:
                // The payload has shot_id and score.
                // Displaying "Score" in history is fine. But "Coords" will be missing.
                //
                // OPTION 2 is safest for consistency: Re-fetch shots.
                // Optimized: Fetch only the new shots? Backend doesn't support "get shot by id" easily yet?
                // Actually, re-fetching the list isn't too expensive for a single session.

                // Let's re-fetch for now to ensure we have full data (x, y).
                // In the future, we should improve the backend notification payload.

                if (data.content_type === 'shot_created') {
                    void fetchShots(slotId)
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

    return {
        shots,
        loading,
        error,
        createShot,
        fetchShots,
        subscribeToShots,
    } as const
}
