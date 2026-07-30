import { api } from '@/lib/api'
import type { TInstallHealthTimeline } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getInstallHealthTimeline = ({
  orgId,
  installId,
  days,
}: {
  orgId: string
  installId: string
  days?: number
}) =>
  api<TInstallHealthTimeline>({
    orgId,
    path: `installs/${installId}/health/timeline${buildQueryParams({ days })}`,
  })
