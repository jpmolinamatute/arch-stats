import type { components } from '@/types/types.generated';

type HTTPValidationError = components['schemas']['HTTPValidationError'];
type ValidationError = components['schemas']['ValidationError'];
type ErrorJson = HTTPValidationError | { detail?: string | ValidationError[] };

export class ApiError extends Error {
    public status: number;
    public data: unknown;

    constructor(message: string, status: number, data?: unknown) {
        super(message);
        this.status = status;
        this.data = data;
        this.name = 'ApiError';
    }
}

function extractErrorMessage(err: ErrorJson | null | undefined): string | null {
    if (!err) return null;

    // Handle array of validation errors
    if ('detail' in err && Array.isArray(err.detail)) {
        const first = err.detail[0];
        if (first && typeof first.msg === 'string') {
            return first.msg;
        }
    }

    // Handle simple string detail
    if ('detail' in err && typeof err.detail === 'string') {
        return err.detail;
    }

    return null;
}

async function handleResponse<T>(response: Response): Promise<T> {
    if (!response.ok) {
        let errorMessage = `Request failed: ${response.status}`;
        let errorData: unknown;

        try {
            errorData = await response.json();
            const extracted = extractErrorMessage(errorData as ErrorJson);
            if (extracted) {
                errorMessage = extracted;
            }
        } catch {
            // If JSON parsing fails, try text
            try {
                const text = await response.text();
                if (text) errorMessage = `${errorMessage} - ${text}`;
            } catch {
                // Ignore text parsing error
            }
        }

        throw new ApiError(errorMessage, response.status, errorData);
    }

    // Handle 204 No Content
    if (response.status === 204) {
        return undefined as T;
    }

    return response.json() as Promise<T>;
}

const BASE_URL = '/api/v0';

interface RequestOptions extends RequestInit {
    params?: Record<string, string | number | boolean | undefined>;
}

async function request<T>(endpoint: string, options: RequestOptions = {}): Promise<T> {
    const { params, ...init } = options;

    let url = `${BASE_URL}${endpoint}`;
    if (params) {
        const searchParams = new URLSearchParams();
        Object.entries(params).forEach(([key, value]) => {
            if (value !== undefined) {
                searchParams.append(key, String(value));
            }
        });
        const queryString = searchParams.toString();
        if (queryString) {
            url += `?${queryString}`;
        }
    }

    const headers = new Headers(init.headers);
    if (!headers.has('Content-Type') && !(init.body instanceof FormData)) {
        headers.set('Content-Type', 'application/json');
    }

    const config: RequestInit = {
        ...init,
        headers,
        credentials: 'include', // Always include credentials for auth
    };

    const response = await fetch(url, config);
    return handleResponse<T>(response);
}

export const api = {
    get: <T>(endpoint: string, options?: RequestOptions) =>
        request<T>(endpoint, { ...options, method: 'GET' }),
    post: <T>(endpoint: string, body?: unknown, options?: RequestOptions) =>
        request<T>(endpoint, {
            ...options,
            method: 'POST',
            body: body ? JSON.stringify(body) : undefined,
        }),
    put: <T>(endpoint: string, body?: unknown, options?: RequestOptions) =>
        request<T>(endpoint, {
            ...options,
            method: 'PUT',
            body: body ? JSON.stringify(body) : undefined,
        }),
    patch: <T>(endpoint: string, body?: unknown, options?: RequestOptions) =>
        request<T>(endpoint, {
            ...options,
            method: 'PATCH',
            body: body ? JSON.stringify(body) : undefined,
        }),
    delete: <T>(endpoint: string, options?: RequestOptions) =>
        request<T>(endpoint, { ...options, method: 'DELETE' }),
    createShot: (payload: components['schemas']['ShotCreate']) =>
        request<components['schemas']['ShotId']>('/shot', {
            method: 'POST',
            body: JSON.stringify(payload),
        }),
};
