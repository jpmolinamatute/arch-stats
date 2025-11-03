<script setup lang="ts">
    import { onMounted, watch } from 'vue';
    import { useAuth } from '@/composables/useAuth';

    const { isAuthenticated, user, logout, disableGoogleAutoSelect } = useAuth();

    // If the user is already authenticated on mount, disable Google's auto-select
    onMounted(() => {
        if (isAuthenticated.value) disableGoogleAutoSelect();
    });

    // Also react to auth state changes (e.g., after login completes)
    watch(isAuthenticated, (authed) => {
        if (authed) disableGoogleAutoSelect();
    });
</script>

<template>
    <header
        class="w-full flex items-center justify-between px-4 py-3 border-b border-slate-800 bg-transparent"
    >
        <h1 class="text-lg font-semibold tracking-tight">Arch Stats</h1>
        <div class="flex items-center gap-3">
            <!-- When authenticated, show a compact identity + logout. Otherwise One Tap will prompt automatically. -->
            <div v-if="isAuthenticated" class="flex items-center gap-2">
                <span class="text-sm text-slate-300">
                    {{ user?.first_name ?? '' }} {{ user?.last_name ?? '' }}
                </span>
                <button
                    class="px-2 py-1 text-xs rounded border border-slate-700 text-slate-200 hover:bg-slate-800"
                    @click="logout"
                >
                    Logout
                </button>
            </div>
        </div>
    </header>
</template>

<style scoped></style>
