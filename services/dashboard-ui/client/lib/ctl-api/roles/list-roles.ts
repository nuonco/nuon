import { api } from '@/lib/api'
import type { TRoleInfo } from '@/types'

export const listRoles = ({ orgId }: { orgId: string }) =>
  api<TRoleInfo[]>({
    path: `roles`,
    orgId,
  })
