import { api } from '@/lib/api'

export const deleteDatadogEventSubscription = ({
  orgId,
  subscriptionId,
}: {
  orgId: string
  subscriptionId: string
}) =>
  api<void>({
    method: 'DELETE',
    orgId,
    path: `orgs/${orgId}/datadog/event-subscriptions/${subscriptionId}`,
  })
