import { api } from '@/lib/api'
import type {
  TCreateSlackChannelSubscriptionBody,
  TSlackChannelSubscription,
} from '@/types'

export const createSlackChannelSubscription = ({
  body,
  orgId,
}: {
  body: TCreateSlackChannelSubscriptionBody
  orgId: string
}) =>
  api<TSlackChannelSubscription>({
    body,
    method: 'POST',
    orgId,
    path: `orgs/${orgId}/slack/channel-subscriptions`,
  })
