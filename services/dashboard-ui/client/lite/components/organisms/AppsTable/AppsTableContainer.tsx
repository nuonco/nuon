import { keepPreviousData, useQueries, useQuery } from '@tanstack/react-query'
import { getAppBranches, getApps } from '@/lib'
import { useListQueryState } from '../../../hooks/use-list-query-state'
import { useOrg } from '../../../providers/org-provider'
import {
  appSetupDescriptor,
  appSetupStateFromBranches,
} from '../../../utils/app-setup'
import { isWizardComplete } from '../../../utils/wizard'
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

  const apps = result?.data ?? []
  const branchQueries = useQueries({
    queries: apps.map((app) => ({
      queryKey: ['app-branches', orgId, app?.id, 'setup-complete'],
      queryFn: () =>
        getAppBranches({
          orgId: orgId!,
          appId: app.id!,
          limit: 50,
          offset: 0,
        }),
      enabled: !!orgId && !!app?.id,
      staleTime: 60_000,
    })),
  })

  const incompleteIds = new Set(
    apps.flatMap((app, index) => {
      const query = branchQueries[index]
      if (!app?.id || !query || query.isPending || query.isError) return []
      const complete = isWizardComplete(
        appSetupDescriptor,
        appSetupStateFromBranches(query.data?.data)
      )
      return complete ? [] : [app.id]
    })
  )

  return (
    <AppsTable
      apps={apps}
      orgId={orgId ?? ''}
      search={list.search}
      onSearchChange={list.setSearch}
      offset={list.offset}
      pageSize={list.pageSize}
      hasNext={result?.pagination?.hasNext ?? false}
      onOffsetChange={list.setOffset}
      loading={isLoading}
      fetching={isPlaceholderData}
      incompleteIds={incompleteIds}
      error={error}
    />
  )
}
