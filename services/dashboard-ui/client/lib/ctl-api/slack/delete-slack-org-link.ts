import { api } from '@/lib/api'

export const deleteSlackOrgLink = ({
  linkId,
  orgId,
}: {
  linkId: string
  orgId: string
}) =>
  api({
    method: 'DELETE',
    orgId,
    path: `orgs/${orgId}/slack/org-links/${linkId}`,
  })
