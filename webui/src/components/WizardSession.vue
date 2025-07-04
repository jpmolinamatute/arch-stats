<script setup lang="ts">
    import Wizard, { createStep, type Step } from './widgets/Wizard.vue';
    import NewSession from './forms/NewSession.vue';
    import CalibrateTarget from './forms/CalibrateTarget.vue';

    const { step: sessionStep } = createStep('Open Session', NewSession);
    const { step: targetStep } = createStep('Calibrate Target', CalibrateTarget);
    const steps: Step[] = [sessionStep, targetStep];

    function wizardCompleted() {
        console.log('Wizard flow complete!');
    }
</script>

<template>
    <Wizard :steps="steps" @done="wizardCompleted" />
</template>
