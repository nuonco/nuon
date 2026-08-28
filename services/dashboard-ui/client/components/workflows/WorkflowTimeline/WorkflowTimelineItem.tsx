import type { ReactNode } from 'react'
import type { IBadge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Duration } from '@/components/common/Duration'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import {
  TimelineEvent,
  type ITimelineEvent,
} from '@/components/common/TimelineEvent'

export interface IWorkflowTimelineItem {
  id: string
  title: string
  href?: string
  onSelect?: () => void
  status: ITimelineEvent['status']
  createdAt: string
  updatedAt?: string
  finishedAt?: string
  finished?: boolean
  durationNanoseconds?: number
  createdBy?: string
  titleBadges?: ReactNode
  additionalCaption?: ReactNode
  metadata?: ReactNode
  actions?: ReactNode
  badge?: IBadge
}

export const WorkflowTimelineItem = ({
  id,
  title,
  href,
  onSelect,
  status,
  createdAt,
  updatedAt,
  finishedAt,
  finished,
  durationNanoseconds,
  createdBy,
  titleBadges,
  additionalCaption,
  metadata,
  actions,
  badge,
}: IWorkflowTimelineItem) => {
  const titleContent = href ? (
    <Link className="inline-flex gap-2 items-center" href={href}>
      {title}
    </Link>
  ) : onSelect ? (
    <Button variant="ghost" className="!h-auto !p-0" onClick={onSelect}>
      {title}
    </Button>
  ) : (
    title
  )

  return (
    <TimelineEvent
      actions={actions}
      additionalCaption={additionalCaption}
      badge={badge}
      caption={<ID>{id}</ID>}
      createdAt={createdAt}
      createdBy={createdBy}
      status={status}
      title={
        <span className="flex items-center gap-4 mb-1">
          {titleContent}
          {titleBadges}
        </span>
      }
      underline={
        <span className="flex items-center gap-6 mt-1 flex-wrap">
          {metadata}
          <Text
            flex
            className="gap-1"
            variant="subtext"
            theme="neutral"
            title="Created"
          >
            <Icon variant="CalendarBlankIcon" />
            <Time time={createdAt} variant="subtext" />
          </Text>
          {updatedAt || finishedAt ? (
            <Text
              flex
              className="gap-1"
              variant="subtext"
              theme="neutral"
              title={finished ? 'Finished' : 'Last updated'}
            >
              <Icon variant="ClockClockwiseIcon" />
              <Time
                time={finished && finishedAt ? finishedAt : updatedAt}
                variant="subtext"
                format="relative"
              />
            </Text>
          ) : null}
          {finished && (durationNanoseconds || finishedAt) ? (
            <Text
              flex
              className="gap-1"
              variant="subtext"
              theme="neutral"
              title="Duration"
            >
              <Icon variant="TimerIcon" />
              <Duration
                nanoseconds={durationNanoseconds}
                beginTime={durationNanoseconds ? undefined : createdAt}
                endTime={durationNanoseconds ? undefined : finishedAt}
                variant="subtext"
              />
            </Text>
          ) : null}
        </span>
      }
    />
  )
}
