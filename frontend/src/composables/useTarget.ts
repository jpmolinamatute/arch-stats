import type { components } from '../types/types.generated';

type Target = components['schemas']['TargetsRead'];
type TargetsCreate = components['schemas']['TargetsCreate'];

export async function createTarget(payload: TargetsCreate): Promise<Target | null> {
    try {
        const response = await fetch('/api/v0/target', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(payload),
        });
        const json = await response.json();
        if (response.ok && json.data) {
            return json.data as Target;
        }
        console.error('Failed to create target:', json.errors ?? 'Unknown error');
        return null;
    } catch (err) {
        console.error('Error creating target:', err);
        return null;
    }
}

/**
 * List targets with optional filters (e.g., session_id, max_x, human_identifier).
 */
export async function listTargets(
    params?: Record<string, string | number | boolean | undefined | null>,
): Promise<Target[]> {
    const qs = new URLSearchParams();
    if (params) {
        for (const [k, v] of Object.entries(params)) {
            if (v === undefined || v === null) continue;
            qs.set(k, String(v));
        }
    }
    const url = `/api/v0/target${qs.toString() ? `?${qs.toString()}` : ''}`;
    const res = await fetch(url);
    const body = await res.json();
    if (!res.ok) {
        throw new Error(body?.errors?.join(', ') ?? 'Failed to list targets');
    }
    return (body?.data as Target[] | null) ?? [];
}

/** Convenience filter for targets by session id. */
export async function getTargetsBySessionId(sessionId: string): Promise<Target[]> {
    return listTargets({ session_id: sessionId });
}

/**
 * Fetch a single target by id.
 */
export async function getTargetById(targetId: string): Promise<Target | null> {
    const res = await fetch(`/api/v0/target/${encodeURIComponent(targetId)}`);
    if (res.status === 404) {
        return null;
    }
    const body = await res.json().catch(() => ({}));
    if (!res.ok) {
        throw new Error(body?.errors?.join(', ') ?? 'Failed to fetch target');
    }
    return (body?.data as Target | null) ?? null;
}

/**
 * Delete a target by id.
 */
export async function deleteTarget(targetId: string): Promise<void> {
    const res = await fetch(`/api/v0/target/${encodeURIComponent(targetId)}`, {
        method: 'DELETE',
    });
    if (res.status === 204) return;
    // Some handlers still return a JSON envelope; try to surface errors.
    const body = await res.json().catch(() => ({}));
    if (!res.ok) {
        throw new Error(body?.errors?.join(', ') ?? 'Failed to delete target');
    }
}

/**
 * Call backend calibration endpoint to get a suggested Target-like payload.
 */
export async function calibrateTarget(): Promise<Target | null> {
    const res = await fetch('/api/v0/target/calibrate');
    const body = await res.json().catch(() => ({}));
    if (!res.ok) {
        throw new Error(body?.errors?.join(', ') ?? 'Failed to calibrate target');
    }
    return (body?.data as Target | null) ?? null;
}
