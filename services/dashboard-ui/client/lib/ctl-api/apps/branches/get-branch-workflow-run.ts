import { api } from '@/lib/api'
import type { TInstallWorkflow } from '@/types'

export const getBranchWorkflowRun = ({
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
  api<TInstallWorkflow>({
    path: `apps/${appId}/branches/${branchId}/runs/${runId}`,
    orgId,
  })
