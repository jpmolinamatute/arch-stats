<script setup lang="ts">
import type { components } from '@/types/types.generated'
import { computed } from 'vue'

type ShotRead = components['schemas']['ShotRead']

const props = withDefaults(defineProps<{
    shots: ShotRead[]
    shotPerRound?: number
}>(), {
    shotPerRound: 6,
})

interface RoundData {
    roundNumber: number
    shots: (ShotRead | null)[]
    total: number
    cumulative: number
}

const groupedRounds = computed<RoundData[]>(() => {
    // Sort shots oldest to newest to calculate cumulative score correctly
    const sorted = [...props.shots].sort((a, b) => {
        return new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
    })

    const rounds: RoundData[] = []
    let currentRoundShots: (ShotRead | null)[] = []
    let currentRoundTotal = 0
    let runningCumulative = 0
    let roundIndex = 1

    for (const shot of sorted) {
        currentRoundShots.push(shot)
        currentRoundTotal += (shot.score ?? 0)

        // If round is full, push it
        if (currentRoundShots.length === props.shotPerRound) {
            runningCumulative += currentRoundTotal
            rounds.push({
                roundNumber: roundIndex++,
                shots: currentRoundShots,
                total: currentRoundTotal,
                cumulative: runningCumulative,
            })
            currentRoundShots = []
            currentRoundTotal = 0
        }
    }

    // Handle partial last round
    if (currentRoundShots.length > 0) {
        runningCumulative += currentRoundTotal
        // Fill remaining slots with null to keep layout consistent
        while (currentRoundShots.length < props.shotPerRound) {
            currentRoundShots.push(null)
        }
        rounds.push({
            roundNumber: roundIndex,
            shots: currentRoundShots,
            total: currentRoundTotal,
            cumulative: runningCumulative,
        })
    }

    // Return newest rounds first for display
    return rounds.reverse()
})

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
                        <th scope="col" class="px-2 py-3 text-center text-slate-500 w-10">
                            #
                        </th>
                        <!-- Dynamic Columns 1..N -->
                        <th
                            v-for="n in shotPerRound"
                            :key="n"
                            scope="col"
                            class="px-2 py-3 text-center w-12"
                        >
                            {{ n }}
                        </th>
                        <th scope="col" class="px-4 py-3 text-center w-20 bg-slate-800/80">
                            Total
                        </th>
                        <th scope="col" class="px-4 py-3 text-center w-20 font-bold text-slate-100 bg-slate-800">
                            Sum
                        </th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-if="shots.length === 0">
                        <td :colspan="shotPerRound + 3" class="px-4 py-8 text-center text-slate-500 italic">
                            No confirmed shots yet.
                        </td>
                    </tr>
                    <tr
                        v-for="round in groupedRounds"
                        :key="round.roundNumber"
                        class="border-b border-slate-800/50 last:border-0 hover:bg-slate-800/30 transition-colors"
                    >
                        <!-- Round Number -->
                        <td class="px-2 py-3 text-center text-slate-600 font-mono text-xs">
                            {{ round.roundNumber }}
                        </td>

                        <!-- Shot Scores -->
                        <td
                            v-for="(shot, index) in round.shots"
                            :key="index"
                            class="px-2 py-3 text-center font-bold text-lg"
                        >
                            <span v-if="shot" :class="getScoreColor(shot.score ?? 0)">
                                {{ shot.is_x ? 'X' : shot.score }}
                            </span>
                            <span v-else class="text-slate-700">-</span>
                        </td>

                        <!-- Round Total -->
                        <td class="px-4 py-3 text-center font-mono text-slate-300 bg-slate-800/30">
                            {{ round.total }}
                        </td>

                        <!-- Cumulative Total -->
                        <td class="px-4 py-3 text-center font-bold text-lg text-white bg-slate-800/50">
                            {{ round.cumulative }}
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>
