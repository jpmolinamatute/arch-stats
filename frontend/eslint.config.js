import antfu from '@antfu/eslint-config'

export default antfu({
    vue: true,
    typescript: true,
    ignores: [
        'README.md',
    ],
    stylistic: {
        indent: 4,
    },
})
