import { api } from '@/lib/api'
import type { TInstallDeploymentSummariesResponse } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getInstallDeploymentSummaries = async ({
  installId,
  orgId,
  offset = 0,
  limit = 20,
  status,
  type,
  resource,
  createdAtGte,
  search,
  state,
  sort,
  cursor,
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
  state?: 'active' | 'finished'
  sort?: 'attention'
  cursor?: string
}) => {
  const fetchPage = (pageLimit: number, pageCursor?: string) =>
    api<TInstallDeploymentSummariesResponse>({
      path: `installs/${installId}/deployment-summaries${buildQueryParams({
        offset: pageCursor ? undefined : offset,
        limit: pageLimit,
        status: status || undefined,
        type: type || undefined,
        resource: resource || undefined,
        created_at_gte: createdAtGte || undefined,
        search: search || undefined,
        state,
        sort,
        cursor: pageCursor,
      })}`,
      orgId,
    })

  const window = await fetchPage(Math.min(limit, 100), cursor)
  const seen = new Set(window.deployments.map((deployment) => deployment.id))
  while (
    window.deployments.length < limit &&
    window.has_more &&
    window.next_cursor
  ) {
    const next = await fetchPage(
      Math.min(limit - window.deployments.length, 100),
      window.next_cursor
    )
    for (const deployment of next.deployments) {
      if (!seen.has(deployment.id)) {
        seen.add(deployment.id)
        window.deployments.push(deployment)
      }
    }
    window.has_more = next.has_more
    window.next_cursor = next.next_cursor
  }
  return { ...window, limit }
}
