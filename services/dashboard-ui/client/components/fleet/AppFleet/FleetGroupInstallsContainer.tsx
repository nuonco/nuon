import { useInfiniteQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { FLEET_NONE, getAppFleetInstalls } from '@/lib'
import type { TAppFleetFilters } from '@/types'
import { FleetGroupInstalls } from './FleetGroupInstalls'

const PAGE_SIZE = 50

export interface IFleetGroupInstallsContainer {
  appId: string
  branchId?: string
  groupName?: string
  filters: TAppFleetFilters
  labelColors?: Record<string, string>
}

export const FleetGroupInstallsContainer = ({
  appId,
  branchId,
  groupName,
  filters,
  labelColors,
}: IFleetGroupInstallsContainer) => {
  const { org } = useOrg()

  const { data, fetchNextPage, hasNextPage, isFetchingNextPage, isLoading } =
    useInfiniteQuery({
      queryKey: [
        'app-fleet-installs',
        org?.id,
        appId,
        branchId,
        groupName,
        filters,
      ],
      queryFn: ({ pageParam }) =>
        getAppFleetInstalls({
          orgId: org!.id,
          appId,
          ...filters,
          app_branch_id: branchId ?? FLEET_NONE,
          app_branch_group: groupName ?? FLEET_NONE,
          limit: PAGE_SIZE,
          offset: pageParam,
        }),
      initialPageParam: 0,
      getNextPageParam: (lastPage) =>
        lastPage.pagination.hasNext
          ? lastPage.pagination.offset + lastPage.data.length
          : undefined,
      enabled: !!org?.id && !!appId,
    })

  return (
    <FleetGroupInstalls
      installs={data?.pages.flatMap((page) => page.data) ?? []}
      orgId={org?.id}
      appId={appId}
      labelColors={labelColors}
      isLoading={isLoading}
      isLoadingMore={isFetchingNextPage}
      onLoadMore={hasNextPage ? () => void fetchNextPage() : undefined}
    />
  )
}
