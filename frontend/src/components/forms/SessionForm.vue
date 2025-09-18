<script setup lang="ts">
    import { ref } from 'vue';
    import { useSession } from '@/composables/useSession';
    import { useAuth } from '@/composables/useAuth';

    const emit = defineEmits<{
        sessionCreated: [sessionId: string];
    }>();

    const { createSession, loading, error } = useSession();
    const { user } = useAuth();

    const sessionLocation = ref<string>('');
    const isIndoor = ref<boolean>(false);
    const formError = ref<string | null>(null);

    async function handleSubmit() {
        formError.value = null;

        // Validate
        if (!sessionLocation.value.trim()) {
            formError.value = 'Session location is required';
            return;
        }

        if (!user.value?.archer_id) {
            formError.value = 'User not authenticated';
            return;
        }

        try {
            const sessionId = await createSession({
                owner_archer_id: user.value.archer_id,
                session_location: sessionLocation.value.trim(),
                is_indoor: isIndoor.value,
                is_opened: true,
            });

            // Emit event but don't redirect yet - parent will show slot assignment form
            emit('sessionCreated', sessionId);
        } catch (e) {
            console.error('Error creating session:', e);
            formError.value = e instanceof Error ? e.message : 'Failed to create session';
        }
    }
</script>

<template>
    <div class="w-full max-w-md mx-auto">
        <form
            @submit.prevent="handleSubmit"
            class="space-y-4 p-6 bg-transparent rounded-lg border border-slate-800 text-slate-200"
        >
            <div class="text-left">
                <h3 class="text-lg font-semibold text-slate-100">Start a New Session</h3>
                <p class="text-xs text-slate-400 mt-1">
                    Create a shooting session to start tracking your performance
                </p>
            </div>

            <div v-if="formError || error" class="text-sm text-red-600 text-left">
                {{ formError || error }}
            </div>

            <label class="block text-left text-xs text-slate-300">
                Session Location
                <input
                    v-model="sessionLocation"
                    type="text"
                    placeholder="e.g., Range A, Indoor Hall, Field"
                    class="mt-1 w-full border border-slate-700 p-2 rounded bg-slate-900 text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    required
                    :disabled="loading"
                />
            </label>

            <label class="flex items-center text-left text-sm text-slate-300 cursor-pointer">
                <input
                    v-model="isIndoor"
                    type="checkbox"
                    class="mr-2 w-4 h-4 border-slate-700 rounded bg-slate-900 text-blue-600 focus:ring-2 focus:ring-blue-500"
                    :disabled="loading"
                />
                Indoor session
            </label>

            <div class="flex gap-3">
                <button
                    type="submit"
                    class="flex-1 px-4 py-2 text-sm rounded bg-green-600 hover:bg-green-700 text-white disabled:opacity-50 disabled:cursor-not-allowed transition-colors duration-200"
                    :disabled="loading"
                >
                    {{ loading ? 'Creating...' : 'Start Session' }}
                </button>
            </div>
        </form>

        <div class="mt-4 text-center text-xs text-slate-400">
            <p>After creating the session, you'll set up your shooting lane</p>
        </div>
    </div>
</template>
