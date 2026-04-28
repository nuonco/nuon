import { api } from '@/lib/api'

export const deleteCurrentOrgWebhook = ({
  orgId,
  webhookId,
}: {
  orgId: string
  webhookId: string
}) =>
  api({
    method: 'DELETE',
    orgId,
    path: `orgs/current/webhooks/${webhookId}`,
  })
