import { api } from '@/lib/api'
import type { TInstallActivityResponse } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getInstallActivity = ({
  installId,
  orgId,
  offset = 0,
  limit = 20,
  status,
  type,
  createdAtGte,
  search,
}: {
  installId: string
  orgId: string
  offset?: number
  limit?: number
  status?: string
  type?: string
  createdAtGte?: string
  search?: string
}) =>
  api<TInstallActivityResponse>({
    path: `installs/${installId}/activity${buildQueryParams({
      offset,
      limit,
      status: status || undefined,
      type: type || undefined,
      created_at_gte: createdAtGte || undefined,
      search: search || undefined,
    })}`,
    orgId,
  })
