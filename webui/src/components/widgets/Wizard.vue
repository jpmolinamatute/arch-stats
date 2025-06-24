<script setup lang="ts">
    import { computed, ref, type Component, type Ref } from 'vue';

    type Step = {
        name: string;
        component: Component;
        isValid: Ref<boolean>;
        onComplete: Ref<() => Promise<{ success: boolean; error?: string }>>;
    };

    const props = defineProps<{
        steps: Step[];
    }>();

    const emit = defineEmits<{ done: [] }>();

    const activeStepIndex = ref(0);
    const errorMessage = ref<string | null>(null);

    const currentStep = computed(() => props.steps[activeStepIndex.value]);

    async function handleNext() {
        errorMessage.value = null;
        const result = await currentStep.value.onComplete.value();
        if (result.success) {
            if (activeStepIndex.value < props.steps.length - 1) {
                activeStepIndex.value++;
            } else {
                emit('done');
            }
        } else {
            errorMessage.value = result.error || 'An error occurred';
        }
    }

    function handleClose() {
        emit('done');
    }
</script>

<template>
    <div class="wizard max-w-lg mx-auto p-4 bg-white rounded shadow">
        <h2 class="text-xl font-bold mb-2">{{ currentStep.name }}</h2>
        <component :is="currentStep.component" :key="currentStep.name" />
        <p v-if="errorMessage" class="text-red-600 text-sm mt-2">{{ errorMessage }}</p>
        <div class="flex gap-2 mt-4">
            <button
                @click="handleNext"
                :disabled="!currentStep.isValid.value"
                class="bg-blue-600 text-white px-4 py-2 rounded disabled:opacity-50"
            >
                {{ activeStepIndex < props.steps.length - 1 ? 'Next' : 'Finish' }}
            </button>

            <button @click="handleClose" class="bg-gray-400 text-white px-4 py-2 rounded">
                Close
            </button>
        </div>
    </div>
</template>
