# Task 008: Frontend Composables Refactoring and Slot/WebSocket Removal

## Git Branch

`feature/008-frontend-composables-simplify`

## Objective

Refactor the frontend state management and composables in `frontend/src/composables/` to match
the single-archer architecture. Delete `useSlot.ts` completely. Refactor `useSession.ts` to
operate directly on session endpoints without slot indirection. Refactor `useShot.ts` to bind
to `session_id`, remove all WebSocket code, and implement standard HTTP GET polling/fetch for
live stats. Remove obsolete registration method `registerNewArcher` from `useAuth.ts`. Add
Vitest unit tests verifying composable behavior and error handling.

## Dependencies

- Task 007 (OpenAPI specs and frontend type generation).

## Acceptance Criteria

- [ ] `frontend/src/composables/useSlot.ts` deleted
- [ ] `frontend/src/composables/useAuth.ts` cleaned up:
    - [ ] `registerNewArcher` method removed
    - [ ] Obsolete imports and references to `AuthRegistrationRequest` removed
- [ ] `frontend/src/composables/useSession.ts` refactored:
    - [ ] Import and references to `useSlot` completely removed
    - [ ] Typed with generated interfaces: `SessionRead`, `SessionCreate`, `SessionClose`
    - [ ] `checkForOpenSession(archerId: string, force?: boolean)`:
        - Calls `api.get('/session/open')`
        - Updates `currentSession` ref and local cache (`localStorage`)
    - [ ] `createSession(payload: SessionCreate)`:
        - Calls `api.post('/session', payload)`
        - Sets `currentSession` and returns created `sessionId`
    - [ ] `closeSession(sessionId: string, reflections: SessionClose)`:
        - Calls `api.patch('/session/${sessionId}/close', reflections)` requiring mandatory `status`
        - Resets `currentSession` to null and clears cache
    - [ ] `deleteSession(sessionId: string)`:
        - Calls `api.delete('/session/${sessionId}')`
    - [ ] `listSessions()`:
        - Calls `api.get<SessionRead[]>('/session')`
    - [ ] Reactive state: `currentSession`, `hasOpenSession`, `loading`, `error`
- [ ] `frontend/src/composables/useShot.ts` refactored:
    - [ ] WebSocket infrastructure (`subscribeToShots`, `WebSocketMessage`) removed
    - [ ] `createShot(payload: ShotCreate | ShotCreate[])`:
        - Sends payload containing `session_id` to `api.post('/shot', payload)`
        - Handles single shot and array batching (3 to 10 shots)
    - [ ] `fetchShots(sessionId: string)`:
        - Calls `api.get<ShotRead[]>('/shot/by-session/${sessionId}')`
    - [ ] `fetchShotCount(sessionId: string)`:
        - Calls `api.get<number>('/shot/count-by-session/${sessionId}')`
    - [ ] `fetchLiveStats(sessionId: string)`:
        - Calls `api.get<SessionLiveStatsRead>('/session/${sessionId}/live-stats')`
        - Updates `shots.value` and `stats.value`
    - [ ] `deleteShot(shotId: string, sessionId: string)`:
        - Calls `api.delete('/shot/${shotId}')`
- [ ] `frontend/src/api/client.ts` updated:
    - [ ] `createShot` method signature and route updated for `session_id` payloads
- [ ] Vitest unit tests added:
    - [ ] `frontend/tests/composables/useSession.spec.ts` verifies session lifecycle, caching,
          and close reflection payloads
    - [ ] `frontend/tests/composables/useShot.spec.ts` verifies single/batch shot creation,
          live stats fetching, and error mapping
    - [ ] `frontend/tests/composables/useAuth.spec.ts` verifies cleaned auth composable state
- [ ] `cd frontend && npm run test` passes.

## Files to Create/Modify/Remove

| Action | Path |
| ------ | ---- |
| Modify | `frontend/src/composables/useSession.ts` |
| Modify | `frontend/src/composables/useShot.ts` |
| Modify | `frontend/src/composables/useAuth.ts` |
| Modify | `frontend/src/api/client.ts` |
| Create | `frontend/tests/composables/useSession.spec.ts` |
| Create | `frontend/tests/composables/useShot.spec.ts` |
| Create | `frontend/tests/composables/useAuth.spec.ts` |
| Delete | `frontend/src/composables/useSlot.ts` |

## Reference

- [PRD.md](../PRD.md)
- [useSession.ts](../../../frontend/src/composables/useSession.ts)
- [useShot.ts](../../../frontend/src/composables/useShot.ts)
- [useAuth.ts](../../../frontend/src/composables/useAuth.ts)

## Steps

- [ ] **Step 1: Write failing Vitest tests for composables**

  Create `tests/composables/useSession.spec.ts`, `tests/composables/useShot.spec.ts`, and
  `tests/composables/useAuth.spec.ts`:
    - Mock `api` calls with `vi.spyOn`
    - Test `checkForOpenSession` loads and sets `currentSession`
    - Test `createSession` posts canonical payload and updates state
    - Test `closeSession` sends reflection fields and clears cache
    - Test `createShot` sends `session_id` payload
    - Test `fetchLiveStats` updates `shots` and `stats`
    - Test `useAuth` handles authentication states without `registerNewArcher`

- [ ] **Step 2: Run Vitest to verify tests fail**

  ```bash
  cd frontend && npm run test
  ```

- [ ] **Step 3: Update `frontend/src/composables/useSession.ts`**

  Remove `useSlot`. Implement clean session methods targeting `/session` endpoints.

- [ ] **Step 4: Update `frontend/src/composables/useShot.ts`**

  Remove all WebSocket code. Implement HTTP live stats fetching and session-based shot methods.

- [ ] **Step 5: Clean up `frontend/src/composables/useAuth.ts`**

  Remove `registerNewArcher` method and unused `AuthRegistrationRequest` type imports.

- [ ] **Step 6: Update `frontend/src/api/client.ts`**

  Adjust helper methods for session-based shot endpoints.

- [ ] **Step 7: Delete `useSlot.ts`**

  ```bash
  rm frontend/src/composables/useSlot.ts
  ```

- [ ] **Step 8: Run Vitest to verify composable tests pass**

  ```bash
  cd frontend && npm run test
  ```

- [ ] **Step 9: Commit changes**

  ```bash
  git add frontend/src/composables/ frontend/src/api/ frontend/tests/
  git commit -m "feat(frontend): refactor composables; remove useSlot, WebSockets, and legacy auth"
  ```

## Verification

- `cd frontend && npm run test` passes with 0 failures.
- No imports of `useSlot` or WebSocket constructors remain in `frontend/src/composables/`.
- No references to `registerNewArcher` in `frontend/src/composables/`.
