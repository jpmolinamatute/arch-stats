<script setup lang="ts">
    import { ref } from 'vue';
    import Wizard, { createStep, type Step } from './widgets/Wizard.vue';
    import NewSession from './forms/NewSession.vue';
    import CalibrateTarget from './forms/CalibrateTarget.vue';
    import FaceSetup from './forms/FaceSetup.vue';
    import { uiManagerStore } from '../state/uiManagerStore';
    import { sessionOpened } from '../state/session';

    const { step: sessionStep } = createStep('Open Session', NewSession);
    const { step: targetStep } = createStep('Calibrate Target', CalibrateTarget);
    const { step: facesStep } = createStep('Add Faces (Optional)', FaceSetup);

    // Explicitly control Next/Finish labels via hasNext on each step
    const hasNextSession = ref(true); // first -> Next
    const hasNextTarget = ref(true); // target now mandatory -> Next to faces
    const hasNextFaces = ref(false); // final -> Finish

    const origRegisterSession = sessionStep.register;
    sessionStep.register = (opts) => origRegisterSession({ ...opts, hasNext: hasNextSession });

    const origRegisterTarget = targetStep.register;
    targetStep.register = (opts) => origRegisterTarget({ ...opts, hasNext: hasNextTarget });
    const origRegisterFaces = facesStep.register;
    facesStep.register = (opts) => origRegisterFaces({ ...opts, hasNext: hasNextFaces });
    const steps: Step[] = [sessionStep, targetStep, facesStep];

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
