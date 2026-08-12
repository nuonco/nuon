import { api } from '@/lib/api'
import type { TRoleInfo } from '@/types'

export const getRole = ({
  roleId,
  orgId,
}: {
  roleId: string
  orgId: string
}) =>
  api<TRoleInfo>({
    path: `roles/${roleId}`,
    orgId,
  })
