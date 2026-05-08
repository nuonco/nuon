// Mirror of pkg/labels.ParseLabelsQuery — comma-separated `key=value`
// (or `key:value`) pairs, with bare keys treated as wildcards (`key=*`).
//
// Stay in lockstep with services/dashboard-ui/client/components/match/types.ts
// and pkg/labels/labels.go.

import type { Labels } from './types'

// parseLabelsQuery parses `env=prod, tier=critical, owner=*` into a Labels
// map. Returns an empty object for empty input. Mirrors the Go semantics:
//   - splits on commas
//   - trims whitespace around each part, key, and value
//   - tries `:` first, then `=` as the key-value separator (first occurrence wins)
//   - bare keys with no separator are treated as wildcards (value = "*")
//   - empty parts are skipped
//   - empty keys are skipped
export const parseLabelsQuery = (raw: string): Labels => {
  const out: Labels = {}
  const trimmed = raw.trim()
  if (!trimmed) return out

  for (const rawPart of trimmed.split(',')) {
    const part = rawPart.trim()
    if (!part) continue

    let key: string
    let value: string
    // Match the Go helper: try `:` first, then `=`. Splits on the FIRST
    // separator only so values may contain colons or equals signs.
    let sepIdx = part.indexOf(':')
    if (sepIdx === -1) sepIdx = part.indexOf('=')
    if (sepIdx === -1) {
      key = part
      value = '*'
    } else {
      key = part.slice(0, sepIdx)
      value = part.slice(sepIdx + 1)
    }

    key = key.trim()
    value = value.trim()
    if (!key) continue
    out[key] = value
  }

  return out
}

// labelsToQueryString is re-exported from `./types` (where describeMatch
// already needs it) to keep the round-trip API together. Sorted-keys
// invariant matches the Go labelsToQueryString helper so two semantically
// equal Labels render the same string.
export { labelsToQueryString } from './types'
