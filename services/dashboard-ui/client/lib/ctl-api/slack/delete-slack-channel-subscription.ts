import { api } from '@/lib/api'

export const deleteSlackChannelSubscription = ({
  orgId,
  subId,
}: {
  orgId: string
  subId: string
}) =>
  api({
    method: 'DELETE',
    orgId,
    path: `orgs/${orgId}/slack/channel-subscriptions/${subId}`,
  })
