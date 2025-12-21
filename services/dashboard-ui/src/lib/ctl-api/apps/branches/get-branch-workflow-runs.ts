import { api } from '@/lib/api'
import type { TInstallWorkflow } from '@/types'

export const getBranchWorkflowRuns = ({
  appId,
  branchId,
  orgId,
  limit,
  offset,
}: {
  appId: string
  branchId: string
  orgId: string
  limit?: number
  offset?: number
}) =>
  api<TInstallWorkflow[]>({
    path: `apps/${appId}/branches/${branchId}/runs`,
    orgId,
  })