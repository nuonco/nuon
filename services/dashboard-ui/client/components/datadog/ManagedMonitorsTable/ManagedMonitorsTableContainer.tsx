import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import {
  getDatadogConnections,
  getDatadogManagedMonitors,
} from '@/lib'
import { ManagedMonitorsTable } from './ManagedMonitorsTable'

export const ManagedMonitorsTableContainer = ({
  shouldPoll = false,
}: {
  shouldPoll?: boolean
}) => {
  const { org } = useOrg()

  const connectionsQuery = useQuery({
    queryKey: ['datadog-connections', org?.id],
    queryFn: () => getDatadogConnections({ orgId: org!.id }),
    enabled: !!org?.id,
  })

  const monitorsQuery = useQuery({
    queryKey: ['datadog-managed-monitors', org?.id],
    queryFn: () => getDatadogManagedMonitors({ orgId: org!.id }),
    enabled: !!org?.id,
    refetchInterval: shouldPoll ? 5000 : false,
  })

  return (
    <ManagedMonitorsTable
      data={monitorsQuery.data ?? []}
      connections={connectionsQuery.data ?? []}
      isLoading={monitorsQuery.isLoading || connectionsQuery.isLoading}
    />
  )
}
