export interface ILabelSelector {
  match_labels: Record<string, string>
}

export type InstallSelectionMode = 'manual' | 'labels'

export interface IInstallGroup {
  id: string
  name: string
  install_ids: string[]
  label_selector?: ILabelSelector | null
  selection_mode: InstallSelectionMode
  order: number
  max_parallel: number
  requires_approval: boolean
  rollback_on_failure: boolean
  use_for_previews: boolean
}
