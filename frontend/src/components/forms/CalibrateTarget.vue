<script setup lang="ts">
    import { ref, computed, watchEffect } from 'vue';
    import type { StepRegistration } from '../widgets/Wizard.vue';
    import { createTarget, fetchTargetCalibration } from '../../composables/useTarget';
    import type { components } from '../../types/types.generated';
    import { sessionOpened } from '../../state/session';

    const props = defineProps<{ register: (options: StepRegistration) => void }>();
    const success = ref(false);
    const maxX = ref<number>(0);
    const maxY = ref<number>(0);
    const distance = ref<number>(18);
    const errorMessage = ref<string>('');
    const loading = ref(false);
    const isValid = computed(
        () => success.value && Number.isFinite(distance.value) && distance.value > 0,
    );
    const onComplete = ref<() => Promise<{ success: boolean; error?: string }>>(async () => {
        if (!success.value) {
            return { success: false, error: errorMessage.value || 'Calibration not completed' };
        }
        if (!sessionOpened.id) {
            return { success: false, error: 'No open session found.' };
        }

        const payload: components['schemas']['TargetsCreate'] = {
            max_x: maxX.value,
            max_y: maxY.value,
            session_id: sessionOpened.id,
            distance: distance.value,
        };
        const created = await createTarget(payload);
        if (!created) {
            return { success: false, error: 'Failed to save target configuration' };
        }
        return { success: true };
    });

    watchEffect(() => {
        props.register({ isValid, onComplete });
    });

    async function calibrate() {
        loading.value = true;
        success.value = false;
        errorMessage.value = '';
        try {
            const cal = await fetchTargetCalibration();
            maxX.value = cal.max_x;
            maxY.value = cal.max_y;
            success.value = true;
        } catch (e) {
            errorMessage.value = (e as Error).message;
        } finally {
            loading.value = false;
        }
    }
</script>

<template>
    <div class="p-4 bg-slate-800/60 rounded border border-slate-700">
        <button
            @click="calibrate"
            :disabled="loading"
            class="px-4 py-2 bg-primary text-white rounded hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed"
        >
            {{ loading ? 'Calibrating…' : 'Calibrate' }}
        </button>

        <div v-if="errorMessage" class="mt-2 text-danger">
            {{ errorMessage }}
        </div>

        <div v-else-if="success" class="mt-4">
            <div class="flex space-x-4 mb-4 text-slate-200">
                <label class="flex flex-col font-semibold">
                    Max X:
                    <input
                        type="number"
                        :value="maxX"
                        readonly
                        class="bg-slate-700 border border-slate-600 p-2 rounded text-slate-100"
                    />
                </label>
                <label class="flex flex-col font-semibold">
                    Max Y:
                    <input
                        type="number"
                        :value="maxY"
                        readonly
                        class="bg-slate-700 border border-slate-600 p-2 rounded text-slate-100"
                    />
                </label>
                <label class="flex flex-col font-semibold">
                    Distance (m):
                    <input
                        type="number"
                        v-model.number="distance"
                        min="1"
                        step="1"
                        class="bg-slate-900 border border-slate-600 p-2 rounded text-slate-100"
                        aria-description="User provided distance; sensor does not currently supply it"
                    />
                </label>
            </div>
        </div>
    </div>
</template>
