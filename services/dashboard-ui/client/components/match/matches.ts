import type { Labels, Selector } from './types'

export const matchesSelector = (
  labels: Labels | undefined,
  selector?: Selector | null
): boolean => {
  if (!selector) return true
  const set = labels ?? {}

  for (const [key, value] of Object.entries(selector.match_labels ?? {})) {
    const got = set[key]
    if (got === undefined) return false
    if (value === '*') continue
    if (got !== value) return false
  }

  for (const [key, value] of Object.entries(selector.not_match_labels ?? {})) {
    const got = set[key]
    if (got === undefined) continue
    if (value === '*') return false
    if (got === value) return false
  }

  return true
}
