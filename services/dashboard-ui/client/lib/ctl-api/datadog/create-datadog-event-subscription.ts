import { api } from '@/lib/api'
import type {
  TCreateDatadogEventSubscriptionBody,
  TDatadogEventSubscription,
} from '@/types'

export const createDatadogEventSubscription = ({
  orgId,
  body,
}: {
  orgId: string
  body: TCreateDatadogEventSubscriptionBody
}) =>
  api<TDatadogEventSubscription>({
    body,
    method: 'POST',
    orgId,
    path: `orgs/${orgId}/datadog/event-subscriptions`,
  })
