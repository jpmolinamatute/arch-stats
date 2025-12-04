<script setup lang="ts">
import { onMounted } from 'vue'
import AuthGate from '@/components/auth/AuthGate.vue'
import AppHeader from '@/components/layout/AppHeader.vue'
import SessionManager from '@/components/SessionManager.vue'
import { useAuth } from '@/composables/useAuth'

const { bootstrapAuth, isAuthenticated } = useAuth()

// Initialize auth when the /app route is accessed
onMounted(async () => {
  try {
    await bootstrapAuth()
  }
  catch (e) {
    // Non-fatal; component will still render
    console.warn('Auth bootstrap failed', e)
  }
})
</script>

<template>
  <AppHeader />
  <main class="flex-1 flex flex-col items-center justify-center p-6 text-center gap-4">
    <!-- Show AuthGate until user is authenticated -->
    <template v-if="!isAuthenticated">
      <h2 class="text-2xl font-bold">
        Track Every Arrow
      </h2>
      <p class="max-w-md text-gray-600 text-sm">
        Welcome to Arch Stats. Sign in to start logging sessions and analyzing your
        performance.
      </p>
      <AuthGate />
    </template>

    <!-- Show SessionManager once authenticated -->
    <template v-else>
      <h2 class="text-2xl font-bold">
        Session Management
      </h2>
      <p class="max-w-md text-slate-400 text-sm">
        Create or manage your shooting sessions
      </p>
      <SessionManager />
    </template>
  </main>
</template>
