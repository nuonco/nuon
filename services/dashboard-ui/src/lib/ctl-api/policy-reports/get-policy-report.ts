import { api } from '@/lib/api'
import type { TPolicyReport } from '@/types/ctl-api.types'

export const getPolicyReport = ({
  orgId,
  reportId,
}: {
  orgId: string
  reportId: string
}) =>
  api<TPolicyReport>({
    path: `policy-reports/${reportId}`,
    orgId,
  })
