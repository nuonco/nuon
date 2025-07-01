import React from 'react'
import { cn } from '@/stratus/components/helpers'
import {
  Badge,
  type IBadge,
  Pagination,
  type IPagination,
  Text,
  Time,
} from '@/stratus/components/common'
import { Status, type TStatusType } from '@/stratus/components/statuses'
import {
  type IHasCreatedAt,
  formatToRelativeDay,
  parseActivityTimeline,
} from './helpers'
import './Timeline.css'

export interface ITimeline<T extends IHasCreatedAt>
  extends Omit<React.HTMLAttributes<HTMLDivElement>, 'children'> {
  events: Array<T>
  pagination: Omit<IPagination, 'position'>
  renderEvent?: (event: T, idx: number) => React.ReactNode
}

export const Timeline = <T extends IHasCreatedAt>({
  className,
  events,
  pagination,
  renderEvent,
  ...props
}: ITimeline<T>) => {
  const groupedEvents = parseActivityTimeline(events)
  const dates = Object.keys(groupedEvents).sort((a, b) => b.localeCompare(a))

  // TODO(nnnat): remove this new class once the old timeline is gone
  return (
    <div className={cn('timeline new', className)} {...props}>
      {dates.map((date) => (
        <div key={date} className="timeline-group">
          <Text className="timeline-date">{formatToRelativeDay(date)}</Text>
          <div className="timeline-events">
            {groupedEvents[date].map((event, idx) => (
              <React.Fragment key={event.created_at}>
                {renderEvent ? renderEvent(event, idx) : null}
              </React.Fragment>
            ))}
          </div>
        </div>
      ))}
      <Pagination {...pagination} />
    </div>
  )
}

export interface ITimelineEvent
  extends Omit<React.HTMLAttributes<HTMLDivElement>, 'children' | 'title'> {
  additionalCaption?: string
  badge?: IBadge
  caption?: string
  createdAt: string
  createdBy?: string
  status: TStatusType
  title: React.ReactNode | string
}

export const TimelineEvent = ({
  additionalCaption,
  badge,
  caption,
  className,
  createdAt,
  createdBy,
  status,
  title,
  ...props
}: ITimelineEvent) => {
  return (
    <div className={cn('timeline-event', className)} {...props}>
      <Status status={status} variant="timeline" isWithoutText />
      <div className="w-full">
        <hgroup className="w-full flex items-center justify-between">
          <Text variant="body" weight="strong">
            {title}
          </Text>

          <Text variant="subtext" theme="muted">
            <Time time={createdAt} format="relative" />{' '}
            {createdBy ? `by ${createdBy}` : null}
          </Text>
        </hgroup>
        <span className="flex items-center gap-2">
          {caption ? (
            <Text variant="subtext" theme="muted">
              {caption}
            </Text>
          ) : null}{' '}
          {additionalCaption ? (
            <Text variant="label" theme="muted">
              {additionalCaption}
            </Text>
          ) : null}
          {badge ? <Badge {...badge} /> : null}
        </span>
      </div>
    </div>
  )
}
