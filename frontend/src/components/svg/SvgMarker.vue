<script setup lang="ts">
    const props = withDefaults(
        defineProps<{
            x: number;
            y: number;
            r?: number;
            fill?: string;
            stroke?: string;
            strokeWidth?: number;
        }>(),
        {
            r: 4,
            fill: '#111',
            stroke: '#fff',
            strokeWidth: 1,
        },
    );
</script>

<template>
    <circle
        :cx="props.x"
        :cy="props.y"
        :r="props.r"
        :fill="props.fill"
        :stroke="props.stroke"
        :stroke-width="props.strokeWidth"
    />
</template>
