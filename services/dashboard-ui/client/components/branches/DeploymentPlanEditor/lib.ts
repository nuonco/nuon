import type { IInstallGroup } from './types'

export const newGroup = (existingCount: number): IInstallGroup => ({
  id: `group-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
  name: '',
  install_ids: [],
  order: existingCount,
  max_parallel: 1,
  requires_approval: false,
  rollback_on_failure: false,
})
