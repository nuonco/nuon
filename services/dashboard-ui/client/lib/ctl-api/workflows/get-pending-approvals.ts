import { api } from '@/lib/api'
import type { TWorkflowStepApproval } from '@/types'

export const getPendingApprovals = ({ orgId }: { orgId: string }) =>
  api<TWorkflowStepApproval[]>({
    path: `workflows/pending-approvals?limit=100`,
    orgId,
  })
