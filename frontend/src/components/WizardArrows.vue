<script setup lang="ts">
    import Wizard, { createStep, type Step } from './widgets/Wizard.vue';
    import ArrowStep from './forms/NewArrowWizardStep.vue';
    const emit = defineEmits<{ submit: [] }>();

    const { step } = createStep('Register Arrow', ArrowStep);
    const steps: Step[] = [step];

    function onDone() {
        // Keep App.vue contract: emit 'submit' so it clears the view
        emit('submit');
    }
</script>

<template>
    <Wizard :steps="steps" @done="onDone" />
</template>
