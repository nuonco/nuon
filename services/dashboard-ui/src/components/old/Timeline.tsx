'use client'

import classNames from 'classnames'
import React, { type FC } from 'react'
import { CaretRightIcon } from '@phosphor-icons/react'
import { Link } from '@/components/old/Link'
import { Time } from '@/components/old/Time'
import { Text } from '@/components/old/Typography'
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
      status === 'approval-denied' ||
      status === 'timed-out',
    'bg-cool-grey-600 dark:bg-cool-grey-500':
      status === 'noop' ||
      status === 'pending' ||
      status === 'inactive' ||
      status === 'cancelled' ||
      status === 'not-attempted' ||
      status === 'discarded' ||
      status === 'deprovisioned' ||
      status === 'auto-skipped',
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
      status === 'available' ||
      status === 'pending-approval' ||
      status === 'drifted',
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
  status = 'waiting',
  underline,
  time,
  href,
  isMostRecent = false,
}) => {
  const Event = (
    <div
      className={classNames('flex items-start justify-between p-4', {
        'border rounded-md shadow-sm': isMostRecent,
      })}
    >
      <div className="flex flex-col w-full">
        <div className="flex items-center justify-between gap-2 w-full">
          <span className="flex items-center gap-3">
            <EventStatus status={status} />
            <Text variant="med-12">{sentanceCase(status)}</Text>
          </span>
          <span className="flex items-center justify-end gap-0.5 min-w-[100px]">
            <Time
              alignment="right"
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
            {href && <CaretRightIcon />}
          </span>
        </div>

        <Text className="flex items-center gap-2 ml-7" variant="reg-12">
          {underline}
        </Text>
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
  emptyContent: React.ReactNode
}

export const Timeline: FC<ITimeline> = ({ events, emptyContent }) => {
  return (
    <div className="flex flex-col gap-2 timeline">
      {events?.length
        ? events.map((event, i) => (
            <TimelineEvent key={event.id} {...event} isMostRecent={i === 0} />
          ))
        : emptyContent}
    </div>
  )
}
