import { ref } from 'vue';
import type { components, operations } from '@/types/types.generated';

interface UserSession {
    archer_id: string;
    email: string;
    first_name?: string | null;
    last_name?: string | null;
    picture_url?: string | null;
}

// Backend /api/v0/auth/google response shape (OpenAPI generated)
type AuthLoginResponseBody =
    | components['schemas']['AuthAuthenticated']
    | components['schemas']['AuthNeedsRegistration'];

function isAuthAuthenticated(v: unknown): v is components['schemas']['AuthAuthenticated'] {
    return (
        !!v &&
        typeof v === 'object' &&
        'archer' in (v as Record<string, unknown>) &&
        'access_token' in (v as Record<string, unknown>)
    );
}

function isAuthNeedsRegistration(v: unknown): v is components['schemas']['AuthNeedsRegistration'] {
    return (
        !!v &&
        typeof v === 'object' &&
        'google_email' in (v as Record<string, unknown>) &&
        'google_subject' in (v as Record<string, unknown>)
    );
}

const isAuthenticated = ref(false);
const user = ref<UserSession | null>(null);
const loading = ref(true);
const initialized = ref(false);
const initError = ref<string | null>(null);
// Prevent multiple concurrent FedCM prompt() calls
const prompting = ref(false);
type PendingRegistration = {
    credential: string; // keep last id token
    google_email: string;
    google_subject: string;
    given_name: string | null;
    family_name: string | null;
    need_first_name: boolean;
    need_last_name: boolean;
};
const pendingRegistration = ref<PendingRegistration | null>(null);

const GOOGLE_SCRIPT_SRC = 'https://accounts.google.com/gsi/client';
const CLIENT_ID = import.meta.env.ARCH_STATS_GOOGLE_OAUTH_CLIENT_ID as string | undefined;

async function loadScript(): Promise<void> {
    // Access Google API dynamically to avoid hard Window typing dependency
    const maybeApi: unknown = (window as unknown as { google?: unknown })?.google;
    if (maybeApi) return;
    await new Promise<void>((resolve, reject) => {
        const s = document.createElement('script');
        s.src = GOOGLE_SCRIPT_SRC;
        s.async = true;
        s.defer = true;
        s.onload = () => resolve();
        s.onerror = () => reject(new Error('Failed to load Google Identity Services script'));
        document.head.appendChild(s);
    });
}

// On app load, we initialize One Tap and prompt; session persistence happens via cookie only.
async function bootstrapAuth(): Promise<void> {
    try {
        loading.value = true;

        // Check if we have a valid auth cookie by calling /auth/me
        try {
            const meResponse = await fetch('/api/v0/auth/me', {
                method: 'GET',
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json',
                },
            });

            if (meResponse.ok) {
                // Successfully authenticated via cookie
                type MeResponse =
                    operations['get_current_user_api_v0_auth_me_get']['responses']['200']['content']['application/json'];
                const authData = (await meResponse.json()) as MeResponse;

                user.value = {
                    archer_id: authData.archer.archer_id,
                    email: authData.archer.email,
                    first_name: authData.archer.first_name ?? null,
                    last_name: authData.archer.last_name ?? null,
                    picture_url: authData.archer.google_picture_url ?? null,
                };
                isAuthenticated.value = true;

                // Initialize Google One Tap but don't prompt
                await initGoogleOneTap(null);
                return;
            } else if (meResponse.status === 401) {
                // Not authenticated, need to prompt
                isAuthenticated.value = false;
                user.value = null;
            } else {
                // Other errors, assume not authenticated
                isAuthenticated.value = false;
                user.value = null;
            }
        } catch (checkError) {
            console.error('[bootstrapAuth] Error checking auth:', checkError);
            isAuthenticated.value = false;
            user.value = null;
        }

        // Initialize Google One Tap
        await initGoogleOneTap(null);
        // Only prompt if not already authenticated
        if (!isAuthenticated.value) {
            promptOnce();
        }
    } finally {
        loading.value = false;
    }
}

function handleCredentialResponse(resp: { credential: string }): void {
    prompting.value = false;
    beginGoogleLogin(resp.credential);
}

function promptOnce(): void {
    if (prompting.value) return;
    prompting.value = true;
    const maybeApi: unknown = (window as unknown as { google?: unknown })?.google;
    const accounts = (maybeApi as { accounts?: unknown })?.accounts as
        | { id?: { prompt?: () => void } }
        | undefined;
    accounts?.id?.prompt?.();
    window.setTimeout(() => {
        prompting.value = false;
    }, 15000);
}

async function initGoogleOneTap(container?: HTMLElement | null): Promise<void> {
    if (!CLIENT_ID) {
        initError.value = 'Missing ARCH_STATS_GOOGLE_OAUTH_CLIENT_ID';
        return;
    }
    try {
        await loadScript();
        // Initialize once per page load
        if (!initialized.value) {
            const maybeApi: unknown = (window as unknown as { google?: unknown })?.google;
            const idApi = (maybeApi as { accounts?: { id?: Record<string, unknown> } })?.accounts
                ?.id as
                | {
                      initialize?: (cfg: Record<string, unknown>) => void;
                  }
                | undefined;
            idApi?.initialize?.({
                client_id: CLIENT_ID,
                callback: handleCredentialResponse,
                auto_select: true,
                cancel_on_tap_outside: true,
                // Always use FedCM (modern privacy-preserving federated auth)
                use_fedcm_for_prompt: true,
                itp_support: true,
            });
            initialized.value = true;
        }

        // Always (re)render the button if a container is provided, even after initialization
        const maybeApi2: unknown = (window as unknown as { google?: unknown })?.google;
        const rb = (maybeApi2 as { accounts?: { id?: { renderButton?: unknown } } })?.accounts?.id
            ?.renderButton as
            | ((el: HTMLElement, opts: Record<string, unknown>) => void)
            | undefined;
        if (container && rb) {
            try {
                // Clear previous children if any to avoid duplicate renders
                container.innerHTML = '';
                rb(container, {
                    type: 'standard',
                    shape: 'rectangular',
                    theme: 'outline',
                    text: 'signin_with',
                    size: 'large',
                    logo_alignment: 'left',
                });
            } catch (rbErr) {
                console.error('Failed to render Google Sign-In button', rbErr);
            }
        }
    } catch (err) {
        initError.value = (err as Error).message;
        console.error('Google One Tap init error', err);
    }
}

async function beginGoogleLogin(idToken?: string): Promise<void> {
    // If called without token (e.g., button click) trigger prompt again.
    if (!idToken) {
        promptOnce();
        return;
    }
    loading.value = true;
    try {
        const res = await fetch('/api/v0/auth/google', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            body: JSON.stringify({ credential: idToken }),
        });
        if (!res.ok) throw new Error(`Login failed (HTTP ${res.status})`);
        // OpenAPI: returns union of AuthAuthenticated | AuthNeedsRegistration
        const data = (await res.json()) as AuthLoginResponseBody;
        if (isAuthAuthenticated(data)) {
            user.value = {
                archer_id: data.archer.archer_id,
                email: data.archer.email,
                first_name: data.archer.first_name,
                last_name: data.archer.last_name,
                picture_url: data.archer.google_picture_url ?? null,
            };
            isAuthenticated.value = true;
            // Cookie is set by backend; no local storage hints needed
            pendingRegistration.value = null;
        } else if (isAuthNeedsRegistration(data)) {
            // Needs registration: store minimal info and wait for user to submit form
            pendingRegistration.value = {
                credential: idToken,
                google_email: data.google_email,
                google_subject: data.google_subject,
                given_name: data.given_name ?? null,
                family_name: data.family_name ?? null,
                need_first_name: !data.given_name_provided,
                need_last_name: !data.family_name_provided,
            };
            isAuthenticated.value = false;
        } else {
            throw new Error('Unexpected auth response shape');
        }
    } catch (e) {
        console.error('Google login error', e);
    } finally {
        loading.value = false;
    }
}

type GenderType = components['schemas']['GenderType'];
type BowStyleType = components['schemas']['BowStyleType'];

async function registerNewArcher(input: {
    date_of_birth: string; // YYYY-MM-DD
    gender: GenderType;
    bowstyle: BowStyleType;
    draw_weight: number; // in pounds
    first_name?: string;
    last_name?: string;
}): Promise<void> {
    if (!pendingRegistration.value) throw new Error('No pending registration');
    loading.value = true;
    try {
        // Use generated type; draw_weight is now part of the registration request
        type RegistrationPayload = components['schemas']['AuthRegistrationRequest'];
        const payload: RegistrationPayload = {
            credential: pendingRegistration.value.credential,
            date_of_birth: input.date_of_birth,
            gender: input.gender,
            bowstyle: input.bowstyle,
            first_name: input.first_name,
            last_name: input.last_name,
            draw_weight: input.draw_weight,
        };
        const res = await fetch('/api/v0/auth/register', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            body: JSON.stringify(payload),
        });
        if (!res.ok) throw new Error(`Registration failed (HTTP ${res.status})`);
        type RegisterResponse =
            | operations['register_api_v0_auth_register_post']['responses']['201']['content']['application/json']
            | operations['register_api_v0_auth_register_post']['responses']['200']['content']['application/json'];
        const data = (await res.json()) as RegisterResponse;
        if (!isAuthAuthenticated(data)) {
            throw new Error('Unexpected registration response');
        }
        user.value = {
            archer_id: data.archer.archer_id,
            email: data.archer.email,
            first_name: data.archer.first_name,
            last_name: data.archer.last_name,
            picture_url: data.archer.google_picture_url ?? null,
        };
        isAuthenticated.value = true;
        pendingRegistration.value = null;

        // draw_weight persisted during registration; no follow-up PATCH required
    } finally {
        loading.value = false;
    }
}

async function logout(): Promise<void> {
    await fetch('/api/v0/auth/logout', { method: 'POST', credentials: 'include' });
    user.value = null;
    isAuthenticated.value = false;
}

function disableGoogleAutoSelect(): void {
    try {
        // Use dynamic access to avoid TS complaints in our narrowed type
        const maybeApi: unknown = (window as unknown as { google?: unknown })?.google;
        const accounts = (maybeApi as { accounts?: unknown })?.accounts;
        const idApi = (accounts as { id?: unknown })?.id as { disableAutoSelect?: () => void };
        idApi?.disableAutoSelect?.();
    } catch {
        /* noop */
    }
}

export function useAuth() {
    return {
        isAuthenticated,
        user,
        loading,
        initialized,
        initError,
        prompting,
        pendingRegistration,
        initGoogleOneTap,
        initOneTap: initGoogleOneTap,
        beginGoogleLogin,
        registerNewArcher,
        bootstrapAuth,
        clientId: CLIENT_ID,
        logout,
        disableGoogleAutoSelect,
    } as const;
}
