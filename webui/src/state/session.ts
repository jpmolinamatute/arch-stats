import { reactive } from 'vue';
import type { components } from '../types/types.generated';

type SessionsRead = components['schemas']['SessionsRead'];

type Mutable<T> = {
    -readonly [P in keyof T]: T[P];
};

type MutableSessionsRead = Partial<Mutable<SessionsRead>>;

export const openSession = reactive<MutableSessionsRead>({
    id: undefined,
    is_opened: undefined,
    location: undefined,
    start_time: undefined,
    end_time: undefined,
});

export function clearOpenSession() {
    openSession.id = undefined;
    openSession.is_opened = undefined;
    openSession.location = undefined;
    openSession.start_time = undefined;
    openSession.end_time = undefined;
}
export async function fetchOpenSession(): Promise<void> {
    try {
        const response = await fetch('/api/v0/session/open');
        const json = await response.json();
        if (response.ok && json.data) {
            Object.assign(openSession, json.data as SessionsRead);
        } else {
            clearOpenSession();
        }
    } catch (error) {
        console.error('Failed to fetch open session:', error);
        clearOpenSession();
    }
}
