<script context="module" lang="ts">
    import { ref as createRef, watch, type Ref, type Component } from 'vue';

    export type StepRegistration = {
        isValid: Ref<boolean>;
        onComplete: Ref<() => Promise<{ success: boolean; error?: string }>>;
        /** Optional flag allowing a step to control whether the primary button is Next (true) or Finish (false). */
        hasNext?: Ref<boolean>;
    };

    export type Step = {
        name: string;
        component: Component;
        isValid: Ref<boolean>;
        onComplete: Ref<() => Promise<{ success: boolean; error?: string }>>;
        /** Optional: when provided, overrides the default Next/Finish decision. */
        hasNext?: Ref<boolean>;
        /** Internal flag: true when a step actually provided hasNext override. */
        hasNextProvided?: Ref<boolean>;
        register: (options: StepRegistration) => void;
    };

    export function createStep(
        name: string,
        component: Component,
    ): { step: Step; bindings: StepRegistration } {
        const isValid = createRef(false);
        const onComplete = createRef<() => Promise<{ success: boolean; error?: string }>>(
            async () => ({
                success: false,
                error: 'Not implemented',
            }),
        );
        // hasNext is optional; keep an internal ref and mirror child's ref via watch
        const hasNextInternal = createRef<boolean>(false);
        const hasNextProvided = createRef<boolean>(false);

        const step: Step = {
            name,
            component,
            isValid,
            onComplete,
            // important: leave undefined until a child provides hasNext
            hasNext: undefined,
            hasNextProvided,
            register({
                isValid: validRef,
                onComplete: completeRef,
                hasNext: hasNextRef,
            }: StepRegistration) {
                // Mirror validity reactively
                isValid.value = validRef.value;
                watch(validRef, (v) => (isValid.value = v));
                onComplete.value = completeRef.value;

                // Mirror hasNext reactively to internal ref; expose stable ref on step
                if (hasNextRef) {
                    step.hasNext = hasNextInternal;
                    hasNextProvided.value = true;
                    hasNextInternal.value = hasNextRef.value;
                    watch(hasNextRef, (v) => (hasNextInternal.value = v));
                }
            },
        };

        return { step, bindings: { isValid, onComplete } };
    }
</script>

<script setup lang="ts">
    import { ref, computed } from 'vue';
    const props = defineProps<{ steps: Step[] }>();
    const emit = defineEmits<{ (e: 'done'): void }>();

    const activeStepIndex = ref(0);
    const errorMessage = ref<string | null>(null);
    const currentStep = computed(() => props.steps[activeStepIndex.value]!);
    const isNextAction = computed(() => {
        const step = currentStep.value;
        // If a step provides hasNext, honor it; otherwise fall back to index-based logic
        if (step.hasNextProvided && step.hasNextProvided.value && step.hasNext) {
            return !!step.hasNext.value;
        }
        return activeStepIndex.value < props.steps.length - 1;
    });

    function registerStep(options: StepRegistration) {
        currentStep.value.register(options);
    }

    function handleClose() {
        emit('done');
    }

    async function handleNext() {
        errorMessage.value = null;
        const result = await currentStep.value.onComplete.value();
        if (result.success) {
            if (isNextAction.value) {
                // Advance if there is another step; otherwise, stay (single-step repeat flows)
                if (activeStepIndex.value < props.steps.length - 1) {
                    activeStepIndex.value++;
                }
                // If it's the last step but still hasNext, remain on the same step.
            } else {
                handleClose();
            }
        } else {
            errorMessage.value = result.error ?? 'An error occurred';
        }
    }
</script>

<template>
    <div
        class="wizard max-w-lg mx-auto p-4 bg-slate-800/70 backdrop-blur rounded border border-slate-700 shadow text-slate-100"
    >
        <h2 class="text-xl font-bold mb-2 text-slate-100">{{ currentStep.name }}</h2>

        <component :is="currentStep.component" :key="currentStep.name" :register="registerStep" />

        <p v-if="errorMessage" class="text-danger text-sm mt-2">{{ errorMessage }}</p>

        <div class="flex gap-2 mt-4">
            <button
                @click="handleNext"
                :disabled="!currentStep.isValid.value"
                class="bg-primary hover:bg-primary/90 text-white px-4 py-2 rounded disabled:opacity-50 disabled:cursor-not-allowed"
            >
                {{ isNextAction ? 'Next' : 'Finish' }}
            </button>
            <button
                @click="handleClose"
                class="bg-slate-600 hover:bg-slate-500 text-white px-4 py-2 rounded"
            >
                Close
            </button>
        </div>
    </div>
</template>
