<script setup lang="ts">
import type { components } from '@/types/types.generated'

type LiveStat = components['schemas']['LiveStat']

defineProps<{
    stats: LiveStat | null
    xCount: number
}>()
</script>

<template>
    <div class="w-full flex flex-col gap-4 mt-6">
        <h3 class="text-lg font-semibold text-slate-200">
            Stats Table
        </h3>

        <div class="overflow-hidden rounded-lg border border-slate-800 bg-slate-900/50">
            <table class="w-full text-sm text-left text-slate-400">
                <thead class="text-xs uppercase bg-slate-800 text-slate-300">
                    <tr>
                        <th scope="col" class="px-4 py-3 text-center">
                            Mean
                        </th>
                        <th scope="col" class="px-4 py-3 text-center">
                            Max Score
                        </th>
                        <th scope="col" class="px-4 py-3 text-center">
                            # Shots
                        </th>
                        <th scope="col" class="px-4 py-3 text-center">
                            X
                        </th>
                        <th scope="col" class="px-4 py-3 text-center font-bold text-slate-100 bg-slate-800">
                            Total
                        </th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-if="!stats">
                        <td colspan="5" class="px-4 py-8 text-center text-slate-500 italic">
                            No stats available.
                        </td>
                    </tr>
                    <tr
                        v-else
                        class="border-b border-slate-800/50 last:border-0 hover:bg-slate-800/30 transition-colors"
                    >
                        <!-- Mean -->
                        <td class="px-4 py-3 text-center text-lg text-white bg-slate-800/50">
                            {{ stats.mean.toFixed(2) }}
                        </td>

                        <!-- Max Score -->
                        <td class="px-4 py-3 text-center text-lg text-white">
                            {{ stats.max_score }}
                        </td>

                        <!-- Number of Shots -->
                        <td class="px-4 py-3 text-center text-lg text-white">
                            {{ stats.number_of_shots }}
                        </td>

                        <!-- X Count -->
                        <td class="px-4 py-3 text-center text-lg text-white text-yellow-500">
                            {{ xCount }}
                        </td>

                        <!-- Total Score -->
                        <td class="px-4 py-3 text-center text-xl font-bold text-emerald-400 bg-slate-800/50">
                            {{ stats.total_score }}
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>
