type SessionId = string | null;
type SessionListener = (sessionId: SessionId) => void;

class SessionStore {
    private value: SessionId = null;
    private listeners: SessionListener[] = [];
    private loading: boolean = false;

    get(): SessionId {
        return this.value;
    }

    /**
     * Fetches session state from server and notifies listeners.
     * Only one request is in flight at a time.
     */
    async refresh() {
        if (this.loading) return; // prevent concurrent fetches
        this.loading = true;
        let newValue: SessionId = null;
        try {
            const resp = await fetch('/api/v0/session/open');
            if (resp.ok) {
                const json = await resp.json();
                if (json.data && json.data.id) {
                    newValue = json.data.id as string;
                }
            }
        } catch {
            // Ignore error; treat as no open session
        }
        this.loading = false;
        if (this.value !== newValue) {
            this.value = newValue;
            this.listeners.forEach((fn) => fn(this.value));
        }
    }

    /**
     * Subscribe to changes. Listener is called immediately with current value.
     * Returns an unsubscribe function.
     */
    subscribe(fn: SessionListener) {
        this.listeners.push(fn);
        fn(this.value);
        return () => {
            this.listeners = this.listeners.filter((listener) => listener !== fn);
        };
    }
}

export const sessionStore = new SessionStore();
