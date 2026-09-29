import type { Labels } from './types'

export const parseLabelsQuery = (raw: string): Labels => {
  const out: Labels = {}
  const trimmed = raw.trim()
  if (!trimmed) return out

  for (const rawPart of trimmed.split(',')) {
    const part = rawPart.trim()
    if (!part) continue

    let key: string
    let value: string
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

export { labelsToQueryString } from './types'
