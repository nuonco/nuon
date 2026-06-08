import { api } from '@/lib/api'
import type { TNotebook, IInstallScoped } from './types'

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
