import { api } from '@/lib/api'
import type { TAccount, TUpdateServiceAccountBody } from '@/types'

export const updateServiceAccount = ({
  body,
  accountId,
  orgId,
}: {
  body: TUpdateServiceAccountBody
  accountId: string
  orgId: string
}) =>
  api<TAccount>({
    body,
    method: 'PATCH',
    orgId,
    path: `service-accounts/${accountId}`,
  })
