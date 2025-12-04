/* global console */
// Vitest setup file: runs before each test suite
import { config } from '@vue/test-utils'
import { beforeAll } from 'vitest'

// Example: stub out global components or directives if needed
config.global.stubs = {
  // Avoid rendering heavy components by default; override per-test when needed
}

// JSDOM defaults are fine; add any globals/polyfills here if your code needs them.

// --- Test log noise control ---
// Suppress only known, expected warnings/errors produced by composables under
// test when exercising error paths (network failures, non-OK responses, etc.).
// This keeps test output clean without hiding unrelated logs.
beforeAll(() => {
  // Provide a minimal fetch mock for relative API calls in unit tests.
  // This prevents Node's undici from throwing on relative URLs like "/api/v0/auth/me".
  const g = globalThis as unknown as { fetch?: unknown, __fetch_patched__?: boolean }
  if (!g.__fetch_patched__) {
    const originalFetch = g.fetch as unknown
    g.fetch = (async (input: unknown) => {
      const url = typeof input === 'string' ? input : String(input ?? '')
      if (typeof url === 'string' && url.startsWith('/api/')) {
        // Simulate 401 for auth/me and a generic OK empty body for others as needed
        if (url.startsWith('/api/v0/auth/me')) {
          return {
            ok: false,
            status: 401,
            json: async () => ({}),
            text: async () => '{}',
          } as unknown as { ok: boolean, status: number, json: () => Promise<unknown> }
        }
        return {
          ok: true,
          status: 200,
          json: async () => ({}),
          text: async () => '{}',
        } as unknown as { ok: boolean, status: number, json: () => Promise<unknown> }
      }
      if (typeof originalFetch === 'function') {
        return (originalFetch as any)(input)
      }
      throw new Error('No fetch available in test environment')
    }) as typeof globalThis.fetch
    g.__fetch_patched__ = true
  }
  const suppressed = [
    /^A session is already open\b/i,
    /^Failed to create\/open session:/i,
    /^Failed to close session:/i,
    /^Failed to create target:/i,
    /^Error creating target:/i,
  ]

  const shouldSuppress = (args: unknown[]): boolean => {
    const first = args[0]
    return typeof first === 'string' && suppressed.some(rx => rx.test(first))
  }

  const originalWarn = console.warn.bind(console)
  const originalError = console.error.bind(console)

  console.warn = ((...args: unknown[]) => {
    if (shouldSuppress(args))
      return;

    (originalWarn as any)(...args)
  }) as typeof console.warn

  console.error = ((...args: unknown[]) => {
    if (shouldSuppress(args))
      return;

    (originalError as any)(...args)
  }) as typeof console.error
})
