<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuth } from '@/composables/useAuth'
import { isEnvTrue } from '@/utils/env'

const router = useRouter()
const { loginAsDummy } = useAuth()
const isDev = isEnvTrue(import.meta.env.ARCH_STATS_DEV_MODE)

async function handleLogin() {
    if (isDev) {
        await loginAsDummy()
        router.push('/app')
    }
    else {
        router.push('/app')
    }
}
</script>

<template>
    <div class="min-h-screen flex flex-col bg-slate-950 text-slate-300 font-sans selection:bg-blue-500/30">
        <!-- Navbar -->
        <header class="w-full py-2 px-3 md:py-4 md:px-6 border-b border-slate-800/60 bg-slate-950/80 backdrop-blur-md sticky top-0 z-50">
            <div class="max-w-7xl mx-auto flex items-center justify-between gap-2">
                <div class="flex items-center gap-1.5 md:gap-3 flex-wrap">
                    <h1 class="text-base sm:text-lg md:text-2xl font-black tracking-tight text-white flex items-center gap-1 md:gap-2">
                        <svg class="w-4 h-4 sm:w-5 sm:h-5 md:w-6 md:h-6 text-blue-500 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
                        </svg>
                        Arch Stats
                    </h1>
                    <span class="px-1.5 py-0.5 rounded-full text-[8px] md:text-[10px] font-bold tracking-wider bg-blue-500/10 text-blue-400 border border-blue-500/20 uppercase whitespace-nowrap">
                        Alpha
                    </span>
                </div>
                <div>
                    <button
                        data-testid="login-button-header"
                        class="px-2.5 py-1.5 md:px-5 md:py-2 text-[10px] sm:text-xs md:text-sm font-semibold rounded-lg bg-blue-600 hover:bg-blue-500 text-white transition-all duration-300 shadow-[0_0_15px_rgba(37,99,235,0.3)] hover:shadow-[0_0_25px_rgba(37,99,235,0.5)] whitespace-nowrap"
                        @click="handleLogin"
                    >
                        {{ isDev ? 'Login' : 'Sign In' }}
                    </button>
                </div>
            </div>
        </header>

        <main class="flex-1 flex flex-col w-full max-w-full">
            <!-- Hero Section -->
            <section class="relative w-full px-3 py-12 md:px-6 md:py-32 flex flex-col items-center justify-center text-center overflow-hidden">
                <!-- Decorative background elements -->
                <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[150%] max-w-[500px] h-[150%] max-h-[500px] md:max-w-[800px] md:max-h-[800px] bg-blue-600/10 rounded-full blur-[80px] md:blur-[120px] opacity-50 pointer-events-none" />
                <div class="absolute top-0 right-0 w-[120%] max-w-[300px] h-[120%] max-h-[300px] md:max-w-[400px] md:max-h-[400px] bg-indigo-500/10 rounded-full blur-[60px] md:blur-[80px] opacity-30 pointer-events-none" />

                <div class="relative z-10 max-w-4xl space-y-5 md:space-y-8 w-full">
                    <div class="inline-flex items-center gap-1.5 md:gap-2 px-2.5 py-1 md:px-3 md:py-1 rounded-full bg-slate-900 border border-slate-800 mb-2 md:mb-4 animate-fade-in-up text-[10px] sm:text-xs md:text-sm max-w-[90vw] text-left">
                        <span class="w-1.5 h-1.5 md:w-2 md:h-2 rounded-full bg-emerald-500 animate-pulse shrink-0" />
                        <span class="text-slate-400 break-words line-clamp-2">Arch Stats is currently in Alpha</span>
                    </div>

                    <h2 class="text-4xl sm:text-5xl md:text-7xl lg:text-8xl font-black text-transparent bg-clip-text bg-gradient-to-br from-white via-slate-200 to-slate-500 tracking-tight leading-tight mix-blend-plus-lighter px-1 sm:px-2 break-words text-wrap">
                        Track Every <br> Arrow.
                    </h2>

                    <p class="text-sm sm:text-base md:text-xl lg:text-2xl text-slate-400 max-w-2xl mx-auto font-light leading-relaxed px-2 sm:px-4">
                        Professional archery performance tracking. Log your practice, analyze deep metrics, and continuously improve your game.
                    </p>

                    <div class="pt-5 md:pt-8 w-full px-2 sm:px-0">
                        <button
                            data-testid="login-button-hero"
                            class="w-full sm:w-auto px-4 py-3 sm:px-6 sm:py-4 md:px-10 md:py-5 text-sm md:text-lg font-bold rounded-2xl bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-500 hover:to-indigo-500 text-white transition-all duration-300 shadow-[0_0_20px_rgba(37,99,235,0.4)] md:shadow-[0_0_30px_rgba(37,99,235,0.4)] hover:shadow-[0_0_30px_rgba(37,99,235,0.6)] hover:-translate-y-1 block mx-auto break-words"
                            @click="handleLogin"
                        >
                            {{ isDev ? 'Try the Alpha' : 'Start Tracking Now' }}
                        </button>
                    </div>
                </div>
            </section>

            <!-- Core Features -->
            <section class="w-full px-3 py-12 md:px-6 md:py-24 bg-slate-900 border-y border-slate-800/50">
                <div class="max-w-7xl mx-auto space-y-10 md:space-y-16">
                    <div class="text-center space-y-3 md:space-y-4 px-2 sm:px-4">
                        <h3 class="text-2xl sm:text-3xl md:text-4xl font-bold text-white leading-tight">
                            What You Can Do Today
                        </h3>
                        <p class="text-slate-400 max-w-xl mx-auto text-sm sm:text-base md:text-lg">
                            The foundational tools you need to organize your practice sessions.
                        </p>
                    </div>

                    <div class="grid grid-cols-1 md:grid-cols-3 gap-5 sm:gap-6 md:gap-8">
                        <!-- Card 1 -->
                        <div class="group p-5 sm:p-6 md:p-8 rounded-2xl sm:rounded-3xl bg-slate-950/50 border border-slate-800 hover:border-slate-700 transition-colors backdrop-blur-sm mx-1 sm:mx-0">
                            <div class="w-10 h-10 md:w-12 md:h-12 bg-blue-500/10 rounded-xl flex items-center justify-center mb-4 md:mb-6 text-blue-400 group-hover:scale-110 transition-transform">
                                <svg class="w-5 h-5 md:w-6 md:h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" /></svg>
                            </div>
                            <h4 class="text-lg sm:text-xl font-bold text-white mb-2 md:mb-3">
                                Track Sessions
                            </h4>
                            <p class="text-slate-400 leading-relaxed text-xs sm:text-sm md:text-base">
                                Record every practice session with detailed shot data and environmental conditions. Never lose a score sheet again.
                            </p>
                        </div>

                        <!-- Card 2 -->
                        <div class="group p-5 sm:p-6 md:p-8 rounded-2xl sm:rounded-3xl bg-slate-950/50 border border-slate-800 hover:border-slate-700 transition-colors backdrop-blur-sm mx-1 sm:mx-0">
                            <div class="w-10 h-10 md:w-12 md:h-12 bg-indigo-500/10 rounded-xl flex items-center justify-center mb-4 md:mb-6 text-indigo-400 group-hover:scale-110 transition-transform">
                                <svg class="w-5 h-5 md:w-6 md:h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 12l3-3 3 3 4-4M8 21l4-4 4 4M3 4h18M4 4h16v12a1 1 0 01-1 1H5a1 1 0 01-1-1V4z" /></svg>
                            </div>
                            <h4 class="text-lg sm:text-xl font-bold text-white mb-2 md:mb-3">
                                Analyze Performance
                            </h4>
                            <p class="text-slate-400 leading-relaxed text-xs sm:text-sm md:text-base">
                                View detailed statistics, trends, and real-time insights during your sessions to identify areas for immediate improvement.
                            </p>
                        </div>

                        <!-- Card 3 -->
                        <div class="group p-5 sm:p-6 md:p-8 rounded-2xl sm:rounded-3xl bg-slate-950/50 border border-slate-800 hover:border-slate-700 transition-colors backdrop-blur-sm mx-1 sm:mx-0">
                            <div class="w-10 h-10 md:w-12 md:h-12 bg-emerald-500/10 rounded-xl flex items-center justify-center mb-4 md:mb-6 text-emerald-400 group-hover:scale-110 transition-transform">
                                <svg class="w-5 h-5 md:w-6 md:h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" /></svg>
                            </div>
                            <h4 class="text-lg sm:text-xl font-bold text-white mb-2 md:mb-3">
                                Improve Your Game
                            </h4>
                            <p class="text-slate-400 leading-relaxed text-xs sm:text-sm md:text-base">
                                Set concrete goals, monitor your progress over an entire season, and take your archery skills consistently to the next level.
                            </p>
                        </div>
                    </div>
                </div>
            </section>

            <!-- On the Horizon (Roadmap) -->
            <section class="relative w-full px-3 py-12 md:px-6 md:py-24 overflow-hidden">
                <div class="absolute inset-0 bg-[radial-gradient(ellipse_at_top,_var(--tw-gradient-stops))] from-slate-900 via-slate-950 to-slate-950" />
                <div class="relative z-10 max-w-5xl mx-auto space-y-10 md:space-y-16">
                    <div class="text-center space-y-2 md:space-y-4 px-2 sm:px-4">
                        <span class="text-blue-500 font-bold uppercase tracking-widest text-[10px] sm:text-xs md:text-sm">The Roadmap</span>
                        <h3 class="text-2xl sm:text-3xl md:text-5xl font-bold text-white leading-tight">
                            On the Horizon
                        </h3>
                        <p class="text-slate-400 max-w-2xl mx-auto text-sm sm:text-base md:text-lg">
                            We are building the ultimate analytical engine for archers. Here is a sneak peek at what is coming next.
                        </p>
                    </div>

                    <div class="space-y-5 md:space-y-8 px-1 sm:px-0">
                        <!-- Roadmap Item 1 -->
                        <div class="flex flex-col md:flex-row gap-5 md:gap-8 items-start p-5 sm:p-6 md:p-10 rounded-2xl sm:rounded-3xl bg-slate-900/40 border border-slate-800/80 backdrop-blur-md relative overflow-hidden group mx-1 sm:mx-0">
                            <div class="absolute right-0 top-0 w-32 h-32 sm:w-48 sm:h-48 md:w-64 md:h-64 bg-blue-600/5 rounded-full blur-2xl sm:blur-3xl group-hover:bg-blue-600/10 transition-colors" />

                            <div class="flex-shrink-0 w-12 h-12 md:w-16 md:h-16 bg-slate-800 rounded-xl md:rounded-2xl flex items-center justify-center text-blue-400 border border-slate-700">
                                <svg class="w-6 h-6 md:w-8 md:h-8" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" /></svg>
                            </div>
                            <div class="space-y-3 sm:space-y-4 flex-1">
                                <h4 class="text-lg sm:text-xl md:text-2xl font-bold text-slate-100">
                                    Deep Historical Analysis
                                </h4>
                                <p class="text-slate-400 leading-relaxed text-xs sm:text-sm md:text-lg">
                                    Uncover true trends in your performance. You'll be able to track results across different locations, distances, and target faces.
                                </p>
                                <ul class="grid grid-cols-1 sm:grid-cols-2 gap-2 md:gap-4 text-slate-300 text-xs sm:text-sm md:text-base">
                                    <li class="flex items-start gap-2 sm:gap-3">
                                        <svg class="w-4 h-4 md:w-5 md:h-5 text-blue-500 shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" /></svg>
                                        <span>Track X's vs Misses</span>
                                    </li>
                                    <li class="flex items-start gap-2 sm:gap-3">
                                        <svg class="w-4 h-4 md:w-5 md:h-5 text-blue-500 shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" /></svg>
                                        <span>Customidable chart visualizations</span>
                                    </li>
                                    <li class="flex items-start gap-2 sm:gap-3">
                                        <svg class="w-4 h-4 md:w-5 md:h-5 text-blue-500 shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" /></svg>
                                        <span>Zoom time (Year/Month/Week)</span>
                                    </li>
                                    <li class="flex items-start gap-2 sm:gap-3">
                                        <svg class="w-4 h-4 md:w-5 md:h-5 text-blue-500 shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" /></svg>
                                        <span>Full CSV Data Export</span>
                                    </li>
                                </ul>
                            </div>
                        </div>

                        <!-- Roadmap Item 2 -->
                        <div class="flex flex-col md:flex-row gap-5 md:gap-8 items-start p-5 sm:p-6 md:p-10 rounded-2xl sm:rounded-3xl bg-slate-900/40 border border-slate-800/80 backdrop-blur-md relative overflow-hidden group mx-1 sm:mx-0">
                            <div class="absolute left-0 bottom-0 w-32 h-32 sm:w-48 sm:h-48 md:w-64 md:h-64 bg-indigo-600/5 rounded-full blur-2xl sm:blur-3xl group-hover:bg-indigo-600/10 transition-colors" />

                            <div class="flex-shrink-0 w-12 h-12 md:w-16 md:h-16 bg-slate-800 rounded-xl md:rounded-2xl flex items-center justify-center text-indigo-400 border border-slate-700">
                                <svg class="w-6 h-6 md:w-8 md:h-8" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3" /></svg>
                            </div>
                            <div class="space-y-2 sm:space-y-3 md:space-y-4 flex-1">
                                <h4 class="text-lg sm:text-xl md:text-2xl font-bold text-slate-100">
                                    Arrow Lifecycle Tracking
                                </h4>
                                <p class="text-slate-400 leading-relaxed text-xs sm:text-sm md:text-lg">
                                    Not all arrows are created equal. Track exactly how each individual arrow performs over its lifespan. Register the number of arrows per session, and precisely document when an arrow is damaged, lost, or replaced to keep your inventory accurate.
                                </p>
                            </div>
                        </div>
                    </div>
                </div>
            </section>

            <!-- Maker Section -->
            <section class="w-full px-3 py-12 md:px-6 md:py-24 bg-slate-950 border-t border-slate-800">
                <div class="max-w-3xl mx-auto text-center space-y-5 sm:space-y-6 md:space-y-8 px-2 sm:px-0">
                    <div class="inline-flex items-center justify-center w-12 h-12 sm:w-16 sm:h-16 md:w-20 md:h-20 rounded-full bg-slate-900 border-2 border-slate-800 shadow-xl mb-1 sm:mb-2 md:mb-4 text-xl sm:text-2xl md:text-3xl">
                        🎯
                    </div>
                    <h3 class="text-xl sm:text-2xl md:text-3xl font-bold text-white">
                        Built by an Archer.
                    </h3>
                    <p class="text-sm sm:text-base md:text-xl text-slate-400 leading-relaxed px-1 sm:px-2 md:px-0">
                        Hi, I'm the solo developer behind Arch Stats, and more importantly, a passionate archer. I built this app because I couldn't find a tool that gave me the deep insights I needed to improve my game. Every feature here is built to solve real problems on the range.
                    </p>
                </div>
            </section>
        </main>

        <!-- Footer -->
        <footer class="w-full py-5 sm:py-6 md:py-8 px-3 sm:px-4 md:px-6 border-t border-slate-800/80 bg-slate-950 text-slate-500 text-center md:text-left">
            <div class="max-w-7xl mx-auto flex flex-col md:flex-row items-center justify-between gap-3 sm:gap-4 md:gap-4">
                <p class="text-[10px] sm:text-xs md:text-sm">
                    © {{ new Date().getFullYear() }} Arch Stats. All rights reserved.
                </p>
                <div class="flex items-center justify-center gap-1.5 sm:gap-2 text-[10px] sm:text-xs md:text-sm">
                    <span>Proudly designed and built in Canada</span>
                    <span class="text-sm sm:text-base md:text-lg">🇨🇦</span>
                </div>
            </div>
        </footer>
    </div>
</template>

<style scoped>
/* Custom animations for the landing page */
@keyframes fadeInUp {
    from {
        opacity: 0;
        transform: translateY(20px);
    }
    to {
        opacity: 1;
        transform: translateY(0);
    }
}
.animate-fade-in-up {
    animation: fadeInUp 0.8s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}
</style>
