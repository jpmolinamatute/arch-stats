import { describe, it, expect, vi, beforeEach } from 'vitest';
import { mount } from '@vue/test-utils';
import { ref } from 'vue';
import AppContainer from '@/components/AppContainer.vue';
import { useAuth } from '@/composables/useAuth';

interface UserSession {
    archer_id: string;
    email: string;
    first_name?: string | null;
    last_name?: string | null;
    picture_url?: string | null;
    is_admin?: boolean;
}

// Mock useAuth
vi.mock('@/composables/useAuth', () => ({
    useAuth: vi.fn(),
}));

// Mock child components
vi.mock('@/components/layout/AppHeader.vue', () => ({
    default: { template: '<div data-testid="app-header"></div>' },
}));
vi.mock('@/components/auth/AuthGate.vue', () => ({
    default: { template: '<div data-testid="auth-gate"></div>' },
}));
vi.mock('@/components/SessionManager.vue', () => ({
    default: { template: '<div data-testid="session-manager"></div>' },
}));

describe('AppContainer', () => {
    const mockBootstrapAuth = vi.fn();

    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('calls bootstrapAuth on mount', () => {
        vi.mocked(useAuth).mockReturnValue({
            bootstrapAuth: mockBootstrapAuth,
            isAuthenticated: ref(false),
            loading: ref(false),
            pendingRegistration: ref(null),
            user: ref(null),
            logout: vi.fn(),
            disableGoogleAutoSelect: vi.fn(),
            registerNewArcher: vi.fn(),
            initialized: ref(true),
            initError: ref(null),
            prompting: ref(false),
            initOneTap: vi.fn(),
            beginGoogleLogin: vi.fn(),
            clientId: 'test-client-id',
        });

        mount(AppContainer);
        expect(mockBootstrapAuth).toHaveBeenCalled();
    });

    it('renders AuthGate when not authenticated', () => {
        vi.mocked(useAuth).mockReturnValue({
            bootstrapAuth: mockBootstrapAuth,
            isAuthenticated: ref(false),
            loading: ref(false),
            pendingRegistration: ref(null),
            user: ref(null),
            logout: vi.fn(),
            disableGoogleAutoSelect: vi.fn(),
            registerNewArcher: vi.fn(),
            initialized: ref(true),
            initError: ref(null),
            prompting: ref(false),
            initOneTap: vi.fn(),
            beginGoogleLogin: vi.fn(),
            clientId: 'test-client-id',
        });

        const wrapper = mount(AppContainer);
        expect(wrapper.find('[data-testid="auth-gate"]').exists()).toBe(true);
        expect(wrapper.find('[data-testid="session-manager"]').exists()).toBe(false);
        expect(wrapper.text()).toContain('Track Every Arrow');
    });

    it('renders SessionManager when authenticated', () => {
        vi.mocked(useAuth).mockReturnValue({
            bootstrapAuth: mockBootstrapAuth,
            isAuthenticated: ref(true),
            loading: ref(false),
            pendingRegistration: ref(null),
            user: ref<UserSession>({
                first_name: 'John',
                last_name: 'Doe',
                archer_id: 'archer_1',
                email: 'john@example.com',
                is_admin: false,
            }),
            logout: vi.fn(),
            disableGoogleAutoSelect: vi.fn(),
            registerNewArcher: vi.fn(),
            initialized: ref(true),
            initError: ref(null),
            prompting: ref(false),
            initOneTap: vi.fn(),
            beginGoogleLogin: vi.fn(),
            clientId: 'test-client-id',
        });

        const wrapper = mount(AppContainer);
        expect(wrapper.find('[data-testid="auth-gate"]').exists()).toBe(false);
        expect(wrapper.find('[data-testid="session-manager"]').exists()).toBe(true);
        expect(wrapper.text()).toContain('Session Management');
    });
});
