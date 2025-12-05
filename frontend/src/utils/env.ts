/**
 * Checks if an environment variable value is considered "true".
 * Accepted truthy values: "true", "True", "TRUE", "T", "t", "1", 1, true.
 * Everything else is false.
 */
export function isEnvTrue(value: unknown): boolean {
  let is_true = false
  if (typeof value === 'boolean' && value) {
    is_true = true
  }
  if (typeof value === 'number') {
    is_true = value === 1
  }
  if (typeof value === 'string') {
    const v = value.toLowerCase()
    is_true = v === 'true' || v === 't' || v === '1'
  }
  return is_true
}
