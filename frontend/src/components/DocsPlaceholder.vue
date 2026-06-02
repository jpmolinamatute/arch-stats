<script setup lang="ts">
import { RouterLink } from 'vue-router'
</script>

<template>
    <div class="page-shell">
        <header class="page-header">
            <RouterLink to="/" class="back-link" data-testid="docs-back-link">
                <svg class="back-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
                </svg>
                Back to Home
            </RouterLink>
        </header>

        <main class="page-content">
            <div class="placeholder-badge">
                Coming Soon
            </div>
            <h1 class="page-title">
                Documentation
            </h1>
            <p class="page-description">
                We're working on comprehensive documentation to help you get the most out of Arch Stats.
                Guides, tutorials, and API references are on the way.
            </p>
            <div class="placeholder-features">
                <div class="placeholder-item">
                    <svg class="placeholder-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
                    </svg>
                    <span>Getting Started Guide</span>
                </div>
                <div class="placeholder-item">
                    <svg class="placeholder-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01" />
                    </svg>
                    <span>Session Logging Tutorial</span>
                </div>
                <div class="placeholder-item">
                    <svg class="placeholder-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
                    </svg>
                    <span>API Reference</span>
                </div>
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
    justify-content: center;
    text-align: center;
    padding: 3rem 1.5rem;
    max-width: 40rem;
    margin: 0 auto;
}

.placeholder-badge {
    display: inline-block;
    padding: 0.25rem 0.75rem;
    border-radius: 9999px;
    font-size: 0.6875rem;
    font-weight: 700;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--landing-accent-warm);
    background: rgba(245, 158, 11, 0.1);
    border: 1px solid rgba(245, 158, 11, 0.2);
    margin-bottom: 1.5rem;
}

.page-title {
    font-size: 2.5rem;
    font-weight: 800;
    color: white;
    margin: 0 0 1rem;
    letter-spacing: -0.02em;
}

.page-description {
    font-size: 1.125rem;
    line-height: 1.75;
    color: var(--landing-text-muted);
    margin: 0 0 2.5rem;
}

.placeholder-features {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    width: 100%;
    max-width: 20rem;
}

.placeholder-item {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.875rem 1.25rem;
    border-radius: 0.75rem;
    background: var(--landing-surface);
    border: 1px solid var(--landing-border);
    font-size: 0.875rem;
    color: var(--landing-text-muted);
}

.placeholder-icon {
    width: 1.25rem;
    height: 1.25rem;
    color: var(--landing-accent);
    flex-shrink: 0;
}
</style>
