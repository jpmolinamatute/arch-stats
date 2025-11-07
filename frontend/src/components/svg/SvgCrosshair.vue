<script setup lang="ts">
    const props = withDefaults(
        defineProps<{
            size?: number; // half-length of crosshair lines in local units
            color?: string;
            strokeWidth?: number;
        }>(),
        {
            size: 8,
            color: '#555',
            strokeWidth: 1,
        },
    );
</script>

<template>
    <line
        :x1="-props.size"
        y1="0"
        :x2="props.size"
        y2="0"
        :stroke="props.color"
        :stroke-width="props.strokeWidth"
    />
    <line
        x1="0"
        :y1="-props.size"
        x2="0"
        :y2="props.size"
        :stroke="props.color"
        :stroke-width="props.strokeWidth"
    />
</template>
