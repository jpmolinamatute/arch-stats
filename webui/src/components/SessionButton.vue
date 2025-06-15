<script setup lang="ts">
    import { openSession } from '../state/session';
    import { computed } from 'vue';
    import { closeSession } from '../composables/useSession';

    const emit = defineEmits<{ (e: 'requestOpenForm'): void }>();

    const buttonText = computed(() => {
        return openSession.id ? 'Close Session' : 'Open Session';
    });

    function handleClick() {
        if (openSession.id) {
            closeSession();
        } else {
            emit('requestOpenForm');
        }
    }
</script>

<template>
    <button @click="handleClick">
        {{ buttonText }}
    </button>
</template>
