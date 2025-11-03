<script setup lang="ts">
    import { onMounted, ref } from 'vue';
    import { useRouter } from 'vue-router';
    import { useAuth } from '@/composables/useAuth';
    import { useSession } from '@/composables/useSession';
    import { useSlot } from '@/composables/useSlot';
    import AppHeader from '@/components/layout/AppHeader.vue';
    import ConfirmModal from '@/components/widgets/ConfirmModal.vue';
    import SlotJoinForm from '@/components/forms/SlotJoinForm.vue';

    const router = useRouter();
    const { user, isAuthenticated, bootstrapAuth } = useAuth();
    const { currentSession, closeSession, loading, checkForOpenSession } = useSession();
    const { currentSlot, getSlot } = useSlot();

    const showCloseModal = ref(false);
    const initializing = ref(true);

    onMounted(async () => {
        try {
            // First, ensure auth is bootstrapped
            await bootstrapAuth();

            // Wait for authentication to complete
            if (!isAuthenticated.value || !user.value) {
                await router.push('/app');
                return;
            }

            // If no current session in state but user is authenticated, try to fetch it
            if (!currentSession.value) {
                await checkForOpenSession(user.value.archer_id);
            }

            // If still no session after checking, redirect back to app
            if (!currentSession.value && !loading.value) {
                await router.push('/app');
                return;
            }

            // Fetch the archer's current slot in this session (if any)
            if (currentSession.value && user.value.archer_id) {
                try {
                    const slot = await getSlot(
                        currentSession.value.session_id,
                        user.value.archer_id,
                    );
                    currentSlot.value = slot;
                } catch (slotError) {
                    // 404 means archer has no active slot - show SlotJoinForm to complete setup
                    console.log('[LiveSession] Archer has no active slot, showing join form');
                    currentSlot.value = null;
                }
            }
        } catch (e) {
            console.error('[LiveSession] Error during initialization:', e);
            await router.push('/app');
        } finally {
            initializing.value = false;
        }
    });

    function openCloseModal() {
        showCloseModal.value = true;
    }

    function cancelClose() {
        showCloseModal.value = false;
    }

    async function confirmClose() {
        if (!currentSession.value || !user.value) return;

        showCloseModal.value = false;

        try {
            // Close the session (this will automatically leave the slot first)
            await closeSession(currentSession.value.session_id, user.value.archer_id);

            // Redirect back to app without triggering re-authentication
            // The router will handle showing the form since no session exists
            await router.push('/app');
        } catch (e) {
            console.error('Failed to close session:', e);
            // Show error to user with a proper error modal or toast in the future
            const errorMessage =
                e instanceof Error ? e.message : 'Failed to close session. Please try again.';
            alert(errorMessage);
        }
    }

    async function handleSlotAssigned() {
        // After slot is assigned, fetch the complete slot details
        if (currentSession.value && user.value) {
            try {
                const slot = await getSlot(currentSession.value.session_id, user.value.archer_id);
                currentSlot.value = slot;
                console.log('[LiveSession] Slot assigned successfully:', currentSlot.value);
            } catch (e) {
                console.error('[LiveSession] Failed to fetch slot after assignment:', e);
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
                    <div class="h-8 w-48 bg-slate-800 rounded mx-auto"></div>
                    <div class="h-32 w-96 bg-slate-800 rounded mx-auto"></div>
                </div>
                <p class="mt-4 text-sm text-slate-400">Loading session...</p>
            </div>
        </main>

        <!-- Main content once initialized -->
        <main v-else class="flex-1 p-6">
            <div class="max-w-6xl mx-auto">
                <div class="mb-6 flex items-center justify-between">
                    <div>
                        <h1 class="text-2xl font-bold text-slate-100">Live Session</h1>
                        <p class="text-sm text-slate-400 mt-1" v-if="currentSession">
                            {{ currentSession.session_location }}
                            <span class="mx-2">•</span>
                            {{ currentSession.is_indoor ? 'Indoor' : 'Outdoor' }}
                        </p>
                    </div>

                    <button
                        @click="openCloseModal"
                        class="px-4 py-2 text-sm rounded bg-red-600 hover:bg-red-700 text-white disabled:opacity-50 transition-colors duration-200"
                        :disabled="loading"
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
                <div
                    v-else-if="currentSlot"
                    class="p-8 rounded-lg border border-slate-800 bg-slate-900/50 text-center text-slate-400"
                >
                    <h2 class="text-xl font-semibold text-slate-200 mb-2">Session Active</h2>
                    <p class="text-sm mb-4">
                        Your shooting session is now active. This is where shot tracking and
                        real-time updates will appear.
                    </p>
                    <p class="text-xs text-slate-500">
                        (Live session features will be implemented in the next phase)
                    </p>

                    <div v-if="currentSession" class="mt-8 text-left max-w-md mx-auto">
                        <h3 class="text-sm font-semibold text-slate-300 mb-3">Session Details</h3>
                        <dl class="space-y-2 text-sm">
                            <div class="flex justify-between">
                                <dt class="text-slate-400">Session ID:</dt>
                                <dd class="text-slate-200 font-mono text-xs">
                                    {{ currentSession.session_id }}
                                </dd>
                            </div>
                            <div class="flex justify-between">
                                <dt class="text-slate-400">Location:</dt>
                                <dd class="text-slate-200">
                                    {{ currentSession.session_location }}
                                </dd>
                            </div>
                            <div class="flex justify-between">
                                <dt class="text-slate-400">Environment:</dt>
                                <dd class="text-slate-200">
                                    {{ currentSession.is_indoor ? 'Indoor' : 'Outdoor' }}
                                </dd>
                            </div>
                            <div class="flex justify-between">
                                <dt class="text-slate-400">Status:</dt>
                                <dd class="text-green-400">Open</dd>
                            </div>
                        </dl>

                        <!-- Slot Details Section -->
                        <div v-if="currentSlot" class="mt-6 pt-6 border-t border-slate-700">
                            <h3 class="text-sm font-semibold text-slate-300 mb-3">
                                Your Shooting Lane
                            </h3>
                            <dl class="space-y-2 text-sm">
                                <div class="flex justify-between">
                                    <dt class="text-slate-400">Slot:</dt>
                                    <dd class="text-slate-200 font-semibold text-lg">
                                        {{ currentSlot.slot }}
                                    </dd>
                                </div>
                                <div class="flex justify-between">
                                    <dt class="text-slate-400">Bow Style:</dt>
                                    <dd class="text-slate-200 capitalize">
                                        {{ currentSlot.bowstyle }}
                                    </dd>
                                </div>
                                <div class="flex justify-between">
                                    <dt class="text-slate-400">Draw Weight:</dt>
                                    <dd class="text-slate-200">
                                        {{ currentSlot.draw_weight }} lbs
                                    </dd>
                                </div>
                                <div class="flex justify-between">
                                    <dt class="text-slate-400">Target Face:</dt>
                                    <dd class="text-slate-200 capitalize">
                                        {{ currentSlot.face_type.replace('_', ' ') }}
                                    </dd>
                                </div>
                            </dl>
                        </div>
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
