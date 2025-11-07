import { ref } from 'vue';
import type { components } from '@/types/types.generated';

type FaceMinimal = components['schemas']['FaceMinimal'];
type HTTPValidationError = components['schemas']['HTTPValidationError'];
type ErrorJson = HTTPValidationError | { detail?: string };

const faces = ref<FaceMinimal[]>([]);
const loading = ref(false);
const error = ref<string | null>(null);

export function useFaces() {
    async function listFaces(): Promise<FaceMinimal[]> {
        loading.value = true;
        error.value = null;
        try {
            const response = await fetch('/api/v0/faces', {
                method: 'GET',
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json',
                },
            });

            if (!response.ok) {
                let errorMessage = `Failed to fetch faces: ${response.status}`;
                try {
                    const errorData = (await response.json()) as ErrorJson;
                    // 422 validation error from FastAPI
                    if (Array.isArray((errorData as HTTPValidationError)?.detail)) {
                        const first = (errorData as HTTPValidationError).detail?.[0];
                        if (first && typeof first.msg === 'string') {
                            errorMessage = first.msg;
                        }
                    } else if (typeof errorData.detail === 'string') {
                        // Common FastAPI error shape for 401/403/404
                        errorMessage = errorData.detail;
                    }
                } catch (parseError) {
                    console.error('Error parsing error response:', parseError);
                }

                if (response.status === 401) {
                    throw new Error('Not authenticated. Please sign in again.');
                }
                throw new Error(errorMessage);
            }

            const data = (await response.json()) as FaceMinimal[];
            faces.value = Array.isArray(data) ? data : [];
            return faces.value;
        } catch (e) {
            error.value = e instanceof Error ? e.message : 'Failed to fetch faces';
            throw e;
        } finally {
            loading.value = false;
        }
    }

    return {
        loading,
        error,
        faces,
        listFaces,
    } as const;
}
