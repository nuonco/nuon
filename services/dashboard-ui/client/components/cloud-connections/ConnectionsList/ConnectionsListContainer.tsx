import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useSearchParams } from 'react-router'
import { useOrg } from '@/hooks/use-org'
import { getCloudConnections } from '@/lib'
import { ConnectionsList } from './ConnectionsList'

const LIMIT = 20

export const ConnectionsListContainer = () => {
  const { org } = useOrg()
  const [searchParams] = useSearchParams()
  const offset = Number(searchParams.get('offset') ?? 0)
  const q = searchParams.get('q') ?? ''
  const query = useQuery({
    queryKey: ['cloud-connections', org?.id, offset, q],
    queryFn: () =>
      getCloudConnections({ orgId: org.id, offset, limit: LIMIT, q }),
    enabled: !!org?.id,
    placeholderData: keepPreviousData,
  })
  return (
    <ConnectionsList
      orgId={org?.id}
      connections={query.data?.data ?? []}
      isLoading={query.isLoading}
      error={query.error}
      pagination={{
        hasNext: query.data?.pagination?.hasNext ?? false,
        offset,
        limit: LIMIT,
      }}
    />
  )
}
