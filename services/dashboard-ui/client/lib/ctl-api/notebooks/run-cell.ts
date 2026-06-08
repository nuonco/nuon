import { api } from '@/lib/api'
import type { TNotebookCellRun, INotebookScoped } from './types'

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
