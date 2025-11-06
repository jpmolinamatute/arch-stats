<script setup lang="ts">
    import { computed, onMounted, ref, watch } from 'vue';
    import type { components } from '@/types/types.generated';
    import { useTargetFace } from '@/composables/useTargetFace';

    type TargetFace = components['schemas']['TargetFace'];
    type TargetFaceType = components['schemas']['TargetFaceType'];

    type ShotPayload = {
        xMm: number; // +right (global, relative to canvas center)
        yMm: number; // +up (screen y inverted)
        score: number; // 0..10
    };

    const props = withDefaults(
        defineProps<{
            targetId?: TargetFaceType; // e.g., 'wa_122cm_full'
            targetFace?: TargetFace; // if already fetched
            sizePx?: number; // base size in pixels (used when container is unknown)
            maxSizePx?: number; // optional clamp for very large screens
            showCrosshair?: boolean | 'auto'; // 'auto' -> follow targetFace.render_cross
            lineThicknessScale?: number; // visual only (stroke px = mm * pxPerMm * scale)
            minZoom?: number;
            maxZoom?: number;
        }>(),
        {
            sizePx: 360,
            maxSizePx: Infinity,
            showCrosshair: 'auto',
            lineThicknessScale: 1.0,
            minZoom: 1,
            maxZoom: 5,
        },
    );

    const emit = defineEmits<{
        (e: 'shot', payload: ShotPayload): void;
        (e: 'preview', payload: ShotPayload): void;
    }>();

    // Data loading
    const { fetchTargetFace } = useTargetFace();
    const face = ref<TargetFace | null>(props.targetFace ?? null);

    async function ensureFaceLoaded() {
        if (props.targetFace) {
            face.value = props.targetFace;
            return;
        }
        if (props.targetId) {
            face.value = await fetchTargetFace(props.targetId);
        } else {
            face.value = null;
        }
    }

    watch(
        () => [props.targetId, props.targetFace] as const,
        async () => {
            await ensureFaceLoaded();
        },
        { immediate: true },
    );

    // Responsive sizing
    const containerRef = ref<HTMLDivElement | null>(null);
    const containerWidth = ref<number>(0);
    const renderSizePx = computed(() => {
        // prefer container width if available; keep square canvas
        const w = containerWidth.value > 0 ? containerWidth.value : props.sizePx;
        return Math.min(w, props.maxSizePx ?? Infinity);
    });

    // Derived geometry
    const svgRef = ref<SVGSVGElement | null>(null);

    // Compute the maximum extent in mm so we scale the whole arrangement to fit the viewport
    const maxExtentMm = computed(() => {
        const f = face.value;
        if (!f || f.spots.length === 0) return 1;
        let maxX = 0;
        let maxY = 0;
        for (const s of f.spots) {
            const r = s.diameter / 2;
            maxX = Math.max(maxX, Math.abs(s.x_offset) + r);
            maxY = Math.max(maxY, Math.abs(s.y_offset) + r);
        }
        // Fit the larger of width/height
        return Math.max(maxX, maxY);
    });

    const pxPerMmBase = computed(() => {
        const f = face.value;
        if (!f) return 1;
        const scale = renderSizePx.value / (2 * maxExtentMm.value);
        return scale * (f.svg_scale_factor ?? 1);
    });

    // Zoom state (applied on top of base px/mm)
    const zoom = ref<number>(1);
    const pxPerMm = computed(() => pxPerMmBase.value * zoom.value);

    // For drawing bands: sort rings by radius (outermost first)
    const ringsDescByRadius = computed(() => {
        const f = face.value;
        if (!f) return [] as TargetFace['rings'];
        return [...f.rings].sort((a, b) => b.scoring_zone_radius - a.scoring_zone_radius);
    });

    // For scoring: sort rings by radius (innermost -> outermost)
    const ringsAscByRadius = computed(() => {
        const f = face.value;
        if (!f) return [] as TargetFace['rings'];
        return [...f.rings].sort((a, b) => a.scoring_zone_radius - b.scoring_zone_radius);
    });

    const drawCrosshair = computed(() => {
        if (props.showCrosshair === 'auto') return !!face.value?.render_cross;
        return !!props.showCrosshair;
    });

    // Interaction
    const marker = ref<{ xPx: number; yPx: number } | null>(null);
    const activePointers = new Map<number, { x: number; y: number }>();
    const pinch = ref<{ baseZoom: number; startDist: number } | null>(null);

    function clientToLocalMm(evt: PointerEvent): { xMm: number; yMm: number } {
        const svg = svgRef.value!;
        const rect = svg.getBoundingClientRect();
        const xPx = evt.clientX - rect.left;
        const yPx = evt.clientY - rect.top;

        // Center coordinates in px
        const cxPx = renderSizePx.value / 2;
        const cyPx = renderSizePx.value / 2;

        const dxPx = xPx - cxPx;
        const dyPx = yPx - cyPx;

        const xMm = dxPx / (pxPerMmBase.value * zoom.value);
        const yMm = -(dyPx / (pxPerMmBase.value * zoom.value)); // invert y to make +up
        return { xMm, yMm };
    }

    function scoreFromPoint(xMm: number, yMm: number): number {
        const f = face.value;
        if (!f) return 0;
        if (f.spots.length === 0 || f.rings.length === 0) return 0;

        const ringsAsc = ringsAscByRadius.value;
        const outermostR = ringsAsc[ringsAsc.length - 1]?.scoring_zone_radius ?? 0;

        let bestScore = 0;

        for (const s of f.spots) {
            const sx = xMm - s.x_offset;
            const sy = yMm - s.y_offset;
            const d = Math.hypot(sx, sy);

            // Quick reject if outside the spot diameter (when provided), else use outermost ring
            const spotRadius = (s.diameter ?? outermostR * 2) / 2;
            if (d > spotRadius) {
                continue;
            }

            // On a boundary? Credit the higher (inner) score when touching the line
            let scored = 0;
            for (const r of ringsAsc) {
                const boundaryR = r.scoring_zone_radius;
                const t = r.outer_line_thickness ?? 0;
                const onLine = Math.abs(d - boundaryR) <= t / 2;
                if (onLine) {
                    scored = Math.max(scored, r.score);
                    break;
                }
                if (d <= boundaryR) {
                    scored = Math.max(scored, r.score);
                    break;
                }
            }

            bestScore = Math.max(bestScore, scored);
        }

        return bestScore;
    }

    function handlePointer(evt: PointerEvent, commit: boolean) {
        const { xMm, yMm } = clientToLocalMm(evt);
        const score = scoreFromPoint(xMm, yMm);

        // Marker in px for visualization
        const xPx = xMm * pxPerMm.value + renderSizePx.value / 2;
        const yPx = -yMm * pxPerMm.value + renderSizePx.value / 2;
        marker.value = { xPx, yPx };

        const payload: ShotPayload = { xMm, yMm, score };
        if (commit) emit('shot', payload);
        else emit('preview', payload);
    }

    function onPointerDown(e: PointerEvent) {
        // Track pointers for pinch
        activePointers.set(e.pointerId, { x: e.clientX, y: e.clientY });
        if (activePointers.size === 2) {
            const vals = [...activePointers.values()];
            const a = vals[0];
            const b = vals[1];
            if (a && b) {
                const dx = a.x - b.x;
                const dy = a.y - b.y;
                pinch.value = { baseZoom: zoom.value, startDist: Math.hypot(dx, dy) };
            }
            return; // do not shoot while starting pinch
        }
        // Single-pointer shot
        handlePointer(e, true);
    }
    function onPointerMove(e: PointerEvent) {
        if (activePointers.has(e.pointerId)) {
            activePointers.set(e.pointerId, { x: e.clientX, y: e.clientY });
            // If pinching, update zoom
            if (pinch.value && activePointers.size >= 2) {
                const vals = [...activePointers.values()];
                const a = vals[0];
                const b = vals[1];
                if (a && b) {
                    const dx = a.x - b.x;
                    const dy = a.y - b.y;
                    const dist = Math.hypot(dx, dy);
                    const raw = (dist / pinch.value.startDist) * pinch.value.baseZoom;
                    zoom.value = Math.min(Math.max(raw, props.minZoom), props.maxZoom);
                }
                return;
            }
        }
        // Hover preview when not pinching
        if (e.pressure > 0 || e.buttons) return;
        handlePointer(e, false);
    }
    function onPointerUp(e: PointerEvent) {
        activePointers.delete(e.pointerId);
        if (activePointers.size < 2) {
            pinch.value = null;
        }
    }
    function onWheel(e: WheelEvent) {
        e.preventDefault();
        const step = e.deltaY > 0 ? 0.9 : 1.1;
        const raw = zoom.value * step;
        zoom.value = Math.min(Math.max(raw, props.minZoom), props.maxZoom);
    }

    onMounted(() => {
        // Manage our own gestures; disable native panning/zooming on the svg element
        svgRef.value?.setAttribute('touch-action', 'none');
        // Observe container width for responsive sizing
        const updateWidth = () => {
            if (containerRef.value) {
                const w = containerRef.value.clientWidth;
                containerWidth.value = w;
            }
        };
        updateWidth();
        const ro = new ResizeObserver(() => updateWidth());
        if (containerRef.value) ro.observe(containerRef.value);
    });
</script>

<template>
    <div ref="containerRef" class="w-full flex flex-col items-center">
        <svg
            ref="svgRef"
            :width="renderSizePx"
            :height="renderSizePx"
            :viewBox="`0 0 ${renderSizePx} ${renderSizePx}`"
            @pointerdown="onPointerDown"
            @pointermove="onPointerMove"
            @pointerup="onPointerUp"
            @pointercancel="onPointerUp"
            @wheel="onWheel"
            class="select-none"
        >
            <!-- Background -->
            <rect :width="renderSizePx" :height="renderSizePx" fill="#f5f5f5" />

            <!-- Drawing origin at canvas center -->
            <g :transform="`translate(${renderSizePx / 2}, ${renderSizePx / 2})`">
                <!-- Zoom wrapper -->
                <g :transform="`scale(${zoom})`">
                    <!-- Spots (each spot is a full target replicated at an offset) -->
                    <template
                        v-for="s in face?.spots ?? []"
                        :key="`spot-${s.x_offset}-${s.y_offset}-${s.diameter}`"
                    >
                        <g
                            :transform="`translate(${s.x_offset * pxPerMm}, ${-s.y_offset * pxPerMm})`"
                        >
                            <!-- Bands: draw outermost -> innermost as filled circles -->
                            <template
                                v-for="r in ringsDescByRadius"
                                :key="`ring-fill-${s.x_offset}-${s.y_offset}-${r.scoring_zone_radius}`"
                            >
                                <circle
                                    :r="r.scoring_zone_radius * pxPerMm"
                                    :fill="r.scoring_zone_color"
                                />
                            </template>
                            <!-- Boundary lines at the outer edge of each scoring zone -->
                            <template
                                v-for="r in ringsDescByRadius"
                                :key="`ring-line-${s.x_offset}-${s.y_offset}-${r.scoring_zone_radius}`"
                            >
                                <circle
                                    :r="r.scoring_zone_radius * pxPerMm"
                                    fill="none"
                                    :stroke="r.outer_line_color"
                                    :stroke-width="
                                        Math.max(
                                            1,
                                            (r.outer_line_thickness ?? 0) *
                                                pxPerMm *
                                                lineThicknessScale,
                                        )
                                    "
                                />
                            </template>
                        </g>
                    </template>

                    <!-- Crosshair at global center -->
                    <template v-if="drawCrosshair">
                        <line x1="-8" y1="0" x2="8" y2="0" stroke="#555" stroke-width="1" />
                        <line x1="0" y1="-8" x2="0" y2="8" stroke="#555" stroke-width="1" />
                    </template>

                    <!-- Click marker (global coords) -->
                    <circle
                        v-if="marker"
                        :cx="marker.xPx - renderSizePx / 2"
                        :cy="marker.yPx - renderSizePx / 2"
                        r="4"
                        fill="#111"
                        stroke="#fff"
                        stroke-width="1"
                    />
                </g>
            </g>
        </svg>
    </div>
</template>

<style scoped>
    svg {
        touch-action: none;
    }
</style>
