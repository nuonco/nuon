import { api } from '@/lib/api'
import type { TOrgInvite } from '@/types'

export type TInviteUserBody = {
  email: string
}

export const inviteUser = ({
  body,
  orgId,
}: {
  body: TInviteUserBody
  orgId: string
}) =>
  api<TOrgInvite>({
    body,
    method: 'POST',
    orgId,
    path: `orgs/current/invites`,
  })
