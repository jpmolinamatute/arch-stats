<script setup lang="ts">
import type { components } from '@/types/types.generated'
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import Face from '@/components/Face.vue'
import SlotJoinForm from '@/components/forms/SlotJoinForm.vue'
import AppHeader from '@/components/layout/AppHeader.vue'
import ConfirmModal from '@/components/widgets/ConfirmModal.vue'
import LiveScore from '@/components/widgets/LiveScore.vue'
import LiveStatsTable from '@/components/widgets/LiveStatsTable.vue'
import MiniTable from '@/components/widgets/MiniTable.vue'
import { useAuth } from '@/composables/useAuth'
import { useFaces } from '@/composables/useFaces'
import { useSession } from '@/composables/useSession'
import { useShot } from '@/composables/useShot'
import { useSlot } from '@/composables/useSlot'

type FaceModel = components['schemas']['Face']

const router = useRouter()
const { user, isAuthenticated, bootstrapAuth } = useAuth()
const { currentSession, closeSession, loading, checkForOpenSession } = useSession()
const { currentSlot, getSlot, getSlotCached } = useSlot()
const { createShot, fetchShots, subscribeToShots, shots: historyShots, stats: currentStats, loading: shotLoading } = useShot()
const { fetchFace } = useFaces()

const xCount = computed(() => historyShots.value.filter(s => s.is_x).length)

const showCloseModal = ref(false)
const initializing = ref(true)
const showTarget = ref(true)

// Review / Draft State
const draftShots = ref<{ score: number, x: number, y: number, is_x: boolean, color: string }[]>([])
const face = ref<FaceModel | null>(null)
let wsSocket: WebSocket | null = null

// Persistence Key
const storageKey = computed(() => {
    if (!currentSession.value || !currentSlot.value)
        return null
    return `session_${currentSession.value.session_id}_slot_${currentSlot.value.slot_id}_draft`
})

// Load draft from storage
function loadDraft() {
    if (!storageKey.value)
        return
    const stored = localStorage.getItem(storageKey.value)
    if (stored) {
        try {
            draftShots.value = JSON.parse(stored)
        }
        catch {
            // ignore invalid json
        }
    }
}

// Save draft to storage
watch(draftShots, (newShots) => {
    if (storageKey.value) {
        localStorage.setItem(storageKey.value, JSON.stringify(newShots))
    }
}, { deep: true })

// Clear storage when key changes (e.g. slot changes) or load new
watch(storageKey, (newKey) => {
    if (newKey) {
        loadDraft()
    }
    else {
        draftShots.value = []
    }
})

// Fetch Face when slot face_type changes
watch(() => currentSlot.value?.face_type, async (newFaceId) => {
    if (newFaceId && newFaceId !== 'none') {
        try {
            face.value = await fetchFace(newFaceId)
        }
        catch (e) {
            console.error('Failed to load face:', e)
        }
    }
    else {
        face.value = null
    }
}, { immediate: true })

async function setupShotSubscription(slotId: string) {
    await fetchShots(slotId)
    if (wsSocket) {
        wsSocket.close()
        wsSocket = null
    }
    wsSocket = subscribeToShots(slotId)
}

onMounted(async () => {
    try {
    // First, ensure auth is bootstrapped
        await bootstrapAuth()

        // Wait for authentication to complete
        if (!isAuthenticated.value || !user.value) {
            await router.push('/app')
            return
        }

        // If no current session in state but user is authenticated, try to fetch it
        if (!currentSession.value) {
            await checkForOpenSession(user.value.archer_id)
        }

        // If still no session after checking, redirect back to app
        if (!currentSession.value && !loading.value) {
            await router.push('/app')
            return
        }

        // Fetch the archer's current slot in this session (if any)
        if (currentSession.value && user.value.archer_id) {
            try {
                const slot = getSlotCached() ?? (await getSlot())
                currentSlot.value = slot
                // Initial load of draft
                loadDraft()

                // Initialize Shot Logic (History + WS)
                if (slot && slot.slot_id) {
                    await setupShotSubscription(slot.slot_id)
                }
            }
            catch {
                currentSlot.value = null
            }
        }
    }
    catch (e) {
        console.error('[LiveSession] Error during initialization:', e)
        await router.push('/app')
    }
    finally {
        initializing.value = false
    }
})

onUnmounted(() => {
    if (wsSocket) {
        wsSocket.close()
        wsSocket = null
    }
})

function openCloseModal() {
    showCloseModal.value = true
}

function cancelClose() {
    showCloseModal.value = false
}

async function confirmClose() {
    if (!currentSession.value || !user.value)
        return

    showCloseModal.value = false

    try {
    // Close the session (this will automatically leave the slot first)
        await closeSession(currentSession.value.session_id)
        // Clear any draft
        if (storageKey.value)
            localStorage.removeItem(storageKey.value)

        // Redirect back to app without triggering re-authentication
        // The router will handle showing the form since no session exists
        await router.push('/app')
    }
    catch (e) {
        console.error('Failed to close session:', e)
        // Show error to user with a proper error modal or toast in the future
        const errorMessage
            = e instanceof Error ? e.message : 'Failed to close session. Please try again.'
        // eslint-disable-next-line no-alert
        alert(errorMessage)
    }
}

async function handleSlotAssigned() {
    // After slot is assigned, fetch the complete slot details
    if (currentSession.value && user.value) {
        try {
            const slot = getSlotCached() ?? (await getSlot(true))
            currentSlot.value = slot

            if (slot && slot.slot_id) {
                // Initialize shot logic for newly assigned slot
                await setupShotSubscription(slot.slot_id)
            }
        }
        catch (e) {
            console.error('[LiveSession] Failed to fetch slot after assignment:', e)
        }
    }
}

function handleShotDraft(payload: { score: number, x: number, y: number, is_x: boolean, color: string }) {
    if (!currentSlot.value)
        return

    const limit = currentSession.value?.shot_per_round ?? 6
    if (draftShots.value.length >= limit)
        return
    draftShots.value.push(payload)
}

function handleDraftDelete(index: number) {
    draftShots.value.splice(index, 1)
}

function handleDraftClear() {
    draftShots.value = []
    if (storageKey.value)
        localStorage.removeItem(storageKey.value)
}

async function handleConfirmRound() {
    if (!currentSlot.value || !currentSlot.value.slot_id)
        return

    if (draftShots.value.length === 0)
        return

    try {
        const shotsPayload = draftShots.value.map(s => ({
            slot_id: currentSlot.value!.slot_id,
            score: s.score,
            x: s.x,
            y: s.y,
            is_x: s.is_x,
        }))

        await createShot(shotsPayload)

        // Clear draft on success
        handleDraftClear()
    }
    catch (e: any) {
        console.error('Failed to confirm round:', e)

        // Robust Recovery: If slot not found in backend, try to refresh it
        if (e instanceof Error && e.message === 'Slot not found') {
            try {
                // Force refresh slot from backend
                const refreshedSlot = await getSlot(true)
                if (refreshedSlot && refreshedSlot.slot_id) {
                    currentSlot.value = refreshedSlot

                    // Re-connnect WS
                    await setupShotSubscription(refreshedSlot.slot_id)

                    // Retry submission with new slot_id
                    const newPayload = draftShots.value.map(s => ({
                        slot_id: currentSlot.value!.slot_id,
                        score: s.score,
                        x: s.x,
                        y: s.y,
                        is_x: s.is_x,
                    }))

                    await createShot(newPayload)
                    handleDraftClear()
                    return
                }
            }
            catch (recoveryError) {
                console.error('Recovery failed:', recoveryError)
            }

            // If recovery fails, session/slot invalid
            // eslint-disable-next-line no-alert
            alert('Session appears to be invalid or expired. Please start a new session.')
            await router.push('/app')
        }
    }
}
</script>

<template>
    <div class="min-h-screen flex flex-col">
        <AppHeader />

        <!-- Loading state during initialization -->
        <main v-if="initializing" class="flex-1 flex items-center justify-center p-6">
            <div class="text-center">
                <div class="animate-pulse space-y-4">
                    <div class="h-8 w-48 bg-slate-800 rounded mx-auto" />
                    <div class="h-32 w-96 bg-slate-800 rounded mx-auto" />
                </div>
                <p class="mt-4 text-sm text-slate-400">
                    Loading session...
                </p>
            </div>
        </main>

        <!-- Main content once initialized -->
        <main v-else class="flex-1 p-6">
            <div class="max-w-6xl mx-auto">
                <div class="mb-6 flex items-center justify-between">
                    <div>
                        <h1 class="text-2xl font-bold text-slate-100">
                            Live Session
                        </h1>
                        <p v-if="currentSession" class="text-sm text-slate-400 mt-1">
                            {{ currentSession.session_location }}
                            <span class="mx-2">•</span>
                            {{ currentSession.is_indoor ? 'Indoor' : 'Outdoor' }}
                        </p>
                    </div>

                    <button
                        class="px-4 py-2 text-sm rounded bg-red-600 hover:bg-red-700 text-white disabled:opacity-50 transition-colors duration-200"
                        :disabled="loading || shotLoading"
                        data-testid="close-session-btn"
                        @click="openCloseModal"
                    >
                        Close Session
                    </button>
                </div>

                <!-- Show SlotJoinForm if no slot assigned -->
                <SlotJoinForm
                    v-if="!currentSlot && currentSession"
                    :session-id="currentSession.session_id"
                    @slot-assigned="handleSlotAssigned"
                />

                <!-- Show Live Session details only if slot is assigned -->
                <div v-else-if="currentSlot">
                    <!-- View Switcher -->
                    <div class="flex justify-center mb-6">
                        <div class="bg-slate-800 p-1 rounded-lg inline-flex">
                            <button
                                data-testid="view-target-btn"
                                class="px-4 py-2 text-sm rounded-md transition-colors duration-200"
                                :class="showTarget ? 'bg-indigo-600 text-white shadow' : 'text-slate-400 hover:text-slate-200'"
                                @click="showTarget = true"
                            >
                                Target
                            </button>
                            <button
                                data-testid="view-shots-btn"
                                class="px-4 py-2 text-sm rounded-md transition-colors duration-200"
                                :class="!showTarget ? 'bg-indigo-600 text-white shadow' : 'text-slate-400 hover:text-slate-200'"
                                @click="showTarget = false"
                            >
                                Shots
                            </button>
                        </div>
                    </div>

                    <!-- Target View -->
                    <div v-show="showTarget" class="max-w-3xl mx-auto" data-testid="target-view">
                        <div class="p-6 md:p-8 rounded-lg border border-slate-800 bg-slate-900/50 text-center text-slate-400">
                            <h2 class="text-xl font-semibold text-slate-200 mb-3">
                                Session Active
                            </h2>
                            <p class="text-sm mb-6">
                                Tap the target face to record a shot.
                            </p>

                            <div class="w-full flex flex-col items-center">
                                <MiniTable
                                    :shots="draftShots"
                                    :face="face"
                                    :max-shots="currentSession?.shot_per_round ?? 6"
                                    @delete="handleDraftDelete"
                                    @clear="handleDraftClear"
                                    @confirm="handleConfirmRound"
                                />

                                <p v-if="shotLoading" class="mb-2 text-xs text-emerald-400 animate-pulse">
                                    Saving round...
                                </p>

                                <Face
                                    v-if="currentSlot.face_type && currentSlot.face_type !== 'none'"
                                    :face="face"
                                    :shots="draftShots"
                                    @shot="handleShotDraft"
                                />
                            </div>
                        </div>
                    </div>

                    <!-- Shots List View -->
                    <div v-show="!showTarget" class="max-w-3xl mx-auto" data-testid="shots-view">
                        <LiveScore
                            :shots="historyShots"
                            :shot-per-round="currentSession?.shot_per_round ?? 6"
                        />
                        <LiveStatsTable
                            :stats="currentStats"
                            :x-count="xCount"
                        />
                    </div>
                </div>
            </div>
        </main>

        <!-- Close Session Confirmation Modal -->
        <ConfirmModal
            :show="showCloseModal"
            title="Close Session"
            message="Are you sure you want to close this session? This will end your current shooting session and you'll need to create a new one to continue tracking."
            confirm-text="Close Session"
            cancel-text="Keep Session"
            @confirm="confirmClose"
            @cancel="cancelClose"
        />
    </div>
</template>
