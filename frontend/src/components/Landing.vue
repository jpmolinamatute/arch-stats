<script setup lang="ts">
    import { useRouter } from 'vue-router';

    const router = useRouter();

    function goToApp() {
        router.push('/app');
    }
</script>

<template>
    <div class="min-h-screen flex flex-col">
        <header class="w-full py-4 px-6 border-b border-slate-800">
            <div class="max-w-6xl mx-auto flex items-center justify-between">
                <h1 class="text-xl font-bold text-slate-100">Arch Stats</h1>
            </div>
        </header>

        <main class="flex-1 flex flex-col items-center justify-center p-6 text-center">
            <div class="max-w-3xl space-y-8">
                <div class="space-y-4">
                    <h2 class="text-4xl md:text-5xl font-bold text-slate-100">Track Every Arrow</h2>
                    <p class="text-lg md:text-xl text-slate-400 max-w-2xl mx-auto">
                        Professional archery performance tracking and analysis. Log your sessions,
                        analyze your performance, and improve your game.
                    </p>
                </div>

                <div class="space-y-6">
                    <button
                        @click="goToApp"
                        class="px-8 py-4 text-lg font-semibold rounded-lg bg-blue-600 hover:bg-blue-700 text-white transition-colors duration-200 shadow-lg hover:shadow-xl"
                    >
                        Login
                    </button>

                    <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mt-12 text-left">
                        <div class="p-6 rounded-lg border border-slate-800 bg-slate-900/50">
                            <h3 class="text-lg font-semibold text-slate-100 mb-2">
                                Track Sessions
                            </h3>
                            <p class="text-sm text-slate-400">
                                Record every practice session with detailed shot data and
                                environmental conditions.
                            </p>
                        </div>

                        <div class="p-6 rounded-lg border border-slate-800 bg-slate-900/50">
                            <h3 class="text-lg font-semibold text-slate-100 mb-2">
                                Analyze Performance
                            </h3>
                            <p class="text-sm text-slate-400">
                                View detailed statistics, trends, and insights to identify areas for
                                improvement.
                            </p>
                        </div>

                        <div class="p-6 rounded-lg border border-slate-800 bg-slate-900/50">
                            <h3 class="text-lg font-semibold text-slate-100 mb-2">
                                Improve Your Game
                            </h3>
                            <p class="text-sm text-slate-400">
                                Set goals, monitor progress, and take your archery skills to the
                                next level.
                            </p>
                        </div>
                    </div>
                </div>
            </div>
        </main>

        <footer
            class="w-full py-6 px-6 border-t border-slate-800 text-center text-sm text-slate-500"
        >
            <p>© 2025 Arch Stats. All rights reserved.</p>
        </footer>
    </div>
</template>
