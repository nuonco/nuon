import { api } from '@/lib/api'
import type { TAvailableRolesResponse, TOperationType, TPrincipalType } from '@/types'

export async function getAvailableRoles({
  installId,
  operationType,
  principalType,
  orgId,
}: {
  installId: string
  operationType: TOperationType
  principalType: TPrincipalType
  orgId: string
}) {
  return api<TAvailableRolesResponse>({
    path: `installs/${installId}/available-roles?principal_type=${principalType}&operation_type=${operationType}`,
    orgId,
  })
}
