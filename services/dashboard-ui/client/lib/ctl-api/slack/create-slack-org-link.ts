import { api } from '@/lib/api'
import type { TCreateSlackOrgLinkBody, TSlackOrgLink } from '@/types'

export const createSlackOrgLink = ({
  body,
  orgId,
}: {
  body: TCreateSlackOrgLinkBody
  orgId: string
}) =>
  api<TSlackOrgLink>({
    body,
    method: 'POST',
    orgId,
    path: `orgs/${orgId}/slack/org-links`,
  })
