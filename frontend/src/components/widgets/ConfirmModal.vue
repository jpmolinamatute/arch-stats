<script setup lang="ts">
    defineProps<{
        show: boolean;
        title?: string;
        message?: string;
        confirmText?: string;
        cancelText?: string;
        confirmClass?: string;
    }>();

    const emit = defineEmits<{
        confirm: [];
        cancel: [];
    }>();

    function handleConfirm() {
        emit('confirm');
    }

    function handleCancel() {
        emit('cancel');
    }
</script>

<template>
    <Teleport to="body">
        <Transition name="modal">
            <div
                v-if="show"
                class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm"
                @click.self="handleCancel"
            >
                <div
                    class="bg-slate-900 border border-slate-800 rounded-lg shadow-xl max-w-md w-full p-6 space-y-4"
                    role="dialog"
                    aria-modal="true"
                >
                    <h2 class="text-xl font-semibold text-slate-100">
                        {{ title || 'Confirm Action' }}
                    </h2>

                    <p class="text-sm text-slate-400">
                        {{ message || 'Are you sure you want to proceed?' }}
                    </p>

                    <div class="flex gap-3 justify-end">
                        <button
                            @click="handleCancel"
                            class="px-4 py-2 text-sm rounded bg-slate-800 hover:bg-slate-700 text-slate-200 transition-colors duration-200"
                        >
                            {{ cancelText || 'Cancel' }}
                        </button>
                        <button
                            @click="handleConfirm"
                            :class="
                                confirmClass ||
                                'px-4 py-2 text-sm rounded bg-red-600 hover:bg-red-700 text-white transition-colors duration-200'
                            "
                        >
                            {{ confirmText || 'Confirm' }}
                        </button>
                    </div>
                </div>
            </div>
        </Transition>
    </Teleport>
</template>

<style scoped>
    .modal-enter-active,
    .modal-leave-active {
        transition: opacity 0.2s ease;
    }

    .modal-enter-from,
    .modal-leave-to {
        opacity: 0;
    }

    .modal-enter-active > div,
    .modal-leave-active > div {
        transition: transform 0.2s ease;
    }

    .modal-enter-from > div,
    .modal-leave-to > div {
        transform: scale(0.95);
    }
</style>
