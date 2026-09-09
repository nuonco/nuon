import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { getApps } from '@/lib'
import { useListQueryState } from '../../../hooks/use-list-query-state'
import { useOrg } from '../../../providers/org-provider'
import { AppsTable } from './AppsTable'

const PAGE_SIZE = 20

export const AppsTableContainer = () => {
  const { orgId } = useOrg()
  const list = useListQueryState({ pageSize: PAGE_SIZE })
  const {
    data: result,
    isLoading,
    isPlaceholderData,
    error,
  } = useQuery({
    queryKey: ['apps', orgId, ...list.queryKey],
    queryFn: () =>
      getApps({
        orgId: orgId!,
        q: list.search || undefined,
        offset: list.offset,
        limit: list.pageSize,
      }),
    enabled: !!orgId,
    placeholderData: keepPreviousData,
    refetchInterval: 15_000,
  })

  return (
    <AppsTable
      apps={result?.data ?? []}
      orgId={orgId ?? ''}
      search={list.search}
      onSearchChange={list.setSearch}
      offset={list.offset}
      pageSize={list.pageSize}
      hasNext={result?.pagination?.hasNext ?? false}
      onOffsetChange={list.setOffset}
      loading={isLoading}
      fetching={isPlaceholderData}
      error={error}
    />
  )
}
