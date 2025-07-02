<script setup lang="ts">
    import { computed } from 'vue';
    import { uiManagerStore } from './state/uiManagerStore';
    import { sessionOpened } from './state/session';
    import type { ViewName } from './state/uiManagerStore';

    import NewArrow from './components/forms/NewArrow.vue';
    import WizardSession from './components/WizardSession.vue';
    import CalibrateTarget from './components/forms/CalibrateTarget.vue';
    import ShotsTable from './components/ShotsTable.vue';
    import { closeSession } from './composables/useSession';

    // Map view names to components
    const componentsMap: Record<ViewName, any> = {
        ArrowForm: NewArrow,
        SessionForm: WizardSession,
        TargetForm: CalibrateTarget,
        ShotsTable: ShotsTable,
    };

    // Compute the active component to render
    const currentComponent = computed(() => {
        const view = uiManagerStore.currentView;
        return view ? componentsMap[view] : null;
    });

    function handleSubmit() {
        uiManagerStore.clearView();
    }
    const buttonText = computed(() => {
        return sessionOpened.is_opened === true ? 'Close Session' : 'Open Session';
    });

    function handleClick() {
        if (sessionOpened.is_opened === true) {
            closeSession();
        } else {
            uiManagerStore.setView('SessionForm');
        }
    }
</script>

<template>
    <header>
        <button @click="handleClick()">{{ buttonText }}</button>
        <button @click="uiManagerStore.setView('ArrowForm')">Register Arrow</button>
        <button @click="uiManagerStore.setView('TargetForm')">Register Target</button>
        <button @click="uiManagerStore.setView('ShotsTable')">View Shots</button>
    </header>

    <main>
        <component :is="currentComponent" v-if="currentComponent" @submit="handleSubmit" />
        <p v-else>Open a session</p>
    </main>
</template>
