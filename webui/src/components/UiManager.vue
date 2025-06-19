<template>
    <component :is="currentComponent" v-if="currentComponent" @submit="handleSubmit" />
</template>

<script setup lang="ts">
    import { computed } from 'vue';
    import { uiManagerStore } from '../state/uiManagerStore';

    import SessionForm from './SessionForm.vue';
    import ShotsTable from './ShotsTable.vue';

    const viewMap: Record<string, any> = {
        SessionForm,
        ShotsTable,
    };

    const currentComponent = computed(() => {
        return viewMap[uiManagerStore.currentView] || null;
    });

    function handleSubmit() {
        uiManagerStore.setView('ShotsTable');
    }
</script>
