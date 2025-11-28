import { ref, computed } from 'vue';
import type { components } from '@/types/types.generated';
import { useSlot } from './useSlot';
import { api, ApiError } from '@/api/client';

type SessionRead = components['schemas']['SessionRead'];
type SessionCreate = components['schemas']['SessionCreate'];
type SessionId = components['schemas']['SessionId'];

const currentSession = ref<SessionRead | null>(null);
const loading = ref(false);
const error = ref<string | null>(null);

export function useSession() {
    const hasOpenSession = computed(() => currentSession.value !== null);

    /**
     * Check if the authenticated archer has an open session.
     * Returns the session if found, null otherwise.
     */
    async function checkForOpenSession(archerId: string): Promise<SessionRead | null> {
        loading.value = true;
        error.value = null;
        try {
            // First, check if the archer has an open session
            let idData: SessionId;
            try {
                idData = await api.get<SessionId>(`/session/archer/${archerId}/open-session`);
            } catch (e) {
                if (e instanceof ApiError && (e.status === 404 || e.status === 422)) {
                    // No open session found
                    currentSession.value = null;
                    return null;
                }
                throw e;
            }

            // If ${archerId} doesn't have an open session, return null
            if (!idData.session_id) {
                currentSession.value = null;
                return null;
            }

            // Now fetch the specific session details using the new endpoint
            const session = await api.get<SessionRead>(`/session/${idData.session_id}`);

            currentSession.value = session;
            return session;
        } catch (e) {
            error.value = e instanceof Error ? e.message : 'Failed to check for open session';
            console.error('[useSession] Error checking for open session:', e);
            return null;
        } finally {
            loading.value = false;
        }
    }

    /**
     * Create a new shooting session.
     * Returns the created session ID.
     */
    async function createSession(payload: SessionCreate): Promise<string> {
        loading.value = true;
        error.value = null;
        try {
            const data = await api.post<SessionId>('/session', payload);

            if (!data.session_id) {
                throw new Error('No session ID returned from server');
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
            };
            currentSession.value = session;

            return data.session_id;
        } catch (e) {
            if (e instanceof ApiError) {
                if (e.status === 409) {
                    error.value = 'You already have an open session';
                } else if (e.status === 403) {
                    error.value = 'Not authorized to create session for this archer';
                } else if (e.status === 401) {
                    error.value = 'Not authenticated. Please sign in again.';
                } else {
                    error.value = e.message;
                }
            } else {
                error.value = e instanceof Error ? e.message : 'Failed to create session';
            }
            throw e;
        } finally {
            loading.value = false;
        }
    }

    /**
     * Close the current session.
     * This will automatically leave the slot before closing.
     * @param sessionId - The session to close
     * @param archerId - The archer leaving and closing the session
     */
    async function closeSession(sessionId: string): Promise<void> {
        loading.value = true;
        error.value = null;
        try {
            // First, leave the slot if the archer is in one (required before closing)
            const { leaveSession, getSlot, clearSlotCache } = useSlot();
            try {
                // Try to get the slot for this archer in this session
                const slot = await getSlot();
                if (slot && slot.slot_id) {
                    await leaveSession(slot.slot_id);
                }
            } catch (leaveError) {
                // Silently ignore 404/409 errors - archer not in slot is acceptable when closing
                if (
                    leaveError instanceof Error &&
                    (leaveError.message.includes('not participating') ||
                        leaveError.message.includes('not found'))
                ) {
                    // Not in slot, proceed
                } else {
                    // Re-throw any other errors
                    throw leaveError;
                }
                console.log('[useSession] Archer not in slot, proceeding to close session');
            }

            // Then close the session
            try {
                await api.patch('/session/close', { session_id: sessionId });
            } catch (e) {
                if (e instanceof ApiError) {
                    if (e.status === 422) {
                        throw new Error('Cannot close session with active participants');
                    }
                    if (e.status === 404) {
                        throw new Error('Session not found');
                    }
                }
                throw e;
            }

            // Clear current session
            currentSession.value = null;

            // Also clear any cached slot data to avoid stale state on next load
            try {
                clearSlotCache();
            } catch {
                /* ignore */
            }
        } catch (e) {
            error.value = e instanceof Error ? e.message : 'Failed to close session';
            throw e;
        } finally {
            loading.value = false;
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
    } as const;
}
