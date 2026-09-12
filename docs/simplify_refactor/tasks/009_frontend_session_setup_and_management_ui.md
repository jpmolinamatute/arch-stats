# Task 009: Frontend Session Setup and Management UI

## Git Branch

`feature/009-frontend-session-setup-ui`

## Objective

Refactor the session creation and management interface in `frontend/src/components/`. Delete the
obsolete `SlotJoinForm.vue`. Refactor `SessionForm.vue` into a comprehensive session setup form that
loads the archer's registered bows, captures distance, face type (including volume session),
cappable shots per end, interval, and optional goal. Refactor `SessionManager.vue` to display active
sessions and past session history with soft deletion, eliminating slot assignment workflows.

## Dependencies

- Task 008 (Frontend composables refactoring and slot removal).

## Acceptance Criteria

- [ ] `frontend/src/components/forms/SlotJoinForm.vue` deleted
- [ ] `frontend/src/components/forms/SessionForm.vue` refactored:
    - [ ] Bow selector: loads registered bows; requires selection of active `bow_id`
    - [ ] Location: required text input (1-255 characters)
    - [ ] Environment: indoor / outdoor toggle
    - [ ] Distance: number input bounded between 1 and 100 meters
    - [ ] Face Type: selector supporting WA faces and `none` (Volume session)
    - [ ] Shots per End: number input (min 3); if archer has registered arrows, max is
          capped at `in_use` registered arrow count with explanatory helper text
    - [ ] Interval Seconds: number input bounded between 1 and 100 (default 20)
    - [ ] Goal: optional textarea (max 2000 characters)
    - [ ] Submission: calls `createSession` and immediately routes to `/app/live-session`
          without displaying any slot join UI
- [ ] `frontend/src/components/SessionManager.vue` refactored:
    - [ ] Slot joining tabs and lane configuration modals removed
    - [ ] Active session banner: when an open session exists, displays session details and a
          "Resume Session" button
    - [ ] New session view: renders `SessionForm.vue` when no open session exists
    - [ ] Session history list: renders closed sessions with soft-delete action using
          `ConfirmModal.vue`
- [ ] Component unit tests added:
    - [ ] `frontend/tests/components/SessionForm.spec.ts` verifies form rendering, bow selection,
          arrow capping validation, and submit routing
- [ ] `cd frontend && npm run test` passes.

## Files to Create/Modify/Remove

| Action | Path |
| ------ | ---- |
| Modify | `frontend/src/components/forms/SessionForm.vue` |
| Modify | `frontend/src/components/SessionManager.vue` |
| Create | `frontend/tests/components/SessionForm.spec.ts` |
| Delete | `frontend/src/components/forms/SlotJoinForm.vue` |

## Reference

- [PRD.md](../PRD.md)
- [SessionForm.vue](../../../frontend/src/components/forms/SessionForm.vue)
- [SessionManager.vue](../../../frontend/src/components/SessionManager.vue)

## Steps

- [ ] **Step 1: Write failing component tests for `SessionForm.vue`**

  Create `frontend/tests/components/SessionForm.spec.ts`:
    - Mount `SessionForm` with mocked `useSession`, `useAuth`, and bow list
    - Test bow dropdown populates with user bows
    - Test shots_per_end input reflects arrow count limit when arrows are present
    - Test form submission calls `createSession` with complete payload and emits event

- [ ] **Step 2: Run Vitest to verify failure**

  ```bash
  cd frontend && npm run test
  ```

- [ ] **Step 3: Implement refactored `SessionForm.vue`**

  Build the comprehensive session creation form. Include reactive arrow count capping and
  clean styling aligned with TailwindCSS tokens.

- [ ] **Step 4: Refactor `SessionManager.vue`**

  Remove all slot join buttons and imports of `SlotJoinForm.vue`. Implement active session card
  and session history table with delete action.

- [ ] **Step 5: Delete `SlotJoinForm.vue`**

  ```bash
  rm frontend/src/components/forms/SlotJoinForm.vue
  ```

- [ ] **Step 6: Run Vitest and type-check**

  ```bash
  cd frontend && npm run type-check && npm run test
  ```

- [ ] **Step 7: Commit changes**

  ```bash
  git add frontend/src/components/ frontend/tests/
  git commit -m "feat(frontend): refactor SessionForm and SessionManager; delete SlotJoinForm"
  ```

## Verification

- `cd frontend && npm run type-check && npm run test` passes with 0 failures.
- Navigating to `/app` presents the direct session form without any slot assignment references.
