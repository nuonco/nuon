import { api } from '@/lib/api'
import type { TInstallUpdatesResponse } from '@/types'

export const getInstallUpdates = ({
  installId,
  orgId,
  page = 0,
  limit = 20,
}: {
  installId: string
  orgId: string
  page?: number
  limit?: number
}) =>
  api<TInstallUpdatesResponse>({
    path: `installs/${installId}/updates?page=${page}&limit=${limit}`,
    orgId,
  })
