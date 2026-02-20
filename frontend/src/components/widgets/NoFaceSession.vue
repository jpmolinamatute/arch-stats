<script setup lang="ts">
import { ref } from 'vue'

defineProps<{
    totalShots: number
    loading: boolean
}>()

const emit = defineEmits<{
    (e: 'addShots', count: number): void
}>()

const shotsInput = ref<number | null>(null)

function handleSend() {
    if (shotsInput.value && shotsInput.value > 0) {
        emit('addShots', shotsInput.value)
        shotsInput.value = null
    }
}
</script>

<template>
    <div class="max-w-lg mx-auto bg-slate-900/50 rounded-lg border border-slate-800 p-6">
        <div class="grid grid-cols-2 gap-4 text-center">
            <!-- Headers -->
            <div class="text-sm font-semibold text-slate-400 uppercase tracking-wider pb-2 border-b border-slate-800">
                Shots
            </div>
            <div class="text-sm font-semibold text-slate-400 uppercase tracking-wider pb-2 border-b border-slate-800">
                Control
            </div>

            <!-- Data -->
            <div class="flex items-center justify-center">
                <span class="text-4xl font-bold text-slate-100">{{ totalShots }}</span>
            </div>

            <div class="flex items-center justify-center gap-2">
                <input
                    v-model.number="shotsInput"
                    type="number"
                    min="1"
                    placeholder="#"
                    class="w-20 bg-slate-800 border border-slate-700 rounded px-3 py-2 text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    @keyup.enter="handleSend"
                >
                <button
                    class="bg-indigo-600 hover:bg-indigo-700 text-white px-4 py-2 rounded transition-colors duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
                    :disabled="!shotsInput || shotsInput <= 0 || loading"
                    @click="handleSend"
                >
                    Send
                </button>
            </div>
        </div>
    </div>
</template>
