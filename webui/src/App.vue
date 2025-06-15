<script setup lang="ts">
    import { ref, watch } from 'vue';
    import { openSession } from './state/session';
    import SessionButton from './components/SessionButton.vue';
    import SessionForm from './components/SessionForm.vue';
    import ShotsTable from './components/ShotsTable.vue';

    const showForm = ref(false);
    function handleFormSubmit() {
        showForm.value = false;
    }

    watch(
        () => openSession.id,
        (id) => {
            if (id) {
                showForm.value = false;
            }
        },
    );
</script>

<template>
    <h1>Arch Stats App</h1>
    <nav>
        <SessionButton @requestOpenForm="showForm = true" />
    </nav>
    <main>
        <SessionForm v-if="showForm" @submit="handleFormSubmit" />
        <ShotsTable v-if="openSession.id" />
    </main>
</template>
