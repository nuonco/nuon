import { api } from '@/lib/api'
import type { TInstallComponentHealthTimeline } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getInstallComponentHealthTimeline = ({
  orgId,
  installId,
  componentId,
  days,
}: {
  orgId: string
  installId: string
  componentId: string
  days?: number
}) =>
  api<TInstallComponentHealthTimeline>({
    orgId,
    path: `installs/${installId}/components/${componentId}/health/timeline${buildQueryParams({ days })}`,
  })
