import type { components } from '../types/types.generated';

export interface FaceCalibration {
    x: number;
    y: number;
    radii: number[];
}

export type Face = components['schemas']['FacesRead'];
export type FaceCreate = components['schemas']['Face']; // creation payload shape (no id, target_id injected server-side)

function isRecord(v: unknown): v is Record<string, unknown> {
    return typeof v === 'object' && v !== null;
}

function isFace(v: unknown): v is Face {
    if (!isRecord(v)) return false;
    return typeof v.id === 'string' && typeof v.target_id === 'string';
}

function isFaceArray(v: unknown): v is Face[] {
    return Array.isArray(v) && v.every(isFace);
}

/**
 * Create one or more faces for a target (1..3). Endpoint returns list[UUID].
 * Caller does not currently need hydrated faces; fetch separately if needed.
 */
export async function createFacesForTarget(
    targetId: string,
    faces: FaceCreate[],
): Promise<string[]> {
    const res = await fetch(`/api/v0/target/${encodeURIComponent(targetId)}/face`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(faces),
    });
    const body: unknown = await res.json().catch(() => ({}));
    if (!res.ok) {
        const errs = isRecord(body) ? (body.errors as unknown) : undefined;
        const msg = Array.isArray(errs) ? errs.join(', ') : 'Failed to create faces';
        throw new Error(msg);
    }
    const data = isRecord(body) ? (body.data as unknown) : null;
    return Array.isArray(data) && data.every((v) => typeof v === 'string')
        ? (data as string[])
        : [];
}

/** List faces for a given target id. */
export async function listFacesForTarget(targetId: string): Promise<Face[]> {
    const res = await fetch(`/api/v0/target/${encodeURIComponent(targetId)}/face`);
    const body: unknown = await res.json().catch(() => ({}));
    if (!res.ok) {
        const errs = isRecord(body) ? (body.errors as unknown) : undefined;
        const msg = Array.isArray(errs) ? errs.join(', ') : 'Failed to list faces';
        throw new Error(msg);
    }
    const data = isRecord(body) ? (body.data as unknown) : null;
    return isFaceArray(data) ? data : [];
}

/** List all faces with optional filtering params. */
export async function listAllFaces(
    params?: Record<string, string | number | boolean | undefined | null>,
): Promise<Face[]> {
    const qs = new URLSearchParams();
    if (params) {
        for (const [k, v] of Object.entries(params)) {
            if (v === undefined || v === null) continue;
            qs.set(k, String(v));
        }
    }
    const url = `/api/v0/face${qs.toString() ? `?${qs.toString()}` : ''}`;
    const res = await fetch(url);
    const body: unknown = await res.json().catch(() => ({}));
    if (!res.ok) {
        const errs = isRecord(body) ? (body.errors as unknown) : undefined;
        const msg = Array.isArray(errs) ? errs.join(', ') : 'Failed to list faces';
        throw new Error(msg);
    }
    const data = isRecord(body) ? (body.data as unknown) : null;
    return isFaceArray(data) ? data : [];
}

/** Delete a face by id. Endpoint: DELETE /api/v0/face/{face_id} */
export async function deleteFace(faceId: string): Promise<void> {
    const res = await fetch(`/api/v0/face/${encodeURIComponent(faceId)}`, {
        method: 'DELETE',
    });
    if (res.ok) return;
    let msg = 'Failed to delete face';
    const body = await res.json().catch(() => null);
    if (isRecord(body) && Array.isArray(body.errors)) {
        msg = body.errors.join(', ');
    }
    throw new Error(msg);
}

/** Fetch sensor-provided calibration for a single face proposal. */
export async function fetchFaceCalibration(): Promise<FaceCalibration> {
    const res = await fetch('/api/v0/face/calibrate');
    const body: unknown = await res.json().catch(() => ({}));
    if (!res.ok) {
        throw new Error('Failed to calibrate face');
    }
    if (typeof body !== 'object' || body === null) {
        throw new Error('Malformed face calibration envelope');
    }
    const data: unknown = (body as Record<string, unknown>).data;
    interface RawFaceCal {
        x?: unknown;
        y?: unknown;
        radii?: unknown;
    }
    const raw = data as RawFaceCal | null;
    if (
        !raw ||
        typeof raw.x !== 'number' ||
        typeof raw.y !== 'number' ||
        !Array.isArray(raw.radii)
    ) {
        throw new Error('Malformed face calibration payload');
    }
    if (!raw.radii.every((r) => typeof r === 'number')) {
        throw new Error('Invalid radii types in face calibration');
    }
    return { x: raw.x, y: raw.y, radii: raw.radii };
}
