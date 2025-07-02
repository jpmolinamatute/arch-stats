<script setup lang="ts">
    import { ref, watch } from 'vue';
    import { type Ref } from 'vue';
    import { createSession } from '../../composables/useSession';
    const props = defineProps<{
        register: (options: {
            isValid: Ref<boolean>;
            onComplete: () => Promise<{ success: boolean; error?: string }>;
        }) => void;
    }>();

    const location = ref('club');
    const startTime = ref(new Date().toISOString().slice(0, 16));
    const isValid = ref(false);
    const errors = ref<{ location?: string; startTime?: string }>({});
    watch([location, startTime], () => {
        errors.value = {};
        let valid = true;

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
            } else if (inputDate > now) {
                errors.value.startTime = 'Start time cannot be in the future';
                valid = false;
            }
        }

        isValid.value = valid;
    });

    async function onComplete() {
        if (!isValid.value) {
            return { success: false, error: 'Session form is invalid' };
        }
        try {
            await createSession({
                location: location.value.trim(),
                start_time: new Date(startTime.value).toISOString(),
                is_opened: true,
            });
            return { success: true };
        } catch (err) {
            return { success: false, error: 'Failed to create session' };
        }
    }

    props.register({ isValid, onComplete });
</script>

<template>
    <div>
        <label class="block mb-1 font-medium text-gray-700">Location</label>
        <input
            v-model="location"
            type="text"
            class="w-full border rounded p-2 mb-1 text-gray-900 bg-white"
        />
        <p v-if="errors.location" class="text-red-600 text-sm">{{ errors.location }}</p>

        <label class="block mb-1 mt-2 font-medium text-gray-700">Start Time</label>
        <input
            v-model="startTime"
            type="datetime-local"
            class="w-full border rounded p-2 mb-1 text-gray-900 bg-white"
        />
        <p v-if="errors.startTime" class="text-red-600 text-sm">{{ errors.startTime }}</p>
    </div>
</template>
