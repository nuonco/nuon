'use client'

import { DateTime } from 'luxon'
import React, { type FC, useEffect, useState } from 'react'
import {
  GoArrowLeft,
  GoKebabHorizontal,
  GoCheckCircleFill,
  GoClockFill,
  GoXCircleFill,
  GoInfo,
} from 'react-icons/go'
import { Heading, Link, Text } from '@/components'
import type { TInstallEvent } from '@/types'

export const EventStatus = ({ status }) =>
  status === 'finished' ? (
    <GoCheckCircleFill className="text-green-700 dark:text-green-500" />
  ) : status === 'failed' ? (
    <GoXCircleFill className="text-red-600 dark:text-red-500" />
  ) : (
    <GoClockFill className="text-yellow-600 dark:text-yellow-500" />
  )

export interface IEvent {
  event: TInstallEvent
  feedId: string
  orgId: string
}

export const Event: FC<IEvent> = ({ event, feedId, orgId }) => {
  const payload = JSON.parse(atob(event?.payload))
  const eventUrl =
    (event?.operation === 'deploy' &&
      `/dashboard/${orgId}/${feedId}/components/${payload?.install_component_id}/deploys/${payload?.id}`) ||
    ((event?.operation === 'provision' || event?.operation === 'reprovision') &&
      `/dashboard/${orgId}/${feedId}/runs/${payload?.id}`) ||
    `/dashboard/${orgId}/${feedId}/events/${event?.id}`
  return (
    <div className="flex flex-wrap items-center gap-6 py-4">
      <EventStatus status={event?.operation_status} />
      <div className="flex flex-col flex-auto">
        <span className="text-xs text-gray-600 dark:text-gray-300">
          {DateTime.fromISO(event?.created_at).toRelative()}
        </span>
        <span className="font-semibold text-sm">{`${event?.operation_name} ${event?.operation_status}`}</span>
        {event?.operation === 'deploy' ? (
          <span className="text-xs">{payload?.component_name}</span>
        ) : null}
        {event?.operation_status === 'finished' ||
        event?.operation_status === 'failed' ? (
          <Link className="text-xs" href={eventUrl}>
            Details
          </Link>
        ) : null}
      </div>
    </div>
  )
}

export interface IEventsTimeline {
  feedId: string
  orgId: string
  initEvents?: Array<TInstallEvent>
}

export const EventsTimeline: FC<IEventsTimeline> = ({
  feedId,
  orgId,
  initEvents = [],
}) => {
  const [events, setEvents] = useState(initEvents)

  const fetchEvents = () => {
    fetch(`/api/${orgId}/${feedId}/events`)
      .then((res) => res.json().then((e) => setEvents(e)))
      .catch(console.error)
  }

  useEffect(() => {
    fetchEvents()
  }, [])

  let pollEvents
  useEffect(() => {
    pollEvents = setInterval(fetchEvents, 3000)
    return () => clearInterval(pollEvents)
  }, [events])

  return (
    <div className="flex flex-col divide-y">
      {events.length ? (
        events.map(
          (e) =>
            e?.operation !== 'deploy_components' && (
              <Event key={e.id} event={e} feedId={feedId} orgId={orgId} />
            )
        )
      ) : (
        <>No events</>
      )}
    </div>
  )
}
