import { api } from '@/lib/api'
import type { TRoleInfo, TUpdateRoleBody } from '@/types'

export const updateRole = ({
  roleId,
  body,
  orgId,
}: {
  roleId: string
  body: TUpdateRoleBody
  orgId: string
}) =>
  api<TRoleInfo>({
    path: `roles/${roleId}`,
    method: 'PATCH',
    body,
    orgId,
  })
