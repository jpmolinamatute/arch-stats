<script setup lang="ts">
    import { ref, computed, watchEffect } from 'vue';
    import { createSession } from '../../composables/useSession';
    import type { StepRegistration } from '../widgets/Wizard.vue';

    const props = defineProps<{ register: (opts: StepRegistration) => void }>();

    const location = ref('Club');
    const startTime = ref(new Date().toISOString().slice(0, 16));
    const errors = ref({ location: '', startTime: '' });

    const isValid = computed(() => {
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

    const onComplete = ref<() => Promise<{ success: boolean; error?: string }>>(async () => {
        if (!isValid.value) {
            return { success: false, error: 'Form is invalid' };
        }
        await createSession({
            location: location.value,
            start_time: new Date(startTime.value).toISOString(),
            is_opened: true,
        });
        return { success: true };
    });

    watchEffect(() => {
        props.register({ isValid, onComplete });
    });
</script>

<template>
    <div class="p-4">
        <label class="block mb-1 font-medium text-gray-700">Location</label>
        <input
            v-model="location"
            type="text"
            class="w-full border rounded p-2 mb-1 text-gray-900 bg-white"
        />
        <p v-if="errors.location" class="text-red-600 text-sm">{{ errors.location }}</p>

        <label class="block mb-1 mt-4 font-medium text-gray-700">Start Time</label>
        <input
            v-model="startTime"
            type="datetime-local"
            class="w-full border rounded p-2 mb-1 text-gray-900 bg-white"
        />
        <p v-if="errors.startTime" class="text-red-600 text-sm">{{ errors.startTime }}</p>
    </div>
</template>
