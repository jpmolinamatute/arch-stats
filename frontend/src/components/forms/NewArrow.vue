<script setup lang="ts">
    import { ref, onMounted, computed } from 'vue';

    const emit = defineEmits<{ submit: [] }>();

    const form = ref({
        id: '',
        length: null as number | null,
        human_identifier: '',
        is_programmed: false,
        label_position: null as number | null,
        weight: null as number | null,
        diameter: null as number | null,
        spine: null as number | null,
    });

    const errors = ref({
        length: '',
        human_identifier: '',
        label_position: '',
        weight: '',
        diameter: '',
        spine: '',
    });

    const loadError = ref('');

    function validateOptionalFloat(
        fieldValue: number | null,
        fieldName: keyof typeof errors.value,
        fieldLabel: string,
    ): boolean {
        if (fieldValue !== null) {
            if (typeof fieldValue !== 'number' || isNaN(fieldValue) || fieldValue <= 0) {
                errors.value[fieldName] = `${fieldLabel} must be a positive number`;
                return false;
            }
        }
        return true;
    }

    const isFormValid = computed(() => {
        let valid = true;
        errors.value = {
            length: '',
            human_identifier: '',
            label_position: '',
            weight: '',
            diameter: '',
            spine: '',
        };
        if (!form.value.id) {
            console.error('Arrow ID is missing');
            valid = false;
        }
        if (!form.value.human_identifier.trim()) {
            errors.value.human_identifier = 'Name is required';
            valid = false;
        }

        if (!validateOptionalFloat(form.value.length, 'length', 'Length')) valid = false;
        if (!validateOptionalFloat(form.value.label_position, 'label_position', 'Label position'))
            valid = false;
        if (!validateOptionalFloat(form.value.weight, 'weight', 'Weight')) valid = false;
        if (!validateOptionalFloat(form.value.diameter, 'diameter', 'Diameter')) valid = false;
        if (!validateOptionalFloat(form.value.spine, 'spine', 'Spine')) valid = false;

        return valid;
    });

    async function fetchNewArrowId() {
        try {
            const response = await fetch('/api/v0/arrow/new_arrow_uuid');
            const data = await response.json();

            if (!response.ok) {
                loadError.value = data['errors'] || 'Failed to fetch new arrow ID';
                return;
            }

            form.value.id = data.id;
        } catch (err) {
            console.error(err);
            loadError.value = 'Network or server error while fetching arrow ID';
        }
    }

    async function handleSubmit() {
        if (!isFormValid.value) {
            console.log('Validation failed:', errors.value);
            return;
        }

        const payload = {
            id: form.value.id,
            length: form.value.length,
            human_identifier: form.value.human_identifier.trim(),
            is_programmed: form.value.is_programmed,
            label_position: form.value.label_position,
            weight: form.value.weight,
            diameter: form.value.diameter,
            spine: form.value.spine,
        };

        try {
            const response = await fetch('/api/v0/arrow', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload),
            });
            if (!response.ok) throw new Error('Failed to submit arrow');
            emit('submit');
        } catch (err) {
            console.error('Submission failed', err);
        }
    }

    onMounted(fetchNewArrowId);
</script>

<template>
    <div
        v-if="loadError"
        class="max-w-md mx-auto p-4 bg-danger/20 border border-danger text-danger rounded"
    >
        <p>Error: {{ loadError }}</p>
    </div>
    <form
        @submit.prevent="handleSubmit"
        class="max-w-md mx-auto p-4 bg-slate-800/70 backdrop-blur rounded-lg shadow-md space-y-4 border border-slate-700 text-slate-100"
    >
        <div>
            <label class="block mb-1 font-medium text-slate-200"
                >Arrow ID:
                <input
                    type="text"
                    :value="form.id"
                    required
                    readonly
                    class="w-full border border-slate-600 rounded px-3 py-2 bg-slate-700 text-slate-300"
                />
            </label>
        </div>

        <div>
            <label class="block mb-1 font-medium text-slate-200"
                >Name:
                <input
                    type="text"
                    v-model="form.human_identifier"
                    required
                    class="w-full border border-slate-600 rounded px-3 py-2 bg-slate-900 text-slate-100 focus:outline-none focus:ring-2 focus:ring-primary"
                />
                <p v-if="errors.human_identifier" class="text-danger text-sm mt-1">
                    {{ errors.human_identifier }}
                </p>
            </label>
        </div>

        <div>
            <label class="block mb-1 font-medium text-slate-200">Length:</label>
            <input
                type="number"
                step="any"
                v-model.number="form.length"
                class="w-full border border-slate-600 rounded px-3 py-2 bg-slate-900 text-slate-100 focus:outline-none focus:ring-2 focus:ring-primary"
            />
            <p v-if="errors.length" class="text-danger text-sm mt-1">{{ errors.length }}</p>
        </div>

        <div>
            <label class="block mb-1 font-medium text-slate-200">Label Position:</label>
            <input
                type="number"
                step="any"
                v-model.number="form.label_position"
                class="w-full border border-slate-600 rounded px-3 py-2 bg-slate-900 text-slate-100 focus:outline-none focus:ring-2 focus:ring-primary"
            />
            <p v-if="errors.label_position" class="text-danger text-sm mt-1">
                {{ errors.label_position }}
            </p>
        </div>

        <div>
            <label class="block mb-1 font-medium text-slate-200">Weight:</label>
            <input
                type="number"
                step="any"
                v-model.number="form.weight"
                class="w-full border border-slate-600 rounded px-3 py-2 bg-slate-900 text-slate-100 focus:outline-none focus:ring-2 focus:ring-primary"
            />
            <p v-if="errors.weight" class="text-danger text-sm mt-1">{{ errors.weight }}</p>
        </div>

        <div>
            <label class="block mb-1 font-medium text-slate-200">Diameter:</label>
            <input
                type="number"
                step="any"
                v-model.number="form.diameter"
                class="w-full border border-slate-600 rounded px-3 py-2 bg-slate-900 text-slate-100 focus:outline-none focus:ring-2 focus:ring-primary"
            />
            <p v-if="errors.diameter" class="text-danger text-sm mt-1">{{ errors.diameter }}</p>
        </div>

        <div>
            <label class="block mb-1 font-medium text-slate-200">Spine:</label>
            <input
                type="number"
                step="any"
                v-model.number="form.spine"
                class="w-full border border-slate-600 rounded px-3 py-2 bg-slate-900 text-slate-100 focus:outline-none focus:ring-2 focus:ring-primary"
            />
            <p v-if="errors.spine" class="text-danger text-sm mt-1">{{ errors.spine }}</p>
        </div>

        <div>
            <label class="block mb-1 font-medium text-slate-200">Is Programmed:</label>
            <input
                type="text"
                :value="form.is_programmed ? 'Yes' : 'No'"
                required
                readonly
                class="w-full border border-slate-600 rounded px-3 py-2 bg-slate-700 text-slate-300"
            />
        </div>

        <button
            type="submit"
            class="w-full bg-primary hover:bg-primary/90 text-white py-2 rounded transition-colors"
        >
            Register Arrow
        </button>
    </form>
</template>
