<script setup lang="ts">
    import { ref } from 'vue';
    import { createSession } from '../composables/useSession';
    const emit = defineEmits<{ submit: [] }>();

    const location = ref('club');
    const startTime = ref(new Date().toISOString().slice(0, 16)); // For input[type="datetime-local"]

    async function handleSubmit() {
        await createSession({
            location: location.value,
            start_time: new Date(startTime.value).toISOString(),
            is_opened: true,
        });
        emit('submit');
    }
</script>

<template>
    <form @submit.prevent="handleSubmit">
        <label>
            Location:
            <input type="text" v-model="location" required />
        </label>
        <label>
            Start Time:
            <input type="datetime-local" v-model="startTime" required />
        </label>
        <button type="submit">Start Session</button>
    </form>
</template>
