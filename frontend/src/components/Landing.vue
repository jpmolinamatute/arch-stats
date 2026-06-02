<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { useAuth } from '@/composables/useAuth'
import { isEnvTrue } from '@/utils/env'

const router = useRouter()
const { loginAsDummy } = useAuth()
const isDev = isEnvTrue(import.meta.env.ARCH_STATS_DEV_MODE)
const mobileMenuOpen = ref(false)

async function handleLogin() {
    if (isDev) {
        await loginAsDummy()
        router.push('/app')
    }
    else {
        router.push('/app')
    }
}

function toggleMobileMenu() {
    mobileMenuOpen.value = !mobileMenuOpen.value
}

function closeMobileMenu() {
    mobileMenuOpen.value = false
}
</script>

<template>
    <div class="landing">
        <!-- ═══════════════════════════════════════════
             NAVIGATION BAR
             ═══════════════════════════════════════════ -->
        <header class="nav">
            <div class="nav__container">
                <div class="nav__brand">
                    <h1 class="nav__logo">
                        <svg class="nav__logo-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
                        </svg>
                        Arch Stats
                    </h1>
                    <span class="nav__badge">Alpha</span>
                </div>

                <!-- Desktop nav links -->
                <nav class="nav__links" aria-label="Main navigation">
                    <RouterLink to="/docs" class="nav__link" data-testid="nav-docs">
                        Docs
                    </RouterLink>
                    <RouterLink to="/about" class="nav__link" data-testid="nav-about">
                        About
                    </RouterLink>
                    <RouterLink to="/feedback" class="nav__link" data-testid="nav-feedback">
                        Feedback
                    </RouterLink>
                    <button
                        class="nav__cta"
                        data-testid="login-button-header"
                        @click="handleLogin"
                    >
                        {{ isDev ? 'Login' : 'Sign In' }}
                    </button>
                </nav>

                <!-- Mobile hamburger button -->
                <button
                    class="nav__hamburger"
                    data-testid="mobile-menu-toggle"
                    :aria-expanded="mobileMenuOpen"
                    aria-controls="mobile-menu"
                    aria-label="Toggle navigation menu"
                    @click="toggleMobileMenu"
                >
                    <svg v-if="!mobileMenuOpen" class="nav__hamburger-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
                    </svg>
                    <svg v-else class="nav__hamburger-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                </button>
            </div>

            <!-- Mobile menu panel -->
            <Transition name="slide-down">
                <nav
                    v-if="mobileMenuOpen"
                    id="mobile-menu"
                    class="nav__mobile"
                    aria-label="Mobile navigation"
                >
                    <RouterLink to="/docs" class="nav__mobile-link" data-testid="mobile-nav-docs" @click="closeMobileMenu">
                        Docs
                    </RouterLink>
                    <RouterLink to="/about" class="nav__mobile-link" data-testid="mobile-nav-about" @click="closeMobileMenu">
                        About
                    </RouterLink>
                    <RouterLink to="/feedback" class="nav__mobile-link" data-testid="mobile-nav-feedback" @click="closeMobileMenu">
                        Feedback
                    </RouterLink>
                    <button
                        class="nav__mobile-cta"
                        data-testid="login-button-mobile"
                        @click="handleLogin"
                    >
                        {{ isDev ? 'Login' : 'Sign In' }}
                    </button>
                </nav>
            </Transition>
        </header>

        <main class="landing__main">
            <!-- ═══════════════════════════════════════════
                 HERO SECTION
                 ═══════════════════════════════════════════ -->
            <section class="hero">
                <div class="hero__glow hero__glow--primary" />
                <div class="hero__glow hero__glow--secondary" />

                <div class="hero__content">
                    <div class="hero__pill">
                        <span class="hero__pill-dot" />
                        <span>Arch Stats is currently in Alpha</span>
                    </div>

                    <h2 class="hero__title">
                        Track Every <br>Arrow.
                    </h2>

                    <p class="hero__subtitle">
                        Take control of your archery journey. Log accurate session data, turn it into
                        measurable progress, and always retain full ownership of your stats.
                    </p>

                    <div class="hero__actions">
                        <button
                            class="hero__cta"
                            data-testid="login-button-hero"
                            @click="handleLogin"
                        >
                            {{ isDev ? 'Try the Alpha' : 'Start Tracking Now' }}
                        </button>
                    </div>
                </div>
            </section>

            <!-- ═══════════════════════════════════════════
                 FEATURES & BENEFITS
                 ═══════════════════════════════════════════ -->
            <section class="features">
                <div class="features__container">
                    <div class="section-header">
                        <h3 class="section-header__title">
                            Everything You Need to Improve
                        </h3>
                        <p class="section-header__subtitle">
                            The tools you need to organize your practice and level up your scores.
                        </p>
                    </div>

                    <div class="features__grid">
                        <!-- Frictionless Data Entry -->
                        <div class="feature-card">
                            <div class="feature-card__icon feature-card__icon--blue">
                                <svg fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
                                </svg>
                            </div>
                            <h4 class="feature-card__title">
                                Frictionless Data Entry
                            </h4>
                            <p class="feature-card__text">
                                Record sessions with minimal taps. No more lost score sheets or
                                forgotten details from the range.
                            </p>
                        </div>

                        <!-- Detailed Session Analytics -->
                        <div class="feature-card">
                            <div class="feature-card__icon feature-card__icon--indigo">
                                <svg fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                                </svg>
                            </div>
                            <h4 class="feature-card__title">
                                Detailed Session Analytics
                            </h4>
                            <p class="feature-card__text">
                                Real-time stats and trends during your sessions to spot what's
                                working and what needs attention.
                            </p>
                        </div>

                        <!-- Full Data Ownership -->
                        <div class="feature-card">
                            <div class="feature-card__icon feature-card__icon--emerald">
                                <svg fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                                </svg>
                            </div>
                            <h4 class="feature-card__title">
                                Full Data Ownership
                            </h4>
                            <p class="feature-card__text">
                                Your data is yours. Export everything as CSV anytime you want.
                                No lock-in, no restrictions.
                            </p>
                        </div>
                    </div>
                </div>
            </section>

            <!-- ═══════════════════════════════════════════
                 PRICING & OPEN SOURCE
                 ═══════════════════════════════════════════ -->
            <section class="pricing">
                <div class="pricing__container">
                    <div class="section-header">
                        <h3 class="section-header__title">
                            Free. Forever. Open Source.
                        </h3>
                        <p class="section-header__subtitle">
                            No subscriptions, no paywalls, no fine print.
                        </p>
                    </div>

                    <div class="pricing__grid">
                        <!-- Free Forever card -->
                        <div class="pricing-card pricing-card--featured">
                            <div class="pricing-card__header">
                                <span class="pricing-card__label">Web App</span>
                                <span class="pricing-card__price">Free Forever</span>
                            </div>
                            <ul class="pricing-card__list">
                                <li class="pricing-card__item">
                                    <svg class="pricing-card__check" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" /></svg>
                                    Unlimited sessions
                                </li>
                                <li class="pricing-card__item">
                                    <svg class="pricing-card__check" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" /></svg>
                                    Full analytics dashboard
                                </li>
                                <li class="pricing-card__item">
                                    <svg class="pricing-card__check" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" /></svg>
                                    CSV data export
                                </li>
                                <li class="pricing-card__item">
                                    <svg class="pricing-card__check" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" /></svg>
                                    100% open source
                                </li>
                            </ul>
                            <a
                                href="https://github.com/jpmolinamatute/arch-stats"
                                target="_blank"
                                rel="noopener noreferrer"
                                class="pricing-card__github"
                                data-testid="github-link"
                            >
                                <svg class="pricing-card__github-icon" viewBox="0 0 24 24" fill="currentColor">
                                    <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" />
                                </svg>
                                View on GitHub
                            </a>
                        </div>

                        <!-- Mobile Coming Soon card -->
                        <div class="pricing-card pricing-card--coming-soon">
                            <div class="pricing-card__header">
                                <span class="pricing-card__label">Mobile App</span>
                                <span class="pricing-card__price pricing-card__price--muted">Coming Soon</span>
                            </div>
                            <ul class="pricing-card__list">
                                <li class="pricing-card__item pricing-card__item--muted">
                                    <svg class="pricing-card__check pricing-card__check--muted" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" /></svg>
                                    Native iOS &amp; Android
                                </li>
                                <li class="pricing-card__item pricing-card__item--muted">
                                    <svg class="pricing-card__check pricing-card__check--muted" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" /></svg>
                                    Compete against other archers.
                                </li>
                                <li class="pricing-card__item pricing-card__item--muted">
                                    <svg class="pricing-card__check pricing-card__check--muted" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" /></svg>
                                    Join Clubes
                                </li>
                                <li class="pricing-card__item pricing-card__item--muted">
                                    <svg class="pricing-card__check pricing-card__check--muted" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" /></svg>
                                    Manage Clubes
                                </li>
                                <li class="pricing-card__item pricing-card__item--muted">
                                    <svg class="pricing-card__check pricing-card__check--muted" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" /></svg>
                                    New Coach and Club admin users
                                </li>
                            </ul>
                            <div class="pricing-card__teaser">
                                Stay tuned for updates
                            </div>
                        </div>
                    </div>
                </div>
            </section>

            <!-- ═══════════════════════════════════════════
                 WHAT'S COMING NEXT (Trimmed Roadmap)
                 ═══════════════════════════════════════════ -->
            <section class="roadmap">
                <div class="roadmap__container">
                    <div class="section-header">
                        <span class="section-header__eyebrow">The Roadmap</span>
                        <h3 class="section-header__title">
                            What's Coming Next
                        </h3>
                    </div>

                    <ul class="roadmap__list">
                        <li class="roadmap__item">
                            <svg class="roadmap__icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                            </svg>
                            <div>
                                <strong>Deep Historical Analysis</strong>
                                <span class="roadmap__desc"> — Track trends across seasons, distances, and target faces.</span>
                            </div>
                        </li>
                        <li class="roadmap__item">
                            <svg class="roadmap__icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3" />
                            </svg>
                            <div>
                                <strong>Arrow Lifecycle Tracking</strong>
                                <span class="roadmap__desc"> — Monitor individual arrow performance and inventory over time.</span>
                            </div>
                        </li>
                        <li class="roadmap__item">
                            <svg class="roadmap__icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                            </svg>
                            <div>
                                <strong>Full CSV Data Export</strong>
                                <span class="roadmap__desc"> — Take your data anywhere, anytime.</span>
                            </div>
                        </li>
                    </ul>
                </div>
            </section>
        </main>

        <!-- ═══════════════════════════════════════════
             FOOTER
             ═══════════════════════════════════════════ -->
        <footer class="footer">
            <div class="footer__container">
                <p class="footer__copyright">
                    © {{ new Date().getFullYear() }} Arch Stats. All rights reserved.
                </p>
                <nav class="footer__links" aria-label="Footer navigation">
                    <RouterLink to="/docs" class="footer__link">
                        Docs
                    </RouterLink>
                    <RouterLink to="/about" class="footer__link">
                        About
                    </RouterLink>
                    <RouterLink to="/feedback" class="footer__link">
                        Feedback
                    </RouterLink>
                </nav>
                <div class="footer__canada">
                    <span>Proudly designed and built in Canada</span>
                    <span class="footer__flag">🇨🇦</span>
                </div>
            </div>
        </footer>
    </div>
</template>

<style scoped>
/* ════════════════════════════════════════════════════════
   LANDING PAGE — Component Styles
   Uses CSS custom properties from style.css (--landing-*)
   ════════════════════════════════════════════════════════ */

.landing {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
    background: var(--landing-bg);
    color: var(--landing-text);
}

/* ─── NAVIGATION ─── */

.nav {
    position: sticky;
    top: 0;
    z-index: 50;
    width: 100%;
    background: rgba(10, 15, 30, 0.85);
    backdrop-filter: blur(16px);
    border-bottom: 1px solid var(--landing-border-subtle);
}

.nav__container {
    max-width: 72rem;
    margin: 0 auto;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.75rem 1.5rem;
}

.nav__brand {
    display: flex;
    align-items: center;
    gap: 0.75rem;
}

.nav__logo {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 1.25rem;
    font-weight: 900;
    color: white;
    letter-spacing: -0.02em;
    margin: 0;
}

.nav__logo-icon {
    width: 1.375rem;
    height: 1.375rem;
    color: var(--landing-accent);
}

.nav__badge {
    padding: 0.125rem 0.5rem;
    border-radius: 9999px;
    font-size: 0.5625rem;
    font-weight: 700;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--landing-accent);
    background: rgba(59, 130, 246, 0.1);
    border: 1px solid rgba(59, 130, 246, 0.2);
}

.nav__links {
    display: flex;
    align-items: center;
    gap: 2rem;
}

.nav__link {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--landing-text-muted);
    text-decoration: none;
    transition: color 0.2s;
}
.nav__link:hover {
    color: white;
}

.nav__cta {
    padding: 0.5rem 1.25rem;
    border: none;
    border-radius: 9999px;
    font-size: 0.8125rem;
    font-weight: 600;
    font-family: inherit;
    color: white;
    background: var(--landing-accent);
    cursor: pointer;
    transition: background 0.2s, box-shadow 0.2s;
    box-shadow: 0 0 15px rgba(59, 130, 246, 0.25);
}
.nav__cta:hover {
    background: var(--landing-accent-hover);
    box-shadow: 0 0 25px rgba(59, 130, 246, 0.45);
}

.nav__hamburger {
    display: none;
    padding: 0.5rem;
    border: none;
    border-radius: 0.5rem;
    background: transparent;
    color: var(--landing-text-muted);
    cursor: pointer;
    transition: color 0.2s;
}
.nav__hamburger:hover {
    color: white;
}

.nav__hamburger-icon {
    width: 1.5rem;
    height: 1.5rem;
}

.nav__mobile {
    display: none;
    flex-direction: column;
    gap: 0.25rem;
    padding: 0.75rem 1.5rem 1rem;
    border-top: 1px solid var(--landing-border-subtle);
}

.nav__mobile-link {
    display: block;
    padding: 0.625rem 0.75rem;
    border-radius: 0.5rem;
    font-size: 0.9375rem;
    font-weight: 500;
    color: var(--landing-text-muted);
    text-decoration: none;
    transition: background 0.2s, color 0.2s;
}
.nav__mobile-link:hover {
    background: var(--landing-surface);
    color: white;
}

.nav__mobile-cta {
    margin-top: 0.5rem;
    padding: 0.75rem;
    border: none;
    border-radius: 0.75rem;
    font-size: 0.9375rem;
    font-weight: 600;
    font-family: inherit;
    color: white;
    background: var(--landing-accent);
    cursor: pointer;
    text-align: center;
}

@media (max-width: 768px) {
    .nav__links {
        display: none;
    }
    .nav__hamburger {
        display: block;
    }
    .nav__mobile {
        display: flex;
    }
}

/* Mobile menu slide transition */
.slide-down-enter-active,
.slide-down-leave-active {
    transition: all 0.25s cubic-bezier(0.16, 1, 0.3, 1);
    overflow: hidden;
}
.slide-down-enter-from,
.slide-down-leave-to {
    opacity: 0;
    max-height: 0;
    padding-top: 0;
    padding-bottom: 0;
}
.slide-down-enter-to,
.slide-down-leave-from {
    opacity: 1;
    max-height: 20rem;
}

/* ─── MAIN ─── */

.landing__main {
    flex: 1;
    display: flex;
    flex-direction: column;
    width: 100%;
}

/* ─── HERO ─── */

.hero {
    position: relative;
    width: 100%;
    padding: 5rem 1.5rem;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    overflow: hidden;
}

.hero__glow {
    position: absolute;
    border-radius: 50%;
    pointer-events: none;
}
.hero__glow--primary {
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    width: min(500px, 150%);
    height: min(500px, 150%);
    background: rgba(59, 130, 246, 0.08);
    filter: blur(100px);
    opacity: 0.6;
}
.hero__glow--secondary {
    top: 0;
    right: -10%;
    width: min(350px, 120%);
    height: min(350px, 120%);
    background: rgba(245, 158, 11, 0.06);
    filter: blur(80px);
    opacity: 0.4;
}

.hero__content {
    position: relative;
    z-index: 10;
    max-width: 48rem;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1.5rem;
}

.hero__pill {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.375rem 0.875rem;
    border-radius: 9999px;
    background: var(--landing-surface);
    border: 1px solid var(--landing-border);
    font-size: 0.75rem;
    color: var(--landing-text-muted);
    animation: fadeInUp 0.8s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}

.hero__pill-dot {
    width: 0.5rem;
    height: 0.5rem;
    border-radius: 50%;
    background: #10b981;
    animation: pulse 2s ease-in-out infinite;
}

.hero__title {
    font-size: clamp(2.5rem, 8vw, 5rem);
    font-weight: 900;
    letter-spacing: -0.03em;
    line-height: 1.05;
    color: transparent;
    background: linear-gradient(135deg, #ffffff, #c8cdd8, #8891a5);
    background-clip: text;
    -webkit-background-clip: text;
    margin: 0;
}

.hero__subtitle {
    font-size: clamp(0.9375rem, 2vw, 1.25rem);
    line-height: 1.7;
    color: var(--landing-text-muted);
    max-width: 36rem;
    margin: 0;
    font-weight: 300;
}

.hero__actions {
    padding-top: 1.5rem;
    width: 100%;
    display: flex;
    justify-content: center;
}

.hero__cta {
    padding: 1rem 2.5rem;
    border: none;
    border-radius: 1rem;
    font-size: 1rem;
    font-weight: 700;
    font-family: inherit;
    color: white;
    background: linear-gradient(135deg, var(--landing-gradient-start), var(--landing-gradient-end));
    cursor: pointer;
    transition: transform 0.2s, box-shadow 0.2s;
    box-shadow: 0 0 25px rgba(59, 130, 246, 0.35);
}
.hero__cta:hover {
    transform: translateY(-2px);
    box-shadow: 0 0 40px rgba(59, 130, 246, 0.55);
}

/* ─── FEATURES ─── */

.features {
    width: 100%;
    padding: 5rem 1.5rem;
    background: var(--landing-bg-alt);
    border-top: 1px solid var(--landing-border-subtle);
    border-bottom: 1px solid var(--landing-border-subtle);
}

.features__container {
    max-width: 72rem;
    margin: 0 auto;
}

.features__grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 1.5rem;
    margin-top: 3rem;
}

.feature-card {
    padding: 2rem;
    border-radius: 1.25rem;
    background: var(--landing-surface);
    border: 1px solid var(--landing-border);
    transition: border-color 0.3s, transform 0.3s;
}
.feature-card:hover {
    border-color: rgba(59, 130, 246, 0.3);
    transform: translateY(-2px);
}

.feature-card__icon {
    width: 2.75rem;
    height: 2.75rem;
    border-radius: 0.75rem;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-bottom: 1.25rem;
    transition: transform 0.3s;
}
.feature-card:hover .feature-card__icon {
    transform: scale(1.1);
}
.feature-card__icon svg {
    width: 1.375rem;
    height: 1.375rem;
}
.feature-card__icon--blue {
    background: rgba(59, 130, 246, 0.1);
    color: #60a5fa;
}
.feature-card__icon--indigo {
    background: rgba(99, 102, 241, 0.1);
    color: #818cf8;
}
.feature-card__icon--emerald {
    background: rgba(16, 185, 129, 0.1);
    color: #34d399;
}

.feature-card__title {
    font-size: 1.125rem;
    font-weight: 700;
    color: white;
    margin: 0 0 0.625rem;
}

.feature-card__text {
    font-size: 0.875rem;
    line-height: 1.7;
    color: var(--landing-text-muted);
    margin: 0;
}

/* ─── PRICING ─── */

.pricing {
    width: 100%;
    padding: 5rem 1.5rem;
}

.pricing__container {
    max-width: 52rem;
    margin: 0 auto;
}

.pricing__grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 1.5rem;
    margin-top: 3rem;
}

.pricing-card {
    padding: 2rem;
    border-radius: 1.25rem;
    background: var(--landing-surface);
    border: 1px solid var(--landing-border);
    display: flex;
    flex-direction: column;
}

.pricing-card--featured {
    border-color: rgba(245, 158, 11, 0.35);
    box-shadow: 0 0 30px rgba(245, 158, 11, 0.08);
}

.pricing-card--coming-soon {
    opacity: 0.65;
}

.pricing-card__header {
    margin-bottom: 1.5rem;
}

.pricing-card__label {
    display: block;
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--landing-text-muted);
    margin-bottom: 0.375rem;
}

.pricing-card__price {
    font-size: 1.5rem;
    font-weight: 800;
    color: white;
}
.pricing-card__price--muted {
    color: var(--landing-text-muted);
}

.pricing-card__list {
    list-style: none;
    padding: 0;
    margin: 0 0 1.5rem;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    flex: 1;
}

.pricing-card__item {
    display: flex;
    align-items: center;
    gap: 0.625rem;
    font-size: 0.875rem;
    color: var(--landing-text);
}
.pricing-card__item--muted {
    color: var(--landing-text-muted);
}

.pricing-card__check {
    width: 1.125rem;
    height: 1.125rem;
    color: var(--landing-accent-warm);
    flex-shrink: 0;
}
.pricing-card__check--muted {
    color: var(--landing-text-dim);
}

.pricing-card__github {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    padding: 0.625rem 1rem;
    border-radius: 0.625rem;
    font-size: 0.8125rem;
    font-weight: 600;
    color: var(--landing-text);
    background: var(--landing-surface-raised);
    border: 1px solid var(--landing-border);
    text-decoration: none;
    transition: border-color 0.2s, background 0.2s;
}
.pricing-card__github:hover {
    border-color: var(--landing-text-muted);
    background: rgba(255, 255, 255, 0.05);
}

.pricing-card__github-icon {
    width: 1rem;
    height: 1rem;
}

.pricing-card__teaser {
    text-align: center;
    font-size: 0.8125rem;
    font-weight: 500;
    color: var(--landing-text-dim);
    padding: 0.625rem;
    border-radius: 0.625rem;
    background: var(--landing-surface-raised);
    border: 1px solid var(--landing-border);
}

/* ─── ROADMAP ─── */

.roadmap {
    width: 100%;
    padding: 4rem 1.5rem;
    background: var(--landing-bg-alt);
    border-top: 1px solid var(--landing-border-subtle);
}

.roadmap__container {
    max-width: 42rem;
    margin: 0 auto;
}

.roadmap__list {
    list-style: none;
    padding: 0;
    margin: 2.5rem 0 0;
    display: flex;
    flex-direction: column;
    gap: 1rem;
}

.roadmap__item {
    display: flex;
    align-items: flex-start;
    gap: 0.875rem;
    padding: 1rem 1.25rem;
    border-radius: 0.75rem;
    background: var(--landing-surface);
    border: 1px solid var(--landing-border);
    font-size: 0.9375rem;
    line-height: 1.6;
    color: var(--landing-text);
}

.roadmap__icon {
    width: 1.25rem;
    height: 1.25rem;
    color: var(--landing-accent);
    flex-shrink: 0;
    margin-top: 0.25rem;
}

.roadmap__desc {
    color: var(--landing-text-muted);
}

/* ─── FOOTER ─── */

.footer {
    width: 100%;
    padding: 2rem 1.5rem;
    border-top: 1px solid var(--landing-border-subtle);
    background: var(--landing-bg);
}

.footer__container {
    max-width: 72rem;
    margin: 0 auto;
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: 1rem;
}

.footer__copyright {
    font-size: 0.75rem;
    color: var(--landing-text-dim);
    margin: 0;
}

.footer__links {
    display: flex;
    gap: 1.5rem;
}

.footer__link {
    font-size: 0.75rem;
    font-weight: 500;
    color: var(--landing-text-muted);
    text-decoration: none;
    transition: color 0.2s;
}
.footer__link:hover {
    color: white;
}

.footer__canada {
    display: flex;
    align-items: center;
    gap: 0.375rem;
    font-size: 0.75rem;
    color: var(--landing-text-dim);
}

.footer__flag {
    font-size: 0.9375rem;
}

/* ─── SHARED SECTION HEADERS ─── */

.section-header {
    text-align: center;
}

.section-header__eyebrow {
    display: block;
    font-size: 0.75rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.15em;
    color: var(--landing-accent);
    margin-bottom: 0.75rem;
}

.section-header__title {
    font-size: clamp(1.5rem, 4vw, 2.25rem);
    font-weight: 800;
    color: white;
    margin: 0 0 0.75rem;
    letter-spacing: -0.02em;
}

.section-header__subtitle {
    font-size: 1rem;
    color: var(--landing-text-muted);
    max-width: 32rem;
    margin: 0 auto;
}

/* ─── RESPONSIVE ─── */

@media (max-width: 768px) {
    .hero {
        padding: 3rem 1rem;
    }

    .hero__cta {
        width: 100%;
        padding: 0.875rem;
    }

    .features {
        padding: 3rem 1rem;
    }

    .features__grid {
        grid-template-columns: 1fr;
    }

    .pricing {
        padding: 3rem 1rem;
    }

    .pricing__grid {
        grid-template-columns: 1fr;
    }

    .roadmap {
        padding: 3rem 1rem;
    }

    .footer__container {
        flex-direction: column;
        text-align: center;
    }
}

/* ─── ANIMATIONS ─── */

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

@keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.5; }
}
</style>
