<script setup lang="ts">
    import { computed } from 'vue';
    import { uiManagerStore } from './state/uiManagerStore';
    // @ts-expect-error TS6133
    import { sessionOpened } from './state/session';
    import type { ViewName } from './state/uiManagerStore';

    // Map view names to components
    // @ts-ignore TS2739
    const componentsMap: Record<ViewName, any> = {};

    // Compute the active component to render
    const currentComponent = computed(() => {
        const view = uiManagerStore.currentView;
        return view ? componentsMap[view] : null;
    });

    function handleSubmit() {
        uiManagerStore.clearView();
    }
</script>

<template>
    <header></header>

    <main>
        <component :is="currentComponent" v-if="currentComponent" @submit="handleSubmit" />
        <p v-else>Open a session</p>
    </main>
</template>
