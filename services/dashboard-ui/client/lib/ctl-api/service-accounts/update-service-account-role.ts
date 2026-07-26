import { api } from '@/lib/api'
import type { TAccount, TUpdateServiceAccountRoleBody } from '@/types'

export const updateServiceAccountRole = ({
  body,
  accountId,
  orgId,
}: {
  body: TUpdateServiceAccountRoleBody
  accountId: string
  orgId: string
}) =>
  api<TAccount>({
    body,
    method: 'PATCH',
    orgId,
    path: `service-accounts/${accountId}/role`,
  })
