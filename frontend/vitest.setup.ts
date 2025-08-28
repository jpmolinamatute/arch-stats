/* global console */
// Vitest setup file: runs before each test suite
import { config } from '@vue/test-utils';
import { beforeAll } from 'vitest';

// Example: stub out global components or directives if needed
config.global.stubs = {
    // Avoid rendering heavy components by default; override per-test when needed
};

// JSDOM defaults are fine; add any globals/polyfills here if your code needs them.

// --- Test log noise control ---
// Suppress only known, expected warnings/errors produced by composables under
// test when exercising error paths (network failures, non-OK responses, etc.).
// This keeps test output clean without hiding unrelated logs.
beforeAll(() => {
    const suppressed = [
        /^A session is already open\b/i,
        /^Failed to create\/open session:/i,
        /^Failed to close session:/i,
        /^Failed to create target:/i,
        /^Error creating target:/i,
    ];

    const shouldSuppress = (args: unknown[]): boolean => {
        const first = args[0];
        return typeof first === 'string' && suppressed.some((rx) => rx.test(first));
    };

    const originalWarn = console.warn.bind(console);
    const originalError = console.error.bind(console);

    // Wrap warn/error to filter specific, expected messages while preserving others.

    console.warn = ((...args: unknown[]) => {
        if (shouldSuppress(args)) return;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        (originalWarn as any)(...args);
    }) as typeof console.warn;

    console.error = ((...args: unknown[]) => {
        if (shouldSuppress(args)) return;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        (originalError as any)(...args);
    }) as typeof console.error;
});
