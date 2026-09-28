import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getCloudConnections } from '@/lib'
import { ConnectionsList } from './ConnectionsList'

export const ConnectionsListContainer = () => {
  const { org } = useOrg()
  const query = useQuery({
    queryKey: ['cloud-connections', org?.id],
    queryFn: () => getCloudConnections({ orgId: org.id }),
    enabled: !!org?.id,
    placeholderData: keepPreviousData,
  })
  return (
    <ConnectionsList
      orgId={org?.id}
      connections={query.data ?? []}
      isLoading={query.isLoading}
      error={query.error}
    />
  )
}
