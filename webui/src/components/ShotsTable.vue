<script setup lang="ts">
    import { onMounted, onUnmounted, ref } from 'vue';
    import { openSession } from '../state/session';
    import type { components } from '../types/types.generated';

    type ShotsRead = components['schemas']['ShotsRead'];

    const shots = ref<ShotsRead[]>([]);
    let socket: WebSocket | null = null;

    // Fetch existing shots
    async function fetchShots(sessionId: string) {
        try {
            const response = await fetch(`/api/v0/shot?session_id=${sessionId}`);
            const json = await response.json();
            if (response.ok && json.data) {
                shots.value = json.data as ShotsRead[];
            } else {
                shots.value = [];
            }
        } catch (err) {
            console.error('Failed to fetch shots:', err);
            shots.value = [];
        }
    }

    // Connect websocket
    function connectSocket() {
        console.debug('Starting WS connection');
        socket = new WebSocket(`ws://${window.location.host}/api/v0/ws/shot`, 'shooting');
        socket.onmessage = (event) => {
            const data: ShotsRead = JSON.parse(event.data);
            shots.value.push(data);
        };
        socket.onclose = () => {
            console.log('Shot WebSocket closed');
        };
        socket.onerror = (err) => {
            console.error('Shot WebSocket error:', err);
        };
    }

    // Cleanup websocket
    function closeSocket() {
        console.debug('Closing WS connection');
        if (socket) {
            socket.close();
            socket = null;
        }
    }

    onMounted(() => {
        if (openSession.id) {
            fetchShots(openSession.id);
            connectSocket();
        }
    });

    onUnmounted(() => {
        closeSocket();
    });
</script>

<template>
    <table v-if="openSession.is_opened === true" class="shot-table">
        <thead>
            <tr>
                <th>ID</th>
                <th>Arrow ID</th>
                <th>Engage Time</th>
                <th>Disengage Time</th>
                <th>Landing Time</th>
                <th>X</th>
                <th>Y</th>
            </tr>
        </thead>
        <tbody>
            <tr v-for="shot in shots" :key="shot.id">
                <td>{{ shot.id }}</td>
                <td>{{ shot.arrow_id }}</td>
                <td>{{ shot.arrow_engage_time }}</td>
                <td>{{ shot.arrow_disengage_time }}</td>
                <td>{{ shot.arrow_landing_time }}</td>
                <td>{{ shot.x_coordinate }}</td>
                <td>{{ shot.y_coordinate }}</td>
            </tr>
        </tbody>
    </table>
</template>

<style scoped>
    .shot-table {
        width: 100%;
        border-collapse: collapse;
    }
    .shot-table th,
    .shot-table td {
        border: 1px solid #ccc;
        padding: 0.5rem;
        text-align: center;
    }
</style>
