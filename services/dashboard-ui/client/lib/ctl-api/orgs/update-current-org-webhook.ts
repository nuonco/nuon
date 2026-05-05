import type { Interests } from '@/components/interests'
import { api } from '@/lib/api'
import type { TWebhook } from '@/types'

// PATCH /v1/orgs/current/webhooks/{webhook_id}
//
// Backend: services/ctl-api/internal/app/orgs/service/current_webhooks.go
// (UpdateCurrentOrgWebhook). Replaces `interests` wholesale and (optionally)
// rotates the signing secret. WebhookURL cannot be changed in place — delete
// + recreate to rename.
//
// Body shape:
//   webhook_secret: string | undefined
//     - undefined → leave unchanged
//     - ""        → clear the existing secret
//     - "..."     → rotate to this value
//   interests: Interests
//     - PUT-style replacement of the entire interests filter
export type TUpdateCurrentOrgWebhookBody = {
  webhook_secret?: string
  interests: Interests
}

export const updateCurrentOrgWebhook = ({
  body,
  orgId,
  webhookId,
}: {
  body: TUpdateCurrentOrgWebhookBody
  orgId: string
  webhookId: string
}) =>
  api<TWebhook>({
    body,
    method: 'PATCH',
    orgId,
    path: `orgs/current/webhooks/${webhookId}`,
  })
