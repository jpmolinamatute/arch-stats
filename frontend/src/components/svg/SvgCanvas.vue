<script setup lang="ts">
    import { ref } from 'vue';

    const props = withDefaults(
        defineProps<{
            widthPx: number;
            heightPx: number;
            backgroundColor?: string;
            viewBox?: string; // optional custom viewBox
        }>(),
        {
            backgroundColor: '#f5f5f5',
        },
    );

    const emit = defineEmits<{
        (e: 'pointerdown', evt: PointerEvent): void;
        (e: 'pointermove', evt: PointerEvent): void;
        (e: 'pointerup', evt: PointerEvent): void;
        (e: 'pointercancel', evt: PointerEvent): void;
        (e: 'wheel', evt: WheelEvent): void;
    }>();

    const svgRef = ref<SVGSVGElement | null>(null);

    function getBounds(): DOMRect {
        const el = svgRef.value;
        if (!el) return new DOMRect(0, 0, props.widthPx, props.heightPx);
        return el.getBoundingClientRect();
    }

    defineExpose({ getBounds, svgRef });
</script>

<template>
    <svg
        ref="svgRef"
        :width="props.widthPx"
        :height="props.heightPx"
        :viewBox="props.viewBox ?? `0 0 ${props.widthPx} ${props.heightPx}`"
        class="select-none"
        style="touch-action: none"
        @pointerdown="(e) => emit('pointerdown', e)"
        @pointermove="(e) => emit('pointermove', e)"
        @pointerup="(e) => emit('pointerup', e)"
        @pointercancel="(e) => emit('pointercancel', e)"
        @wheel.prevent="(e) => emit('wheel', e)"
    >
        <!-- Background full area rectangle -->
        <rect :width="props.widthPx" :height="props.heightPx" :fill="props.backgroundColor" />

        <!-- Content -->
        <slot />
    </svg>
</template>
