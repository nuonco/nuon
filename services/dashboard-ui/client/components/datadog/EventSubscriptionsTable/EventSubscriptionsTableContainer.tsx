import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import {
  getDatadogConnections,
  getDatadogEventSubscriptions,
} from '@/lib'
import { EventSubscriptionsTable } from './EventSubscriptionsTable'

export const EventSubscriptionsTableContainer = ({
  shouldPoll = false,
}: {
  shouldPoll?: boolean
}) => {
  const { org } = useOrg()

  // Subscriptions and connections are fetched in parallel; the table
  // joins them client-side to render connection names. Doing the join on
  // the client (rather than denormalizing on the backend) keeps the
  // backend list endpoint cheap and avoids a Preload that the lifecycle
  // hot path already doesn't pay.
  const connectionsQuery = useQuery({
    queryKey: ['datadog-connections', org?.id],
    queryFn: () => getDatadogConnections({ orgId: org!.id }),
    enabled: !!org?.id,
  })

  const subsQuery = useQuery({
    queryKey: ['datadog-event-subscriptions', org?.id],
    queryFn: () => getDatadogEventSubscriptions({ orgId: org!.id }),
    enabled: !!org?.id,
    refetchInterval: shouldPoll ? 5000 : false,
  })

  return (
    <EventSubscriptionsTable
      data={subsQuery.data ?? []}
      connections={connectionsQuery.data ?? []}
      isLoading={subsQuery.isLoading || connectionsQuery.isLoading}
    />
  )
}
