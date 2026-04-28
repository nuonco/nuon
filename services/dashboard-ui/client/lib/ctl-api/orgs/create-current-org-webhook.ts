import { api } from '@/lib/api'
import type { TCreateWebhookBody, TWebhook } from '@/types'

export const createCurrentOrgWebhook = ({
  body,
  orgId,
}: {
  body: TCreateWebhookBody
  orgId: string
}) =>
  api<TWebhook>({
    body,
    method: 'POST',
    orgId,
    path: `orgs/current/webhooks`,
  })
