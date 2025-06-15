import { defineConfig } from 'vite';
import tailwindcss from '@tailwindcss/vite';
import vue from '@vitejs/plugin-vue';

export default defineConfig({
    plugins: [vue(), tailwindcss()],
    build: {
        outDir: '../backend/server/frontend',
    },
    server: {
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
