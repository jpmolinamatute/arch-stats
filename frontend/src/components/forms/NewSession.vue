<script setup lang="ts">
    import { ref, computed, watchEffect } from 'vue';
    import { createSession } from '../../composables/useSession';
    import type { StepRegistration } from '../widgets/Wizard.vue';
    import type { components } from '../../types/types.generated';

    type SessionCreate = components['schemas']['SessionsCreate'];

    const props = defineProps<{ register: (opts: StepRegistration) => void }>();

    // --- Form state (strongly typed from OpenAPI) ---
    type Mutable<T> = { -readonly [K in keyof T]: T[K] };
    const form = ref<Mutable<SessionCreate>>({
        is_opened: true,
        start_time: new Date().toISOString(), // API expects ISO 8601
        location: 'Club',
        is_indoor: false,
        distance: 18, // meters
    });

    // UI uses datetime-local; keep a local value in "YYYY-MM-DDTHH:mm"
    const startLocal = ref<string>(new Date().toISOString().slice(0, 16));

    // Mirror datetime-local -> ISO 8601 for the API
    watchEffect(() => {
        // startLocal is local time without seconds; construct ISO from it
        // If the user clears it, avoid setting an invalid date
        if (startLocal.value && startLocal.value.length >= 16) {
            const d = new Date(startLocal.value);
            if (!Number.isNaN(d.getTime())) {
                form.value.start_time = d.toISOString();
            }
        }
    });

    // --- Validation ---
    const errors = ref<{ location?: string; distance?: string; start?: string }>({});

    const isValid = computed(() => {
        errors.value = {};
        let ok = true;

        if (!form.value.location?.trim()) {
            errors.value.location = 'Location is required.';
            ok = false;
        }
        if (!Number.isFinite(form.value.distance) || form.value.distance <= 0) {
            errors.value.distance = 'Distance must be a number greater than 0.';
            ok = false;
        }
        const t = Date.parse(form.value.start_time);
        if (Number.isNaN(t)) {
            errors.value.start = 'Start time is invalid.';
            ok = false;
        }
        return ok;
    });

    // --- Wizard registration: ONLY the known properties ---
    const onComplete = ref<() => Promise<{ success: boolean; error?: string }>>(async () => {
        try {
            await createSession(form.value); // expects SessionsCreate
            return { success: true };
        } catch (e: unknown) {
            const msg = e instanceof Error ? e.message : 'Failed to create session';
            return { success: false, error: msg };
        }
    });

    props.register({ isValid, onComplete });
</script>

<template>
    <div class="p-4">
        <!-- Location -->
        <label class="block mb-1 font-medium">Location</label>
        <input
            v-model.trim="form.location"
            type="text"
            class="w-full border rounded p-2 mb-1 text-gray-900 bg-white"
            placeholder="Club, Backyard, Range..."
        />
        <p v-if="errors.location" class="text-red-600 text-sm">{{ errors.location }}</p>

        <!-- Environment -->
        <div class="mt-4">
            <span class="block mb-1 font-medium">Environment</span>
            <div class="flex gap-4 items-center">
                <label class="inline-flex items-center gap-2">
                    <input
                        type="radio"
                        name="env"
                        :checked="form.is_indoor === false"
                        @change="form.is_indoor = false"
                    />
                    <span>Outdoor</span>
                </label>
                <label class="inline-flex items-center gap-2">
                    <input
                        type="radio"
                        name="env"
                        :checked="form.is_indoor === true"
                        @change="form.is_indoor = true"
                    />
                    <span>Indoor</span>
                </label>
            </div>
        </div>

        <!-- Distance -->
        <label class="block mb-1 mt-4 font-medium">Distance (m)</label>
        <input
            v-model.number="form.distance"
            type="number"
            inputmode="decimal"
            min="1"
            step="0.5"
            class="w-full border rounded p-2 mb-1 text-gray-900 bg-white"
            placeholder="e.g. 18"
        />
        <p v-if="errors.distance" class="text-red-600 text-sm">{{ errors.distance }}</p>

        <!-- Start Time -->
        <label class="block mb-1 mt-4 font-medium">Start Time</label>
        <input
            v-model="startLocal"
            type="datetime-local"
            class="w-full border rounded p-2 mb-1 text-gray-900 bg-white"
        />
        <p v-if="errors.start" class="text-red-600 text-sm">{{ errors.start }}</p>

        <!-- Is Opened (read-only for new sessions, but shown for clarity) -->
        <div class="mt-4 text-sm text-gray-300">
            <span class="font-medium">Session State:</span>
            <span class="ml-1">{{ form.is_opened ? 'Open' : 'Closed' }}</span>
        </div>
    </div>
</template>
