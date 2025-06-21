<script setup lang="ts">
    import { ref, computed } from 'vue';
    import { createSession } from '../../composables/useSession';

    const emit = defineEmits<{ submit: [] }>();

    const location = ref('club');
    const startTime = ref(new Date().toISOString().slice(0, 16));

    const errors = ref({
        location: '',
        startTime: '',
    });

    const isFormValid = computed(() => {
        let valid = true;
        errors.value.location = '';
        errors.value.startTime = '';

        if (!location.value.trim()) {
            errors.value.location = 'Location is required';
            valid = false;
        }

        if (!startTime.value) {
            errors.value.startTime = 'Start time is required';
            valid = false;
        } else {
            const inputDate = new Date(startTime.value);
            const now = new Date();
            if (isNaN(inputDate.getTime())) {
                errors.value.startTime = 'Start time is invalid';
                valid = false;
            } else if (inputDate < now) {
                errors.value.startTime = 'Start time cannot be in the past';
                valid = false;
            }
        }

        return valid;
    });

    async function handleSubmit() {
        if (!isFormValid.value) {
            console.log('Validation failed:', errors.value);
            return;
        }

        await createSession({
            location: location.value,
            start_time: new Date(startTime.value).toISOString(),
            is_opened: true,
        });

        emit('submit');
    }
</script>

<template>
    <form
        @submit.prevent="handleSubmit"
        class="max-w-md mx-auto p-4 bg-white rounded-lg shadow-md space-y-4"
    >
        <div>
            <label class="block mb-1 font-medium text-gray-700"> Location: </label>
            <input
                type="text"
                v-model="location"
                required
                class="w-full border border-gray-300 rounded px-3 py-2 text-gray-900 bg-white focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <p v-if="errors.location" class="text-red-600 text-sm mt-1">{{ errors.location }}</p>
        </div>

        <div>
            <label class="block mb-1 font-medium text-gray-700"> Start Time: </label>
            <input
                type="datetime-local"
                v-model="startTime"
                required
                class="w-full border border-gray-300 rounded px-3 py-2 text-gray-900 bg-white focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <p v-if="errors.startTime" class="text-red-600 text-sm mt-1">{{ errors.startTime }}</p>
        </div>

        <button
            type="submit"
            class="w-full bg-blue-600 text-white py-2 rounded hover:bg-blue-700 transition-colors"
        >
            Start Session
        </button>
    </form>
</template>
