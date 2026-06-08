import { api } from '@/lib/api'
import type { TNotebook, INotebookScoped } from './types'

export const getNotebook = ({ orgId, installId, notebookId }: INotebookScoped) =>
  api<TNotebook>({
    orgId,
    path: `installs/${installId}/notebooks/${notebookId}`,
  })
