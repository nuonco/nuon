export interface ILabelSelector {
  match_labels?: Record<string, string>
  not_match_labels?: Record<string, string>
}

export type InstallSelectionMode = 'labels' | 'default'

export interface IInstallGroup {
  id: string
  name: string
  label_selector?: ILabelSelector | null
  selection_mode: InstallSelectionMode
  order: number
  max_parallel: number
  auto_approve_on_policies_passing: boolean
}
