# Task 010: Frontend Live Session, Arrow Tagging, and Scoring UI

## Git Branch

`feature/010-frontend-live-session-scoring-ui`

## Objective

Refactor the live session view (`frontend/src/components/LiveSession.vue`) and related widgets to
directly consume the active `SessionRead` state without slot or WebSocket dependencies. Support
both scored target faces (`Face.vue`) and volume sessions (`NoFaceSession.vue`). Implement arrow
tagging UI where archers select their physical `arrow_number` and the frontend translates it to
`arrow_id`. Derive end grouping dynamically from `shots_per_end`. Fetch live stats over HTTP, and
implement a session completion modal asking "How was the session?" with mandatory status selection
(`bad`, `neutral`, `good`) alongside optional reflection inputs (`was_goal_achieved`, `did_well`,
`need_work`).

## Dependencies

- Task 008 (Frontend composables refactoring and slot removal).
- Task 009 (Frontend session setup and management UI).

## Acceptance Criteria

- [ ] `frontend/src/components/LiveSession.vue` refactored:
    - [ ] Binds directly to `useSession().currentSession` (all references to `useSlot` removed)
    - [ ] Mode switching derived from `currentSession.face_type`:
        - If `face_type !== 'none'`: renders `Face.vue` interactive target face
        - If `face_type === 'none'`: renders `NoFaceSession.vue` for volume arrow tracking
    - [ ] Arrow tagging selector:
        - Checks if the archer has `in_use` registered arrows
        - If `in_use` registered arrows exist: displays arrow number selector chips (`arrow_number`
          1..N); mandatory selection before plotting shot; translates chosen `arrow_number` to its
          corresponding `arrow_id` UUID in the shot payload
        - If no `in_use` registered arrows exist: arrow tagging UI is hidden and `arrow_id` is sent
          as null
    - [ ] Implicit end grouping:
        - Automatically groups recorded shots into ends of size `currentSession.shots_per_end`
        - Displays current end counter and progress indicator (e.g. "Shot 2 of 6 — End 3")
    - [ ] Live stats integration:
        - Replaces WebSocket listener with HTTP GET `fetchLiveStats(sessionId)` triggered on mount
          and after each shot submission
    - [ ] Close session modal:
        - Modal opens when clicking "Finish Session"
        - Prompts "How was the session?" with 3 selectable options: `bad`, `neutral`, `good`
          (mandatory; submit button is disabled until one option is chosen)
        - If session has a `goal`: displays "Was goal achieved?" toggle (optional)
        - Text areas for "What went well?" (`did_well`, optional) and "What needs work?"
          (`need_work`, optional)
        - Submitting calls `closeSession` with `{ status, was_goal_achieved, did_well, need_work }`
          and redirects to `/app`
- [ ] Scoring widgets updated:
    - [ ] `LiveScore.vue`, `LiveStatsTable.vue`, and `MiniTable.vue` bind directly to session stats
- [ ] Component unit tests added:
    - [ ] `frontend/tests/components/LiveSession.spec.ts` verifies scored vs volume rendering,
          arrow number translation, end calculations, and session close submission
- [ ] `cd frontend && npm run test` passes.

## Files to Create/Modify

| Action | Path |
| ------ | ---- |
| Modify | `frontend/src/components/LiveSession.vue` |
| Modify | `frontend/src/components/widgets/LiveScore.vue` |
| Modify | `frontend/src/components/widgets/LiveStatsTable.vue` |
| Modify | `frontend/src/components/widgets/MiniTable.vue` |
| Create | `frontend/tests/components/LiveSession.spec.ts` |

## Reference

- [PRD.md](../PRD.md)
- [LiveSession.vue](../../../frontend/src/components/LiveSession.vue)
- [Face.vue](../../../frontend/src/components/Face.vue)

## Steps

- [ ] **Step 1: Write failing component tests for `LiveSession.vue`**

  Create `frontend/tests/components/LiveSession.spec.ts`:
    - Test scored session mounts `Face.vue`
    - Test volume session mounts `NoFaceSession.vue`
    - Test arrow chips appear when arrows are registered and translate `arrow_number` to `arrow_id`
    - Test end grouping calculations for N shots per end
    - Test close modal sends reflections and redirects

- [ ] **Step 2: Run Vitest to verify failure**

  ```bash
  cd frontend && npm run test
  ```

- [ ] **Step 3: Refactor `LiveSession.vue`**

  Remove slot and WebSocket imports. Bind directly to `currentSession`. Implement arrow tagging
  selector, end grouping, and HTTP live stats polling.

- [ ] **Step 4: Update scoring widgets**

  Ensure `LiveScore.vue`, `LiveStatsTable.vue`, and `MiniTable.vue` consume session statistics.

- [ ] **Step 5: Implement session close reflection modal**

  Add close modal with mandatory "How was the session?" rating (`bad`, `neutral`, `good`),
  conditional goal check, and optional reflection text fields. Wire submission to `closeSession`.

- [ ] **Step 6: Run Vitest and type-check**

  ```bash
  cd frontend && npm run type-check && npm run test
  ```

- [ ] **Step 7: Commit changes**

  ```bash
  git add frontend/src/components/ frontend/tests/
  git commit -m "feat(frontend): refactor LiveSession for single-archer model and arrow tagging"
  ```

## Verification

- `cd frontend && npm run type-check && npm run test` passes with 0 failures.
- Live session scoring works without WebSockets or slot assignments.
