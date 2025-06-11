import { defineConfig } from 'vite';

export default defineConfig({
    build: {
        outDir: '../backend/server/frontend',
    },
    server: {
        proxy: {
            '/api': 'http://localhost:8000',
        },
        strictPort: true,
    },
});
