<script setup lang="ts">
    import { computed } from 'vue';
    import type { Component } from 'vue';
    import { uiManagerStore } from './state/uiManagerStore';
    import { sessionOpened } from './state/session';
    import type { ViewName } from './state/uiManagerStore';

    import WizardArrows from './components/WizardArrows.vue';
    import WizardSession from './components/WizardSession.vue';
    import ShotsTable from './components/ShotsTable.vue';
    import { closeSession } from './composables/useSession';

    // Map view names to components
    const componentsMap: Record<ViewName, Component> = {
        ArrowForm: WizardArrows,
        SessionForm: WizardSession,
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
    <div class="min-h-screen w-full bg-slate-900 text-slate-100 flex flex-col">
        <header
            class="w-full border-b border-slate-700 px-4 py-3 flex gap-2 flex-wrap bg-slate-800/60 backdrop-blur"
        >
            <button @click="handleClick()" class="">{{ buttonText }}</button>
            <button @click="uiManagerStore.setView('ArrowForm')">Register Arrow</button>
            <button @click="uiManagerStore.setView('ShotsTable')">View Shots</button>
        </header>

        <main class="flex-1 p-4">
            <component :is="currentComponent" v-if="currentComponent" @submit="handleSubmit" />
            <p v-else class="text-slate-400">Open a session</p>
        </main>
    </div>
</template>
