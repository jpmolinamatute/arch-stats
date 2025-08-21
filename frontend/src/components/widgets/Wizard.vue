<script context="module" lang="ts">
    import { ref as createRef, type Ref, type Component } from 'vue';

    export type StepRegistration = {
        isValid: Ref<boolean>;
        onComplete: Ref<() => Promise<{ success: boolean; error?: string }>>;
    };

    export type Step = {
        name: string;
        component: Component;
        isValid: Ref<boolean>;
        onComplete: Ref<() => Promise<{ success: boolean; error?: string }>>;
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

        const step: Step = {
            name,
            component,
            isValid,
            onComplete,
            register({ isValid: validRef, onComplete: completeRef }: StepRegistration) {
                isValid.value = validRef.value;
                onComplete.value = completeRef.value;
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

    function registerStep(options: { isValid: Step['isValid']; onComplete: Step['onComplete'] }) {
        currentStep.value.register(options);
    }

    function handleClose() {
        emit('done');
    }

    async function handleNext() {
        errorMessage.value = null;
        const result = await currentStep.value.onComplete.value();
        if (result.success) {
            if (activeStepIndex.value < props.steps.length - 1) {
                activeStepIndex.value++;
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
                {{ activeStepIndex < props.steps.length - 1 ? 'Next' : 'Finish' }}
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
