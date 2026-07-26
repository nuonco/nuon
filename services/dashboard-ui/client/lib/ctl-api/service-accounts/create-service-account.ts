import { api } from '@/lib/api'
import type { TAccount, TCreateServiceAccountBody } from '@/types'

export const createServiceAccount = ({
  body,
  orgId,
}: {
  body: TCreateServiceAccountBody
  orgId: string
}) =>
  api<TAccount>({
    body,
    method: 'POST',
    orgId,
    path: `service-accounts`,
  })
