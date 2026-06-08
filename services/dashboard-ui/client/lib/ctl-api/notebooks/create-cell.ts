import { api } from '@/lib/api'
import type { TNotebookCell, INotebookScoped } from './types'

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
