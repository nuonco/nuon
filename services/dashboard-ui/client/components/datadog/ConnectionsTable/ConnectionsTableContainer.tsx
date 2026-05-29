import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getDatadogConnections } from '@/lib'
import { ConnectionsTable } from './ConnectionsTable'

export const ConnectionsTableContainer = ({
  shouldPoll = false,
}: {
  shouldPoll?: boolean
}) => {
  const { org } = useOrg()

  const { data, isLoading } = useQuery({
    queryKey: ['datadog-connections', org?.id],
    queryFn: () => getDatadogConnections({ orgId: org!.id }),
    enabled: !!org?.id,
    refetchInterval: shouldPoll ? 5000 : false,
  })

  return <ConnectionsTable data={data ?? []} isLoading={isLoading} />
}
