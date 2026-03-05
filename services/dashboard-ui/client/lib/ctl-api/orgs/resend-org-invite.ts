import { api } from '@/lib/api'
import type { TOrgInvite } from '@/types'

export const resendOrgInvite = ({
  inviteId,
  orgId,
}: {
  inviteId: string
  orgId: string
}) =>
  api<TOrgInvite>({
    method: 'POST',
    orgId,
    path: `orgs/current/invites/${inviteId}/resend`,
  })
