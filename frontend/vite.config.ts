import { defineConfig } from 'vite';
import tailwindcss from '@tailwindcss/vite';
import vue from '@vitejs/plugin-vue';
import { fileURLToPath, URL } from 'node:url';

export default defineConfig({
    plugins: [vue(), tailwindcss()],
    resolve: {
        alias: {
            '@': fileURLToPath(new URL('./src', import.meta.url)),
        },
    },
    build: {
        outDir: '../backend/src/server/frontend',
    },
    server: {
        headers: {
            'Referrer-Policy': 'no-referrer-when-downgrade',
            // Avoid cross-origin isolation in dev; Google Identity popups/iframes need opener access
            'Cross-Origin-Opener-Policy': 'unsafe-none',
            'Cross-Origin-Embedder-Policy': 'unsafe-none',
            // Allow FedCM / Identity Credential API in dev
            'Permissions-Policy': 'identity-credentials-get=(*)',
        },
        hmr: true,
        proxy: {
            '/api/v0': {
                target: 'http://localhost:8000',
                changeOrigin: true,
                ws: true,
            },
        },
        strictPort: true,
    },
});
