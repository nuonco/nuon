import { api } from '@/lib/api'
import type { TPolicyReport } from '@/types'

export const getPolicyReport = ({
  reportId,
  orgId,
}: {
  reportId: string
  orgId: string
}) =>
  api<TPolicyReport>({
    path: `policy-reports/${reportId}`,
    orgId,
  })
