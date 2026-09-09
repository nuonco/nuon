import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useListQueryState } from '../../../hooks/use-list-query-state'
import { APPS_PAGE_SIZE, appsListQuery } from '../../../queries/apps'
import { useOrg } from '../../../providers/org-provider'
import { AppsTable } from './AppsTable'

export const AppsTableContainer = () => {
  const { orgId } = useOrg()
  const list = useListQueryState({ pageSize: APPS_PAGE_SIZE })
  const {
    data: result,
    isLoading,
    isPlaceholderData,
    error,
  } = useQuery({
    ...appsListQuery({
      orgId: orgId!,
      search: list.search,
      offset: list.offset,
      pageSize: list.pageSize,
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
