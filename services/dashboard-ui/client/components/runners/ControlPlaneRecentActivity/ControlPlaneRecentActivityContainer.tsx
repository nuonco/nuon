import { useQuery } from '@tanstack/react-query'
import { useSearchParams } from 'react-router'
import type { ITimeline } from '@/components/common/Timeline'
import { useOrg } from '@/hooks/use-org'
import {
  getOrgComponentBuildHistory,
  type TOrgComponentBuildHistoryItem,
} from '@/lib'
import { ControlPlaneRecentActivity } from './ControlPlaneRecentActivity'

const BUILD_HISTORY_CURSOR_PARAM = 'build_cursor'
const BUILD_HISTORY_LIMIT = 10

interface IControlPlaneRecentActivityContainer
  extends Omit<
    ITimeline<TOrgComponentBuildHistoryItem & { created_at: string }>,
    'events' | 'renderEvent' | 'pagination'
  > {
  shouldPoll?: boolean
  pollInterval?: number
}

export const ControlPlaneRecentActivityContainer = ({
  shouldPoll = false,
  pollInterval = 20000,
  ...props
}: IControlPlaneRecentActivityContainer) => {
  const { org } = useOrg()
  const [searchParams, setSearchParams] = useSearchParams()
  const cursor = searchParams.get(BUILD_HISTORY_CURSOR_PARAM) ?? undefined

  const {
    data: result,
    isFetching,
    isLoading,
  } = useQuery({
    queryKey: ['org-component-build-history', org?.id, cursor],
    queryFn: () =>
      getOrgComponentBuildHistory({
        orgId: org!.id,
        limit: BUILD_HISTORY_LIMIT,
        cursor,
      }),
    refetchInterval: shouldPoll && !cursor ? pollInterval : false,
    enabled: !!org?.id,
  })

  const setCursor = (nextCursor: string | null) => {
    const params = new URLSearchParams(searchParams)
    if (nextCursor) {
      params.set(BUILD_HISTORY_CURSOR_PARAM, nextCursor)
    } else {
      params.delete(BUILD_HISTORY_CURSOR_PARAM)
    }
    setSearchParams(params)
  }

  return (
    <ControlPlaneRecentActivity
      activity={result?.items ?? []}
      orgId={org?.id}
      isFetching={isFetching}
      isLoading={isLoading}
      nextCursor={result?.next_cursor ?? null}
      previousCursor={result?.previous_cursor ?? null}
      onCursorChange={setCursor}
      {...props}
    />
  )
}
