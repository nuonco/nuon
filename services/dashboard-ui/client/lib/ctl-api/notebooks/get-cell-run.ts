import { api } from '@/lib/api'
import type { TNotebookCellRun, INotebookScoped } from './types'

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
