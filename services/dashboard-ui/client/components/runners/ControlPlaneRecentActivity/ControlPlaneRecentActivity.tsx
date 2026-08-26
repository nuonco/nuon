import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Timeline, type ITimeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import { TimelineSkeleton } from '@/components/common/TimelineSkeleton'
import { Text } from '@/components/common/Text'
import type { TOrgComponentBuildHistoryItem } from '@/lib'

type TComponentBuildHistoryEvent = TOrgComponentBuildHistoryItem & {
  created_at: string
}

interface IControlPlaneRecentActivity
  extends Omit<
    ITimeline<TComponentBuildHistoryEvent>,
    'events' | 'renderEvent' | 'pagination'
  > {
  activity: TOrgComponentBuildHistoryItem[]
  orgId?: string
  isLoading: boolean
  isFetching?: boolean
  nextCursor: string | null
  previousCursor: string | null
  onCursorChange: (cursor: string | null) => void
}

export const ControlPlaneRecentActivity = ({
  activity,
  orgId,
  isLoading,
  isFetching = false,
  nextCursor,
  previousCursor,
  onCursorChange,
  ...props
}: IControlPlaneRecentActivity) => {
  if (isLoading) {
    return (
      <>
        <Skeleton height="24px" width="110px" />
        <TimelineSkeleton eventCount={10} />
      </>
    )
  }

  const events = activity.map((item) => ({
    ...item,
    created_at: item?.build?.created_at ?? '',
  }))

  return (
    <div className="flex flex-col">
      {events.length === 0 ? (
        <Text theme="neutral">No component builds yet.</Text>
      ) : null}
      <Timeline<TComponentBuildHistoryEvent>
        events={events}
        getEventKey={(item) => item?.build?.id}
        pagination={{ hasNext: false, offset: 0 }}
        renderEvent={(item) => {
          const build = item?.build
          const href = `/${orgId ?? ''}/apps/${item?.app_id ?? ''}/components/${item?.component_id ?? ''}/builds/${build?.id ?? ''}`
          return (
            <TimelineEvent
              key={build?.id}
              caption={<ID>{build?.id}</ID>}
              createdAt={item?.created_at}
              status={build?.status_v2?.status ?? build?.status}
              title={
                <Link href={href} variant="inline">
                  {item?.component_name ?? 'Component'} build
                </Link>
              }
              underline={
                build?.status_description ? (
                  <Text variant="label" theme="neutral">
                    {build.status_description}
                  </Text>
                ) : null
              }
            />
          )
        }}
        {...props}
      />
      {previousCursor || nextCursor ? (
        <div className="flex items-center gap-3 self-center">
          <Button
            disabled={!previousCursor || isFetching}
            onClick={() => onCursorChange(previousCursor)}
          >
            <Icon variant="ArrowLeftIcon" />
            Newer
          </Button>
          <Button
            disabled={!nextCursor || isFetching}
            onClick={() => onCursorChange(nextCursor)}
          >
            Older
            <Icon variant="ArrowRightIcon" />
          </Button>
        </div>
      ) : null}
    </div>
  )
}
