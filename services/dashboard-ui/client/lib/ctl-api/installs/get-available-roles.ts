import { api } from '@/lib/api'
import type { TAvailableRolesResponse, TOperationType, TPrincipalType } from '@/types'

export async function getAvailableRoles({
  installId,
  operationType,
  principalType,
  principalId,
  orgId,
}: {
  installId: string
  operationType?: TOperationType
  principalType?: TPrincipalType
  principalId?: string
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
  return api<TAvailableRolesResponse>({
    path: `installs/${installId}/available-roles?${params.toString()}`,
    orgId,
  })
}
