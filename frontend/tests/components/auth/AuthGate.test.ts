import { describe, it, expect, vi, beforeEach } from 'vitest';
import { mount } from '@vue/test-utils';
import { ref } from 'vue';
import AuthGate from '@/components/auth/AuthGate.vue';
import { useAuth } from '@/composables/useAuth';

interface UserSession {
    archer_id: string;
    email: string;
    first_name?: string | null;
    last_name?: string | null;
    picture_url?: string | null;
    is_admin?: boolean;
}
type ArcherRegistration = {
    credential: string;
    google_email: string;
    google_subject: string;
    given_name: string | null;
    family_name: string | null;
    need_first_name: boolean;
    need_last_name: boolean;
};

// Mock useAuth
vi.mock('@/composables/useAuth', () => ({
    useAuth: vi.fn(),
}));

describe('AuthGate', () => {
    const mockRegisterNewArcher = vi.fn();

    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('renders loading state', () => {
        vi.mocked(useAuth).mockReturnValue({
            loading: ref(true),
            isAuthenticated: ref(false),
            pendingRegistration: ref(null),
            user: ref(null),
            registerNewArcher: mockRegisterNewArcher,
            bootstrapAuth: vi.fn(),
            logout: vi.fn(),
            disableGoogleAutoSelect: vi.fn(),
            initialized: ref(true),
            initError: ref(null),
            prompting: ref(false),
            initOneTap: vi.fn(),
            beginGoogleLogin: vi.fn(),
            clientId: 'test-client-id',
        });

        const wrapper = mount(AuthGate);
        expect(wrapper.find('.animate-pulse').exists()).toBe(false); // It doesn't use animate-pulse class but has skeleton structure
        expect(wrapper.findAll('.bg-gray-100').length).toBe(3); // Check for skeleton lines
    });

    it('renders welcome message when authenticated', () => {
        vi.mocked(useAuth).mockReturnValue({
            loading: ref(false),
            isAuthenticated: ref(true),
            pendingRegistration: ref(null),
            user: ref<UserSession>({
                first_name: 'Jane',
                last_name: 'Doe',
                archer_id: '123',
                email: 'jane@example.com',
                is_admin: false,
            }),
            registerNewArcher: mockRegisterNewArcher,
            bootstrapAuth: vi.fn(),
            logout: vi.fn(),
            disableGoogleAutoSelect: vi.fn(),
            initialized: ref(true),
            initError: ref(null),
            prompting: ref(false),
            initOneTap: vi.fn(),
            beginGoogleLogin: vi.fn(),
            clientId: 'test-client-id',
        });

        const wrapper = mount(AuthGate);
        expect(wrapper.text()).toContain('Welcome Jane Doe');
    });

    it('renders registration form when pendingRegistration is present', () => {
        vi.mocked(useAuth).mockReturnValue({
            loading: ref(false),
            isAuthenticated: ref(false),
            pendingRegistration: ref<ArcherRegistration>({
                need_first_name: true,
                need_last_name: true,
                google_email: 'test@example.com',
                google_subject: '123',
                credential: 'token',
                given_name: null,
                family_name: null,
            }),
            user: ref(null),
            registerNewArcher: mockRegisterNewArcher,
            bootstrapAuth: vi.fn(),
            logout: vi.fn(),
            disableGoogleAutoSelect: vi.fn(),
            initialized: ref(true),
            initError: ref(null),
            prompting: ref(false),
            initOneTap: vi.fn(),
            beginGoogleLogin: vi.fn(),
            clientId: 'test-client-id',
        });

        const wrapper = mount(AuthGate);
        expect(wrapper.find('form').exists()).toBe(true);
        expect(wrapper.text()).toContain('Complete your registration');
        expect(wrapper.find('input[type="text"]').exists()).toBe(true); // First name input
    });

    it('renders sign in message when not authenticated and no pending registration', () => {
        vi.mocked(useAuth).mockReturnValue({
            loading: ref(false),
            isAuthenticated: ref(false),
            pendingRegistration: ref(null),
            user: ref(null),
            registerNewArcher: mockRegisterNewArcher,
            bootstrapAuth: vi.fn(),
            logout: vi.fn(),
            disableGoogleAutoSelect: vi.fn(),
            initialized: ref(true),
            initError: ref(null),
            prompting: ref(false),
            initOneTap: vi.fn(),
            beginGoogleLogin: vi.fn(),
            clientId: 'test-client-id',
        });

        const wrapper = mount(AuthGate);
        expect(wrapper.text()).toContain('Sign in to start');
        expect(wrapper.find('form').exists()).toBe(false);
    });

    it('submits registration form with correct data', async () => {
        vi.mocked(useAuth).mockReturnValue({
            loading: ref(false),
            isAuthenticated: ref(false),
            pendingRegistration: ref<ArcherRegistration>({
                need_first_name: true,
                need_last_name: true,
                google_email: 'test@example.com',
                google_subject: '123',
                credential: 'token',
                given_name: null,
                family_name: null,
            }),
            user: ref(null),
            registerNewArcher: mockRegisterNewArcher,
            bootstrapAuth: vi.fn(),
            logout: vi.fn(),
            disableGoogleAutoSelect: vi.fn(),
            initialized: ref(true),
            initError: ref(null),
            prompting: ref(false),
            initOneTap: vi.fn(),
            beginGoogleLogin: vi.fn(),
            clientId: 'test-client-id',
        });

        const wrapper = mount(AuthGate);

        // Fill form
        await wrapper.findAll('input[type="text"]')[0].setValue('John'); // First Name
        await wrapper.findAll('input[type="text"]')[1].setValue('Doe'); // Last Name
        await wrapper.find('input[type="date"]').setValue('1990-01-01');
        await wrapper.findAll('select')[0].setValue('male'); // Gender
        await wrapper.findAll('select')[1].setValue('recurve'); // Bowstyle
        await wrapper.find('input[type="number"]').setValue(40); // Draw weight

        await wrapper.find('form').trigger('submit');

        expect(mockRegisterNewArcher).toHaveBeenCalledWith({
            first_name: 'John',
            last_name: 'Doe',
            date_of_birth: '1990-01-01',
            gender: 'male',
            bowstyle: 'recurve',
            draw_weight: 40,
        });
    });

    it('shows validation error if bowstyle is missing', async () => {
        vi.mocked(useAuth).mockReturnValue({
            loading: ref(false),
            isAuthenticated: ref(false),
            pendingRegistration: ref<ArcherRegistration>({
                need_first_name: false,
                need_last_name: true,
                google_email: 'test@example.com',
                google_subject: '123',
                credential: 'token',
                given_name: null,
                family_name: null,
            }),
            user: ref(null),
            registerNewArcher: mockRegisterNewArcher,
            bootstrapAuth: vi.fn(),
            logout: vi.fn(),
            disableGoogleAutoSelect: vi.fn(),
            initialized: ref(true),
            initError: ref(null),
            prompting: ref(false),
            initOneTap: vi.fn(),
            beginGoogleLogin: vi.fn(),
            clientId: 'test-client-id',
        });

        const wrapper = mount(AuthGate);

        // Don't select bowstyle
        await wrapper.find('input[type="number"]').setValue(40);

        await wrapper.find('form').trigger('submit');

        expect(mockRegisterNewArcher).not.toHaveBeenCalled();
        expect(wrapper.text()).toContain('Please select a bowstyle');
    });
});
