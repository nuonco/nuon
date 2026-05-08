import { api } from '@/lib/api'
import type { TSlackChannelSubscription } from '@/types'

export const getSlackChannelSubscriptions = ({ orgId }: { orgId: string }) =>
  api<TSlackChannelSubscription[]>({
    orgId,
    path: `orgs/${orgId}/slack/channel-subscriptions`,
  })
