import { api } from '@/lib/api'
import type { TNotebook, TNotebookStatus, INotebookScoped } from './types'

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
