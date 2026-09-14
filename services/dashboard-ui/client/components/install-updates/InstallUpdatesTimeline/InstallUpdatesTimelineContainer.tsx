import { useInfiniteQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallUpdates } from '@/lib'
import { InstallUpdatesTimeline } from './InstallUpdatesTimeline'

const PAGE_SIZE = 20

export const InstallUpdatesTimelineContainer = ({
  pollInterval = 20000,
  shouldPoll,
}: {
  pollInterval?: number
  shouldPoll?: boolean
}) => {
  const { org } = useOrg()
  const { install } = useInstall()

  const { data, fetchNextPage, hasNextPage, isLoading } = useInfiniteQuery({
    queryKey: ['install-updates', org?.id, install?.id],
    queryFn: ({ pageParam }) =>
      getInstallUpdates({
        orgId: org!.id,
        installId: install!.id,
        page: pageParam,
        limit: PAGE_SIZE,
      }),
    initialPageParam: 0,
    getNextPageParam: (lastPage) =>
      lastPage.has_more ? lastPage.page + 1 : undefined,
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!install?.id,
  })

  return (
    <InstallUpdatesTimeline
      updates={data?.pages.flatMap((page) => page.updates) ?? []}
      isLoading={isLoading}
      hasMore={hasNextPage}
      onLoadMore={hasNextPage ? () => void fetchNextPage() : undefined}
      orgId={org?.id}
      installId={install?.id}
      appId={install?.app_id}
    />
  )
}
