import { describe, it, expect, beforeEach, vi } from 'vitest';
import type { components } from '../src/types/types.generated';
import {
    listArrows,
    getArrowById,
    createArrow,
    patchArrow,
    deleteArrow,
    getNewArrowUuid,
} from '../src/composables/useArrow';

// Fetch mock helpers
let fetchMock: ReturnType<typeof vi.fn>;
const okEnvelope = (data: unknown) => ({ ok: true, status: 200, json: async () => ({ data }) });
const notOkEnvelope = (status = 500, errors?: string[]) => ({
    ok: false,
    status,
    json: async () => ({ errors: errors ?? ['boom'] }),
});

beforeEach(() => {
    vi.resetAllMocks();
    fetchMock = vi.fn();
    vi.stubGlobal('fetch', fetchMock);
});

describe('listArrows', () => {
    it('builds query string, filters out nullish, and returns array', async () => {
        const data = [{ id: 'a1' }, { id: 'a2' }];
        fetchMock.mockResolvedValue(okEnvelope(data));
        const res = await listArrows({
            human_identifier: 'A',
            is_programmed: true,
            skip: undefined,
            n: null,
        });
        expect(res).toEqual(data);
        const calledUrl: string = fetchMock.mock.calls[0][0] as string;
        expect(calledUrl).toBe('/api/v0/arrow?human_identifier=A&is_programmed=true');
    });

    it('returns [] when ok but data is missing', async () => {
        fetchMock.mockResolvedValue({ ok: true, status: 200, json: async () => ({}) });
        const res = await listArrows();
        expect(res).toEqual([]);
    });

    it('throws aggregated errors on non-ok', async () => {
        fetchMock.mockResolvedValue(notOkEnvelope(400, ['x', 'y']));
        await expect(listArrows()).rejects.toThrow('x, y');
    });

    it('throws default message when body is not JSON', async () => {
        fetchMock.mockResolvedValue({
            ok: false,
            status: 500,
            json: async () => {
                throw new Error('bad');
            },
        });
        await expect(listArrows()).rejects.toThrow('Failed to list arrows');
    });

    it('stringifies unusual param values without crashing (invalid types)', async () => {
        const data: unknown[] = [];
        fetchMock.mockResolvedValue(okEnvelope(data));
        // Cast through unknown to satisfy TS while testing runtime stringification
        const invalidParams = { weird: { a: 1 }, arr: [1, 2] } as unknown as Record<
            string,
            string | number | boolean | undefined | null
        >;
        const res = await listArrows(invalidParams);
        expect(res).toEqual([]);
        const calledUrl: string = fetchMock.mock.calls[0][0] as string;
        expect(calledUrl).toBe('/api/v0/arrow?weird=%5Bobject+Object%5D&arr=1%2C2');
    });
});

describe('getArrowById', () => {
    it('returns null on 404', async () => {
        fetchMock.mockResolvedValue({ status: 404, ok: false, json: async () => ({}) });
        const res = await getArrowById('missing');
        expect(res).toBeNull();
    });

    it('throws on non-ok (not 404)', async () => {
        fetchMock.mockResolvedValue(notOkEnvelope(500, ['kaboom']));
        await expect(getArrowById('bad')).rejects.toThrow('kaboom');
    });

    it('returns Arrow on ok', async () => {
        const a = { id: 'a9' };
        fetchMock.mockResolvedValue(okEnvelope(a));
        const res = await getArrowById('a9');
        expect(res).toEqual(a);
    });

    it('returns null when ok but data missing', async () => {
        fetchMock.mockResolvedValue({ ok: true, status: 200, json: async () => ({}) });
        const res = await getArrowById('a10');
        expect(res).toBeNull();
    });

    it('encodes id in URL', async () => {
        const a = { id: 'with space' };
        fetchMock.mockResolvedValue(okEnvelope(a));
        await getArrowById('with space');
        const calledUrl: string = fetchMock.mock.calls[0][0] as string;
        expect(calledUrl).toBe('/api/v0/arrow/with%20space');
    });
});

describe('createArrow', () => {
    it('returns new id string on success', async () => {
        fetchMock.mockResolvedValue(okEnvelope('new-id'));
        type ArrowsCreate = components['schemas']['ArrowsCreate'];
        const payload = { human_identifier: 'H1' } as unknown as ArrowsCreate;
        const res = await createArrow(payload);
        expect(res).toBe('new-id');
        expect(fetchMock).toHaveBeenCalledWith(
            '/api/v0/arrow',
            expect.objectContaining({ method: 'POST' }),
        );
    });

    it('returns null when ok but no data', async () => {
        fetchMock.mockResolvedValue({ ok: true, status: 200, json: async () => ({}) });
        const res = await createArrow({} as unknown as components['schemas']['ArrowsCreate']);
        expect(res).toBeNull();
    });

    it('throws on non-ok and surfaces errors', async () => {
        fetchMock.mockResolvedValue(notOkEnvelope(422, ['invalid']));
        await expect(
            createArrow({} as unknown as components['schemas']['ArrowsCreate']),
        ).rejects.toThrow('invalid');
    });

    it('throws generic message when body is not JSON', async () => {
        fetchMock.mockResolvedValue({
            ok: false,
            status: 500,
            json: async () => {
                throw new Error('bad');
            },
        });
        await expect(
            createArrow({} as unknown as components['schemas']['ArrowsCreate']),
        ).rejects.toThrow('Failed to create arrow');
    });
});

describe('patchArrow', () => {
    it('resolves on 202 Accepted', async () => {
        fetchMock.mockResolvedValue({ status: 202, ok: true, json: async () => ({}) });
        await expect(
            patchArrow('a1', {} as unknown as components['schemas']['ArrowsUpdate']),
        ).resolves.toBeUndefined();
    });

    it('resolves silently on ok non-202 (e.g., 200)', async () => {
        fetchMock.mockResolvedValue({ status: 200, ok: true, json: async () => ({ data: null }) });
        await expect(
            patchArrow('a2', {} as unknown as components['schemas']['ArrowsUpdate']),
        ).resolves.toBeUndefined();
    });

    it('throws on non-ok with error aggregation', async () => {
        fetchMock.mockResolvedValue(notOkEnvelope(400, ['bad update']));
        await expect(
            patchArrow('a3', {} as unknown as components['schemas']['ArrowsUpdate']),
        ).rejects.toThrow('bad update');
    });

    it('throws generic message when body is not JSON', async () => {
        fetchMock.mockResolvedValue({
            ok: false,
            status: 500,
            json: async () => {
                throw new Error('bad');
            },
        });
        await expect(
            patchArrow('a4', {} as unknown as components['schemas']['ArrowsUpdate']),
        ).rejects.toThrow('Failed to update arrow');
    });

    it('encodes id in URL', async () => {
        fetchMock.mockResolvedValue({ status: 202, ok: true, json: async () => ({}) });
        await patchArrow('with space', {} as unknown as components['schemas']['ArrowsUpdate']);
        const calledUrl: string = fetchMock.mock.calls[0][0] as string;
        expect(calledUrl).toBe('/api/v0/arrow/with%20space');
    });
});

describe('deleteArrow', () => {
    it('resolves on 204 No Content', async () => {
        fetchMock.mockResolvedValue({ status: 204, ok: true, json: async () => ({}) });
        await expect(deleteArrow('a1')).resolves.toBeUndefined();
    });

    it('resolves on ok JSON envelope (non-204)', async () => {
        fetchMock.mockResolvedValue({ status: 200, ok: true, json: async () => ({ data: null }) });
        await expect(deleteArrow('a2')).resolves.toBeUndefined();
    });

    it('throws on non-ok and surfaces errors', async () => {
        fetchMock.mockResolvedValue(notOkEnvelope(409, ['conflict']));
        await expect(deleteArrow('a3')).rejects.toThrow('conflict');
    });

    it('throws generic message when body is not JSON', async () => {
        fetchMock.mockResolvedValue({
            ok: false,
            status: 500,
            json: async () => {
                throw new Error('bad');
            },
        });
        await expect(deleteArrow('a4')).rejects.toThrow('Failed to delete arrow');
    });

    it('encodes id in URL', async () => {
        fetchMock.mockResolvedValue({ status: 204, ok: true, json: async () => ({}) });
        await deleteArrow('with space');
        const calledUrl: string = fetchMock.mock.calls[0][0] as string;
        expect(calledUrl).toBe('/api/v0/arrow/with%20space');
    });
});

describe('getNewArrowUuid', () => {
    it('returns string on ok', async () => {
        fetchMock.mockResolvedValue(okEnvelope('uuid-123'));
        const res = await getNewArrowUuid();
        expect(res).toBe('uuid-123');
        expect(fetchMock).toHaveBeenCalledWith('/api/v0/arrow/new_arrow_uuid');
    });

    it('throws on non-ok and surfaces errors', async () => {
        fetchMock.mockResolvedValue(notOkEnvelope(503, ['unavailable']));
        await expect(getNewArrowUuid()).rejects.toThrow('unavailable');
    });

    it('throws generic message when body is not JSON', async () => {
        fetchMock.mockResolvedValue({
            ok: false,
            status: 500,
            json: async () => {
                throw new Error('bad');
            },
        });
        await expect(getNewArrowUuid()).rejects.toThrow('Failed to get new arrow UUID');
    });
});
