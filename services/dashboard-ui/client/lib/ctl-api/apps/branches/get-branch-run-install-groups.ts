import { api } from '@/lib/api'
import type { TInstallAppConfigVersion } from '@/types'

export const getBranchRunInstallGroups = ({
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
  api<TInstallAppConfigVersion[]>({
    path: `apps/${appId}/branches/${branchId}/runs/${runId}/install-groups`,
    orgId,
  })
