import { api } from '@/lib/api'
import type { INotebookScoped } from './types'

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
