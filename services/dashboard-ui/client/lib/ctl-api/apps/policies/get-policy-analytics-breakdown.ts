import { api } from '@/lib/api'
import type { TPolicyAnalyticsBreakdown } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getPolicyAnalyticsBreakdown = ({
  appId,
  orgId,
  dimension,
  start,
  end,
  limit,
  installId,
  policyId,
}: {
  appId: string
  orgId: string
  dimension: string
  start?: string
  end?: string
  limit?: number
  installId?: string
  policyId?: string
}) =>
  api<TPolicyAnalyticsBreakdown>({
    path: `apps/${appId}/policy-analytics/breakdown${buildQueryParams({ dimension, start, end, limit, install_id: installId, policy_id: policyId })}`,
    orgId,
  })
