import { api } from '@/lib/api'
import type { TAccount } from '@/types'

export type TUpdateAccountRoleBody = {
  role_type: string
}

export const updateAccountRole = ({
  body,
  accountId,
  orgId,
}: {
  body: TUpdateAccountRoleBody
  accountId: string
  orgId: string
}) =>
  api<TAccount>({
    body,
    method: 'PATCH',
    orgId,
    path: `orgs/current/accounts/${accountId}/role`,
  })
