<script setup lang="ts">
    import { ref, computed, watchEffect } from 'vue';
    import type { StepRegistration } from '../widgets/Wizard.vue';
    import { createTarget } from '../../composables/useTarget';
    import type { components } from '../../types/types.generated';
    import { sessionOpened } from '../../state/session';

    const props = defineProps<{ register: (options: StepRegistration) => void }>();
    interface FaceInput {
        x: number;
        y: number;
        radii: number[];
        humanIdentifier: string;
        points: string;
    }

    const success = ref(false);
    const maxX = ref<number>(0);
    const maxY = ref<number>(0);
    const faces = ref<FaceInput[]>([]);
    const errorMessage = ref<string>('');
    const isValid = computed(() => {
        if (!success.value) return false;
        if (faces.value.length === 0) return true;
        return faces.value.every((f) => {
            if (!f.humanIdentifier.trim()) {
                return false;
            }

            const pts = f.points.trim();
            if (pts) {
                const parts = pts.split(',').map((p) => p.trim());
                if (parts.length !== f.radii.length) {
                    return false;
                }
                return parts.every((p) => !isNaN(Number(p)));
            }

            return true;
        });
    });
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
            faces: faces.value.map((f) => ({
                x: f.x,
                y: f.y,
                radii: f.radii,
                human_identifier: f.humanIdentifier,
                points: f.points.split(',').map((p) => Number(p.trim())),
            })),
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

    function calibrate() {
        // The calibrate() function is a dummy function for now. Think of it as a placeholder for
        // future work.
        // The function will make a GET HTTP request, and the server will return:

        // * max_x
        // * max_y
        // * faces

        // faces will be optional; if the server returns one or more faces, then it will include:

        // * x
        // * y
        // * radii[]

        // And then the user must add a human_readable name per face and optionally add points.
        // If the user provides points, the points array length must match the radii array length.
        // The user can click as many times as they want the calibrate button until they are happy
        // with the values that the server returns. This means that the number of faces (and
        // number of rows as well) may change, so the validation must also be reactive.
        success.value = false;
        errorMessage.value = '';
        faces.value = [];

        if (Math.random() < 0.2) {
            errorMessage.value = 'Calibration failed. Please try again.';
            return;
        }

        maxX.value = Math.floor(Math.random() * 201) + 100;
        maxY.value = Math.floor(Math.random() * 201) + 100;

        const count = Math.floor(Math.random() * 6);
        const newFaces: FaceInput[] = [];
        for (let i = 0; i < count; i++) {
            const x = Math.floor(Math.random() * maxX.value);
            const y = Math.floor(Math.random() * maxY.value);
            const radii = Array.from({ length: 3 }, () => Math.floor(Math.random() * 41) + 10);
            newFaces.push({ x, y, radii, humanIdentifier: '', points: '' });
        }
        faces.value = newFaces;
        success.value = true;
    }
</script>

<template>
    <div class="p-4 bg-slate-800/60 rounded border border-slate-700">
        <button
            @click="calibrate"
            class="px-4 py-2 bg-primary text-white rounded hover:bg-primary/90"
        >
            Calibrate
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
            </div>

            <table class="min-w-full border-collapse text-slate-100 text-sm">
                <thead>
                    <tr>
                        <th class="border border-slate-700 p-2 text-left bg-slate-800/70">X</th>
                        <th class="border border-slate-700 p-2 text-left bg-slate-800/70">Y</th>
                        <th class="border border-slate-700 p-2 text-left bg-slate-800/70">Radii</th>
                        <th class="border border-slate-700 p-2 text-left bg-slate-800/70">
                            Human Identifier
                        </th>
                        <th class="border border-slate-700 p-2 text-left bg-slate-800/70">
                            Points (CSV)
                        </th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="(face, index) in faces" :key="index">
                        <td class="border border-slate-700 p-2">
                            <input
                                type="number"
                                :value="face.x"
                                readonly
                                class="bg-slate-700 border border-slate-600 p-2 rounded w-full text-slate-100"
                            />
                        </td>
                        <td class="border border-slate-700 p-2">
                            <input
                                type="number"
                                :value="face.y"
                                readonly
                                class="bg-slate-700 border border-slate-600 p-2 rounded w-full text-slate-100"
                            />
                        </td>
                        <td class="border border-slate-700 p-2">
                            <input
                                type="text"
                                :value="face.radii.join(',')"
                                readonly
                                class="bg-slate-700 border border-slate-600 p-2 rounded w-full text-slate-100"
                            />
                        </td>
                        <td class="border border-slate-700 p-2">
                            <input
                                type="text"
                                v-model="face.humanIdentifier"
                                placeholder="e.g. face1"
                                class="border border-slate-600 bg-slate-900 p-2 rounded w-full text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-600"
                            />
                        </td>
                        <td class="border border-slate-700 p-2">
                            <input
                                type="text"
                                v-model="face.points"
                                placeholder="10,9,8"
                                class="border border-slate-600 bg-slate-900 p-2 rounded w-full text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-600"
                            />
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>
