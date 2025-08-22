import { describe, it, expect, beforeEach, vi } from 'vitest';
import type { components } from '../src/types/types.generated';
import * as sessionState from '../src/state/session';
import {
    createSession,
    closeSession,
    listSessions,
    getSessionById,
    deleteSession,
} from '../src/composables/useSession';

// Fetch envelope helpers
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
    // Reset session state
    sessionState.sessionOpened.id = undefined;
    sessionState.sessionOpened.is_opened = undefined;
    sessionState.sessionOpened.location = undefined;
    sessionState.sessionOpened.start_time = undefined;
    sessionState.sessionOpened.end_time = undefined;
});

describe('createSession', () => {
    it('returns early when a session is already open (no network call)', async () => {
        sessionState.sessionOpened.is_opened = true;
        await createSession({} as unknown as components['schemas']['SessionsCreate']);
        expect(fetchMock).not.toHaveBeenCalled();
    });

    it('POSTs and calls fetchOpenSession on success', async () => {
        fetchMock.mockResolvedValue(okEnvelope({ id: 's1' }));
        const fetchOpenSpy = vi.spyOn(sessionState, 'fetchOpenSession').mockResolvedValue();
        await createSession({} as unknown as components['schemas']['SessionsCreate']);
        expect(fetchMock).toHaveBeenCalledWith(
            '/api/v0/session',
            expect.objectContaining({ method: 'POST' }),
        );
        expect(fetchOpenSpy).toHaveBeenCalled();
    });

    it('handles non-ok response with errors without throwing', async () => {
        fetchMock.mockResolvedValue(notOkEnvelope(400, ['bad payload']));
        await expect(
            createSession({} as unknown as components['schemas']['SessionsCreate']),
        ).resolves.toBeUndefined();
    });

    it('handles fetch throw without throwing', async () => {
        fetchMock.mockRejectedValue(new Error('network'));
        await expect(
            createSession({} as unknown as components['schemas']['SessionsCreate']),
        ).resolves.toBeUndefined();
    });
});

describe('closeSession', () => {
    it('returns early when no session is open', async () => {
        sessionState.sessionOpened.is_opened = false;
        await closeSession();
        expect(fetchMock).not.toHaveBeenCalled();
    });

    it('PATCHes and clears session on success', async () => {
        sessionState.sessionOpened.is_opened = true;
        sessionState.sessionOpened.id = 's1';
        const clearSpy = vi.spyOn(sessionState, 'clearOpenSession').mockImplementation(() => {});
        fetchMock.mockResolvedValue({ ok: true, status: 200, json: async () => ({ data: null }) });
        await closeSession();
        const calledUrl: string = fetchMock.mock.calls[0][0] as string;
        expect(calledUrl).toBe('/api/v0/session/s1');
        const options = fetchMock.mock.calls[0][1] as unknown;
        expect((options as { method?: string }).method).toBe('PATCH');
        expect(clearSpy).toHaveBeenCalled();
    });

    it('uses raw id on non-ok (no encode in implementation)', async () => {
        sessionState.sessionOpened.is_opened = true;
        sessionState.sessionOpened.id = 'with space';
        fetchMock.mockResolvedValue(notOkEnvelope(500, ['kaboom']));
        await closeSession();
        const calledUrl: string = fetchMock.mock.calls[0][0] as string;
        expect(calledUrl).toBe('/api/v0/session/with space');
    });

    it('handles fetch throw without throwing', async () => {
        sessionState.sessionOpened.is_opened = true;
        sessionState.sessionOpened.id = 's1';
        fetchMock.mockRejectedValue(new Error('network'));
        await expect(closeSession()).resolves.toBeUndefined();
    });
});

describe('listSessions', () => {
    it('builds query string and returns array', async () => {
        const data = [{ id: 's1' }, { id: 's2' }];
        fetchMock.mockResolvedValue(okEnvelope(data));
        const res = await listSessions({ location: 'A', active: true, skip: undefined, n: null });
        expect(res).toEqual(data);
        const calledUrl: string = fetchMock.mock.calls[0][0] as string;
        expect(calledUrl).toBe('/api/v0/session?location=A&active=true');
    });

    it('returns [] when ok but data missing', async () => {
        fetchMock.mockResolvedValue({ ok: true, status: 200, json: async () => ({}) });
        const res = await listSessions();
        expect(res).toEqual([]);
    });

    it('throws on non-ok with aggregated errors', async () => {
        fetchMock.mockResolvedValue(notOkEnvelope(400, ['x', 'y']));
        await expect(listSessions()).rejects.toThrow('x, y');
    });

    it('stringifies unusual param values without crashing (invalid types)', async () => {
        fetchMock.mockResolvedValue(okEnvelope([]));
        const invalid = { weird: { a: 1 }, arr: [1, 2] } as unknown as Record<
            string,
            string | number | boolean | undefined | null
        >;
        await listSessions(invalid);
        const calledUrl: string = fetchMock.mock.calls[0][0] as string;
        expect(calledUrl).toBe('/api/v0/session?weird=%5Bobject+Object%5D&arr=1%2C2');
    });
});

describe('getSessionById', () => {
    it('throws on non-ok with error aggregation (404 included)', async () => {
        fetchMock.mockResolvedValue(notOkEnvelope(404, ['not found']));
        await expect(getSessionById('missing')).rejects.toThrow('not found');
    });

    it('returns Session on ok', async () => {
        const s = { id: 's9' };
        fetchMock.mockResolvedValue(okEnvelope(s));
        const res = await getSessionById('s9');
        expect(res).toEqual(s);
    });

    it('returns null when ok but data missing', async () => {
        fetchMock.mockResolvedValue({ ok: true, status: 200, json: async () => ({}) });
        const res = await getSessionById('s10');
        expect(res).toBeNull();
    });

    it('encodes id in URL', async () => {
        fetchMock.mockResolvedValue(okEnvelope({ id: 'with space' }));
        await getSessionById('with space');
        const calledUrl: string = fetchMock.mock.calls[0][0] as string;
        expect(calledUrl).toBe('/api/v0/session/with%20space');
    });
});

describe('deleteSession', () => {
    it('resolves on ok and clears open session when ids match', async () => {
        sessionState.sessionOpened.id = 's1';
        const clearSpy = vi.spyOn(sessionState, 'clearOpenSession').mockImplementation(() => {});
        fetchMock.mockResolvedValue({ ok: true, status: 204, json: async () => ({}) });
        await expect(deleteSession('s1')).resolves.toBeUndefined();
        expect(clearSpy).toHaveBeenCalled();
    });

    it('resolves on ok and does not clear when ids differ', async () => {
        sessionState.sessionOpened.id = 's1';
        const clearSpy = vi.spyOn(sessionState, 'clearOpenSession').mockImplementation(() => {});
        fetchMock.mockResolvedValue({ ok: true, status: 200, json: async () => ({ data: null }) });
        await expect(deleteSession('s2')).resolves.toBeUndefined();
        expect(clearSpy).not.toHaveBeenCalled();
    });

    it('throws on non-ok and surfaces errors', async () => {
        fetchMock.mockResolvedValue(notOkEnvelope(409, ['conflict']));
        await expect(deleteSession('s3')).rejects.toThrow('conflict');
    });

    it('throws generic message when body is not JSON', async () => {
        fetchMock.mockResolvedValue({
            ok: false,
            status: 500,
            json: async () => {
                throw new Error('bad');
            },
        });
        await expect(deleteSession('s4')).rejects.toThrow('Failed to delete session');
    });

    it('encodes id in URL', async () => {
        fetchMock.mockResolvedValue({ ok: true, status: 204, json: async () => ({}) });
        await deleteSession('with space');
        const calledUrl: string = fetchMock.mock.calls[0][0] as string;
        expect(calledUrl).toBe('/api/v0/session/with%20space');
    });
});
