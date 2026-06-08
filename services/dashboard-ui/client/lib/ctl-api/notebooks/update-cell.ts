import { api } from '@/lib/api'
import type { TNotebookCell, INotebookScoped } from './types'
import type { ICreateCellBody } from './create-cell'

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
