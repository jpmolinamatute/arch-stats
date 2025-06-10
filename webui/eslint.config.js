import js from '@eslint/js';
import prettier from 'eslint-config-prettier';
import tseslint from 'typescript-eslint';
import prettierPlugin from 'eslint-plugin-prettier';

export default [
    js.configs.recommended,
    ...tseslint.configs.recommended,
    {
        plugins: { prettier: prettierPlugin },
        rules: {
            'prettier/prettier': 'error',
        },
    },
    prettier,
];
