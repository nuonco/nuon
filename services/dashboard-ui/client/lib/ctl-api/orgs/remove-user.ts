import { api } from '@/lib/api'
import type { TAccount } from '@/types'

export type TRemoveUserBody = {
  user_id: string
}

export const removeUser = ({
  body,
  orgId,
}: {
  body: TRemoveUserBody
  orgId: string
}) =>
  api<TAccount>({
    body,
    method: 'POST',
    orgId,
    path: `orgs/current/remove-user`,
  })
