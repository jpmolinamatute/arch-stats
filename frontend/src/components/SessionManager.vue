<script setup lang="ts">
    import { onMounted, ref } from 'vue';
    import { useRouter } from 'vue-router';
    import { useAuth } from '@/composables/useAuth';
    import { useSession } from '@/composables/useSession';
    import SessionForm from '@/components/forms/SessionForm.vue';
    import SlotJoinForm from '@/components/forms/SlotJoinForm.vue';

    const router = useRouter();
    const { user, isAuthenticated } = useAuth();
    const { checkForOpenSession } = useSession();

    const checking = ref(true);
    const showSessionForm = ref(false);
    const showSlotForm = ref(false);
    const createdSessionId = ref<string | null>(null);

    onMounted(async () => {
        // Wait for user to be authenticated
        if (!isAuthenticated.value || !user.value) {
            checking.value = false;
            return;
        }

        // Check if user has an open session
        try {
            const openSession = await checkForOpenSession(user.value.archer_id);

            if (openSession) {
                // Redirect to live session
                await router.push('/app/live-session');
            } else {
                // Show the form to create a new session
                showSessionForm.value = true;
            }
        } catch (e) {
            console.error('[SessionManager] Error checking for open session:', e);
            // On error, show the form
            showSessionForm.value = true;
        } finally {
            checking.value = false;
        }
    });

    function handleSessionCreated(sessionId: string) {
        // Store the session ID and show slot assignment form
        createdSessionId.value = sessionId;
        showSessionForm.value = false;
        showSlotForm.value = true;
    }

    async function handleSlotAssigned() {
        // Redirect to live session view after slot is assigned
        await router.push('/app/live-session');
    }
</script>

<template>
    <div class="flex flex-col items-center justify-center gap-6">
        <!-- Show slot assignment form after session is created (highest priority) -->
        <div v-if="showSlotForm && createdSessionId && isAuthenticated">
            <SlotJoinForm :session-id="createdSessionId" @slot-assigned="handleSlotAssigned" />
        </div>

        <!-- Show session creation form when authenticated and no open session -->
        <div v-else-if="showSessionForm && isAuthenticated">
            <SessionForm @session-created="handleSessionCreated" />
        </div>

        <!-- Loading state while checking for open session -->
        <div v-else-if="checking" class="text-center">
            <div class="animate-pulse space-y-4">
                <div class="h-8 w-48 bg-slate-800 rounded mx-auto"></div>
                <div class="h-32 w-96 bg-slate-800 rounded mx-auto"></div>
            </div>
            <p class="mt-4 text-sm text-slate-400">Checking for active sessions...</p>
        </div>

        <!-- Unauthenticated message -->
        <div v-else class="text-center text-slate-400">
            <p class="text-sm">Please sign in to manage sessions</p>
        </div>
    </div>
</template>
