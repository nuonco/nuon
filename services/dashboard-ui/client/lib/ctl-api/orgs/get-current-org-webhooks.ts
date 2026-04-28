import { api } from '@/lib/api'
import type { TWebhook } from '@/types'

export const getCurrentOrgWebhooks = ({ orgId }: { orgId: string }) =>
  api<TWebhook[]>({
    orgId,
    path: `orgs/current/webhooks`,
  })
