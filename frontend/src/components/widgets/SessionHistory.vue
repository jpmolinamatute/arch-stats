<script setup lang="ts">
import type { components } from '@/types/types.generated'
import { computed } from 'vue'

type ShotRead = components['schemas']['ShotRead']

const props = defineProps<{
    shots: ShotRead[]
}>()

// Sort shots by creation time (newest first) or by arrow/end?
// Usually history is shown chronologically. Let's show newest first.
const sortedShots = computed(() => {
    return [...props.shots].sort((a, b) => {
        return new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
    })
})

function formatDate(dateString: string) {
    return new Date(dateString).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

function getScoreColor(score: number | null): string {
    if (score === null)
        return 'text-slate-500'
    if (score >= 9)
        return 'text-yellow-400'
    if (score >= 7)
        return 'text-red-400'
    if (score >= 5)
        return 'text-blue-400'
    if (score >= 3)
        return 'text-black bg-white/10 px-1 rounded' // Black on dark theme needs background
    return 'text-white'
}
</script>

<template>
    <div class="w-full flex flex-col gap-4">
        <h3 class="text-lg font-semibold text-slate-200">
            Session History
        </h3>

        <div class="overflow-hidden rounded-lg border border-slate-800 bg-slate-900/50">
            <table class="w-full text-sm text-left text-slate-400">
                <thead class="text-xs uppercase bg-slate-800 text-slate-300">
                    <tr>
                        <th scope="col" class="px-4 py-3">
                            Time
                        </th>
                        <th scope="col" class="px-4 py-3 text-center">
                            Score
                        </th>
                        <th scope="col" class="px-4 py-3 text-center">
                            X
                        </th>
                        <th scope="col" class="px-4 py-3 text-right">
                            Coords
                        </th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-if="shots.length === 0">
                        <td colspan="4" class="px-4 py-8 text-center text-slate-500 italic">
                            No confirmed shots yet.
                        </td>
                    </tr>
                    <tr
                        v-for="shot in sortedShots"
                        :key="shot.shot_id"
                        class="border-b border-slate-800/50 last:border-0 hover:bg-slate-800/30 transition-colors"
                    >
                        <td class="px-4 py-3 font-mono text-xs">
                            {{ formatDate(shot.created_at) }}
                        </td>
                        <td class="px-4 py-3 text-center font-bold text-lg">
                            <span :class="getScoreColor(shot.score ?? null)">
                                {{ shot.score === 0 ? 'M' : (shot.score ?? '-') }}
                            </span>
                        </td>
                        <td class="px-4 py-3 text-center font-bold">
                            <span v-if="shot.is_x" class="text-yellow-400">X</span>
                            <span v-else class="text-slate-700">-</span>
                        </td>
                        <td class="px-4 py-3 text-right font-mono text-xs text-slate-500">
                            <span v-if="shot.x !== undefined && shot.y !== undefined && shot.x !== null && shot.y !== null">
                                ({{ shot.x.toFixed(1) }}, {{ shot.y.toFixed(1) }})
                            </span>
                            <span v-else>-</span>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>
