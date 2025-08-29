<script setup lang="ts">
    import { ref, computed, onMounted } from 'vue';
    import type { StepRegistration } from '../widgets/Wizard.vue';
    import { createArrow, getNewArrowUuid } from '../../composables/useArrow';
    import type { components } from '../../types/types.generated';

    type ArrowsCreate = components['schemas']['ArrowsCreate'];

    const props = defineProps<{ register: (opts: StepRegistration) => void }>();

    // Local form state matching ArrowsCreate (id required + optional fields)
    type Mutable<T> = { -readonly [K in keyof T]: T[K] };
    const form = ref<Mutable<ArrowsCreate>>({
        id: '',
        human_identifier: '',
        length: 0,
        registration_date: new Date().toISOString(),
        is_programmed: false,
        is_active: true,
        voided_date: null,
        weight: null,
        diameter: null,
        spine: null,
        label_position: null,
    });

    const loadError = ref<string | null>(null);
    const submitting = ref(false);

    // Ask user if they want to continue registering arrows; default Yes
    const wantAnother = ref<boolean>(true);

    async function fetchUuid() {
        try {
            loadError.value = null;
            form.value.id = await getNewArrowUuid();
        } catch (e: unknown) {
            loadError.value = e instanceof Error ? e.message : 'Failed to get new arrow UUID';
        }
    }

    function resetFormForNext() {
        form.value = {
            id: '',
            human_identifier: '',
            length: 0,
            registration_date: new Date().toISOString(),
            is_programmed: false,
            is_active: true,
            voided_date: null,
            weight: null,
            diameter: null,
            spine: null,
            label_position: null,
        } as Mutable<ArrowsCreate>;
    }

    onMounted(fetchUuid);

    // Validation
    const errors = ref<Record<string, string>>({});
    const touchedName = ref(false);
    const touchedLength = ref(false);
    const showAllErrors = ref(false);

    function validateNumberPositive(val: number | null, label: string): string {
        if (val === null || val === undefined) return '';
        if (typeof val !== 'number' || Number.isNaN(val) || val <= 0) {
            return `${label} must be a positive number`;
        }
        return '';
    }

    function computeValidityOnly(): boolean {
        if (!form.value.id) return false;
        if (!form.value.human_identifier?.trim()) return false;
        if (!Number.isFinite(form.value.length) || form.value.length <= 0) return false;
        const optionals: Array<[keyof ArrowsCreate, string]> = [
            ['label_position', 'Label position'],
            ['weight', 'Weight'],
            ['diameter', 'Diameter'],
            ['spine', 'Spine'],
        ];
        for (const [key, label] of optionals) {
            const msg = validateNumberPositive(form.value[key] as number | null, label);
            if (msg) return false;
        }
        return true;
    }

    function updateErrors(displayAll: boolean): void {
        const showName = displayAll || touchedName.value;
        const showLength = displayAll || touchedLength.value;
        const next: Record<string, string> = {};

        if (showName && !form.value.human_identifier?.trim()) {
            next.human_identifier = 'Name is required';
        }
        if (showLength && (!Number.isFinite(form.value.length) || form.value.length <= 0)) {
            next.length = 'Length is required and must be > 0';
        }
        const checks: Array<[keyof ArrowsCreate, string, boolean]> = [
            ['label_position', 'Label position', displayAll],
            ['weight', 'Weight', displayAll],
            ['diameter', 'Diameter', displayAll],
            ['spine', 'Spine', displayAll],
        ];
        for (const [key, label, shouldShow] of checks) {
            if (!shouldShow) continue; // only show optionals on full submit or if we later track touched per-field
            const msg = validateNumberPositive(form.value[key] as number | null, label);
            if (msg) next[key as string] = msg;
        }
        errors.value = next;
    }

    const isValid = computed(() => computeValidityOnly());

    // Register with Wizard: provide isValid, onComplete, and dynamic hasNext tied to wantAnother
    const onComplete = ref<() => Promise<{ success: boolean; error?: string }>>(async () => {
        try {
            // Force showing all errors if invalid on submit
            if (!computeValidityOnly()) {
                showAllErrors.value = true;
                updateErrors(true);
                return { success: false, error: 'Please fix the highlighted fields' };
            }
            submitting.value = true;
            await createArrow(form.value);
            submitting.value = false;

            if (wantAnother.value) {
                // Prepare for next cycle: clear fields and fetch a fresh UUID
                resetFormForNext();
                await fetchUuid();
                // Reset UX flags
                touchedName.value = false;
                touchedLength.value = false;
                showAllErrors.value = false;
                errors.value = {};
            }

            return { success: true };
        } catch (e: unknown) {
            submitting.value = false;
            const msg = e instanceof Error ? e.message : 'Failed to register arrow';
            return { success: false, error: msg };
        }
    });

    props.register({ isValid, onComplete, hasNext: wantAnother });
</script>

<template>
    <div class="p-4 bg-slate-800/60 rounded border border-slate-700">
        <div v-if="loadError" class="mb-3 text-danger text-sm">{{ loadError }}</div>

        <!-- Arrow ID (read-only) -->
        <label class="block mb-1 font-medium text-slate-200">Arrow ID</label>
        <input
            type="text"
            :value="form.id"
            readonly
            class="w-full border border-slate-600 rounded p-2 mb-3 bg-slate-700 text-slate-300"
        />

        <!-- Human Identifier -->
        <label class="block mb-1 font-medium text-slate-200">Name</label>
        <input
            v-model.trim="form.human_identifier"
            type="text"
            @blur="
                touchedName = true;
                updateErrors(showAllErrors);
            "
            class="w-full border border-slate-600 rounded p-2 mb-1 bg-slate-900 text-slate-100"
            placeholder="Arrow nickname"
        />
        <p v-if="errors.human_identifier" class="text-danger text-sm mb-2">
            {{ errors.human_identifier }}
        </p>

        <!-- Optional numeric fields -->
        <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
            <div>
                <label class="block mb-1 font-medium text-slate-200">Length</label>
                <input
                    v-model.number="form.length"
                    type="number"
                    step="any"
                    @blur="
                        touchedLength = true;
                        updateErrors(showAllErrors);
                    "
                    class="w-full border border-slate-600 rounded p-2 bg-slate-900 text-slate-100"
                />
                <p v-if="errors.length" class="text-danger text-sm">{{ errors.length }}</p>
            </div>
            <div>
                <label class="block mb-1 font-medium text-slate-200">Label Position</label>
                <input
                    v-model.number="form.label_position"
                    type="number"
                    step="any"
                    class="w-full border border-slate-600 rounded p-2 bg-slate-900 text-slate-100"
                />
                <p v-if="errors.label_position" class="text-danger text-sm">
                    {{ errors.label_position }}
                </p>
            </div>
            <div>
                <label class="block mb-1 font-medium text-slate-200">Weight</label>
                <input
                    v-model.number="form.weight"
                    type="number"
                    step="any"
                    class="w-full border border-slate-600 rounded p-2 bg-slate-900 text-slate-100"
                />
                <p v-if="errors.weight" class="text-danger text-sm">{{ errors.weight }}</p>
            </div>
            <div>
                <label class="block mb-1 font-medium text-slate-200">Diameter</label>
                <input
                    v-model.number="form.diameter"
                    type="number"
                    step="any"
                    class="w-full border border-slate-600 rounded p-2 bg-slate-900 text-slate-100"
                />
                <p v-if="errors.diameter" class="text-danger text-sm">{{ errors.diameter }}</p>
            </div>
            <div>
                <label class="block mb-1 font-medium text-slate-200">Spine</label>
                <input
                    v-model.number="form.spine"
                    type="number"
                    step="any"
                    class="w-full border border-slate-600 rounded p-2 bg-slate-900 text-slate-100"
                />
                <p v-if="errors.spine" class="text-danger text-sm">{{ errors.spine }}</p>
            </div>
        </div>

        <!-- Is Programmed (read-only for now) -->
        <div class="mt-4">
            <span class="block mb-1 font-medium text-slate-200"> Is the arrow programmed? </span>
            <div class="flex gap-4 items-center">
                <label class="inline-flex items-center gap-2">
                    <input
                        type="radio"
                        name="programmed"
                        v-model="form.is_programmed"
                        :value="true"
                    />
                    <span>Yes</span>
                </label>
                <label class="inline-flex items-center gap-2">
                    <input
                        type="radio"
                        name="programmed"
                        v-model="form.is_programmed"
                        :value="false"
                    />
                    <span>No</span>
                </label>
            </div>
        </div>

        <!-- Yes/No toggle to continue -->
        <div class="mt-4">
            <span class="block mb-1 font-medium text-slate-200">
                Do you want to register another Arrow?
            </span>
            <div class="flex gap-4 items-center">
                <label class="inline-flex items-center gap-2">
                    <input type="radio" name="another" v-model="wantAnother" :value="true" />
                    <span>Yes</span>
                </label>
                <label class="inline-flex items-center gap-2">
                    <input type="radio" name="another" v-model="wantAnother" :value="false" />
                    <span>No</span>
                </label>
            </div>
        </div>
        <div v-if="submitting" class="mt-2 text-slate-400 text-sm">Submitting…</div>
    </div>
</template>
