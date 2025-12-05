import { describe, expect, it } from 'vitest'
import { isEnvTrue } from '../../src/utils/env'

describe('isEnvTrue', () => {
  it('returns true for boolean true', () => {
    expect(isEnvTrue(true)).toBe(true)
  })

  it('returns true for number 1', () => {
    expect(isEnvTrue(1)).toBe(true)
  })

  it('returns true for string "1"', () => {
    expect(isEnvTrue('1')).toBe(true)
  })

  it('returns true for string "true" (case insensitive)', () => {
    expect(isEnvTrue('true')).toBe(true)
    expect(isEnvTrue('True')).toBe(true)
    expect(isEnvTrue('TRUE')).toBe(true)
  })

  it('returns true for string "t" (case insensitive)', () => {
    expect(isEnvTrue('t')).toBe(true)
    expect(isEnvTrue('T')).toBe(true)
  })

  it('returns false for boolean false', () => {
    expect(isEnvTrue(false)).toBe(false)
  })

  it('returns false for number 0', () => {
    expect(isEnvTrue(0)).toBe(false)
  })

  it('returns false for string "false"', () => {
    expect(isEnvTrue('false')).toBe(false)
  })

  it('returns false for undefined/null', () => {
    expect(isEnvTrue(undefined)).toBe(false)
    expect(isEnvTrue(null)).toBe(false)
  })

  it('returns false for other strings', () => {
    expect(isEnvTrue('random')).toBe(false)
    expect(isEnvTrue('')).toBe(false)
  })
})
