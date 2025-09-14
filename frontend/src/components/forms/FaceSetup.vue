<script setup lang="ts">
    import { ref, computed, watchEffect, onMounted } from 'vue';
    import type { StepRegistration } from '../widgets/Wizard.vue';
    import { sessionOpened } from '../../state/session';
    import { getTargetsBySessionId } from '../../composables/useTarget';
    import {
        createFacesForTarget,
        fetchFaceCalibration,
        type FaceCalibration,
    } from '../../composables/useFace';
    import type { components } from '../../types/types.generated';

    interface EditableFace {
        x: number; // sensor provided (read-only)
        y: number; // sensor provided (read-only)
        radii: number[]; // sensor provided (read-only)
        humanIdentifier: string; // user provided
        points: string; // CSV points (optional user provided)
    }

    const props = defineProps<{ register: (o: StepRegistration) => void }>();

    const targets = ref<components['schemas']['TargetsRead'][]>([]);
    const selectedTargetId = ref<string | null>(null);
    const faces = ref<EditableFace[]>([]);
    const loadingTargets = ref(false);
    const errorMessage = ref<string>('');
    const saving = ref(false);

    async function loadTargets() {
        if (!sessionOpened.id) return;
        loadingTargets.value = true;
        errorMessage.value = '';
        try {
            targets.value = await getTargetsBySessionId(sessionOpened.id);
            if (targets.value.length === 1 && targets.value[0]) {
                selectedTargetId.value = targets.value[0]?.id;
            }
        } catch (e) {
            errorMessage.value = (e as Error).message;
        } finally {
            loadingTargets.value = false;
        }
    }

    onMounted(loadTargets);

    const MAX_FACES = 3;
    async function addCalibrationFace() {
        if (!selectedTargetId.value) return;
        if (faces.value.length >= MAX_FACES) return;
        try {
            const cal: FaceCalibration = await fetchFaceCalibration();
            faces.value.push({
                x: cal.x,
                y: cal.y,
                radii: cal.radii,
                humanIdentifier: '',
                points: '',
            });
        } catch (e) {
            errorMessage.value = (e as Error).message;
        }
    }
    function removeFaceRow(idx: number) {
        faces.value.splice(idx, 1);
    }

    const isValid = computed(() => {
        if (!selectedTargetId.value) return true; // allow skip
        return faces.value.every((f) => {
            if (!f.humanIdentifier.trim()) return false;
            if (f.points.trim()) {
                const pts = f.points
                    .split(',')
                    .map((p) => p.trim())
                    .filter(Boolean);
                if (pts.length !== f.radii.length) return false;
                if (pts.some((p) => isNaN(Number(p)))) return false;
            }
            return true;
        });
    });

    const onComplete = ref<() => Promise<{ success: boolean; error?: string }>>(async () => {
        // Faces step is optional entirely
        if (!selectedTargetId.value || faces.value.length === 0) {
            return { success: true };
        }
        saving.value = true;
        try {
            await createFacesForTarget(
                selectedTargetId.value,
                faces.value.map((f) => ({
                    x: f.x,
                    y: f.y,
                    radii: f.radii,
                    human_identifier: f.humanIdentifier.trim(),
                    points: f.points
                        ? f.points
                              .split(',')
                              .map((p) => Number(p.trim()))
                              .filter((n) => !isNaN(n))
                        : undefined,
                })),
            );
            return { success: true };
        } catch (e) {
            return { success: false, error: (e as Error).message };
        } finally {
            saving.value = false;
        }
    });

    watchEffect(() => props.register({ isValid, onComplete }));
</script>

<template>
    <div class="p-4 bg-slate-800/60 rounded border border-slate-700">
        <p class="text-sm text-slate-400 mb-2">Faces are optional; skip this step to finish.</p>
        <div v-if="errorMessage" class="text-danger mb-2">{{ errorMessage }}</div>

        <div class="mb-4">
            <label class="block mb-1 font-medium text-slate-200">Select Target</label>
            <select
                v-model="selectedTargetId"
                class="w-full bg-slate-900 border border-slate-600 p-2 rounded text-slate-100"
            >
                <option :value="null">-- Skip (no faces) --</option>
                <option v-for="t in targets" :key="t.id" :value="t.id">
                    {{ t.id.slice(0, 8) }} • dist {{ t.distance }}m ({{ t.max_x }}x{{ t.max_y }})
                </option>
            </select>
        </div>

        <div v-if="selectedTargetId">
            <div class="flex items-center justify-between mb-2">
                <h3 class="font-semibold text-slate-200">Faces</h3>
                <button
                    type="button"
                    @click="addCalibrationFace"
                    :disabled="faces.length >= 3"
                    class="px-2 py-1 bg-primary text-white rounded text-sm hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                    {{ faces.length >= 3 ? 'Max faces reached' : 'Fetch Face Calibration' }}
                </button>
            </div>
            <table class="min-w-full border-collapse text-slate-100 text-xs">
                <thead>
                    <tr>
                        <th class="border border-slate-700 p-1">X</th>
                        <th class="border border-slate-700 p-1">Y</th>
                        <th class="border border-slate-700 p-1">Radii (CSV)</th>
                        <th class="border border-slate-700 p-1">Identifier</th>
                        <th class="border border-slate-700 p-1">Points (CSV)</th>
                        <th class="border border-slate-700 p-1"></th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="(f, idx) in faces" :key="idx">
                        <td class="border border-slate-700 p-1 text-xs text-slate-300">
                            {{ f.x }}
                        </td>
                        <td class="border border-slate-700 p-1 text-xs text-slate-300">
                            {{ f.y }}
                        </td>
                        <td class="border border-slate-700 p-1 text-xs text-slate-300">
                            {{ f.radii.join(',') }}
                        </td>
                        <td class="border border-slate-700 p-1">
                            <input
                                v-model="f.humanIdentifier"
                                type="text"
                                placeholder="face1"
                                class="w-32 bg-slate-900 border border-slate-600 p-1 rounded"
                            />
                        </td>
                        <td class="border border-slate-700 p-1">
                            <input
                                v-model="f.points"
                                type="text"
                                placeholder="10,9,8"
                                class="w-32 bg-slate-900 border border-slate-600 p-1 rounded"
                            />
                        </td>
                        <td class="border border-slate-700 p-1">
                            <button @click="removeFaceRow(idx)" class="text-danger text-xs">
                                Remove
                            </button>
                        </td>
                    </tr>
                    <tr v-if="faces.length === 0">
                        <td colspan="6" class="text-center text-slate-500 p-2">No faces added.</td>
                    </tr>
                </tbody>
            </table>
        </div>
        <div v-if="saving" class="mt-2 text-slate-400 text-sm">Saving...</div>
    </div>
</template>
