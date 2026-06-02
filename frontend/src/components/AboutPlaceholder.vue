<script setup lang="ts">
import { RouterLink } from 'vue-router'
</script>

<template>
    <div class="page-shell">
        <header class="page-header">
            <RouterLink to="/" class="back-link" data-testid="about-back-link">
                <svg class="back-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
                </svg>
                Back to Home
            </RouterLink>
        </header>

        <main class="page-content">
            <div class="maker-avatar">
                🎯
            </div>

            <h1 class="page-title">
                Built by an Archer.
            </h1>

            <p class="page-description">
                Hi, I'm the solo developer behind Arch Stats, and more importantly, a passionate archer.
                I built this app because I couldn't find a tool that gave me the deep insights I needed
                to improve my game. Every feature here is built to solve real problems on the range.
            </p>

            <div class="values-grid">
                <div class="value-card">
                    <svg class="value-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
                    </svg>
                    <h3 class="value-title">
                        Open Source
                    </h3>
                    <p class="value-text">
                        The entire codebase is open. Inspect it, fork it, contribute to it.
                    </p>
                </div>
                <div class="value-card">
                    <svg class="value-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                    </svg>
                    <h3 class="value-title">
                        Privacy First
                    </h3>
                    <p class="value-text">
                        Your data stays yours. No ads, no tracking, no data selling. Ever.
                    </p>
                </div>
                <div class="value-card">
                    <svg class="value-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
                    </svg>
                    <h3 class="value-title">
                        Archer-Driven
                    </h3>
                    <p class="value-text">
                        Every feature is shaped by real archery needs, not market trends.
                    </p>
                </div>
            </div>

            <div class="canada-badge">
                <span>Proudly designed and built in Canada</span>
                <span class="flag">🇨🇦</span>
            </div>
        </main>
    </div>
</template>

<style scoped>
.page-shell {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
    background: var(--landing-bg);
    color: var(--landing-text);
}

.page-header {
    padding: 1rem 1.5rem;
    border-bottom: 1px solid var(--landing-border-subtle);
}

.back-link {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--landing-text-muted);
    text-decoration: none;
    transition: color 0.2s;
}
.back-link:hover {
    color: var(--landing-accent);
}

.back-icon {
    width: 1rem;
    height: 1rem;
}

.page-content {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    padding: 4rem 1.5rem;
    max-width: 48rem;
    margin: 0 auto;
}

.maker-avatar {
    width: 5rem;
    height: 5rem;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    background: var(--landing-surface);
    border: 2px solid var(--landing-border);
    font-size: 2rem;
    box-shadow: 0 0 40px rgba(59, 130, 246, 0.1);
    margin-bottom: 2rem;
}

.page-title {
    font-size: 2.5rem;
    font-weight: 800;
    color: white;
    margin: 0 0 1.5rem;
    letter-spacing: -0.02em;
}

.page-description {
    font-size: 1.125rem;
    line-height: 1.8;
    color: var(--landing-text-muted);
    margin: 0 0 3rem;
    max-width: 36rem;
}

.values-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(14rem, 1fr));
    gap: 1.25rem;
    width: 100%;
    margin-bottom: 3rem;
}

.value-card {
    padding: 1.5rem;
    border-radius: 1rem;
    background: var(--landing-surface);
    border: 1px solid var(--landing-border);
    text-align: left;
    transition: border-color 0.3s;
}
.value-card:hover {
    border-color: var(--landing-accent);
}

.value-icon {
    width: 1.5rem;
    height: 1.5rem;
    color: var(--landing-accent);
    margin-bottom: 0.75rem;
}

.value-title {
    font-size: 1rem;
    font-weight: 700;
    color: white;
    margin: 0 0 0.5rem;
}

.value-text {
    font-size: 0.8125rem;
    line-height: 1.6;
    color: var(--landing-text-muted);
    margin: 0;
}

.canada-badge {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.875rem;
    color: var(--landing-text-dim);
}

.flag {
    font-size: 1.125rem;
}
</style>
