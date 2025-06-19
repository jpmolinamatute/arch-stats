<template>
    <SessionButton @requestOpenForm="uiManagerStore.setView('SessionForm')" />
    <button @click="uiManagerStore.setView('ShotsTable')">View Shots</button>
    <button @click="uiManagerStore.clearView()">Clear</button>
</template>

<script setup lang="ts">
    import { uiManagerStore } from '../state/uiManagerStore';
    import SessionButton from './SessionButton.vue';
</script>
