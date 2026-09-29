import type { Interests } from '@/components/interests/types'
import type { SubscriptionMatch } from '@/components/match/types'
import { api } from '@/lib/api'
import type { TSlackChannelSubscription } from '@/types'

export interface UpdateSlackChannelSubscriptionBody {
  channel_id?: string
  channel_name?: string
  match?: SubscriptionMatch | null
  interests?: Interests
}

export const updateSlackChannelSubscription = ({
  body,
  orgId,
  subId,
}: {
  body: UpdateSlackChannelSubscriptionBody
  orgId: string
  subId: string
}) =>
  api<TSlackChannelSubscription>({
    body,
    method: 'PATCH',
    orgId,
    path: `orgs/${orgId}/slack/channel-subscriptions/${subId}`,
  })
