export let session_id: string | null = null;

export async function checkSessionOpen(): Promise<boolean> {
    try {
        const resp = await fetch('/api/v0/session/open');
        if (!resp.ok) {
            session_id = null;
            return false;
        }
        const json = await resp.json();
        if (json.data && json.data.id) {
            session_id = json.data.id;
            return true;
        } else {
            session_id = null;
            return false;
        }
    } catch {
        session_id = null;
        return false;
    }
}

export async function closeSessionFlow(): Promise<void> {
    if (!session_id) {
        // No session to close; optionally show an error or return silently
        return;
    }

    const now = new Date().toISOString();

    const payload = {
        is_opened: false,
        end_time: now,
    };

    await fetch(`/api/v0/session/${session_id}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
    });
    session_id = null;
}

export async function openSessionFlow(container: HTMLElement, onSuccess?: () => void) {
    const static_path = '/src/components/sessions_new/sessions_new';
    const html = await fetch(`${static_path}.html`).then((r) => r.text());
    container.innerHTML = html;

    // Inject CSS if not already present
    const styleId = 'session-css';
    if (!document.getElementById(styleId)) {
        const cssText = await fetch(`${static_path}.css`).then((r) => r.text());
        const style = document.createElement('style');
        style.id = styleId;
        style.textContent = cssText;
        document.head.appendChild(style);
    }

    // Pre-fill start_time to now (ISO format for input[type=datetime-local])
    const startTimeInput = document.getElementById('start_time') as HTMLInputElement;
    if (startTimeInput) {
        const now = new Date();
        const pad = (n: number) => n.toString().padStart(2, '0');
        const local = `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())}T${pad(now.getHours())}:${pad(now.getMinutes())}`;
        startTimeInput.value = local;
    }

    // Handle form submit
    const form = document.getElementById('session-form') as HTMLFormElement | null;
    const errorDiv = document.getElementById('session-form-error');
    if (form) {
        form.onsubmit = async (event) => {
            event.preventDefault();
            if (errorDiv) errorDiv.textContent = '';

            // Collect values
            const data = {
                is_opened: true,
                start_time: (form.elements.namedItem('start_time') as HTMLInputElement).value,
                location: (form.elements.namedItem('location') as HTMLInputElement).value,
            };

            if (!data.start_time || !data.location) {
                if (errorDiv) errorDiv.textContent = 'Start time and location are required.';
                return;
            }

            let startTime = data.start_time;
            if (startTime.length === 16) startTime += ':00'; // Add seconds if not present

            try {
                const resp = await fetch('/api/v0/session', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data),
                });
                if (!resp.ok) {
                    const err = await resp.json().catch(() => ({}));
                    if (errorDiv)
                        errorDiv.textContent =
                            (err.errors && err.errors[0]) || 'Failed to create session.';
                    return;
                }
                if (onSuccess) {
                    onSuccess();
                }
                container.innerHTML = '';
            } catch (e) {
                if (errorDiv) errorDiv.textContent = 'Network error, please try again.';
                console.error(e);
            }
        };
    }
}
