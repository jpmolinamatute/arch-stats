<script setup lang="ts">
    import { onMounted, ref } from 'vue';
    import { useRouter } from 'vue-router';
    import { useSlot } from '@/composables/useSlot';
    import { useArcher } from '@/composables/useArcher';
    import { useAuth } from '@/composables/useAuth';
    import type { components } from '@/types/types.generated';

    type TargetFaceType = components['schemas']['TargetFaceType'];
    type BowStyleType = components['schemas']['BowStyleType'];

    const props = defineProps<{
        sessionId: string;
    }>();

    const emit = defineEmits<{
        slotAssigned: [slotId: string];
    }>();

    const router = useRouter();
    const { currentSlot, joinSession, getSlot, loading: slotLoading, error: slotError } = useSlot();
    const { getArcher, loading: archerLoading } = useArcher();
    const { user } = useAuth();

    const faceType = ref<TargetFaceType>('40cm_full');
    const bowstyle = ref<BowStyleType>('recurve');
    const drawWeight = ref<number>(25);
    const distance = ref<number>(18);
    const formError = ref<string | null>(null);

    // Target face options
    const faceTypeOptions: { value: TargetFaceType; label: string }[] = [
        { value: '40cm_full', label: '40cm Full Face' },
        { value: '60cm_full', label: '60cm Full Face' },
        { value: '80cm_full', label: '80cm Full Face' },
        { value: '122cm_full', label: '122cm Full Face' },
        { value: '40cm_6rings', label: '40cm 6-Ring' },
        { value: '60cm_6rings', label: '60cm 6-Ring' },
        { value: '80cm_6rings', label: '80cm 6-Ring' },
        { value: '122cm_6rings', label: '122cm 6-Ring' },
        { value: '40cm_triple_vertical', label: '40cm Triple Vertical' },
        { value: '60cm_triple_triangular', label: '60cm Triple Triangular' },
        { value: 'none', label: 'None' },
    ];

    // Bowstyle options
    const bowstyleOptions: { value: BowStyleType; label: string }[] = [
        { value: 'recurve', label: 'Recurve' },
        { value: 'compound', label: 'Compound' },
        { value: 'barebow', label: 'Barebow' },
        { value: 'longbow', label: 'Longbow' },
    ];

    const loading = ref(false);

    // Fetch archer data on mount to pre-fill bowstyle and draw_weight
    onMounted(async () => {
        if (!user.value?.archer_id) {
            formError.value = 'User not authenticated';
            return;
        }

        try {
            loading.value = true;
            const archerData = await getArcher(user.value.archer_id);
            bowstyle.value = archerData.bowstyle;
            drawWeight.value = archerData.draw_weight;
        } catch (e) {
            console.error('Error fetching archer data:', e);
            // Continue with default values if fetch fails
        } finally {
            loading.value = false;
        }
    });

    async function handleSubmit() {
        formError.value = null;

        // Validate
        if (!user.value?.archer_id) {
            formError.value = 'User not authenticated';
            return;
        }

        if (distance.value < 1 || distance.value > 100) {
            formError.value = 'Distance must be between 1 and 100 meters';
            return;
        }

        if (drawWeight.value <= 0) {
            formError.value = 'Draw weight must be positive';
            return;
        }

        try {
            const result = await joinSession({
                archer_id: user.value.archer_id,
                session_id: props.sessionId,
                face_type: faceType.value,
                is_shooting: true,
                bowstyle: bowstyle.value,
                draw_weight: drawWeight.value,
                distance: distance.value,
                club_id: null,
            });

            // Fetch the full slot details and store in session state
            const slotDetails = await getSlot(props.sessionId, user.value.archer_id);
            currentSlot.value = slotDetails;

            emit('slotAssigned', result.slot_id);
            // Redirect to live session view
            await router.push('/app/live-session');
        } catch (e) {
            console.error('Error joining session:', e);
            formError.value = e instanceof Error ? e.message : 'Failed to join session';
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
                <h3 class="text-lg font-semibold text-slate-100">Join Shooting Lane</h3>
                <p class="text-xs text-slate-400 mt-1">
                    Configure your equipment and distance to get assigned a target and slot
                </p>
            </div>

            <!-- Loading state -->
            <div v-if="loading || archerLoading" class="text-center py-4">
                <p class="text-sm text-slate-400">Loading archer data...</p>
            </div>

            <!-- Form fields -->
            <template v-else>
                <div v-if="formError || slotError" class="text-sm text-red-600 text-left">
                    {{ formError || slotError }}
                </div>

                <!-- Target Face Type -->
                <label class="block text-left text-xs text-slate-300">
                    Target Face Type
                    <select
                        v-model="faceType"
                        class="mt-1 w-full border border-slate-700 p-2 rounded bg-slate-900 text-slate-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        required
                        :disabled="slotLoading"
                    >
                        <option
                            v-for="option in faceTypeOptions"
                            :key="option.value"
                            :value="option.value"
                        >
                            {{ option.label }}
                        </option>
                    </select>
                </label>

                <!-- Bowstyle -->
                <label class="block text-left text-xs text-slate-300">
                    Bow Style
                    <select
                        v-model="bowstyle"
                        class="mt-1 w-full border border-slate-700 p-2 rounded bg-slate-900 text-slate-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        required
                        :disabled="slotLoading"
                    >
                        <option
                            v-for="option in bowstyleOptions"
                            :key="option.value"
                            :value="option.value"
                        >
                            {{ option.label }}
                        </option>
                    </select>
                </label>

                <!-- Draw Weight -->
                <label class="block text-left text-xs text-slate-300">
                    Draw Weight (lbs)
                    <input
                        v-model.number="drawWeight"
                        type="number"
                        step="0.5"
                        min="1"
                        placeholder="e.g., 25"
                        class="mt-1 w-full border border-slate-700 p-2 rounded bg-slate-900 text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        required
                        :disabled="slotLoading"
                    />
                </label>

                <!-- Distance -->
                <label class="block text-left text-xs text-slate-300">
                    Distance (meters)
                    <input
                        v-model.number="distance"
                        type="number"
                        min="1"
                        max="100"
                        placeholder="e.g., 18"
                        class="mt-1 w-full border border-slate-700 p-2 rounded bg-slate-900 text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        required
                        :disabled="slotLoading"
                    />
                    <span class="text-xs text-slate-500 mt-1 block">Distance: 1-100 meters</span>
                </label>

                <div class="flex gap-3">
                    <button
                        type="submit"
                        class="flex-1 px-4 py-2 text-sm rounded bg-green-600 hover:bg-green-700 text-white disabled:opacity-50 disabled:cursor-not-allowed transition-colors duration-200"
                        :disabled="slotLoading"
                    >
                        {{ slotLoading ? 'Assigning...' : 'Join Lane' }}
                    </button>
                </div>
            </template>
        </form>

        <div class="mt-4 text-center text-xs text-slate-400">
            <p>You'll be assigned to an available target and slot</p>
        </div>
    </div>
</template>
