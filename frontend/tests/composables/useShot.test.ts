import { describe, it, expect, beforeEach, vi } from 'vitest';
import type { components } from '../src/types/types.generated';
import {
    listShots,
    getShotsBySessionId,
    getShotsByArrowId,
    getShotById,
    deleteShot,
} from '../src/composables/useShot';

type Shot = components['schemas']['ShotsRead'];

// Helpers
const okEnvelope = (data: unknown) => ({ ok: true, status: 200, json: async () => ({ data }) });
const notOkEnvelope = (status = 500, errors?: string[]) => ({
    ok: false,
    status,
    json: async () => ({ errors: errors ?? ['boom'] }),
});

let fetchMock: ReturnType<typeof vi.fn>;

beforeEach(() => {
    vi.resetAllMocks();
    fetchMock = vi.fn();
    vi.stubGlobal('fetch', fetchMock);
});

describe('listShots', () => {
    it('builds query string correctly and returns array', async () => {
        const data: Shot[] = [{ id: 'sh1' } as Shot, { id: 'sh2' } as Shot];
        fetchMock.mockResolvedValue(okEnvelope(data));
        const res = await listShots({
            session_id: 's1',
            arrow_id: 'a1',
            skip: undefined,
            n: null,
            hit: true,
        });
        expect(res).toEqual(data);
        const calledUrl: string = fetchMock.mock.calls[0][0] as string;
        expect(calledUrl).toBe('/api/v0/shot?session_id=s1&arrow_id=a1&hit=true');
    });

    it('returns [] when ok but data missing', async () => {
        fetchMock.mockResolvedValue({ ok: true, status: 200, json: async () => ({}) });
        const res = await listShots();
        expect(res).toEqual([]);
    });

    it('throws on non-ok with aggregated errors', async () => {
        fetchMock.mockResolvedValue(notOkEnvelope(400, ['x', 'y']));
        await expect(listShots()).rejects.toThrow('x, y');
    });

    it('throws default message when body is not JSON', async () => {
        fetchMock.mockResolvedValue({
            ok: false,
            status: 500,
            json: async () => {
                throw new Error('bad');
            },
        });
        await expect(listShots()).rejects.toThrow('Failed to list shots');
    });

    it('stringifies invalid param values and applies URLSearchParams encoding', async () => {
        fetchMock.mockResolvedValue(okEnvelope([]));
        const invalid = { weird: { a: 1 }, arr: [1, 2] } as unknown as Record<
            string,
            string | number | boolean | undefined | null
        >;
        await listShots(invalid);
        const calledUrl: string = fetchMock.mock.calls[0][0] as string;
        // Object -> "[object Object]" spaces encoded as + ; comma encoded as %2C
        expect(calledUrl).toBe('/api/v0/shot?weird=%5Bobject+Object%5D&arr=1%2C2');
    });
});

describe('getShotsBySessionId', () => {
    it('encodes session id and returns array on ok', async () => {
        const data: Shot[] = [{ id: 's1-1' } as Shot];
        fetchMock.mockResolvedValue(okEnvelope(data));
        const res = await getShotsBySessionId('with space');
        expect(res).toEqual(data);
        const calledUrl: string = fetchMock.mock.calls[0][0] as string;
        expect(calledUrl).toBe('/api/v0/shot/session-id/with%20space');
    });

    it('throws on non-ok with errors', async () => {
        fetchMock.mockResolvedValue(notOkEnvelope(500, ['kaboom']));
        await expect(getShotsBySessionId('s1')).rejects.toThrow('kaboom');
    });

    it('throws default message on non-JSON body', async () => {
        fetchMock.mockResolvedValue({
            ok: false,
            status: 500,
            json: async () => {
                throw new Error('bad');
            },
        });
        await expect(getShotsBySessionId('s1')).rejects.toThrow(
            'Failed to fetch shots by session id',
        );
    });
});

describe('getShotsByArrowId', () => {
    it('delegates to listShots with arrow_id query', async () => {
        fetchMock.mockResolvedValue(okEnvelope([]));
        await getShotsByArrowId('a123');
        const calledUrl: string = fetchMock.mock.calls[0][0] as string;
        expect(calledUrl).toBe('/api/v0/shot?arrow_id=a123');
    });

    it('propagates errors from listShots', async () => {
        fetchMock.mockResolvedValue(notOkEnvelope(400, ['bad']));
        await expect(getShotsByArrowId('a1')).rejects.toThrow('bad');
    });
});

describe('getShotById', () => {
    it('returns null on 404', async () => {
        fetchMock.mockResolvedValue({ status: 404, ok: false, json: async () => ({}) });
        const res = await getShotById('missing');
        expect(res).toBeNull();
    });

    it('throws on non-ok (not 404)', async () => {
        fetchMock.mockResolvedValue(notOkEnvelope(500, ['oops']));
        await expect(getShotById('bad')).rejects.toThrow('oops');
    });

    it('returns Shot on ok', async () => {
        const s = { id: 'sh9' } as Shot;
        fetchMock.mockResolvedValue(okEnvelope(s));
        const res = await getShotById('sh9');
        expect(res).toEqual(s);
    });

    it('returns null when ok but data missing', async () => {
        fetchMock.mockResolvedValue({ ok: true, status: 200, json: async () => ({}) });
        const res = await getShotById('sh10');
        expect(res).toBeNull();
    });

    it('encodes id in URL', async () => {
        fetchMock.mockResolvedValue(okEnvelope({ id: 'with space' }));
        await getShotById('with space');
        const calledUrl: string = fetchMock.mock.calls[0][0] as string;
        expect(calledUrl).toBe('/api/v0/shot/with%20space');
    });
});

describe('deleteShot', () => {
    it('resolves on 204 No Content', async () => {
        fetchMock.mockResolvedValue({ status: 204, ok: true, json: async () => ({}) });
        await expect(deleteShot('sh1')).resolves.toBeUndefined();
    });

    it('resolves on ok JSON envelope (non-204)', async () => {
        fetchMock.mockResolvedValue({ status: 200, ok: true, json: async () => ({ data: null }) });
        await expect(deleteShot('sh2')).resolves.toBeUndefined();
    });

    it('throws on non-ok and surfaces errors', async () => {
        fetchMock.mockResolvedValue(notOkEnvelope(409, ['conflict']));
        await expect(deleteShot('sh3')).rejects.toThrow('conflict');
    });

    it('throws generic message when body is not JSON', async () => {
        fetchMock.mockResolvedValue({
            ok: false,
            status: 500,
            json: async () => {
                throw new Error('bad');
            },
        });
        await expect(deleteShot('sh4')).rejects.toThrow('Failed to delete shot');
    });

    it('encodes id in URL', async () => {
        fetchMock.mockResolvedValue({ status: 204, ok: true, json: async () => ({}) });
        await deleteShot('with space');
        const calledUrl: string = fetchMock.mock.calls[0][0] as string;
        expect(calledUrl).toBe('/api/v0/shot/with%20space');
    });
});
