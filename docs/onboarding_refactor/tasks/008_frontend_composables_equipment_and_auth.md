# Task 008: Frontend Composables for Bows, Arrows, and Onboarding State

## Git Branch

`feature/008-frontend-composables`

## Objective

Build Vue 3 Composition API composables (`useBow`, `useArrow`), update `useArcher`, and refactor
`useAuth` to support multi-step onboarding progression and manage equipment state across the
application.

## Dependencies

- Task 007 (OpenAPI specs and frontend type generation)

## Acceptance Criteria

- [ ] New composable `frontend/src/composables/useBow.ts` provides:
    - [ ] `bows`: reactive `Ref<BowRead[]>`
    - [ ] `hasBows`: `ComputedRef<boolean>` (true if `bows.value.length > 0`)
    - [ ] `loading`: `Ref<boolean>`
    - [ ] `error`: `Ref<string | null>`
    - [ ] `fetchBows(): Promise<BowRead[]>` calling `GET /api/v0/bows`
    - [ ] `registerBow(data: BowCreate): Promise<BowRead>` calling `POST /api/v0/bows`
- [ ] New composable `frontend/src/composables/useArrow.ts` provides:
    - [ ] `arrows`: reactive `Ref<ArrowRead[]>`
    - [ ] `hasArrows`: `ComputedRef<boolean>` (true if `arrows.value.length > 0`)
    - [ ] `inUseArrows`: `ComputedRef<ArrowRead[]>` (filtered for `status === 'in_use'`)
    - [ ] `arrowSets`: `ComputedRef<Record<number, ArrowRead[]>>` grouped by `arrow_set`
    - [ ] `loading`: `Ref<boolean>`
    - [ ] `error`: `Ref<string | null>`
    - [ ] `fetchArrows(): Promise<ArrowRead[]>` calling `GET /api/v0/arrows`
    - [ ] `registerArrowBatch(data: ArrowBatchCreate): Promise<ArrowRead[]>` calling
          `POST /api/v0/arrows` (arrows registered default to `status: 'in_use'`)
- [ ] Refactored `frontend/src/composables/useArcher.ts`:
    - [ ] `createProfile(data: ArcherCreate): Promise<ArcherRead>` calling
          `POST /api/v0/archers`
    - [ ] `fetchProfile(id?: string): Promise<ArcherRead>` calling
          `GET /api/v0/archers/{archer_id}`
- [ ] Refactored `frontend/src/composables/useAuth.ts`:
    - [ ] Decomposed legacy monolithic `registerNewArcher` into step-wise state updates
    - [ ] Reactive `onboardingStep` tracking progress: `'profile' | 'bow' | 'arrows' | 'ready'`
    - [ ] Seamless transition to `isAuthenticated = true` once profile and first bow exist
- [ ] Unit tests for each composable in `frontend/src/composables/__tests__/` using Vitest and MSW.
- [ ] `cd frontend && npm run test` passes.
- [ ] `cd frontend && npm run lint` passes with zero ESLint/Prettier errors.

## Files to Create/Modify

| Action | Path |
| ------ | ---- |
| Create | `frontend/src/composables/useBow.ts` |
| Create | `frontend/src/composables/__tests__/useBow.spec.ts` |
| Create | `frontend/src/composables/useArrow.ts` |
| Create | `frontend/src/composables/__tests__/useArrow.spec.ts` |
| Modify | `frontend/src/composables/useArcher.ts` |
| Modify | `frontend/src/composables/useAuth.ts` |
| Modify | `frontend/src/composables/__tests__/useAuth.spec.ts` |

## Reference

- [PRD.md](../PRD.md)
- [useAuth.ts](../../../frontend/src/composables/useAuth.ts)
- [useArcher.ts](../../../frontend/src/composables/useArcher.ts)

## Steps

- [ ] **Step 1: Write unit tests for `useBow.spec.ts`**

  Test initial state, successful bow fetching, registration error propagation, and reactive update
  to `hasBows`.

- [ ] **Step 2: Implement `frontend/src/composables/useBow.ts`**

  Use generated OpenAPI types and API client:

  ```ts
  import type { components } from '@/types/types.generated'
  import { computed, ref } from 'vue'
  import { api } from '@/api/client'

  export type BowRead = components['schemas']['BowRead']
  export type BowCreate = components['schemas']['BowCreate']

  const bows = ref<BowRead[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  export function useBow() {
      const hasBows = computed(() => bows.value.length > 0)

      async function fetchBows(): Promise<BowRead[]> {
          loading.value = true
          error.value = null
          try {
              const data = await api.get<BowRead[]>('/bows')
              bows.value = data ?? []
              return bows.value
          } catch (e) {
              error.value = e instanceof Error ? e.message : 'Failed to fetch bows'
              throw e
          } finally {
              loading.value = false
          }
      }

      async function registerBow(payload: BowCreate): Promise<BowRead> {
          loading.value = true
          error.value = null
          try {
              const created = await api.post<BowRead>('/bows', payload)
              bows.value.push(created)
              return created
          } catch (e) {
              error.value = e instanceof Error ? e.message : 'Failed to register bow'
              throw e
          } finally {
              loading.value = false
          }
      }

      return { bows, hasBows, loading, error, fetchBows, registerBow }
  }
  ```

- [ ] **Step 3: Implement `frontend/src/composables/useArrow.ts`**

  Implement `useArrow` with `registerArrowBatch`, grouping by `arrow_set`, and error handling.

- [ ] **Step 4: Update `useArcher.ts` and `useAuth.ts`**

  Update `useArcher` with `createProfile` using `ArcherCreate`.
  In `useAuth.ts`, update status detection: compute `onboardingStep` based on whether profile and
  bows are loaded.

- [ ] **Step 5: Run tests and linter**

  ```bash
  cd frontend
  npm run test
  npm run lint
  ```

- [ ] **Step 6: Commit changes**

  ```bash
  git add frontend/src/composables
  git commit -m "feat(frontend): add useBow and useArrow composables, refactor useAuth"
  ```

## Verification

- `cd frontend && npm run test` passes all unit tests.
- `cd frontend && npm run lint` passes with 0 lint errors.
