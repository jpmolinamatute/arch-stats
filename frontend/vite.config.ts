import { defineConfig } from 'vite';
import tailwindcss from '@tailwindcss/vite';
import vue from '@vitejs/plugin-vue';
import { fileURLToPath, URL } from 'node:url';

import { rmSync, readdirSync } from 'node:fs';
import { join } from 'node:path';

export default defineConfig({
    plugins: [
        vue(),
        tailwindcss(),
        {
            name: 'clean-out-dir',
            buildStart() {
                const outDir = fileURLToPath(new URL('../backend/src/frontend', import.meta.url));
                const files = readdirSync(outDir);
                for (const file of files) {
                    if (file !== '.gitkeep') {
                        rmSync(join(outDir, file), { recursive: true, force: true });
                    }
                }
            },
        },
    ],
    resolve: {
        alias: {
            '@': fileURLToPath(new URL('./src', import.meta.url)),
        },
    },
    build: {
        outDir: '../backend/src/frontend',
        emptyOutDir: false,
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
