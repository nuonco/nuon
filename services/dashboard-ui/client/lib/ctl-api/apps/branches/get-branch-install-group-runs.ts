import { api } from '@/lib/api'
import type { TInstallGroupRun } from '@/types'

export const getBranchInstallGroupRuns = ({
  appId,
  branchId,
  runId,
  orgId,
}: {
  appId: string
  branchId: string
  runId: string
  orgId: string
}) =>
  api<TInstallGroupRun[]>({
    path: `apps/${appId}/branches/${branchId}/runs/${runId}/install-group-runs`,
    orgId,
  })
