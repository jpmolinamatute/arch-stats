<script setup lang="ts">
    import { ref } from 'vue';
    import NewSession from './forms/NewSession.vue';
    import CalibrateTarget from './forms/CalibrateTarget.vue';
    import Wizard from './widgets/Wizard.vue';

    const sessionIsValid = ref(false);
    const sessionOnComplete = ref<() => Promise<{ success: boolean; error?: string }>>(
        async () => ({ success: false }),
    );
    const targetIsValid = ref(false);
    const targetOnComplete = ref<() => Promise<{ success: boolean; error?: string }>>(async () => ({
        success: false,
    }));
    const steps = [
        {
            name: 'Open Session',
            component: NewSession,
            isValid: sessionIsValid,
            onComplete: sessionOnComplete,
        },
        {
            name: 'Calibrate Target',
            component: CalibrateTarget,
            isValid: targetIsValid,
            onComplete: targetOnComplete,
        },
    ];

    function wizardCompleted() {
        console.log('Wizard flow complete!');
        // You can reset state or navigate elsewhere if needed
    }
</script>

<template>
    <Wizard :steps="steps" @done="wizardCompleted" />
</template>
