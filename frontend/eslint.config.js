import js from '@eslint/js';
import tsParser from '@typescript-eslint/parser';
import vueParser from 'vue-eslint-parser';
import prettierConfig from 'eslint-config-prettier';
import eslint from '@eslint/js';
import prettierPlugin from 'eslint-plugin-prettier';
import tseslint from 'typescript-eslint';
import { dirname } from 'path';
import { fileURLToPath } from 'url';

const __dirname = dirname(fileURLToPath(import.meta.url));

export default tseslint.config(
    eslint.configs.recommended,
    tseslint.configs.recommended,
    js.configs.recommended,
    prettierConfig,
    {
        files: ['./src/**/*.{js,jsx,ts,tsx,vue}', './tests/**/*.{js,jsx,ts,tsx,vue}'],
        ignores: ['node_modules/**'],
        rules: {
            semi: 'error',
            'prefer-const': 'error',
            'prettier/prettier': 'error',
            'no-unused-vars': 'off',
            '@typescript-eslint/no-unused-vars': 'warn',
        },
        plugins: {
            prettier: prettierPlugin,
        },
        languageOptions: {
            parser: vueParser,
            parserOptions: {
                parser: tsParser,
                extraFileExtensions: ['.vue'],
                ecmaVersion: 2024,
                sourceType: 'module',
                project: ['./tsconfig.app.json', './tsconfig.node.json'],
                tsconfigRootDir: __dirname,
            },
            globals: {
                document: 'readonly',
                window: 'readonly',
                console: 'readonly',
                fetch: 'readonly',
                location: 'readonly',
                setTimeout: 'readonly',
                clearTimeout: 'readonly',
                navigator: 'readonly',
                requestAnimationFrame: 'readonly',
                alert: 'readonly',
            },
        },
    },
);
