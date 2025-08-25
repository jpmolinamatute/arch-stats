import { sessionOpened, clearOpenSession, fetchOpenSession } from '../state/session';
import type { components } from '../types/types.generated';

type Session = components['schemas']['SessionsRead'];
type SessionsCreate = components['schemas']['SessionsCreate'];
type SessionsUpdate = components['schemas']['SessionsUpdate'];

export async function createSession(payload: SessionsCreate): Promise<void> {
    if (sessionOpened.is_opened === true) {
        console.warn('A session is already open. Cannot create a new session.');
        return;
    }
    try {
        const response = await fetch('/api/v0/session', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload),
        });

        if (response.ok) {
            fetchOpenSession();
        } else {
            const json = await response.json();
            throw new Error(json.errors?.join(', ') ?? 'Unknown error');
        }
    } catch (error) {
        console.error('Failed to create/open session:', error);
    }
}

export async function closeSession(): Promise<void> {
    if (!sessionOpened.is_opened) return;
    try {
        const response = await fetch(`/api/v0/session/${sessionOpened.id}`, {
            method: 'PATCH',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                is_opened: false,
                end_time: new Date().toISOString(),
            } as SessionsUpdate),
        });

        if (response.ok) {
            clearOpenSession();
        } else {
            const json = await response.json();
            throw new Error(json.errors?.join(', ') ?? 'Unknown error');
        }
    } catch (error) {
        console.error('Failed to close session:', error);
    }
}

// --- Additional Sessions endpoints ---

/**
 * Fetch a list of sessions. Optional params are serialized as query string.
 */
export async function listSessions(
    params?: Record<string, string | number | boolean | undefined | null>,
): Promise<Session[]> {
    const qs = new URLSearchParams();
    if (params) {
        for (const [k, v] of Object.entries(params)) {
            if (v === undefined || v === null) continue;
            qs.set(k, String(v));
        }
    }

    const url = `/api/v0/session${qs.toString() ? `?${qs.toString()}` : ''}`;
    const res = await fetch(url);
    const body = await res.json();
    if (!res.ok) {
        throw new Error(body?.errors?.join(', ') ?? 'Failed to list sessions');
    }
    // API envelope: { code, data, errors }
    return (body?.data as Session[] | null) ?? [];
}

/**
 * Fetch a single session by id.
 */
export async function getSessionById(sessionId: string): Promise<Session | null> {
    const res = await fetch(`/api/v0/session/${encodeURIComponent(sessionId)}`);
    const body = await res.json();
    if (!res.ok) {
        throw new Error(body?.errors?.join(', ') ?? 'Failed to fetch session');
    }
    return (body?.data as Session | null) ?? null;
}

/**
 * Delete a session by id. If it's the currently open session, clear local state.
 */
export async function deleteSession(sessionId: string): Promise<void> {
    const res = await fetch(`/api/v0/session/${encodeURIComponent(sessionId)}`, {
        method: 'DELETE',
    });
    if (!res.ok) {
        let msg = 'Failed to delete session';
        try {
            const body = await res.json();
            msg = body?.errors?.join(', ') ?? msg;
        } catch {
            // noop
        }
        throw new Error(msg);
    }
    if (sessionOpened.id === sessionId) {
        clearOpenSession();
    }
}
