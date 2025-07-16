import { reactive } from 'vue';
import type { components } from '../types/types.generated';

type SessionsRead = components['schemas']['SessionsRead'];

type Mutable<T> = {
    -readonly [P in keyof T]: T[P];
};

type MutableSessionsRead = Partial<Mutable<SessionsRead>>;

export const sessionOpened = reactive<MutableSessionsRead>({
    id: undefined,
    is_opened: undefined,
    location: undefined,
    start_time: undefined,
    end_time: undefined,
});

export function clearOpenSession() {
    sessionOpened.id = undefined;
    sessionOpened.is_opened = undefined;
    sessionOpened.location = undefined;
    sessionOpened.start_time = undefined;
    sessionOpened.end_time = undefined;
}
export async function fetchOpenSession(): Promise<void> {
    try {
        const response = await fetch('/api/v0/session/open');
        const json = await response.json();
        if (response.ok && json.data) {
            Object.assign(sessionOpened, json.data as SessionsRead);
        } else {
            clearOpenSession();
        }
    } catch (error) {
        console.error('Failed to fetch open session:', error);
        clearOpenSession();
    }
}
