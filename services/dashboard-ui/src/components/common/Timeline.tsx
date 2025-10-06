'use client'

import React, { useEffect } from 'react'
import { PaginationProvider } from '@/providers/pagination-provider'
import { usePagination } from '@/hooks/use-pagination'
import { cn } from '@/utils/classnames'
import {
  type IHasCreatedAt,
  formatToRelativeDay,
  parseActivityTimeline,
} from '@/utils/timeline-utils'
import { Pagination, type IPagination } from './Pagination'
import { Text } from './Text'
import { TimelineSkeleton } from './TimelineSkeleton'

export interface ITimeline<T extends IHasCreatedAt>
  extends Omit<React.HTMLAttributes<HTMLDivElement>, 'children'> {
  events: Array<T>
  eventCount?: number
  pagination: Omit<IPagination, 'position'>
  renderEvent?: (event: T, idx: number) => React.ReactNode
}

const TimelineBase = <T extends IHasCreatedAt>({
  className,
  events,
  eventCount = 10,
  pagination,
  renderEvent,
  ...props
}: ITimeline<T>) => {
  const { isPaginating, setIsPaginating } = usePagination()
  const groupedEvents = parseActivityTimeline(events)
  const dates = Object.keys(groupedEvents).sort((a, b) => b.localeCompare(a))

  useEffect(() => {
    setIsPaginating(false)
  }, [events])

  return (
    <div className={cn('flex flex-col', className)} {...props}>
      {isPaginating ? (
        <TimelineSkeleton eventCount={eventCount} />
      ) : (
        dates.map((date) => (
          <div key={date} className="timeline-group">
            <Text className="timeline-date">{formatToRelativeDay(date)}</Text>
            <div className="timeline-events">
              {groupedEvents[date].map((event, idx) => (
                <React.Fragment key={event.created_at}>
                  {renderEvent ? renderEvent?.(event, idx) : null}
                </React.Fragment>
              ))}
            </div>
          </div>
        ))
      )}
      <Pagination {...pagination} />
    </div>
  )
}

export const Timeline = <T extends IHasCreatedAt>(props: ITimeline<T>) => (
  <PaginationProvider>
    <TimelineBase<T> {...props} />
  </PaginationProvider>
)
