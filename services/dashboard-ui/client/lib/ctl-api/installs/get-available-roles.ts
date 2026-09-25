import { api } from '@/lib/api'
import type {
  TAvailableRolesResponse,
  TOperationType,
  TPrincipalType,
  TWorkflowType,
} from '@/types'

export async function getAvailableRoles({
  installId,
  operationType,
  principalType,
  principalId,
  workflowType,
  orgId,
}: {
  installId: string
  operationType?: TOperationType
  principalType?: TPrincipalType
  principalId?: string
  workflowType?: TWorkflowType
  orgId: string
}) {
  const params = new URLSearchParams()
  if (principalType) {
    params.set('principal_type', principalType)
  }
  if (operationType) {
    params.set('operation_type', operationType)
  }
  if (principalId) {
    params.set('principal_id', principalId)
  }
  if (workflowType) {
    params.set('workflow_type', workflowType)
  }
  return api<TAvailableRolesResponse>({
    path: `installs/${installId}/available-roles?${params.toString()}`,
    orgId,
  })
}
