import type { IInstallGroup } from './types'

export const newGroup = (existingCount: number): IInstallGroup => ({
  id: `group-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
  name: '',
  install_ids: [],
  label_selector: null,
  selection_mode: 'manual',
  order: existingCount,
  max_parallel: 1,
  use_for_previews: false,
})
