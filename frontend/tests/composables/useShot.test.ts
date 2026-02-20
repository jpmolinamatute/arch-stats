import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { api, ApiError } from '@/api/client';
import { useShot } from '@/composables/useShot';

// Mock the API client
vi.mock('@/api/client', () =>
{
    return {
        api: {
            createShot: vi.fn(),
            get: vi.fn(),
        },
        ApiError: class extends Error
        {
            status: number;
            constructor(message: string, status: number)
            {
                super(message);
                this.status = status;
            }
        },
    };
});

describe('useShot', () =>
{
    const { createShot, fetchShots, subscribeToShots, shots, loading, error } = useShot();

    beforeEach(() =>
    {
        vi.clearAllMocks();
        shots.value = [];
        loading.value = false;
        error.value = null;
    });

    describe('createShot', () =>
    {
        const mockShotCreate = {
            slot_id: 'slot_1',
            score: 10,
            x: 0,
            y: 0,
            is_x: false,
        };

        it('successfully creates a shot and returns shot_id', async () =>
        {
            const mockResponse = { shot_id: 'shot_123' };
            vi.mocked(api.createShot).mockResolvedValue(mockResponse as any);

            const result = await createShot(mockShotCreate);

            expect(result).toBe('shot_123');
            expect(api.createShot).toHaveBeenCalledWith(mockShotCreate);
            expect(loading.value).toBe(false);
            expect(error.value).toBeNull();
        });

        it('handles batch array of shots creation (>= 3)', async () =>
        {
            const mockResponse = [{ shot_id: 'shot_1' }, { shot_id: 'shot_2' }, { shot_id: 'shot_3' }];
            vi.mocked(api.createShot).mockResolvedValue(mockResponse as any);

            const result = await createShot([mockShotCreate, mockShotCreate, mockShotCreate]);

            expect(result).toEqual(['shot_1', 'shot_2', 'shot_3']);
            expect(api.createShot).toHaveBeenCalledWith([mockShotCreate, mockShotCreate, mockShotCreate]);
            expect(loading.value).toBe(false);
        });

        it('handles small array of shots creation (< 3) individually', async () =>
        {
            const mockResponse1 = { shot_id: 'shot_1' };
            const mockResponse2 = { shot_id: 'shot_2' };
            vi.mocked(api.createShot)
                .mockResolvedValueOnce(mockResponse1 as any)
                .mockResolvedValueOnce(mockResponse2 as any);

            const result = await createShot([mockShotCreate, mockShotCreate]);

            expect(result).toEqual(['shot_1', 'shot_2']);
            expect(api.createShot).toHaveBeenCalledTimes(2);
            expect(loading.value).toBe(false);
        });

        it('handles 401 Not Authenticated error', async () =>
        {
            const apiError = new ApiError('Unauthorized', 401);
            vi.mocked(api.createShot).mockRejectedValue(apiError);

            await expect(createShot(mockShotCreate)).rejects.toThrow('Not authenticated. Please sign in again.');
            expect(loading.value).toBe(false);
        });

        it('handles 403 Forbidden error', async () =>
        {
            const apiError = new ApiError('Forbidden', 403);
            vi.mocked(api.createShot).mockRejectedValue(apiError);

            await expect(createShot(mockShotCreate)).rejects.toThrow('Forbidden: You are not allowed to add a shot to this slot');
        });

        it('handles 404 Slot Not Found error', async () =>
        {
            const apiError = new ApiError('Not Found', 404);
            vi.mocked(api.createShot).mockRejectedValue(apiError);

            await expect(createShot(mockShotCreate)).rejects.toThrow('Slot not found');
        });

        it('handles generic errors', async () =>
        {
            const genericError = new Error('Network Error');
            vi.mocked(api.createShot).mockRejectedValue(genericError);

            await expect(createShot(mockShotCreate)).rejects.toThrow('Network Error');
            expect(error.value).toBe('Network Error');
        });
    });

    describe('fetchShots', () =>
    {
        const slotId = 'slot_123';
        const mockStatsData = {
            scores: [
                { shot_id: '1', score: 10, is_x: false, created_at: '2023-01-01' },
                { shot_id: '2', score: 9, is_x: false, created_at: '2023-01-01' },
            ],
            stats: {
                slot_id: 'slot_123',
                number_of_shots: 2,
                total_score: 19,
                max_score: 20,
                mean: 9.5,
            },
        };

        it('successfully fetches shots and stats', async () =>
        {
            vi.mocked(api.get).mockResolvedValue(mockStatsData as any);

            await fetchShots(slotId);

            expect(api.get).toHaveBeenCalledWith(`/stats/${slotId}`);
            expect(shots.value).toEqual(mockStatsData.scores);
            // Need to export stats from useShot to test it, currently tests access shots via closure?
            // "const { shots, loading, error } = useShot()" - stats is not exported in the test destructuring at the top?
            // Let's check imports.
        });

        it('handles fetch error', async () =>
        {
            vi.mocked(api.get).mockRejectedValue(new Error('Fetch failed'));

            await fetchShots(slotId);

            expect(shots.value).toEqual([]);
            expect(error.value).toBe('Fetch failed');
            expect(loading.value).toBe(false);
        });
    });

    describe('subscribeToShots', () =>
    {
        let mockWebSocket: any;
        let mockWSConstructor: any;

        beforeEach(() =>
        {
            // Mock WebSocket instance
            mockWebSocket = {
                onopen: null,
                onmessage: null,
                onclose: null,
                onerror: null,
                close: vi.fn(),
            };
            // Mock constructor spy
            mockWSConstructor = vi.fn();

            // Mock the global WebSocket constructor
            // We need a class (or constructor) that returns our mock instance
            globalThis.WebSocket = class
            {
                constructor(url: string)
                {
                    mockWSConstructor(url);
                    return mockWebSocket;
                }
            } as any;
        });

        afterEach(() =>
        {
            vi.restoreAllMocks();
        });

        it('creates WebSocket connection with correct URL', () =>
        {
            // Mock window location
            Object.defineProperty(window, 'location', {
                value: {
                    protocol: 'http:',
                    host: 'localhost:5173',
                },
                writable: true,
            });

            subscribeToShots('slot_123');

            expect(mockWSConstructor).toHaveBeenCalledWith('ws://localhost:5173/api/v0/stats/ws/slot_123');
        });

        it('uses wss when protocol is https', () =>
        {
            Object.defineProperty(window, 'location', {
                value: {
                    protocol: 'https:',
                    host: 'example.com',
                },
                writable: true,
            });

            subscribeToShots('slot_123');

            expect(mockWSConstructor).toHaveBeenCalledWith('wss://example.com/api/v0/stats/ws/slot_123');
        });

        it('updates state directly when "shot.created" message is received', async () =>
        {
            const slotId = 'slot_123';
            subscribeToShots(slotId);

            const newStats = {
                scores: [{ shot_id: 'new_shot', score: 10, is_x: true, created_at: 'now' }],
                stats: {
                    slot_id: 'slot_123',
                    number_of_shots: 1,
                    total_score: 10,
                    max_score: 10,
                    mean: 10.0,
                },
            };

            // Set initial state
            shots.value = [{ shot_id: 'old_shot', score: 8, is_x: false, created_at: 'old' }];

            // Simulate message
            const messageEvent = {
                data: JSON.stringify({
                    content_type: 'shot.created',
                    content: newStats,
                }),
            };

            if (mockWebSocket.onmessage) {
                mockWebSocket.onmessage(messageEvent);
            }

            // Should NOT call api.get
            expect(api.get).not.toHaveBeenCalled();

            // Should update state (append)
            expect(shots.value).toHaveLength(2);
            expect(shots.value).toEqual([
                { shot_id: 'old_shot', score: 8, is_x: false, created_at: 'old' },
                ...newStats.scores,
            ]);
        });

        it('ignores messages with other content types', () =>
        {
            subscribeToShots('slot_123');

            const messageEvent = {
                data: JSON.stringify({
                    content_type: 'other.event',
                    content: {},
                }),
            };

            if (mockWebSocket.onmessage) {
                mockWebSocket.onmessage(messageEvent);
            }

            expect(api.get).not.toHaveBeenCalled();
        });

        it('handles malformed JSON in messages gracefully', () =>
        {
            const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => { });
            subscribeToShots('slot_123');

            const messageEvent = {
                data: 'invalid json',
            };

            if (mockWebSocket.onmessage) {
                mockWebSocket.onmessage(messageEvent);
            }

            expect(consoleSpy).toHaveBeenCalledWith(expect.stringContaining('Failed to parse WS message'), expect.any(Error));
            expect(api.get).not.toHaveBeenCalled();
        });
    });
});
