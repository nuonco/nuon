import type { TCompositeStatus } from '@/types'

export type TNotebookStatus = 'active' | 'archived'

export type TNotebookCellRun = {
  id: string
  created_at?: string
  updated_at?: string
  install_id?: string
  notebook_id?: string
  cell_id?: string
  cell_revision?: number
  install_action_workflow_run_id?: string
  log_stream_id?: string
  runner_job_id?: string
  name?: string
  inline_contents?: string
  command?: string
  env_vars?: Record<string, string>
  triggered_by_id?: string
  triggered_by_type?: string
  status?: string
  status_description?: string
  status_v2?: TCompositeStatus
}

export type TNotebookCell = {
  id: string
  created_at?: string
  updated_at?: string
  notebook_id?: string
  position: number
  revision: number
  name?: string
  inline_contents?: string
  command?: string
  env_vars?: Record<string, string>
  timeout?: number
  role?: string
  enable_kube_config?: boolean
  latest_run?: TNotebookCellRun
}

export type TNotebook = {
  id: string
  created_at?: string
  updated_at?: string
  org_id?: string
  install_id?: string
  name?: string
  description?: string
  status?: TNotebookStatus
  cells?: TNotebookCell[]
  cell_count?: number
  latest_run_at?: string
}

export interface IInstallScoped {
  orgId: string
  installId: string
}

export interface INotebookScoped extends IInstallScoped {
  notebookId: string
}
