import type { TAppBranchInstallGroup } from '@/types'

export type TGroupLabelChange = {
  key: string
  from?: string
  to: string
}

export const labelChangesForGroupMove = (
  labels: Record<string, string> | undefined,
  group?: TAppBranchInstallGroup
): TGroupLabelChange[] => {
  const current = labels ?? {}
  const match = group?.label_selector?.match_labels ?? {}
  const changes: TGroupLabelChange[] = []

  for (const [key, value] of Object.entries(match)) {
    if (!value || value === '*') continue
    if (current[key] === value) continue
    changes.push({ key, from: current[key], to: value })
  }

  return changes
}
