'use client'

import classNames from 'classnames'
import React, { type FC } from 'react'
import { CaretRight } from '@phosphor-icons/react'
import { EmptyStateGraphic } from '@/components/EmptyStateGraphic'
import { Link } from '@/components/Link'
import { Time } from '@/components/Time'
import { Text } from '@/components/Typography'
import { sentanceCase } from '@/utils'

export const EventStatus: FC<{ status?: string }> = ({
  status = 'waiting',
}) => {
  const statusColor = {
    'bg-green-800 dark:bg-green-500':
      status === 'finished' || status === 'active',
    'bg-red-600 dark:bg-red-500':
      status === 'failed' ||
      status === 'error' ||
      status === 'unknown' ||
      status === 'timed-out',
    'bg-cool-grey-600 dark:bg-cool-grey-500':
      status === 'noop' ||
      status === 'pending' ||
      status === 'inactive' ||
      status === 'cancelled' ||
      status === 'not-attempted',
    'bg-orange-800 dark:bg-orange-500':
      status === 'executing' ||
      status === 'waiting' ||
      status === 'started' ||
      status === 'in-progress' ||
      status === 'building' ||
      status === 'queued' ||
      status === 'planning' ||
      status === 'provisioning' ||
      status === 'syncing' ||
      status === 'deploying' ||
      status === 'available',
  }

  return (
    <span
      className={classNames('w-4 h-4 rounded-full relative', {
        'animate-pulse':
          status === 'executing' ||
          status === 'waiting' ||
          status === 'started' ||
          status === 'in-progress' ||
          status === 'building' ||
          status === 'queued' ||
          status === 'planning' ||
          status === 'pending' ||
          status === 'syncing' ||
          status === 'deploying' ||
          status === 'available',
      })}
    >
      <span
        className={classNames(
          'w-2 h-2 top-1 rounded-full absolute',
          statusColor
        )}
      />
      <span
        className={classNames(
          'w-4 h-4 rounded-full top-0 -left-1 opacity-25 absolute',
          statusColor
        )}
      />
    </span>
  )
}

interface ITimelineEvent {
  id: string
  status: string
  underline: React.ReactNode
  time: string
  href: string | null
  isMostRecent?: boolean
}

export const TimelineEvent: FC<ITimelineEvent> = ({
  status,
  underline,
  time,
  href,
  isMostRecent = false,
}) => {
  const Event = (
    <div
      className={classNames('flex items-center justify-between p-4', {
        'border rounded-md shadow-sm': isMostRecent,
      })}
    >
      <div className="flex flex-col">
        <span className="flex items-center gap-3">
          <EventStatus status={status} />
          <Text variant="med-12">{sentanceCase(status)}</Text>
        </span>

        <Text className="flex items-center gap-2 ml-7" variant="reg-12">
          {underline}
        </Text>
      </div>

      <div className="flex items-center gap-2">
        <Time
          time={time}
          format="relative"
          variant="reg-12"
          className={classNames({
            'text-black/60 dark:text-white/60': !Boolean(
              status === 'finished' ||
                status === 'failed' ||
                status === 'active' ||
                status === 'error' ||
                status === 'not-attempted' ||
                status === 'timed-out'
            ),
          })}
        />
        {href && <CaretRight />}
      </div>
    </div>
  )

  return href ? (
    <Link
      className="!block w-full !p-0 timeline-event"
      href={href}
      variant="ghost"
    >
      {Event}
    </Link>
  ) : (
    Event
  )
}

export interface ITimeline {
  events: Array<ITimelineEvent>
  emptyMessage?: string
  emptyTitle?: string
}

export const Timeline: FC<ITimeline> = ({
  events,
  emptyMessage = 'No events to show',
  emptyTitle = 'Nothing to show',
}) => {
  return (
    <div className="flex flex-col gap-2 timeline">
      {events?.length ? (
        events.map((event, i) => (
          <TimelineEvent key={event.id} {...event} isMostRecent={i === 0} />
        ))
      ) : (
        <div className="m-auto flex flex-col items-center max-w-[200px] my-6">
          <EmptyStateGraphic variant="history" />
          <Text className="mt-6" variant="med-14">
            {emptyTitle}
          </Text>
          <Text variant="reg-12" className="text-center">
            {emptyMessage}
          </Text>
        </div>
      )}
    </div>
  )
}
