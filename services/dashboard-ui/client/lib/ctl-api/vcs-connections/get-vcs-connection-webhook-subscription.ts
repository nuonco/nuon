import { api } from '@/lib/api'
import type { TVCSWebhookSubscription } from '@/types'

export async function getVCSConnectionWebhookSubscription({
  orgId,
  connectionId,
}: {
  orgId: string
  connectionId: string
}) {
  return api<TVCSWebhookSubscription>({
    orgId,
    path: `vcs/connections/${connectionId}/webhook-subscription`,
  })
}
