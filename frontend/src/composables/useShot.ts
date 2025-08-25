import type { components } from '../types/types.generated';

type Shot = components['schemas']['ShotsRead'];

/**
 * List shots with optional filters (e.g., arrow_id, x, y, arrow_engage_time).
 */
export async function listShots(
    params?: Record<string, string | number | boolean | undefined | null>,
): Promise<Shot[]> {
    const qs = new URLSearchParams();
    if (params) {
        for (const [k, v] of Object.entries(params)) {
            if (v === undefined || v === null) continue;
            qs.set(k, String(v));
        }
    }
    const url = `/api/v0/shot${qs.toString() ? `?${qs.toString()}` : ''}`;
    const res = await fetch(url);
    const body = await res.json().catch(() => ({}));
    if (!res.ok) {
        throw new Error(body?.errors?.join(', ') ?? 'Failed to list shots');
    }
    return (body?.data as Shot[] | null) ?? [];
}

/** Convenience helper: list shots for a given session id (dedicated endpoint). */
export async function getShotsBySessionId(sessionId: string): Promise<Shot[]> {
    const res = await fetch(`/api/v0/shot/session-id/${encodeURIComponent(sessionId)}`);
    const body = await res.json().catch(() => ({}));
    if (!res.ok) {
        throw new Error(body?.errors?.join(', ') ?? 'Failed to fetch shots by session id');
    }
    return (body?.data as Shot[] | null) ?? [];
}

/** Convenience helper: list shots for a given arrow id via query filter. */
export async function getShotsByArrowId(arrowId: string): Promise<Shot[]> {
    return listShots({ arrow_id: arrowId });
}

/** Fetch a single shot by id. Returns null if 404. */
export async function getShotById(shotId: string): Promise<Shot | null> {
    const res = await fetch(`/api/v0/shot/${encodeURIComponent(shotId)}`);
    if (res.status === 404) {
        return null;
    }
    const body = await res.json().catch(() => ({}));
    if (!res.ok) {
        throw new Error(body?.errors?.join(', ') ?? 'Failed to fetch shot');
    }
    return (body?.data as Shot | null) ?? null;
}

/** Delete a shot by id. */
export async function deleteShot(shotId: string): Promise<void> {
    const res = await fetch(`/api/v0/shot/${encodeURIComponent(shotId)}`, {
        method: 'DELETE',
    });
    if (res.status === 204) return;
    const body = await res.json().catch(() => ({}));
    if (!res.ok) {
        throw new Error(body?.errors?.join(', ') ?? 'Failed to delete shot');
    }
}
