import { api } from '@/lib/api'

export async function createVCSConnectionWebhookSubscription({
  orgId,
  connectionId,
}: {
  orgId: string
  connectionId: string
}) {
  return api<void>({
    method: 'POST',
    orgId,
    path: `vcs/connections/${connectionId}/webhook-subscription`,
  })
}
