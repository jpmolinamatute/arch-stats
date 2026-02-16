<script setup lang="ts">
defineProps<{
    shots: { score: number, x: number, y: number, is_x: boolean, color: string }[]
    maxShots: number
}>()

const emit = defineEmits<{
    (e: 'delete', index: number): void
    (e: 'clear'): void
    (e: 'confirm'): void
}>()

function getShotTextColor(shot: { score: number, color: string }): string {
    return shot.color === '#000000' ? '#FFFFFF' : '#000000'
}
</script>

<template>
    <div class="w-fit mx-auto mb-4 bg-slate-900/50 rounded-lg p-1 border border-slate-800">
        <!-- Row 1: Numbers -->
        <div
            class="grid gap-1 mb-2 text-center"
            :style="{ gridTemplateColumns: `repeat(${maxShots + 2}, minmax(0, 1fr))` }"
        >
            <div
                v-for="i in maxShots"
                :key="`header-${i}`"
                class="text-lg text-slate-400 font-bold font-mono"
            >
                {{ i }}
            </div>
            <!-- Empty column for actions header -->
            <div />
        </div>

        <!-- Row 2: Shots + Actions -->
        <div
            class="grid gap-1"
            :style="{ gridTemplateColumns: `repeat(${maxShots + 2}, minmax(0, 1fr))` }"
        >
            <!-- Shot Slots -->
            <div
                v-for="i in maxShots"
                :key="`slot-${i}`"
                class="relative min-w-0 overflow-hidden w-full"
                style="aspect-ratio: 1 / 1;"
            >
                <button
                    v-if="shots[i - 1]"
                    class="w-full h-full rounded flex items-center justify-center font-bold text-lg leading-none shadow-sm transition-transform hover:scale-105 active:scale-95 border border-black/10 !p-0"
                    :style="{
                        backgroundColor: shots[i - 1]?.color,
                        backgroundImage: 'none',
                        color: getShotTextColor(shots[i - 1]!),
                    }"
                    :title="`Score: ${shots[i - 1]?.score}`"
                    @click="emit('delete', i - 1)"
                >
                    {{ shots[i - 1]?.is_x ? 'X' : (shots[i - 1]?.score === 0 ? 'M' : shots[i - 1]?.score) }}
                </button>
                <div
                    v-else
                    class="w-full h-full rounded bg-slate-800/50 border border-slate-700/50"
                />
            </div>

            <!-- Action Buttons (Clear/Confirm) -->
            <div class="relative min-w-0 overflow-hidden w-full" style="aspect-ratio: 1 / 1;">
                <button
                    class="w-full h-full min-w-0 text-[9px] !p-0 uppercase font-bold rounded border border-red-900/50 bg-red-900/20 text-red-400 hover:bg-red-900/40 disabled:opacity-30 disabled:cursor-not-allowed transition-colors flex items-center justify-center leading-none"
                    title="Clear field"
                    :disabled="shots.length === 0"
                    @click="emit('clear')"
                >
                    CLR
                </button>
            </div>
            <div class="relative min-w-0 overflow-hidden w-full" style="aspect-ratio: 1 / 1;">
                <button
                    class="w-full h-full min-w-0 text-[9px] !p-0 uppercase font-bold rounded border border-emerald-900/50 bg-emerald-900/20 text-emerald-400 hover:bg-emerald-900/40 disabled:opacity-30 disabled:cursor-not-allowed transition-colors flex items-center justify-center leading-none"
                    title="Confirm round"
                    :disabled="shots.length < maxShots"
                    @click="emit('confirm')"
                >
                    OK
                </button>
            </div>
        </div>
    </div>
</template>
