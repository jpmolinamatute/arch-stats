import { describe, it, expect, beforeEach, vi } from 'vitest';
import type { components } from '../../src/types/types.generated';
import {
    createTarget,
    listTargets,
    getTargetsBySessionId,
    getTargetById,
    deleteTarget,
    calibrateTarget,
} from '../../src/composables/useTarget';

// Use any to avoid TS lib dom coupling in Node test env
let fetchMock: ReturnType<typeof vi.fn>;

const okEnvelope = (data: unknown) => ({ ok: true, json: async () => ({ data }) });
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

describe('createTarget', () => {
    it('returns Target on success', async () => {
        const target = { id: 't1', human_identifier: 'face A' };
        fetchMock.mockResolvedValue(okEnvelope(target));
        type TargetsCreate = components['schemas']['TargetsCreate'];
        const payload: TargetsCreate = {
            max_x: 100,
            max_y: 100,
            session_id: 's1',
            faces: [{ x: 50, y: 50, human_identifier: 'face A', radii: [10, 20, 30] }],
        } as unknown as TargetsCreate;
        const res = await createTarget(payload);
        expect(res).toEqual(target);
        expect(fetchMock).toHaveBeenCalledWith('/api/v0/target', expect.any(Object));
    });

    it('returns null on non-ok with errors', async () => {
        fetchMock.mockResolvedValue(notOkEnvelope(400, ['bad payload']));
        type TargetsCreate = components['schemas']['TargetsCreate'];
        const badPayload = {
            // missing required fields intentionally; cast to bypass TS as tests check runtime path
        } as unknown as TargetsCreate;
        const res = await createTarget(badPayload);
        expect(res).toBeNull();
    });

    it('returns null when fetch throws', async () => {
        fetchMock.mockRejectedValue(new Error('network'));
        type TargetsCreate = components['schemas']['TargetsCreate'];
        const res = await createTarget({} as unknown as TargetsCreate);
        expect(res).toBeNull();
    });
});

describe('listTargets', () => {
    it('builds query string correctly and returns array', async () => {
        const data = [{ id: 't1' }, { id: 't2' }];
        fetchMock.mockResolvedValue(okEnvelope(data));
        const res = await listTargets({ session_id: 's1', active: true, skip: undefined, n: null });
        expect(res).toEqual(data);
        const calledUrl: string = fetchMock.mock.calls[0][0] as string;
        // order is deterministic by insertion into URLSearchParams
        expect(calledUrl).toBe('/api/v0/target?session_id=s1&active=true');
    });

    it('returns [] when ok but data is null/absent', async () => {
        fetchMock.mockResolvedValue({ ok: true, json: async () => ({}) });
        const res = await listTargets();
        expect(res).toEqual([]);
    });

    it('throws on non-ok with error aggregation', async () => {
        fetchMock.mockResolvedValue(notOkEnvelope(500, ['x', 'y']));
        await expect(listTargets()).rejects.toThrow('x, y');
    });
});

describe('getTargetsBySessionId', () => {
    it('delegates to listTargets with session_id', async () => {
        const data = [{ id: 't1' }];
        fetchMock.mockResolvedValue(okEnvelope(data));
        const res = await getTargetsBySessionId('abc');
        expect(res).toEqual(data);
        const calledUrl: string = fetchMock.mock.calls[0][0] as string;
        expect(calledUrl).toBe('/api/v0/target?session_id=abc');
    });
});

describe('getTargetById', () => {
    it('returns null on 404', async () => {
        fetchMock.mockResolvedValue({ status: 404, ok: false, json: async () => ({}) });
        const res = await getTargetById('missing');
        expect(res).toBeNull();
    });

    it('throws on non-ok (not 404)', async () => {
        fetchMock.mockResolvedValue(notOkEnvelope(500, ['kaboom']));
        await expect(getTargetById('bad')).rejects.toThrow('kaboom');
    });

    it('returns Target on ok', async () => {
        const t = { id: 't9' };
        fetchMock.mockResolvedValue(okEnvelope(t));
        const res = await getTargetById('t9');
        expect(res).toEqual(t);
    });

    it('returns null when ok but data missing', async () => {
        fetchMock.mockResolvedValue({ ok: true, json: async () => ({}) });
        const res = await getTargetById('t10');
        expect(res).toBeNull();
    });
});

describe('deleteTarget', () => {
    it('resolves on 204 No Content', async () => {
        fetchMock.mockResolvedValue({ status: 204, ok: true, json: async () => ({}) });
        await expect(deleteTarget('t1')).resolves.toBeUndefined();
    });

    it('resolves on ok JSON envelope (non-204)', async () => {
        fetchMock.mockResolvedValue({ status: 200, ok: true, json: async () => ({ data: null }) });
        await expect(deleteTarget('t2')).resolves.toBeUndefined();
    });

    it('throws on non-ok and surfaces errors', async () => {
        fetchMock.mockResolvedValue(notOkEnvelope(409, ['conflict']));
        await expect(deleteTarget('t3')).rejects.toThrow('conflict');
    });

    it('throws generic message when body is not JSON', async () => {
        fetchMock.mockResolvedValue({
            ok: false,
            status: 500,
            json: async () => {
                throw new Error('bad');
            },
        });
        await expect(deleteTarget('t4')).rejects.toThrow('Failed to delete target');
    });
});

describe('calibrateTarget', () => {
    it('returns Target on ok', async () => {
        const t = { id: 'cal' };
        fetchMock.mockResolvedValue(okEnvelope(t));
        const res = await calibrateTarget();
        expect(res).toEqual(t);
    });

    it('returns null when ok but no data', async () => {
        fetchMock.mockResolvedValue({ ok: true, json: async () => ({}) });
        const res = await calibrateTarget();
        expect(res).toBeNull();
    });

    it('throws when non-ok', async () => {
        fetchMock.mockResolvedValue(notOkEnvelope(503, ['unavailable']));
        await expect(calibrateTarget()).rejects.toThrow('unavailable');
    });
});
