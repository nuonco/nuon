import type { IInstallGroup, InstallSelectionMode } from './types'

export const newGroup = (
  existingCount: number,
  selectionMode: InstallSelectionMode = 'labels'
): IInstallGroup => ({
  id: `group-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
  name: '',
  label_selector: null,
  selection_mode: selectionMode,
  order: existingCount,
  max_parallel: 1,
  auto_approve_on_policies_passing: false,
})
