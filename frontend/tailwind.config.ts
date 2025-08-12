import type { Config } from 'tailwindcss';

export default {
    content: ['./index.html', './src/**/*.{vue,ts}'],
    theme: {
        extend: {
            colors: {
                wa: {
                    reflex: '#00209F',
                    pink: '#E5239D',
                    yellow: '#FFC726',
                    green: '#12AD2B',
                    red: '#F42A41',
                    sky: '#00B5E6',
                    black: '#000000',
                },
            },
            borderRadius: {
                DEFAULT: '0.75rem',
                xl: '1rem',
                '2xl': '1.25rem',
            },
        },
    },
    darkMode: 'class',
    plugins: [],
} satisfies Config;
