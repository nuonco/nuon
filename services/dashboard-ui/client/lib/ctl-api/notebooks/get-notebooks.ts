import { api } from '@/lib/api'
import type { TNotebook, IInstallScoped } from './types'

export const getNotebooks = ({ orgId, installId }: IInstallScoped) =>
  api<TNotebook[]>({
    orgId,
    path: `installs/${installId}/notebooks`,
  })
