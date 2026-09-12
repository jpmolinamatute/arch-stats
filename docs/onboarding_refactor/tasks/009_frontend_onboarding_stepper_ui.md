# Task 009: Frontend Onboarding Stepper UI and Component Implementation

## Git Branch

`feature/009-frontend-onboarding-ui`

## Objective

Build a streamlined, accessible multi-step onboarding wizard in Vue 3 that guides newly
authenticated Google users through personal profile completion (mandatory), first bow registration
(mandatory), and optional arrow registration (with a prominent skip option).

## Dependencies

- Task 007 (OpenAPI specs and frontend type generation)
- Task 008 (Frontend composables for bows, arrows, and onboarding state)

## Acceptance Criteria

- [ ] New onboarding components created in `frontend/src/components/onboarding/`:
    - [ ] `OnboardingWizard.vue`: Master container with visual progress indicator showing:
          Step 1 (Profile) -> Step 2 (Bow) -> Step 3 (Arrows, optional).
    - [ ] `StepProfile.vue`: Form for `first_name`, `last_name`, `email`, `date_of_birth`
          (validating archer is >= 10 years old), and `gender`.
    - [ ] `StepBow.vue`: Form for `name`, `bowstyle`, and `draw_weight` (positive number in lbs).
    - [ ] `StepArrows.vue`: Form for `arrow_set`, arrow count (minimum 3), optional spine/length/
          weight (does NOT prompt for arrow status; all arrows default to `'in_use'`), AND a
          prominent secondary button: "Skip for now & start shooting".
- [ ] Modified `frontend/src/components/auth/AuthGate.vue`:
    - [ ] Replaces legacy single registration form with `OnboardingWizard.vue` when `authStatus`
          is `needs_registration`.
    - [ ] Retains Google One Tap login prompt for unauthenticated users.
- [ ] Accessible UI with keyboard navigation, clear error alerts, and loading indicators during
- [ ] Component unit tests in `frontend/src/components/onboarding/__tests__/`
      verify form validation, step transitions, and skip behavior.
- [ ] `cd frontend && npm run test` passes.
- [ ] `cd frontend && npm run lint` passes with 0 errors.

## Files to Create/Modify

| Action | Path |
| ------ | ---- |
| Create | `frontend/src/components/onboarding/OnboardingWizard.vue` |
| Create | `frontend/src/components/onboarding/StepProfile.vue` |
| Create | `frontend/src/components/onboarding/StepBow.vue` |
| Create | `frontend/src/components/onboarding/StepArrows.vue` |
| Create | `frontend/src/components/onboarding/__tests__/OnboardingWizard.spec.ts` |
| Modify | `frontend/src/components/auth/AuthGate.vue` |

## Reference

- [PRD.md](../PRD.md)
- [story_time.md](../../../backend/migrations/story_time.md)
- [AuthGate.vue](../../../frontend/src/components/auth/AuthGate.vue)

## Steps

- [ ] **Step 1: Write component tests for `OnboardingWizard.spec.ts`**

  Test:
    - Renders Step 1 by default when user profile is missing.
    - Prevents proceeding from Step 1 if date of birth is under 10 years old.
    - Advances to Step 2 upon profile submission.
    - Advances to Step 3 upon bow registration.
    - Clicking "Skip for now" on Step 3 completes onboarding and emits `completed`.

- [ ] **Step 2: Implement `StepProfile.vue`**

  Collect First Name, Last Name, Email, DOB, Gender:

  ```vue
  <script setup lang="ts">
  import { ref, computed } from 'vue'
  import { useArcher } from '@/composables/useArcher'

  const emit = defineEmits<{ (e: 'success'): void }>()
  const { createProfile, loading, error } = useArcher()

  const firstName = ref('')
  const lastName = ref('')
  const email = ref('')
  const dob = ref('')
  const gender = ref('unspecified')
  const clientError = ref<string | null>(null)

  function isAtLeastTenYearsOld(birthDateStr: string): boolean {
      const birth = new Date(birthDateStr)
      const cutoff = new Date()
      cutoff.setFullYear(cutoff.getFullYear() - 10)
      return birth <= cutoff
  }

  async function submit() {
      clientError.value = null
      if (!isAtLeastTenYearsOld(dob.value)) {
          clientError.value = 'Archer must be at least 10 years old.'
          return
      }
      await createProfile({
          first_name: firstName.value,
          last_name: lastName.value,
          email: email.value,
          date_of_birth: dob.value,
          gender: gender.value,
      })
      emit('success')
  }
  </script>
  ```

- [ ] **Step 3: Implement `StepBow.vue`**

  Collect bow name, bowstyle (dropdown), draw weight (number > 0).

- [ ] **Step 4: Implement `StepArrows.vue`**

  Provide inputs for arrow set number (SMALLINT) and arrow count with validation `count >= 3`.
  Do not ask the archer for arrow status (all registered arrows default to `'in_use'`).
  Provide secondary action "Skip for now & start shooting" which immediately calls `emit('skip')`.

- [ ] **Step 5: Implement `OnboardingWizard.vue`**

  Wrap steps inside a container with progress pills:

  ```vue
  <script setup lang="ts">
  import { ref } from 'vue'
  import StepProfile from './StepProfile.vue'
  import StepBow from './StepBow.vue'
  import StepArrows from './StepArrows.vue'

  const step = ref<'profile' | 'bow' | 'arrows'>('profile')
  const emit = defineEmits<{ (e: 'finish'): void }>()
  </script>
  ```

- [ ] **Step 6: Update `AuthGate.vue`**

  Replace the legacy inline registration form with `<OnboardingWizard @finish="onComplete" />`.

- [ ] **Step 7: Run frontend test and lint suite**

  ```bash
  cd frontend
  npm run test
  npm run lint
  ```

- [ ] **Step 8: Commit changes**

  ```bash
  git add frontend/src/components/onboarding frontend/src/components/auth
  git commit -m "feat(frontend): implement multi-step onboarding wizard and update AuthGate"
  ```

## Verification

- `cd frontend && npm run test` passes all tests.
- `cd frontend && npm run lint` passes cleanly.
- Browser test verifies smooth step-by-step onboarding and skip arrow functionality.
