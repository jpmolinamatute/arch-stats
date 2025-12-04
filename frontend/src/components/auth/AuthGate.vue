<script setup lang="ts">
import type { components } from '@/types/types.generated'
import { computed, ref } from 'vue'
import { useAuth } from '@/composables/useAuth'

defineProps<{ buttonTarget?: string }>()
const { pendingRegistration, registerNewArcher, isAuthenticated, user, loading } = useAuth()

const firstName = ref<string>('')
const lastName = ref<string>('')
type GenderType = components['schemas']['GenderType']
type BowStyleType = components['schemas']['BowStyleType']
const gender = ref<GenderType>('unspecified')
// Start with empty selection to force the user to choose a valid bowstyle
const bowstyle = ref<BowStyleType | ''>('')
const dob = ref<string>('2000-01-01')
const drawWeight = ref<number | null>(null)
const formError = ref<string | null>(null)

const needsName = computed(
  () =>
    !!pendingRegistration.value
    && (pendingRegistration.value.need_first_name || pendingRegistration.value.need_last_name),
)

async function submitRegistration() {
  formError.value = null
  if (!bowstyle.value) {
    formError.value = 'Please select a bowstyle.'
    return
  }
  if (drawWeight.value == null || Number.isNaN(drawWeight.value) || drawWeight.value <= 0) {
    formError.value = 'Please enter a valid positive draw weight (lbs).'
    return
  }
  try {
    await registerNewArcher({
      date_of_birth: dob.value,
      gender: gender.value,
      bowstyle: bowstyle.value,
      draw_weight: drawWeight.value,
      first_name: needsName.value ? firstName.value : undefined,
      last_name: needsName.value ? lastName.value : undefined,
    })
  }
  catch (e) {
    // Surface a friendly error if registration wasn't properly started
    formError.value
      = e instanceof Error ? e.message : 'Registration failed. Please try again.'
  }
}
</script>

<template>
  <div class="w-full max-w-md mx-auto space-y-4">
    <div v-if="loading" class="p-6 bg-transparent rounded-lg border border-slate-800">
      <div class="h-3 w-32 bg-gray-200 rounded mb-4" />
      <div class="space-y-2">
        <div class="h-9 bg-gray-100 rounded" />
        <div class="h-9 bg-gray-100 rounded" />
        <div class="h-9 bg-gray-100 rounded" />
      </div>
    </div>

    <template v-else>
      <div
        v-if="isAuthenticated"
        class="p-3 rounded border border-green-700/40 bg-transparent text-sm text-green-200"
      >
        Welcome {{ user?.first_name ?? '' }} {{ user?.last_name ?? '' }}
      </div>

      <!-- Registration form is only visible after Google replies needs_registration -->
      <form
        v-else-if="pendingRegistration"
        class="space-y-4 p-6 bg-transparent rounded-lg border border-slate-800 text-slate-200"
        @submit.prevent="submitRegistration"
      >
        <div class="text-left text-slate-200 text-sm">
          <p class="font-semibold">
            Complete your registration
          </p>
          <p class="text-xs text-slate-400">
            We need a few more details before creating your account.
          </p>
        </div>

        <p v-if="formError" class="text-sm text-red-600">
          {{ formError }}
        </p>

        <div v-if="needsName" class="grid grid-cols-2 gap-2">
          <label class="block text-left text-xs text-slate-300">
            First name
            <input
              v-model="firstName"
              type="text"
              class="mt-1 w-full border border-slate-700 p-2 rounded bg-slate-900 text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              required
            >
          </label>
          <label class="block text-left text-xs text-slate-300">
            Last name
            <input
              v-model="lastName"
              type="text"
              class="mt-1 w-full border border-slate-700 p-2 rounded bg-slate-900 text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              required
            >
          </label>
        </div>

        <label class="block text-left text-xs text-slate-300">
          Date of birth
          <input
            v-model="dob"
            type="date"
            class="mt-1 w-full border border-slate-700 p-2 rounded bg-slate-900 text-slate-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            required
          >
        </label>

        <label class="block text-left text-xs text-slate-300">
          Gender
          <select
            v-model="gender"
            class="mt-1 w-full border border-slate-700 p-2 rounded bg-slate-900 text-slate-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          >
            <option value="unspecified">Unspecified</option>
            <option value="male">Male</option>
            <option value="female">Female</option>
            <option value="non_binary">Non-binary</option>
            <option value="other">Other</option>
          </select>
        </label>

        <label class="block text-left text-xs text-slate-300">
          Bowstyle
          <select
            v-model="bowstyle"
            class="mt-1 w-full border border-slate-700 p-2 rounded bg-slate-900 text-slate-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            required
          >
            <option disabled value="">Select a bowstyle</option>
            <option value="recurve">Recurve</option>
            <option value="compound">Compound</option>
            <option value="barebow">Barebow</option>
            <option value="longbow">Longbow</option>
          </select>
        </label>

        <label class="block text-left text-xs text-slate-300">
          Draw weight (lbs)
          <input
            v-model.number="drawWeight"
            type="number"
            min="1"
            step="0.1"
            placeholder="e.g., 36"
            class="mt-1 w-full border border-slate-700 p-2 rounded bg-slate-900 text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            required
          >
        </label>

        <button
          type="submit"
          class="w-full px-4 py-2 text-sm rounded bg-green-600 text-white disabled:opacity-50"
          :disabled="loading"
        >
          Create account and continue
        </button>
      </form>
      <div
        v-else
        class="space-y-2 p-6 bg-transparent rounded-lg border border-slate-800 text-slate-200 text-left"
      >
        <p class="font-semibold text-sm">
          Sign in to start
        </p>
        <p class="text-xs text-gray-600">
          Google One Tap will appear automatically. After you continue, we'll ask for a
          few details to finish creating your account.
        </p>
      </div>
    </template>
  </div>
</template>
