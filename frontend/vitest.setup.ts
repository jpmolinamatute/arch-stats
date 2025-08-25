// Vitest setup file: runs before each test suite
import { config } from '@vue/test-utils';

// Example: stub out global components or directives if needed
config.global.stubs = {
    // Avoid rendering heavy components by default; override per-test when needed
};

// JSDOM defaults are fine; add any globals/polyfills here if your code needs them.
