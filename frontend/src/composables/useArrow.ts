import type { components } from '../types/types.generated';

type Arrow = components['schemas']['ArrowsRead'];
type ArrowsCreate = components['schemas']['ArrowsCreate'];
type ArrowsUpdate = components['schemas']['ArrowsUpdate'];

/** List arrows with optional filters (is_programmed, spine, human_identifier, etc.). */
export async function listArrows(
    params?: Record<string, string | number | boolean | undefined | null>,
): Promise<Arrow[]> {
    const qs = new URLSearchParams();
    if (params) {
        for (const [k, v] of Object.entries(params)) {
            if (v === undefined || v === null) continue;
            qs.set(k, String(v));
        }
    }
    const url = `/api/v0/arrow${qs.toString() ? `?${qs.toString()}` : ''}`;
    const res = await fetch(url);
    const body = await res.json().catch(() => ({}));
    if (!res.ok) {
        throw new Error(body?.errors?.join(', ') ?? 'Failed to list arrows');
    }
    return (body?.data as Arrow[] | null) ?? [];
}

/** Fetch a single arrow by id; returns null on 404. */
export async function getArrowById(arrowId: string): Promise<Arrow | null> {
    const res = await fetch(`/api/v0/arrow/${encodeURIComponent(arrowId)}`);
    if (res.status === 404) return null;
    const body = await res.json().catch(() => ({}));
    if (!res.ok) {
        throw new Error(body?.errors?.join(', ') ?? 'Failed to fetch arrow');
    }
    return (body?.data as Arrow | null) ?? null;
}

/** Create a new arrow; returns created arrow id string on success. */
export async function createArrow(payload: ArrowsCreate): Promise<string | null> {
    const res = await fetch('/api/v0/arrow', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
    });
    const body = await res.json().catch(() => ({}));
    if (!res.ok) {
        throw new Error(body?.errors?.join(', ') ?? 'Failed to create arrow');
    }
    // tests show the API returns the new id in data
    return (body?.data as string | null) ?? null;
}

/** Patch an arrow by id. */
export async function patchArrow(arrowId: string, update: ArrowsUpdate): Promise<void> {
    const res = await fetch(`/api/v0/arrow/${encodeURIComponent(arrowId)}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(update),
    });
    if (res.status === 202) return;
    const body = await res.json().catch(() => ({}));
    if (!res.ok) {
        throw new Error(body?.errors?.join(', ') ?? 'Failed to update arrow');
    }
}

/** Delete an arrow by id. */
export async function deleteArrow(arrowId: string): Promise<void> {
    const res = await fetch(`/api/v0/arrow/${encodeURIComponent(arrowId)}`, {
        method: 'DELETE',
    });
    if (res.status === 204) return;
    const body = await res.json().catch(() => ({}));
    if (!res.ok) {
        throw new Error(body?.errors?.join(', ') ?? 'Failed to delete arrow');
    }
}

/** Request a new UUID for arrow creation/programming. */
export async function getNewArrowUuid(): Promise<string> {
    const res = await fetch('/api/v0/arrow/new_arrow_uuid');
    const body = await res.json().catch(() => ({}));
    if (!res.ok) {
        throw new Error(body?.errors?.join(', ') ?? 'Failed to get new arrow UUID');
    }
    return body?.data as string;
}
