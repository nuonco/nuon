import { api } from '@/lib/api'
import type { TInstallDeploymentsResponse } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getInstallDeployments = ({
  installId,
  orgId,
  offset = 0,
  limit = 20,
  status,
  type,
  resource,
  createdAtGte,
  search,
}: {
  installId: string
  orgId: string
  offset?: number
  limit?: number
  status?: string
  type?: string
  resource?: string
  createdAtGte?: string
  search?: string
}) =>
  api<TInstallDeploymentsResponse>({
    path: `installs/${installId}/deployments${buildQueryParams({
      offset,
      limit,
      status: status || undefined,
      type: type || undefined,
      resource: resource || undefined,
      created_at_gte: createdAtGte || undefined,
      search: search || undefined,
    })}`,
    orgId,
  })
