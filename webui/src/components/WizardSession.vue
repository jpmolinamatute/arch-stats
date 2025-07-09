<script setup lang="ts">
    import Wizard, { createStep, type Step } from './widgets/Wizard.vue';
    import NewSession from './forms/NewSession.vue';
    import CalibrateTarget from './forms/CalibrateTarget.vue';
    import { uiManagerStore } from '../state/uiManagerStore';
    import { sessionOpened } from '../state/session';

    const { step: sessionStep } = createStep('Open Session', NewSession);
    const { step: targetStep } = createStep('Calibrate Target', CalibrateTarget);
    const steps: Step[] = [sessionStep, targetStep];

    function wizardCompleted() {
        if (sessionOpened.id) {
            uiManagerStore.setView('ShotsTable');
        } else {
            uiManagerStore.clearView();
        }
    }
</script>

<template>
    <Wizard :steps="steps" @done="wizardCompleted" />
</template>
