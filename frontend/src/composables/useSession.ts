import { ref, computed } from 'vue';
import type { components } from '@/types/types.generated';
import { useSlot } from './useSlot';

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
            const idResponse = await fetch(`/api/v0/session/archer/${archerId}/open-session`, {
                method: 'GET',
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json',
                },
            });

            if (!idResponse.ok) {
                if (idResponse.status === 404 || idResponse.status === 422) {
                    // No open session found
                    currentSession.value = null;
                    return null;
                }
                throw new Error(`Failed to check for open session: ${idResponse.status}`);
            }

            const idData = (await idResponse.json()) as SessionId;

            // If ${archerId} doesn't have an open session, return null
            if (!idData.session_id) {
                currentSession.value = null;
                return null;
            }

            // Now fetch the specific session details using the new endpoint
            const sessionResponse = await fetch(`/api/v0/session/${idData.session_id}`, {
                method: 'GET',
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json',
                },
            });

            if (!sessionResponse.ok) {
                throw new Error(`Failed to fetch session details: ${sessionResponse.status}`);
            }

            const session = (await sessionResponse.json()) as SessionRead;
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
            const response = await fetch('/api/v0/session', {
                method: 'POST',
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(payload),
            });

            if (!response.ok) {
                // Try to get error details from response
                let errorMessage = `Failed to create session: ${response.status}`;
                try {
                    const errorData = await response.json();
                    if (errorData.detail) {
                        errorMessage = errorData.detail;
                    }
                } catch (parseError) {
                    console.error('Could not parse error response:', parseError);
                }

                if (response.status === 409) {
                    throw new Error('You already have an open session');
                }
                if (response.status === 403) {
                    throw new Error('Not authorized to create session for this archer');
                }
                if (response.status === 401) {
                    throw new Error('Not authenticated. Please sign in again.');
                }
                throw new Error(errorMessage);
            }

            const data = (await response.json()) as SessionId;

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
            error.value = e instanceof Error ? e.message : 'Failed to create session';
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
    async function closeSession(sessionId: string, archerId: string): Promise<void> {
        loading.value = true;
        error.value = null;
        try {
            // First, leave the slot if the archer is in one (required before closing)
            const { leaveSession, getSlot } = useSlot();
            try {
                // Try to get the slot for this archer in this session
                const slot = await getSlot(sessionId, archerId);
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
            const response = await fetch('/api/v0/session/close', {
                method: 'PATCH',
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ session_id: sessionId }),
            });

            if (!response.ok) {
                if (response.status === 422) {
                    throw new Error('Cannot close session with active participants');
                }
                if (response.status === 404) {
                    throw new Error('Session not found');
                }
                throw new Error(`Failed to close session: ${response.status}`);
            }

            // Clear current session
            currentSession.value = null;
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
