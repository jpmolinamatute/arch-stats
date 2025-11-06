import { ref } from 'vue';
import type { components } from '@/types/types.generated';
import { useAuth } from '@/composables/useAuth';

type SlotJoinRequest = components['schemas']['SlotJoinRequest'];
type SlotJoinResponse = components['schemas']['SlotJoinResponse'];
type FullSlotInfo = components['schemas']['FullSlotInfo'];

const currentSlot = ref<FullSlotInfo | null>(null);
const loading = ref(false);
const error = ref<string | null>(null);

export function useSlot() {
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
                    const errorData = await response.json();
                    if (errorData.detail) {
                        errorMessage = errorData.detail;
                    }
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
    async function getSlot(): Promise<FullSlotInfo> {
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
                const meData = (await meResponse.json()) as {
                    archer: { archer_id: string };
                };
                archerId = meData.archer.archer_id;
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
                    const errorData = await response.json();
                    if (errorData.detail) {
                        errorMessage = errorData.detail;
                    }
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

            return (await response.json()) as FullSlotInfo;
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
                    const errorData = await response.json();
                    if (errorData.detail) {
                        errorMessage = errorData.detail;
                    }
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
    } as const;
}
