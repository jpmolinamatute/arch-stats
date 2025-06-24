<script setup lang="ts">
    import { ref, watch } from 'vue';
    import { type Ref } from 'vue';
    import { createTarget } from '../../composables/useTarget';
    import { sessionOpened } from '../../state/session';

    const props = defineProps<{
        register: (options: {
            isValid: Ref<boolean>;
            onComplete: () => Promise<{ success: boolean; error?: string }>;
        }) => void;
    }>();

    const maxX = ref(0);
    const maxY = ref(0);
    const faces = ref<
        {
            x: number;
            y: number;
            radius: number[];
            human_identifier: string;
            points: number[];
        }[]
    >([]);
    const isValid = ref(false);
    const errors = ref<{ maxX?: string; maxY?: string; faces?: string }>({});

    watch([maxX, maxY, faces], () => {
        errors.value = {};
        let valid = true;

        if (maxX.value <= 0) {
            errors.value.maxX = 'Max X must be greater than 0';
            valid = false;
        }

        if (maxY.value <= 0) {
            errors.value.maxY = 'Max Y must be greater than 0';
            valid = false;
        }

        if (faces.value.length === 0) {
            errors.value.faces = 'At least one face is required';
            valid = false;
        }

        isValid.value = valid;
    });

    function calibrate() {
        maxX.value = Math.floor(Math.random() * 1000);
        maxY.value = Math.floor(Math.random() * 1000);
        faces.value = [
            {
                x: Math.floor(Math.random() * maxX.value),
                y: Math.floor(Math.random() * maxY.value),
                radius: [50, 100, 150],
                human_identifier: '',
                points: [0, 0, 0],
            },
        ];
    }

    async function onComplete() {
        if (!isValid.value) {
            return { success: false, error: 'Please correct form errors' };
        }
        if (!sessionOpened.id) {
            return { success: false, error: 'No session ID available' };
        }
        try {
            await createTarget({
                session_id: sessionOpened.id,
                max_x: maxX.value,
                max_y: maxY.value,
                faces: faces.value,
            });
            return { success: true };
        } catch (err) {
            return { success: false, error: 'Failed to create target' };
        }
    }

    function updatePoints(index: number, value: string) {
        faces.value[index].points = value
            .split(',')
            .map((p) => parseInt(p.trim()))
            .filter((p) => !isNaN(p));
    }
    props.register({ isValid, onComplete });
</script>

<template>
    <div>
        <button @click="calibrate" class="bg-blue-500 text-white px-2 py-1 rounded mb-2">
            Calibrate
        </button>

        <label class="block mb-1">Max X</label>
        <input v-model="maxX" type="number" class="w-full border rounded p-2 mb-1" />
        <p v-if="errors.maxX" class="text-red-600 text-sm">{{ errors.maxX }}</p>

        <label class="block mb-1 mt-2">Max Y</label>
        <input v-model="maxY" type="number" class="w-full border rounded p-2 mb-1" />
        <p v-if="errors.maxY" class="text-red-600 text-sm">{{ errors.maxY }}</p>

        <table class="w-full border-collapse border border-gray-300 mt-2">
            <thead>
                <tr class="bg-gray-100">
                    <th class="border border-gray-300 p-2">#</th>
                    <th class="border border-gray-300 p-2">X</th>
                    <th class="border border-gray-300 p-2">Y</th>
                    <th class="border border-gray-300 p-2">Radius</th>
                    <th class="border border-gray-300 p-2">Human Identifier</th>
                    <th class="border border-gray-300 p-2">Points</th>
                </tr>
            </thead>
            <tbody>
                <tr v-for="(face, index) in faces" :key="index">
                    <td class="border border-gray-300 p-2 text-center">{{ index + 1 }}</td>
                    <td class="border border-gray-300 p-2 text-center">{{ face.x }}</td>
                    <td class="border border-gray-300 p-2 text-center">{{ face.y }}</td>
                    <td class="border border-gray-300 p-2 text-center">
                        {{ face.radius.join(', ') }}
                    </td>

                    <td class="border border-gray-300 p-2">
                        <input
                            v-model="face.human_identifier"
                            type="text"
                            class="w-full border rounded p-1"
                        />
                    </td>

                    <td class="border border-gray-300 p-2">
                        <input
                            :value="face.points.join(',')"
                            @input="
                                (e) => updatePoints(index, (e.target as HTMLInputElement).value)
                            "
                            type="text"
                            class="w-full border rounded p-1"
                        />
                    </td>
                </tr>
            </tbody>
        </table>

        <p v-if="errors.faces" class="text-red-600 text-sm mt-1">{{ errors.faces }}</p>
    </div>
</template>
