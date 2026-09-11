export type TLabels = Record<string, string>

export type TLabelSelector = {
  match_labels?: TLabels
  not_match_labels?: TLabels
}

export const matchesSelector = (
  labels: TLabels | undefined,
  selector?: TLabelSelector | null
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

export const hasLabelSelector = (
  selector?: TLabelSelector | null
): boolean => {
  if (!selector) return false
  return (
    Object.keys(selector.match_labels ?? {}).length > 0 ||
    Object.keys(selector.not_match_labels ?? {}).length > 0
  )
}
