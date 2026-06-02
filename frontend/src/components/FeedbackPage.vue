<script setup lang="ts">
import { computed, reactive } from 'vue'
import { RouterLink } from 'vue-router'

const form = reactive({
    name: '',
    email: '',
    message: `What I enjoy about Arch Stats:


What could be better:


Feature ideas or suggestions:


My archery background (optional — helps us understand your needs):
`,
    /** Honeypot field — must stay empty for real users */
    website: '',
})

const errors = reactive({
    name: '',
    email: '',
    message: '',
})

const NAME_MAX = 100
const EMAIL_MAX = 254
const MESSAGE_MIN = 10
const MESSAGE_MAX = 5000

function validateName(): boolean {
    if (!form.name.trim()) {
        errors.name = 'Name is required.'
        return false
    }
    if (form.name.length > NAME_MAX) {
        errors.name = `Name must be ${NAME_MAX} characters or fewer.`
        return false
    }
    errors.name = ''
    return true
}

function validateEmail(): boolean {
    if (!form.email.trim()) {
        errors.email = 'Email is required.'
        return false
    }
    if (form.email.length > EMAIL_MAX) {
        errors.email = `Email must be ${EMAIL_MAX} characters or fewer.`
        return false
    }
    const emailPattern = /^[^\s@]+@[^\s@][^\s.@]*\.[^\s@]+$/
    if (!emailPattern.test(form.email)) {
        errors.email = 'Please enter a valid email address.'
        return false
    }
    errors.email = ''
    return true
}

function validateMessage(): boolean {
    if (!form.message.trim()) {
        errors.message = 'Message is required.'
        return false
    }
    if (form.message.trim().length < MESSAGE_MIN) {
        errors.message = `Message must be at least ${MESSAGE_MIN} characters.`
        return false
    }
    if (form.message.length > MESSAGE_MAX) {
        errors.message = `Message must be ${MESSAGE_MAX} characters or fewer.`
        return false
    }
    errors.message = ''
    return true
}

const isFormValid = computed(() => {
    return (
        form.name.trim().length > 0
        && form.email.trim().length > 0
        && form.message.trim().length >= MESSAGE_MIN
        && /^[^\s@]+@[^\s@][^\s.@]*\.[^\s@]+$/.test(form.email)
    )
})

const messageCharCount = computed(() => form.message.length)

function handleBlur(field: 'name' | 'email' | 'message') {
    if (field === 'name')
        validateName()
    else if (field === 'email')
        validateEmail()
    else if (field === 'message')
        validateMessage()
}

/* Phase 2 will enable this */
const isSubmitEnabled = false
</script>

<template>
    <div class="page-shell">
        <header class="page-header">
            <RouterLink to="/" class="back-link" data-testid="feedback-back-link">
                <svg class="back-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
                </svg>
                Back to Home
            </RouterLink>
        </header>

        <main class="page-content">
            <h1 class="page-title">
                Send Us Feedback
            </h1>
            <p class="page-subtitle">
                Help us make Arch Stats better. Whether it's a bug, a suggestion, or just something
                you'd love to see — we want to hear from you.
            </p>

            <form class="feedback-form" @submit.prevent>
                <!-- Honeypot: invisible to real users, bots will fill it -->
                <div class="honeypot" aria-hidden="true" tabindex="-1">
                    <label for="website">Website</label>
                    <input
                        id="website"
                        v-model="form.website"
                        type="text"
                        name="website"
                        autocomplete="off"
                        tabindex="-1"
                    >
                </div>

                <div class="form-group">
                    <label for="feedback-name" class="form-label">Name</label>
                    <input
                        id="feedback-name"
                        v-model="form.name"
                        type="text"
                        class="form-input"
                        :class="{ 'form-input--error': errors.name }"
                        placeholder="Your name"
                        :maxlength="NAME_MAX"
                        data-testid="feedback-name"
                        @blur="handleBlur('name')"
                    >
                    <p v-if="errors.name" class="form-error" data-testid="feedback-name-error">
                        {{ errors.name }}
                    </p>
                </div>

                <div class="form-group">
                    <label for="feedback-email" class="form-label">Email</label>
                    <input
                        id="feedback-email"
                        v-model="form.email"
                        type="email"
                        class="form-input"
                        :class="{ 'form-input--error': errors.email }"
                        placeholder="you@example.com"
                        :maxlength="EMAIL_MAX"
                        data-testid="feedback-email"
                        @blur="handleBlur('email')"
                    >
                    <p v-if="errors.email" class="form-error" data-testid="feedback-email-error">
                        {{ errors.email }}
                    </p>
                </div>

                <div class="form-group">
                    <label for="feedback-message" class="form-label">
                        Message
                        <span class="char-count" :class="{ 'char-count--warn': messageCharCount > MESSAGE_MAX * 0.9 }">
                            {{ messageCharCount }} / {{ MESSAGE_MAX }}
                        </span>
                    </label>
                    <textarea
                        id="feedback-message"
                        v-model="form.message"
                        class="form-textarea"
                        :class="{ 'form-input--error': errors.message }"
                        rows="14"
                        :maxlength="MESSAGE_MAX"
                        data-testid="feedback-message"
                        @blur="handleBlur('message')"
                    />
                    <p v-if="errors.message" class="form-error" data-testid="feedback-message-error">
                        {{ errors.message }}
                    </p>
                </div>

                <div class="form-actions">
                    <button
                        type="submit"
                        class="submit-btn"
                        :class="{ 'submit-btn--disabled': !isFormValid || !isSubmitEnabled }"
                        :disabled="!isFormValid || !isSubmitEnabled"
                        :title="!isSubmitEnabled ? 'Submission coming soon — backend is being wired up' : ''"
                        data-testid="feedback-submit"
                    >
                        {{ isSubmitEnabled ? 'Send Feedback' : 'Coming Soon' }}
                    </button>
                </div>
            </form>
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
    padding: 3rem 1.5rem;
    max-width: 36rem;
    margin: 0 auto;
    width: 100%;
}

.page-title {
    font-size: 2rem;
    font-weight: 800;
    color: white;
    margin: 0 0 0.75rem;
    letter-spacing: -0.02em;
    text-align: center;
}

.page-subtitle {
    font-size: 1rem;
    line-height: 1.7;
    color: var(--landing-text-muted);
    margin: 0 0 2.5rem;
    text-align: center;
}

/* Honeypot: keep invisible but still in DOM for bots */
.honeypot {
    position: absolute;
    left: -9999px;
    opacity: 0;
    height: 0;
    overflow: hidden;
}

.feedback-form {
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
    position: relative;
}

.form-group {
    display: flex;
    flex-direction: column;
    gap: 0.375rem;
}

.form-label {
    font-size: 0.8125rem;
    font-weight: 600;
    color: var(--landing-text);
    display: flex;
    align-items: baseline;
    justify-content: space-between;
}

.char-count {
    font-size: 0.6875rem;
    font-weight: 400;
    color: var(--landing-text-dim);
    transition: color 0.2s;
}
.char-count--warn {
    color: var(--landing-accent-warm);
}

.form-input,
.form-textarea {
    width: 100%;
    padding: 0.75rem 1rem;
    border-radius: 0.625rem;
    background: var(--landing-surface);
    border: 1px solid var(--landing-border);
    color: var(--landing-text);
    font-size: 0.875rem;
    font-family: inherit;
    transition: border-color 0.2s, box-shadow 0.2s;
    box-sizing: border-box;
}
.form-input:focus,
.form-textarea:focus {
    outline: none;
    border-color: var(--landing-accent);
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.15);
}
.form-input--error {
    border-color: #ef4444;
}
.form-input--error:focus {
    box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.15);
}

.form-textarea {
    resize: vertical;
    min-height: 10rem;
    line-height: 1.7;
}

.form-input::placeholder,
.form-textarea::placeholder {
    color: var(--landing-text-dim);
}

.form-error {
    font-size: 0.75rem;
    color: #ef4444;
    margin: 0;
}

.form-actions {
    padding-top: 0.5rem;
}

.submit-btn {
    width: 100%;
    padding: 0.875rem;
    border: none;
    border-radius: 0.75rem;
    font-size: 0.9375rem;
    font-weight: 700;
    font-family: inherit;
    cursor: pointer;
    background: linear-gradient(135deg, var(--landing-gradient-start), var(--landing-gradient-end));
    color: white;
    transition: opacity 0.2s, transform 0.2s, box-shadow 0.2s;
    box-shadow: 0 0 20px rgba(59, 130, 246, 0.3);
}
.submit-btn:hover:not(:disabled) {
    opacity: 0.92;
    transform: translateY(-1px);
    box-shadow: 0 0 30px rgba(59, 130, 246, 0.5);
}
.submit-btn--disabled {
    opacity: 0.5;
    cursor: not-allowed;
    box-shadow: none;
}
</style>
