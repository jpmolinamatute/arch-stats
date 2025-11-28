<script setup lang="ts">
    import { ref, onMounted, onUnmounted } from 'vue';
    import type { components } from '@/types/types.generated';
    import { useFaces } from '@/composables/useFaces';
    import { api } from '@/api/client';
    // import { useShot } from '@/composables/useShot';
    import { useSlot } from '@/composables/useSlot';

    type Face = components['schemas']['Face'];
    type FaceType = components['schemas']['FaceType'];
    const { fetchFace } = useFaces();
    const face = ref<Face | null>(null);
    const SVGWidth = ref(300); // default value

    function isMobile(): boolean {
        const hasTouch = navigator.maxTouchPoints > 0;
        const isSmallScreen = window.matchMedia('(max-width: 768px)').matches;
        return hasTouch && isSmallScreen;
    }

    function getOrientation(): 'portrait' | 'landscape' {
        return window.matchMedia('(orientation: portrait)').matches ? 'portrait' : 'landscape';
    }

    function updateSVGWidth() {
        const mobile = isMobile();
        let width: number;
        if (mobile) {
            const orientation = getOrientation();
            if (orientation === 'portrait') {
                width = window.innerWidth - 20;
            } else {
                width = window.innerHeight - 20;
            }
        } else {
            width = window.innerWidth - 10;
        }
        SVGWidth.value = Math.max(300, width);
    }

    let resizeObserver: ResizeObserver | null = null;

    async function saving_score(score: number, x: number, y: number) {
        const slot = useSlot();
        const currentSlot = slot.currentSlot.value;

        if (!currentSlot || !currentSlot.session_id || !currentSlot.slot_id) {
            console.error('Missing slot or session ID');
            return;
        }

        try {
            await api.createShot({
                slot_id: currentSlot.slot_id,
                score,
                x,
                y,
                is_x: false,
            });
            console.log('Shot recorded:', { score, x, y });
        } catch (e) {
            console.error('Failed to record shot:', e);
        }
    }

    onMounted(async () => {
        updateSVGWidth();
        const slot = useSlot();
        const faceType: FaceType | undefined = slot.currentSlot.value?.face_type;
        if (!faceType || faceType === 'none') {
            throw new Error('No valid face_type found in current slot');
        }
        const result = await fetchFace(faceType);
        if (!result) {
            throw new Error('Face data could not be loaded');
        }
        face.value = result;
        resizeObserver = new ResizeObserver(() => {
            requestAnimationFrame(updateSVGWidth);
        });
        resizeObserver.observe(document.documentElement);
        window.addEventListener('orientationchange', updateSVGWidth);
        window.addEventListener('resize', updateSVGWidth);
    });

    onUnmounted(() => {
        if (resizeObserver) resizeObserver.disconnect();
        window.removeEventListener('orientationchange', updateSVGWidth);
        window.removeEventListener('resize', updateSVGWidth);
    });
</script>

<template>
    <div class="w-full flex flex-col items-center">
        <svg
            v-if="face"
            id="app"
            :width="SVGWidth"
            :height="SVGWidth"
            :viewBox="`0 0 ${face.viewBox} ${face.viewBox}`"
            preserveAspectRatio="xMidYMid meet"
        >
            <rect
                x="0"
                y="0"
                width="100%"
                height="100%"
                fill="white"
                stroke="black"
                stroke-width="3"
                data-score="0"
                @click="saving_score(0, $event.clientX, $event.clientY)"
                class="cursor-pointer"
            />
            <g
                v-for="(ring, index) in face.rings"
                :key="index"
                @click="saving_score(ring.data_score, $event.clientX, $event.clientY)"
                class="cursor-pointer"
            >
                <circle
                    :cx="face.viewBox / 2"
                    :cy="face.viewBox / 2"
                    :r="ring.r"
                    :fill="ring.fill"
                    :stroke="ring.stroke"
                    :stroke-width="ring.stroke_width"
                    :data-score="ring.data_score"
                />
                <text
                    v-if="!(face.render_cross && index === face.rings.length - 1)"
                    :x="face.viewBox / 2 + (ring.r + (face.rings?.[index + 1]?.r ?? 0)) / 2"
                    :y="face.viewBox / 2"
                    dy="0.35em"
                    :fill="ring.fill"
                    :stroke="ring.stroke"
                    stroke-width="0.5"
                    font-size="16"
                    text-anchor="middle"
                    style="pointer-events: none"
                >
                    {{ ring.data_score }}
                </text>
                <text
                    v-if="face.render_cross && index === face.rings.length - 1"
                    :x="face.viewBox / 2"
                    :y="face.viewBox / 2"
                    font-size="30"
                    text-anchor="middle"
                    fill="#000000"
                    style="pointer-events: none"
                >
                    +
                </text>
            </g>
        </svg>
    </div>
</template>

<style scoped>
    /* Container-level styles only; SVG handled inside SvgCanvas */
</style>
