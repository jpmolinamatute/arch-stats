import vue from '@vitejs/plugin-vue';
import { defineConfig } from 'vitest/config';
import type { PluginOption } from 'vite';
export default defineConfig({
    plugins: [vue() as PluginOption],
    test: {
        environment: 'jsdom',
        globals: true,
        setupFiles: ['./vitest.setup.ts'],
        coverage: {
            provider: 'v8',
            reporter: ['text', 'html'],
        },
        include: ['tests/**/*.{test,spec}.{ts,tsx}'],
        exclude: ['src/**/*.{test,spec}.{ts,tsx}'],
    },
});
