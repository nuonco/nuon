import { api } from '@/lib/api'
import type { TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'
import type { TNotebook, IInstallScoped } from './types'

export const getNotebooks = ({
  orgId,
  installId,
  limit,
  offset,
  q,
}: IInstallScoped & TPaginationParams & { q?: string }) =>
  api<TNotebook[]>({
    orgId,
    path: `installs/${installId}/notebooks${buildQueryParams({ limit, offset, q })}`,
    paginated: true,
  })
