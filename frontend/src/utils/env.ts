/**
 * Checks if an environment variable value is considered "true".
 * Accepted truthy values: "true", "True", "TRUE", "T", "t", "1", 1, true.
 * Everything else is false.
 */
export function isEnvTrue(value: unknown): boolean {
  let isTrue = false
  if (typeof value === 'boolean' && value) {
    isTrue = true
  }
  else if (typeof value === 'number') {
    isTrue = value === 1
  }
  else if (typeof value === 'string') {
    const v = value.toLowerCase()
    isTrue = v === 'true' || v === 't' || v === '1'
  }
  return isTrue
}
