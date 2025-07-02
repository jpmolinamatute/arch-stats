<script setup lang="ts">
    import Wizard, { createStep, type Step } from './widgets/Wizard.vue';
    import NewSession from './forms/NewSession.vue';
    import CalibrateTarget from './forms/CalibrateTarget.vue';

    // Instantiate each step via the shared factory
    const { step: sessionStep } = createStep('Open Session', NewSession);
    const { step: targetStep } = createStep('Calibrate Target', CalibrateTarget);

    // Now just pass these steps into the wizard
    const steps: Step[] = [sessionStep, targetStep];

    function wizardCompleted() {
        console.log('Wizard flow complete!');
        // You can reset state or navigate elsewhere if needed
    }
</script>

<template>
    <Wizard :steps="steps" @done="wizardCompleted" />
</template>
