import { api } from '@/lib/api'
import type { TRoleContext, TRoleInfo } from '@/types'

export const listRoles = ({
  orgId,
  context,
}: {
  orgId: string
  context?: TRoleContext
}) =>
  api<TRoleInfo[]>({
    path: context ? `roles?context=${context}` : `roles`,
    orgId,
  })
