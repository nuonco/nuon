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
  actorId,
  excludeCron,
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
  actorId?: string
  excludeCron?: boolean
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
      actor_id: actorId || undefined,
      exclude_cron: excludeCron || undefined,
      created_at_gte: createdAtGte || undefined,
      search: search || undefined,
    })}`,
    orgId,
  })
