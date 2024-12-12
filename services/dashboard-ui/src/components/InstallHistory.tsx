'use client'

import classNames from 'classnames'
import React, { type FC, useEffect, useState } from 'react'
import { CaretRight } from '@phosphor-icons/react'
import { Link } from '@/components/Link'
import { Time } from '@/components/Time'
import { ToolTip } from '@/components/ToolTip'
import { Text } from '@/components/Typography'
import type { TInstallEvent } from '@/types'
import { SHORT_POLL_DURATION, sentanceCase } from '@/utils'

function parseInstallEventPayload(
  event: TInstallEvent
): Record<string, unknown> {
  return JSON.parse(atob(event?.payload))
}

type TInstallHistoryEvent = {
  component_id?: string
  component_name?: string
  created_at: string
  event_id: string
  install_id: string
  operation: string
  operation_status: string
  operation_name: string
  org_id: string
  payload_id: string
  status: string
  status_description: string
  updated_at?: string
}

function parseInstallHistory(
  events: Array<TInstallEvent>
): Array<TInstallHistoryEvent> {
  return events.reduce((acc, event) => {
    const payload = parseInstallEventPayload(event)

    if (
      event.operation === 'deploy' ||
      event.operation === 'provision' ||
      event.operation === 'reprovision' ||
      event.operation === 'deprovision'
    ) {
      const historyEvent: TInstallHistoryEvent = {
        component_id: payload?.install_component_id as string,
        component_name: payload?.component_name as string,
        created_at: event.created_at,
        event_id: event.id,
        install_id: event.install_id,
        operation: event.operation,
        operation_name: event.operation_name,
        operation_status: event.operation_status,
        org_id: event.org_id,
        payload_id: payload.id as string,
        status: payload.status === '' ? 'waiting' : (payload.status as string),
        status_description: payload?.status_description as string,
        updated_at: event.updated_at,
      }

      acc.push(historyEvent)
      acc = acc.filter(
        (evt, i, self) =>
          self.findIndex((e) => e.payload_id === evt.payload_id) === i
      )
    }

    return acc
  }, [])
}

export interface IInstallHistory {
  initEvents: Array<TInstallEvent>
  installId: string
  orgId: string
  shouldPoll?: boolean
}

export const InstallHistory: FC<IInstallHistory> = ({
  initEvents,
  installId,
  orgId,
  shouldPoll = false,
}) => {
  const [events, setInstallEvents] = useState(parseInstallHistory(initEvents))

  useEffect(() => {
    const fetchInstallEvents = () => {
      fetch(`/api/${orgId}/installs/${installId}/events`)
        .then((res) =>
          res.json().then((e) => setInstallEvents(parseInstallHistory(e)))
        )
        .catch(console.error)
    }

    if (shouldPoll) {
      const pollInstallEvents = setInterval(
        fetchInstallEvents,
        SHORT_POLL_DURATION
      )
      return () => clearInterval(pollInstallEvents)
    }
  }, [events, installId, orgId, shouldPoll])

  return (
    <div className="flex flex-col gap-2">
      {events.length ? (
        events.map((event, i) => (
          <InstallEvent
            key={`${event.payload_id}-${i}`}
            event={event}
            isMostRecent={i === 0}
          />
        ))
      ) : (
        <Text className="text-sm">No install events found</Text>
      )}
    </div>
  )
}

interface IInstallEvent {
  event: TInstallHistoryEvent
  isMostRecent?: boolean
}

const InstallEvent: FC<IInstallEvent> = ({ event, isMostRecent = false }) => {
  const href =
    (event?.operation === 'deploy' &&
      `/${event.org_id}/installs/${event.install_id}/components/${event?.component_id}/deploys/${event.payload_id}`) ||
    ((event?.operation === 'provision' ||
      event?.operation === 'reprovision' ||
      event?.operation === 'deprovision') &&
      `/${event.org_id}/installs/${event.install_id}/runs/${event.payload_id}`) ||
    null

  return (
    <Link className="!block w-full !p-0" href={href} variant="ghost">
      <div
        className={classNames('flex items-center justify-between p-4', {
          'border rounded-md shadow-sm': isMostRecent,
        })}
      >
        <div className="flex flex-col">
          <span className="flex items-center gap-2">
            <InstallEventStatus status={event.operation_status} />
            <Text variant="med-12">{sentanceCase(event.operation_status)}</Text>
          </span>

          <Text className="flex items-center gap-2 ml-3.5" variant="reg-12">
            <span>{event.operation_name}</span>
            {event.operation === 'deploy' && (
              <>
                /{' '}
                <ToolTip tipContent={event.component_name}>
                  <span className="!inline truncate max-w-[100px]">
                    {event.component_name}
                  </span>
                </ToolTip>
              </>
            )}
          </Text>
        </div>

        <div className="flex items-center gap-2">
          <Time
            time={event.updated_at}
            format="relative"
            variant="reg-12"
            className={classNames({
              'text-black/60 dark:text-white/60': !Boolean(
                event.operation_status === 'finished' ||
                  event.operation_status === 'failed'
              ),
            })}
          />
          {href && <CaretRight />}
        </div>
      </div>
    </Link>
  )
}

const InstallEventStatus: FC<{ status?: string }> = ({
  status = 'waiting',
}) => (
  <span
    className={classNames('w-1.5 h-1.5 rounded-full', {
      'bg-green-800 dark:bg-green-500': status === 'finished',
      'bg-red-800 dark:bg-red-500': status === 'failed',
      'bg-cool-grey-600 dark:bg-cool-grey-500': status === 'noop',
      'bg-orange-800 dark:bg-orange-500':
        status === 'waiting' || status === 'started',
    })}
  />
)
