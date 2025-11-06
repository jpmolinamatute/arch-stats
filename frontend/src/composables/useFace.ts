import { ref } from 'vue';
import type { components } from '@/types/types.generated';

type Face = components['schemas']['Face'];
type FaceType = components['schemas']['FaceType'];

// Simple in-memory cache to avoid refetching identical faces repeatedly
const cache = new Map<string, Face>();
const inFlight: Record<string, AbortController> = {};

export function useFace() {
    const data = ref<Face | null>(null);
    const loading = ref(false);
    const error = ref<string | null>(null);

    async function fetchFace(faceId: FaceType): Promise<Face> {
        const key = String(faceId);
        if (cache.has(key)) {
            const cached = cache.get(key)!;
            data.value = cached;
            return cached;
        }

        // Abort any previous request for the same id
        if (inFlight[key]) {
            try {
                inFlight[key].abort();
            } catch {
                // ignore
            }
            delete inFlight[key];
        }

        const controller = new AbortController();
        inFlight[key] = controller;
        loading.value = true;
        error.value = null;

        try {
            const res = await fetch(`/api/v0/faces/${encodeURIComponent(key)}`, {
                method: 'GET',
                signal: controller.signal,
                headers: { Accept: 'application/json' },
            });
            if (!res.ok) {
                const text = await res.text();
                throw new Error(`Failed to fetch face ${key}: ${res.status} ${text}`);
            }
            const body = (await res.json()) as Face;
            cache.set(key, body);
            data.value = body;
            return body;
        } finally {
            loading.value = false;
            delete inFlight[key];
        }
    }

    return { data, loading, error, fetchFace } as const;
}

export async function getFace(faceId: FaceType): Promise<Face> {
    const { fetchFace } = useFace();
    return fetchFace(faceId);
}
