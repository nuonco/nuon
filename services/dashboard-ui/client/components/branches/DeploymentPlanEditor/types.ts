export interface ILabelSelector {
  match_labels?: Record<string, string>
  not_match_labels?: Record<string, string>
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
  use_for_previews: boolean
}
