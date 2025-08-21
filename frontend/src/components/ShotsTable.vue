<script setup lang="ts">
    import { onMounted, onUnmounted, ref } from 'vue';
    import { sessionOpened } from '../state/session';
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
        const scheme = window.location.protocol === 'https:' ? 'wss' : 'ws';
        const url = `${scheme}://${window.location.host}/api/v0/ws/shot`;
        socket = new WebSocket(url);
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
        if (sessionOpened.id) {
            fetchShots(sessionOpened.id);
            connectSocket();
        }
    });

    onUnmounted(() => {
        closeSocket();
    });
</script>

<template>
    <table
        v-if="sessionOpened.is_opened === true"
        class="w-full border-collapse text-sm shadow-sm rounded overflow-hidden"
    >
        <thead class="bg-slate-800 text-slate-200">
            <tr>
                <th class="px-2 py-1 border border-slate-700">ID</th>
                <th class="px-2 py-1 border border-slate-700">Arrow ID</th>
                <th class="px-2 py-1 border border-slate-700">Engage Time</th>
                <th class="px-2 py-1 border border-slate-700">Disengage Time</th>
                <th class="px-2 py-1 border border-slate-700">Landing Time</th>
                <th class="px-2 py-1 border border-slate-700">X</th>
                <th class="px-2 py-1 border border-slate-700">Y</th>
            </tr>
        </thead>
        <tbody>
            <tr
                v-for="shot in shots"
                :key="shot.id"
                class="even:bg-slate-800/40 hover:bg-slate-700/40 transition-colors"
            >
                <td class="px-2 py-1 border border-slate-800">{{ shot.id }}</td>
                <td class="px-2 py-1 border border-slate-800">{{ shot.arrow_id }}</td>
                <td class="px-2 py-1 border border-slate-800">{{ shot.arrow_engage_time }}</td>
                <td class="px-2 py-1 border border-slate-800">{{ shot.arrow_disengage_time }}</td>
                <td class="px-2 py-1 border border-slate-800">{{ shot.arrow_landing_time }}</td>
                <td class="px-2 py-1 border border-slate-800">{{ shot.x }}</td>
                <td class="px-2 py-1 border border-slate-800">{{ shot.y }}</td>
            </tr>
        </tbody>
    </table>
</template>

<style scoped>
    /* Additional scoped adjustments (if any) can go here */
</style>
