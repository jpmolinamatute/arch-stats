<script setup lang="ts">
    import { openSession } from '../state/session';
    import { computed } from 'vue';
    import { closeSession } from '../composables/useSession';

    const emit = defineEmits<{ (e: 'requestOpenForm'): void }>();

    const buttonText = computed(() => {
        return openSession.is_opened === true ? 'Close Session' : 'Open Session';
    });

    function handleClick() {
        if (openSession.is_opened === true) {
            closeSession();
        } else {
            emit('requestOpenForm');
        }
    }
    console.log('SessionButton openSession:', openSession);
</script>

<template>
    <button @click="handleClick">
        {{ buttonText }}
    </button>
</template>
