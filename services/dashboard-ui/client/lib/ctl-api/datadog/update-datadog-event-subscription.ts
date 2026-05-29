import { api } from '@/lib/api'
import type {
  TDatadogEventSubscription,
  TUpdateDatadogEventSubscriptionBody,
} from '@/types'

export const updateDatadogEventSubscription = ({
  orgId,
  subscriptionId,
  body,
}: {
  orgId: string
  subscriptionId: string
  body: TUpdateDatadogEventSubscriptionBody
}) =>
  api<TDatadogEventSubscription>({
    body,
    method: 'PATCH',
    orgId,
    path: `orgs/${orgId}/datadog/event-subscriptions/${subscriptionId}`,
  })
