import { api } from '@/lib/api'

export type AvailableRole = {
  name: string
  arn: string
  role_type: 'custom' | 'break_glass' | 'provision' | 'deprovision' | 'maintenance'
}

export type AvailableRolesResponse = {
  roles: AvailableRole[]
}

export type OperationType =
  | 'provision'
  | 'reprovision'
  | 'deprovision'
  | 'deploy'
  | 'teardown'
  | 'trigger'

export type PrincipalType = 'component' | 'sandbox' | 'action'

export async function getAvailableRoles({
  installId,
  operationType,
  principalType,
  orgId,
}: {
  installId: string
  operationType: OperationType
  principalType: PrincipalType
  orgId: string
}) {
  return api<AvailableRolesResponse>({
    path: `installs/${installId}/available-roles?principal_type=${principalType}&operation_type=${operationType}`,
    orgId,
  })
}
