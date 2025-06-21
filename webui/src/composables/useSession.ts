import { openSession, clearOpenSession, fetchOpenSession } from '../state/session';
import type { components } from '../types/types.generated';

type SessionsCreate = components['schemas']['SessionsCreate'];
type SessionsUpdate = components['schemas']['SessionsUpdate'];

export async function createSession(payload: SessionsCreate): Promise<void> {
    if (openSession.is_opened === true) {
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
    if (!openSession.is_opened) return;
    try {
        const response = await fetch(`/api/v0/session/${openSession.id}`, {
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
