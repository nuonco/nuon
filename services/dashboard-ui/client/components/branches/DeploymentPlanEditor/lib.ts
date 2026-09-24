import type { IInstallGroup, InstallSelectionMode } from './types'

export const newGroup = (
  existingCount: number,
  selectionMode: InstallSelectionMode = 'labels',
  isDefault = false
): IInstallGroup => ({
  id: `group-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
  name: '',
  label_selector: null,
  selection_mode: selectionMode,
  is_default: isDefault,
  order: existingCount,
  max_parallel: 1,
  auto_approve_on_policies_passing: false,
})

export const selectorKey = (
  labels?: Record<string, string> | null
): string | null => {
  if (!labels) return null
  const entries = Object.entries(labels)
    .filter(([key, value]) => key.trim() && value.trim())
    .sort(([a], [b]) => a.localeCompare(b))
  if (entries.length === 0) return null
  return entries.map(([key, value]) => `${key}=${value}`).join('\0')
}

export const duplicateSelectorGroupIds = (
  groups: IInstallGroup[]
): Set<string> => {
  const byKey = new Map<string, string[]>()
  groups.forEach((group) => {
    if (group.selection_mode !== 'labels') return
    const key = selectorKey(group.label_selector?.match_labels)
    if (!key) return
    const ids = byKey.get(key) ?? []
    ids.push(group.id)
    byKey.set(key, ids)
  })
  const duplicates = new Set<string>()
  byKey.forEach((ids) => {
    if (ids.length > 1) ids.forEach((id) => duplicates.add(id))
  })
  return duplicates
}
