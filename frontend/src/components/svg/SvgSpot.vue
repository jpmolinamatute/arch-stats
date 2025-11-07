<script setup lang="ts">
    import type { components } from '@/types/types.generated';

    // Types from generated API
    type Face = components['schemas']['Face'];
    export type FaceRing = Face['rings'][number];

    const props = defineProps<{
        xOffsetMm: number;
        yOffsetMm: number;
        ringsDesc: ReadonlyArray<FaceRing>; // rings sorted by radius desc (outermost first)
        pxPerMm: number;
        lineThicknessScale: number;
    }>();
</script>

<template>
    <g
        :transform="`translate(${props.xOffsetMm * props.pxPerMm}, ${-props.yOffsetMm * props.pxPerMm})`"
    >
        <!-- Filled bands: outermost -> innermost -->
        <template v-for="r in props.ringsDesc" :key="`ring-fill-${r.scoring_zone_radius}`">
            <circle :r="r.scoring_zone_radius * props.pxPerMm" :fill="r.scoring_zone_color" />
        </template>

        <!-- Boundary lines: draw at each scoring zone outer edge -->
        <template v-for="r in props.ringsDesc" :key="`ring-line-${r.scoring_zone_radius}`">
            <circle
                :r="r.scoring_zone_radius * props.pxPerMm"
                fill="none"
                :stroke="r.outer_line_color"
                :stroke-width="
                    Math.max(
                        1,
                        (r.outer_line_thickness ?? 0) * props.pxPerMm * props.lineThicknessScale,
                    )
                "
            />
        </template>
    </g>
</template>
