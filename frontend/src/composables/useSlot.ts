import { ref } from 'vue';
import type { components, operations } from '@/types/types.generated';
import { useAuth } from '@/composables/useAuth';

type SlotJoinRequest = components['schemas']['SlotJoinRequest'];
type SlotJoinResponse = components['schemas']['SlotJoinResponse'];
type FullSlotInfo = components['schemas']['FullSlotInfo'];
type HTTPValidationError = components['schemas']['HTTPValidationError'];
type ValidationError = components['schemas']['ValidationError'];
type ErrorJson = HTTPValidationError | { detail?: string };
type MeResponseOk =
    operations['get_current_user_api_v0_auth_me_get']['responses']['200']['content']['application/json'];

const currentSlot = ref<FullSlotInfo | null>(null);
const loading = ref(false);
const error = ref<string | null>(null);

export function useSlot() {
    // Local cache (localStorage) helpers
    const SLOT_CACHE_PREFIX = 'arch-stats:slot:';
    function getCacheKey(archerId: string): string {
        return `${SLOT_CACHE_PREFIX}${archerId}`;
    }

    function getLS(): Storage | null {
        try {
            return typeof window !== 'undefined' && window.localStorage
                ? window.localStorage
                : null;
        } catch {
            return null;
        }
    }

    function setSlotCache(archerId: string, slot: FullSlotInfo): void {
        try {
            // Store only the slot object; no timestamp needed.
            const ls = getLS();
            ls?.setItem(getCacheKey(archerId), JSON.stringify(slot));
        } catch {
            /* ignore storage errors */
        }
    }

    function readSlotCache(archerId: string): FullSlotInfo | null {
        try {
            const ls = getLS();
            const raw = ls?.getItem(getCacheKey(archerId));
            if (!raw) return null;
            const parsed = JSON.parse(raw) as unknown;
            // Backward compatibility: previously we stored { ts, slot }.
            if (typeof parsed === 'object' && parsed !== null && 'slot' in parsed) {
                const maybe = parsed as { slot?: FullSlotInfo | null };
                return maybe.slot ?? null;
            }
            // New format: the cached value is the FullSlotInfo itself.
            return parsed as FullSlotInfo;
        } catch {
            return null;
        }
    }

    function clearSlotCache(archerId?: string): void {
        try {
            const { user } = useAuth();
            const aid = archerId ?? user.value?.archer_id ?? null;
            if (aid) {
                const ls = getLS();
                ls?.removeItem(getCacheKey(aid));
            }
        } catch {
            /* ignore */
        }
    }

    function extractErrorMessage(err: ErrorJson | null | undefined): string | null {
        if (!err) return null;
        const maybeDetail = (err as { detail?: unknown }).detail;
        if (Array.isArray(maybeDetail)) {
            const first = maybeDetail[0] as ValidationError | undefined;
            if (first && typeof first.msg === 'string' && first.msg.length > 0) {
                return first.msg;
            }
        }
        if (typeof maybeDetail === 'string' && maybeDetail.length > 0) {
            return maybeDetail;
        }
        return null;
    }

    /**
     * Join a session by requesting a slot assignment.
     * Returns the slot ID.
     */
    async function joinSession(payload: SlotJoinRequest): Promise<SlotJoinResponse> {
        loading.value = true;
        error.value = null;
        try {
            const response = await fetch('/api/v0/session/slot', {
                method: 'POST',
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(payload),
            });

            if (!response.ok) {
                // Try to get error details from response
                let errorMessage = `Failed to join session: ${response.status}`;
                try {
                    const errorData = (await response.json()) as ErrorJson;
                    const msg = extractErrorMessage(errorData);
                    if (msg) errorMessage = msg;
                } catch (parseError) {
                    console.error('Could not parse error response:', parseError);
                }

                if (response.status === 409) {
                    throw new Error('You are already participating in an open session');
                }
                if (response.status === 422) {
                    throw new Error('Session not found or is closed');
                }
                if (response.status === 401) {
                    throw new Error('Not authenticated. Please sign in again.');
                }
                if (response.status === 400) {
                    throw new Error(errorMessage);
                }
                throw new Error(errorMessage);
            }

            const data = (await response.json()) as SlotJoinResponse;
            // After joining, try to populate state and cache quickly
            try {
                const { user } = useAuth();
                if (user.value?.archer_id) {
                    const full = await getSlot(true);
                    currentSlot.value = full;
                    setSlotCache(user.value.archer_id, full);
                }
            } catch {
                // non-fatal
            }
            return data;
        } catch (e) {
            error.value = e instanceof Error ? e.message : 'Failed to join session';
            throw e;
        } finally {
            loading.value = false;
        }
    }

    /**
     * Get current slot details for the authenticated archer.
     * Resolves the archer_id from the session (/api/v0/auth/me) to avoid mismatches.
     * Endpoint: GET /api/v0/session/slot/archer/{archer_id}
     */
    async function getSlot(forceRefresh = false): Promise<FullSlotInfo> {
        loading.value = true;
        error.value = null;
        try {
            // Prefer using the current user from the auth composable
            const { user } = useAuth();
            let archerId = user.value?.archer_id;

            // Fallback: fetch /auth/me if user is not initialized yet
            if (!archerId) {
                const meResponse = await fetch('/api/v0/auth/me', {
                    method: 'GET',
                    credentials: 'include',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                });
                if (!meResponse.ok) {
                    if (meResponse.status === 401) {
                        throw new Error('Not authenticated. Please sign in again.');
                    }
                    throw new Error(`Failed to resolve user (HTTP ${meResponse.status})`);
                }
                const meData = (await meResponse.json()) as MeResponseOk;
                archerId = meData.archer.archer_id as string;
            }

            // Try cache first unless forced refresh
            if (!forceRefresh && archerId) {
                const cached = readSlotCache(archerId);
                if (cached) {
                    currentSlot.value = cached;
                    return cached;
                }
            }

            const response = await fetch(`/api/v0/session/slot/archer/${archerId}`, {
                method: 'GET',
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json',
                },
            });

            if (!response.ok) {
                let errorMessage = `Failed to fetch slot: ${response.status}`;
                try {
                    const errorData = (await response.json()) as ErrorJson;
                    const msg = extractErrorMessage(errorData);
                    if (msg) errorMessage = msg;
                } catch (parseError) {
                    console.error('Could not parse error response:', parseError);
                }

                if (response.status === 404) {
                    throw new Error('Slot not found');
                }
                if (response.status === 401) {
                    throw new Error('Not authenticated. Please sign in again.');
                }
                throw new Error(errorMessage);
            }

            const full = (await response.json()) as FullSlotInfo;
            // Update state and cache
            currentSlot.value = full;
            if (archerId) setSlotCache(archerId, full);
            return full;
        } catch (e) {
            error.value = e instanceof Error ? e.message : 'Failed to fetch slot';
            throw e;
        } finally {
            loading.value = false;
        }
    }

    /**
     * Leave a session (remove archer from slot).
     */
    /**
     * Leave a session (remove archer from slot) using slot_id in the URL path.
     */
    async function leaveSession(slotId: string): Promise<void> {
        loading.value = true;
        error.value = null;
        try {
            const response = await fetch(`/api/v0/session/slot/leave/${slotId}`, {
                method: 'PATCH',
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json',
                },
            });

            if (!response.ok) {
                let errorMessage = `Failed to leave session: ${response.status}`;
                try {
                    const errorData = (await response.json()) as ErrorJson;
                    const msg = extractErrorMessage(errorData);
                    if (msg) errorMessage = msg;
                } catch (parseError) {
                    console.error('Could not parse error response:', parseError);
                }

                if (response.status === 409) {
                    throw new Error('Archer is not participating in this session');
                }
                if (response.status === 422) {
                    throw new Error('Session not found or is closed');
                }
                if (response.status === 403) {
                    throw new Error('Not authorized to leave this session');
                }
                if (response.status === 401) {
                    throw new Error('Not authenticated. Please sign in again.');
                }
                throw new Error(errorMessage);
            }

            // Clear current slot after leaving
            currentSlot.value = null;
            clearSlotCache();
        } catch (e) {
            error.value = e instanceof Error ? e.message : 'Failed to leave session';
            throw e;
        } finally {
            loading.value = false;
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
            const { user } = useAuth();
            const archerId = user.value?.archer_id;
            return archerId ? readSlotCache(archerId) : null;
        },
        clearSlotCache,
    } as const;
}
