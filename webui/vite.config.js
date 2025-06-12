import { defineConfig } from 'vite';

export default defineConfig({
    build: {
        outDir: '../backend/server/frontend',
    },
    server: {
        proxy: {
            '/api/v0': 'http://localhost:8000',
        },
        strictPort: true,
    },
});
