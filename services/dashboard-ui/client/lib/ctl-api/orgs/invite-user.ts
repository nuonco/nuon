import { api } from '@/lib/api'
import type { TOrgInvite } from '@/types'

export type TInviteUserBody = {
  email: string
  first_name?: string
  last_name?: string
  role_type?: string
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
