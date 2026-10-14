import { api } from '@/lib/api'
import type { TInstallActivityResponse } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getInstallActivity = ({
  installId,
  orgId,
  offset,
  cursor,
  limit = 20,
  status,
  type,
  createdAtGte,
  search,
}: {
  installId: string
  orgId: string
  offset?: number
  cursor?: string
  limit?: number
  status?: string
  type?: string
  createdAtGte?: string
  search?: string
}) =>
  api<TInstallActivityResponse>({
    path: `installs/${installId}/activity${buildQueryParams({
      offset,
      cursor,
      limit,
      status: status || undefined,
      type: type || undefined,
      created_at_gte: createdAtGte || undefined,
      search: search || undefined,
    })}`,
    orgId,
  })
