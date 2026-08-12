import { api } from '@/lib/api'
import type { TCreateRoleBody, TRoleInfo } from '@/types'

export const createRole = ({
  body,
  orgId,
}: {
  body: TCreateRoleBody
  orgId: string
}) =>
  api<TRoleInfo>({
    path: `roles`,
    method: 'POST',
    body,
    orgId,
  })
