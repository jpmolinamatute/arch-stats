<script setup lang="ts">
    import { reactive, computed } from 'vue';
    import { openSession } from '../../state/session';

    interface Face {
        center_x: number;
        center_y: number;
        radius: number[];
        human_identifier: string;
        points: number[];
    }

    const form = reactive({
        session_id: computed(() => openSession.id),
        max_x_coordinate: 0,
        max_y_coordinate: 0,
        faces: [] as Face[],
    });

    const errors = reactive({
        faces: [] as { human_identifier: string; points: string }[],
    });

    // Load initial data (placeholder for WebSocket data)
    function loadFromTargetReader() {
        form.max_x_coordinate = 500;
        form.max_y_coordinate = 400;
        form.faces = [
            {
                center_x: 123,
                center_y: 111,
                radius: [50, 60, 80, 90, 100],
                human_identifier: '',
                points: [],
            },
            {
                center_x: 200,
                center_y: 180,
                radius: [50, 60, 80, 90, 100],
                human_identifier: '',
                points: [],
            },
            {
                center_x: 300,
                center_y: 250,
                radius: [50, 60, 80, 90, 100],
                human_identifier: '',
                points: [],
            },
        ];
        errors.faces = form.faces.map(() => ({ human_identifier: '', points: '' }));
    }

    loadFromTargetReader();

    const isFormValid = computed(() => {
        let valid = true;
        errors.faces = form.faces.map(() => ({ human_identifier: '', points: '' }));

        form.faces.forEach((face, index) => {
            if (!face.human_identifier.trim()) {
                errors.faces[index].human_identifier = 'Identifier is required';
                valid = false;
            }

            if (face.points.length === 0) {
                errors.faces[index].points = 'At least one point value is required';
                valid = false;
            } else if (face.points.some((p) => isNaN(p) || p < 0)) {
                errors.faces[index].points = 'Points must be non-negative integers';
                valid = false;
            }
        });

        return valid;
    });

    async function handleSubmit() {
        if (!form.session_id) {
            alert('No open session. Please open a session before submitting.');
            return;
        }

        if (!isFormValid.value) {
            console.log('Validation failed:', JSON.stringify(errors, null, 2));
            return;
        }

        const payload = {
            session_id: form.session_id,
            max_x_coordinate: form.max_x_coordinate,
            max_y_coordinate: form.max_y_coordinate,
            faces: form.faces.map((face) => ({
                center_x: face.center_x,
                center_y: face.center_y,
                radius: face.radius,
                human_identifier: face.human_identifier.trim(),
                points: face.points,
            })),
        };

        try {
            const response = await fetch('/api/v0/target', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload),
            });

            if (!response.ok) {
                const data = await response.json();
                console.error('API error:', data);
                alert(`Submission failed: ${data.errors || 'Unknown error'}`);
                return;
            }

            alert('Target submitted successfully!');
            console.log('Submitted payload:', JSON.stringify(payload, null, 2));
        } catch (err) {
            console.error('Submission failed:', err);
            alert('Network or server error during submission.');
        }
    }
</script>

<template>
    <form
        v-if="openSession.is_opened"
        @submit.prevent="handleSubmit"
        class="max-w-md mx-auto p-4 bg-white rounded-lg shadow-md space-y-4"
    >
        <div>
            <label class="block mb-1 font-medium text-gray-700">Session ID</label>
            <input
                type="text"
                :value="form.session_id"
                readonly
                class="w-full border border-gray-300 rounded px-3 py-2 bg-gray-100 text-gray-500"
            />
        </div>

        <div>
            <label class="block mb-1 font-medium text-gray-700">Max X Coordinate</label>
            <input
                type="number"
                :value="form.max_x_coordinate"
                readonly
                class="w-full border border-gray-300 rounded px-3 py-2 bg-gray-100"
            />
        </div>

        <div>
            <label class="block mb-1 font-medium text-gray-700">Max Y Coordinate</label>
            <input
                type="number"
                :value="form.max_y_coordinate"
                readonly
                class="w-full border border-gray-300 rounded px-3 py-2 bg-gray-100"
            />
        </div>

        <div v-for="(face, index) in form.faces" :key="index" class="border p-3 rounded space-y-2">
            <h3 class="font-semibold text-gray-800">Face {{ index + 1 }}</h3>

            <div>
                <label class="block mb-1 text-gray-700">Center X</label>
                <input
                    type="number"
                    :value="face.center_x"
                    readonly
                    class="w-full border border-gray-300 rounded px-3 py-2 bg-gray-100"
                />
            </div>

            <div>
                <label class="block mb-1 text-gray-700">Center Y</label>
                <input
                    type="number"
                    :value="face.center_y"
                    readonly
                    class="w-full border border-gray-300 rounded px-3 py-2 bg-gray-100"
                />
            </div>

            <div>
                <label class="block mb-1 text-gray-700">Radius</label>
                <input
                    type="text"
                    :value="face.radius.join(', ')"
                    readonly
                    class="w-full border border-gray-300 rounded px-3 py-2 bg-gray-100"
                />
            </div>

            <div>
                <label class="block mb-1 text-gray-700">Human Identifier</label>
                <input
                    type="text"
                    v-model="face.human_identifier"
                    class="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
                <p v-if="errors.faces[index]?.human_identifier" class="text-red-600 text-sm mt-1">
                    {{ errors.faces[index].human_identifier }}
                </p>
            </div>

            <div>
                <label class="block mb-1 text-gray-700">Points (comma-separated)</label>
                <input
                    type="text"
                    :value="face.points.join(', ')"
                    @input="
                        (e) =>
                            (face.points = (e.target as HTMLInputElement).value
                                .split(',')
                                .map((s) => parseInt(s.trim()))
                                .filter((n) => !isNaN(n)))
                    "
                    class="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
                <p v-if="errors.faces[index]?.points" class="text-red-600 text-sm mt-1">
                    {{ errors.faces[index].points }}
                </p>
            </div>
        </div>

        <button
            type="submit"
            class="w-full bg-blue-600 text-white py-2 rounded hover:bg-blue-700 transition-colors"
        >
            Submit
        </button>
    </form>
</template>
