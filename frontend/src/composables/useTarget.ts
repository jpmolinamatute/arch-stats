import type { components } from '../types/types.generated';

type Target = components['schemas']['TargetsRead'];
type TargetsCreate = components['schemas']['TargetsCreate'];

function isRecord(v: unknown): v is Record<string, unknown> {
    return typeof v === 'object' && v !== null;
}

function isTarget(v: unknown): v is Target {
    if (!isRecord(v)) return false;
    // Minimal invariant check; extend as schema evolves.
    return typeof v.id === 'string';
}

function isTargetArray(v: unknown): v is Target[] {
    return Array.isArray(v) && v.every(isTarget);
}

export async function createTarget(payload: TargetsCreate): Promise<Target | null> {
    try {
        const response = await fetch('/api/v0/target', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(payload),
        });
        const json: unknown = await response.json();
        const data = isRecord(json) ? json['data'] : undefined;
        if (response.ok && isTarget(data)) {
            return data;
        }
        const errors = isRecord(json) && Array.isArray(json.errors) ? json.errors : 'Unknown error';
        console.error('Failed to create target:', errors);
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
    const body: unknown = await res.json();
    if (!res.ok) {
        const errs = isRecord(body) ? (body['errors'] as unknown) : undefined;
        const msg = Array.isArray(errs) ? errs.join(', ') : 'Failed to list targets';
        throw new Error(msg);
    }
    const data = isRecord(body) ? (body['data'] as unknown) : null;
    return isTargetArray(data) ? data : [];
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
    const body: unknown = await res.json().catch(() => ({}));
    if (!res.ok) {
        const errs = isRecord(body) ? (body['errors'] as unknown) : undefined;
        const msg = Array.isArray(errs) ? errs.join(', ') : 'Failed to fetch target';
        throw new Error(msg);
    }
    const data = isRecord(body) ? (body['data'] as unknown) : null;
    return isTarget(data) ? data : null;
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
    const body: unknown = await res.json().catch(() => ({}));
    if (!res.ok) {
        const errs = isRecord(body) ? (body['errors'] as unknown) : undefined;
        const msg = Array.isArray(errs) ? errs.join(', ') : 'Failed to delete target';
        throw new Error(msg);
    }
}

/**
 * Call backend calibration endpoint to get a suggested Target-like payload.
 */
export async function calibrateTarget(): Promise<Target | null> {
    const res = await fetch('/api/v0/target/calibrate');
    const body: unknown = await res.json().catch(() => ({}));
    if (!res.ok) {
        const errs = isRecord(body) ? (body['errors'] as unknown) : undefined;
        const msg = Array.isArray(errs) ? errs.join(', ') : 'Failed to calibrate target';
        throw new Error(msg);
    }
    const data = isRecord(body) ? (body['data'] as unknown) : null;
    return isTarget(data) ? data : null;
}
