# Task 010: Frontend Integration, Route Guards, and End-to-End Verification

## Git Branch

`feature/010-frontend-e2e-verification`

## Objective

Connect onboarding state into Vue Router navigation guards to ensure un-onboarded archers cannot
access shooting session views, verify the complete end-to-end user experience, and execute full
backend and frontend lint and test suites.

## Dependencies

- Task 006 (Backend integration tests)
- Task 009 (Frontend onboarding stepper UI)

## Acceptance Criteria

- [ ] Navigation guards in `frontend/src/router/index.ts`:
    - [ ] Users in `needs_registration` state attempting to access protected routes (`/app`,
          `/app/live-session`) are redirected to the onboarding wizard.
    - [ ] Completed onboarding immediately unlocks the shooting session dashboard.
- [ ] End-to-end flow verified:
    1. New Google user signs in.
    2. Prompted with Step 1 (Personal profile).
    3. Advanced to Step 2 (Register bow).
    4. Advanced to Step 3 (Register arrows with status `in_use`, or skip).
    5. Landed on `/app` ready to open a shooting session.
- [ ] Full backend checks pass:
    - [ ] `cd backend && go test -race ./... -v`
    - [ ] `./scripts/linting.bash --go`
- [ ] Full frontend checks pass:
    - [ ] `./scripts/linting.bash --frontend` (including `vue-tsc -b`, `npm run lint`,
          `npm run test`, and `vite build`)

## Files to Create/Modify

| Action | Path |
| ------ | ---- |
| Modify | `frontend/src/router/index.ts` |
| Create | `frontend/tests/e2e/onboarding.spec.ts` |

## Reference

- [PRD.md](../PRD.md)
- [router/index.ts](../../../frontend/src/router/index.ts)
- [story_time.md](../../../backend/migrations/story_time.md)

## Steps

- [ ] **Step 1: Update `frontend/src/router/index.ts` with onboarding navigation guard**

  Add `beforeEach` navigation guard:

  ```ts
  import { useAuth } from '@/composables/useAuth'

  router.beforeEach(async (to, from, next) => {
      const { isAuthenticated, pendingRegistration, loading } = useAuth()
      if (to.path.startsWith('/app')) {
          if (!isAuthenticated.value && pendingRegistration.value) {
              next({ name: 'landing' })
              return
          }
      }
      next()
  })
  ```

- [ ] **Step 2: Create E2E test `frontend/tests/e2e/onboarding.spec.ts`**

  Test the full user journey with mocked API responses:
    - Sign in -> Needs Registration.
    - Submit profile -> 201 Created.
    - Submit bow -> 201 Created -> Authenticated.
    - Confirm redirect to `/app`.

- [ ] **Step 3: Run backend test and lint suite**

  ```bash
  cd backend
  go test -race ./... -v
  cd ..
  ./scripts/linting.bash --go
  ```

- [ ] **Step 4: Run frontend test and lint suite**

  ```bash
  ./scripts/linting.bash --frontend
  ```

- [ ] **Step 5: Commit changes**

  ```bash
  git add frontend/src/router frontend/tests
  git commit -m "feat(frontend): add onboarding route guards and end-to-end verification"
  ```

## Verification

- `./scripts/linting.bash --go` exits with code 0.
- `./scripts/linting.bash --frontend` exits with code 0.
- All 10 tasks in `docs/onboarding_refactor/tasks/` are complete, actionable, and verified.
