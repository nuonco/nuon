import { api } from '@/lib/api'
import type { TCompositeStatus } from '@/types'

// Notebook types mirror app.Notebook / app.NotebookCell / app.NotebookCellRun
// in ctl-api. These are declared here (rather than pulled from the generated
// OpenAPI types) so the dashboard UI can ship ahead of an SDK regen.

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
}

interface IInstallScoped {
  orgId: string
  installId: string
}

interface INotebookScoped extends IInstallScoped {
  notebookId: string
}

export const getNotebooks = ({ orgId, installId }: IInstallScoped) =>
  api<TNotebook[]>({
    orgId,
    path: `installs/${installId}/notebooks`,
  })

export const getNotebook = ({ orgId, installId, notebookId }: INotebookScoped) =>
  api<TNotebook>({
    orgId,
    path: `installs/${installId}/notebooks/${notebookId}`,
  })

export interface ICreateNotebookBody {
  name?: string
  description?: string
}

export const createNotebook = ({
  orgId,
  installId,
  body,
}: IInstallScoped & { body: ICreateNotebookBody }) =>
  api<TNotebook>({
    orgId,
    method: 'POST',
    path: `installs/${installId}/notebooks`,
    body,
  })

export interface IUpdateNotebookBody {
  name?: string
  description?: string
  status?: TNotebookStatus
}

export const updateNotebook = ({
  orgId,
  installId,
  notebookId,
  body,
}: INotebookScoped & { body: IUpdateNotebookBody }) =>
  api<TNotebook>({
    orgId,
    method: 'PATCH',
    path: `installs/${installId}/notebooks/${notebookId}`,
    body,
  })

export const deleteNotebook = ({
  orgId,
  installId,
  notebookId,
}: INotebookScoped) =>
  api<void>({
    orgId,
    method: 'DELETE',
    path: `installs/${installId}/notebooks/${notebookId}`,
  })

export interface ICreateCellBody {
  name?: string
  inline_contents?: string
  command?: string
  env_vars?: Record<string, string>
  timeout?: number
  role?: string
  enable_kube_config?: boolean
}

export const createCell = ({
  orgId,
  installId,
  notebookId,
  body,
}: INotebookScoped & { body: ICreateCellBody }) =>
  api<TNotebookCell>({
    orgId,
    method: 'POST',
    path: `installs/${installId}/notebooks/${notebookId}/cells`,
    body,
  })

export type IUpdateCellBody = Partial<ICreateCellBody>

export const updateCell = ({
  orgId,
  installId,
  notebookId,
  cellId,
  body,
}: INotebookScoped & { cellId: string; body: IUpdateCellBody }) =>
  api<TNotebookCell>({
    orgId,
    method: 'PATCH',
    path: `installs/${installId}/notebooks/${notebookId}/cells/${cellId}`,
    body,
  })

export const deleteCell = ({
  orgId,
  installId,
  notebookId,
  cellId,
}: INotebookScoped & { cellId: string }) =>
  api<void>({
    orgId,
    method: 'DELETE',
    path: `installs/${installId}/notebooks/${notebookId}/cells/${cellId}`,
  })

export const runCell = ({
  orgId,
  installId,
  notebookId,
  cellId,
}: INotebookScoped & { cellId: string }) =>
  api<TNotebookCellRun>({
    orgId,
    method: 'POST',
    path: `installs/${installId}/notebooks/${notebookId}/cells/${cellId}/runs`,
    body: {},
  })

export const getCellRun = ({
  orgId,
  installId,
  notebookId,
  runId,
}: INotebookScoped & { runId: string }) =>
  api<TNotebookCellRun>({
    orgId,
    path: `installs/${installId}/notebooks/${notebookId}/runs/${runId}`,
  })
