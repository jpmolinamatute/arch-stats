export default {
    build: {
        outDir: '../backend/server/frontend',
    },
    server: {
        proxy: {
            '/api': 'http://localhost:8000', // Forward /api to your backend
        },
    },
};
