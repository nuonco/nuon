import { api } from '@/lib/api'
import type { TInstallWorkflow } from '@/types'

export const getBranchWorkflowRun = ({
  appId: _appId,
  branchId: _branchId,
  runId,
  orgId,
}: {
  appId: string
  branchId: string
  runId: string
  orgId: string
}) =>
  api<TInstallWorkflow>({
    path: `workflows/${runId}`,
    orgId,
  })
