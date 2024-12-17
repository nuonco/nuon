'use client'

import classNames from 'classnames'
import React, { type FC } from 'react'
import { CaretRight } from '@phosphor-icons/react'
import { Link } from '@/components/Link'
import { Time } from '@/components/Time'
import { Text } from '@/components/Typography'
import { sentanceCase } from '@/utils'

const EventStatus: FC<{ status?: string }> = ({ status = 'waiting' }) => {
  const statusColor = {
    'bg-green-800 dark:bg-green-500':
      status === 'finished' || status === 'active',
    'bg-red-600 dark:bg-red-500': status === 'failed' || status === 'error',
    'bg-cool-grey-600 dark:bg-cool-grey-500': status === 'noop',
    'bg-orange-800 dark:bg-orange-500':
      status === 'waiting' ||
      status === 'started' ||
      status === 'in-progress' ||
      status === 'building' ||
      status === 'queued' ||
      status === 'planning' ||
      status === 'deploying',
  }

  return (
    <span
      className={classNames('w-4 h-4 rounded-full relative', {
        'animate-pulse':
          status === 'waiting' ||
          status === 'started' ||
          status === 'in-progress' ||
          status === 'building' ||
          status === 'queued' ||
          status === 'planning' ||
          status === 'deploying',
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
  return (
    <Link
      className="!block w-full !p-0 timeline-event"
      href={href}
      variant="ghost"
    >
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
                  status === 'error'
              ),
            })}
          />
          {href && <CaretRight />}
        </div>
      </div>
    </Link>
  )
}

export interface ITimeline {
  events: Array<ITimelineEvent>
  emptyMessage?: string
}

export const Timeline: FC<ITimeline> = ({
  events,
  emptyMessage = 'No events to show',
}) => {
  return (
    <div className="flex flex-col gap-2 timeline">
      {events?.length ? (
        events.map((event, i) => (
          <TimelineEvent key={event.id} {...event} isMostRecent={i === 0} />
        ))
      ) : (
        <Text variant="reg-14">{emptyMessage}</Text>
      )}
    </div>
  )
}
