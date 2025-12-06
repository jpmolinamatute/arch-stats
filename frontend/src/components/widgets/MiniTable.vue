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
    return shot.color === '#FFFFFF' ? '#000000' : '#FFFFFF'
}
</script>

<template>
    <div class="w-fit mx-auto mb-4 bg-slate-900/50 rounded-lg p-3 border border-slate-800">
        <!-- Row 1: Numbers -->
        <div
            class="grid gap-2 mb-2 text-center"
            :style="{ gridTemplateColumns: `repeat(${maxShots + 1}, minmax(3.5rem, 1fr))` }"
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
            class="grid gap-2 h-14"
            :style="{ gridTemplateColumns: `repeat(${maxShots + 1}, minmax(3.5rem, 1fr))` }"
        >
            <!-- Shot Slots -->
            <div
                v-for="i in maxShots"
                :key="`slot-${i}`"
                class="relative"
            >
                <button
                    v-if="shots[i - 1]"
                    class="w-full h-full rounded flex items-center justify-center font-bold text-xl shadow-sm transition-transform hover:scale-105 active:scale-95 border border-black/10"
                    :style="{
                        backgroundColor: shots[i - 1]?.color,
                        backgroundImage: 'none',
                        color: getShotTextColor(shots[i - 1]!),
                    }"
                    :title="`Score: ${shots[i - 1]?.score}`"
                    @click="emit('delete', i - 1)"
                >
                    {{ shots[i - 1]?.is_x ? 'X' : shots[i - 1]?.score }}
                </button>
                <div
                    v-else
                    class="w-full h-full rounded bg-slate-800/50 border border-slate-700/50"
                />
            </div>

            <!-- Action Buttons (Clear/Confirm) -->
            <div class="flex flex-row gap-1">
                <button
                    class="flex-1 text-[10px] uppercase font-bold tracking-wider rounded border border-red-900/50 bg-red-900/20 text-red-400 hover:bg-red-900/40 disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
                    title="Clear field"
                    :disabled="shots.length === 0"
                    @click="emit('clear')"
                >
                    CLR
                </button>
                <button
                    class="flex-1 text-[10px] uppercase font-bold tracking-wider rounded border border-emerald-900/50 bg-emerald-900/20 text-emerald-400 hover:bg-emerald-900/40 disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
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
