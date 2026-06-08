import { api } from '@/lib/api'
import type { INotebookScoped } from './types'

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
