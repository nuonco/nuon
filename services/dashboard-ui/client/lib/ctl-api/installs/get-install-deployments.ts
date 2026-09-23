import { api } from '@/lib/api'
import type { TInstallDeploymentsResponse } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getInstallDeployments = ({
  installId,
  orgId,
  page = 0,
  limit = 20,
  status,
  type,
  resource,
  createdAtGte,
  search,
}: {
  installId: string
  orgId: string
  page?: number
  limit?: number
  status?: string
  type?: string
  resource?: string
  createdAtGte?: string
  search?: string
}) =>
  api<TInstallDeploymentsResponse>({
    path: `installs/${installId}/deployments${buildQueryParams({
      page,
      limit,
      status: status || undefined,
      type: type || undefined,
      resource: resource || undefined,
      created_at_gte: createdAtGte || undefined,
      search: search || undefined,
    })}`,
    orgId,
  })
