<script setup lang="ts">
import type { components } from '@/types/types.generated'
import { onMounted, onUnmounted, ref, watch } from 'vue'

type Face = components['schemas']['Face']

const props = defineProps<{
    face: Face | null
    shots: { x: number, y: number }[]
    width?: number
}>()

const emit = defineEmits<{
    (e: 'shot', payload: { score: number, x: number, y: number, is_x: boolean, color: string }): void
}>()

const SVGWidth = ref(300) // default value
const svgRef = ref<SVGSVGElement | null>(null)

function isMobile(): boolean {
    const hasTouch = navigator.maxTouchPoints > 0
    const isSmallScreen = window.matchMedia('(max-width: 768px)').matches
    return hasTouch && isSmallScreen
}

function getOrientation(): 'portrait' | 'landscape' {
    return window.matchMedia('(orientation: portrait)').matches ? 'portrait' : 'landscape'
}

function updateSVGWidth() {
    if (props.width) {
        SVGWidth.value = props.width
        return
    }

    const mobile = isMobile()
    let width: number
    if (mobile) {
        const orientation = getOrientation()
        if (orientation === 'portrait') {
            width = window.innerWidth - 20
        }
        else {
            width = window.innerHeight - 20
        }
    }
    else {
        width = window.innerWidth - 10
    }
    SVGWidth.value = Math.max(300, width)
}

let resizeObserver: ResizeObserver | null = null

function getSVGCoordinates(clientX: number, clientY: number): { x: number, y: number } | null {
    if (!svgRef.value)
        return null
    const svg = svgRef.value
    const pt = svg.createSVGPoint()
    pt.x = clientX
    pt.y = clientY
    const svgP = pt.matrixTransform(svg.getScreenCTM()?.inverse())
    return { x: svgP.x, y: svgP.y }
}

function handleShotClick(score: number, color: string, clientX: number, clientY: number, isX: boolean = false) {
    const svgCoords = getSVGCoordinates(clientX, clientY)
    if (!svgCoords)
        return

    const { x, y } = svgCoords
    emit('shot', { score, x, y, is_x: isX, color })
}

watch(() => props.width, updateSVGWidth)

onMounted(() => {
    updateSVGWidth()

    resizeObserver = new ResizeObserver(() => {
        requestAnimationFrame(updateSVGWidth)
    })
    resizeObserver.observe(document.documentElement)
    window.addEventListener('orientationchange', updateSVGWidth)
    window.addEventListener('resize', updateSVGWidth)
})

onUnmounted(() => {
    if (resizeObserver)
        resizeObserver.disconnect()
    window.removeEventListener('orientationchange', updateSVGWidth)
    window.removeEventListener('resize', updateSVGWidth)
})
</script>

<template>
    <div class="w-full flex flex-col items-center">
        <svg
            v-if="face"
            id="app"
            ref="svgRef"
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
                class="cursor-pointer"
                @click="handleShotClick(0, 'white', $event.clientX, $event.clientY)"
            />
            <g
                v-for="(ring, index) in face.rings"
                :key="index"
                class="cursor-pointer"
                @click="handleShotClick(ring.data_score, ring.fill, $event.clientX, $event.clientY, !!(face.render_cross && index === face.rings.length - 1))"
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
            <!-- Visual feedback for shots -->
            <circle
                v-for="(shot, index) in shots"
                :key="`shot-${index}`"
                :cx="shot.x"
                :cy="shot.y"
                r="2.5"
                fill="transparent"
                stroke="black"
                stroke-width="1"
                pointer-events="none"
            />
        </svg>
    </div>
</template>

<style scoped>
    /* Container-level styles only; SVG handled inside SvgCanvas */
</style>
