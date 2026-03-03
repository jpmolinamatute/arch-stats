import type { ArcherRead } from '@/composables/useAuth'

export function createMockArcher(overrides?: Partial<ArcherRead>): ArcherRead {
    return {
        archer_id: 'archer_1',
        first_name: 'John',
        last_name: 'Doe',
        email: 'john@example.com',
        date_of_birth: '1990-01-01',
        gender: 'unspecified',
        bowstyle: 'recurve',
        draw_weight: 40,
        club_id: null,
        google_picture_url: null,
        google_subject: 'mock-google-sub-123',
        last_login_at: '2023-01-01T00:00:00Z',
        created_at: '2023-01-01T00:00:00Z',
        ...overrides,
    }
}
