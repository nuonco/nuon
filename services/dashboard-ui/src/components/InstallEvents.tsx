'use client'

import { DateTime } from 'luxon'
import React, { type FC, useEffect, useState } from 'react'
import { GoCheckCircleFill, GoClockFill, GoXCircleFill } from 'react-icons/go'
import { Card, Heading, Link, Text } from '@/components'
import { useInstallContext } from '@/context'
import type { TInstallEvent } from '@/types'
import { SHORT_POLL_DURATION } from '@/utils'

export const InstallEventStatus: FC<{ status?: string }> = ({
  status = 'waiting',
}) =>
  status === 'finished' ? (
    <GoCheckCircleFill className="text-green-700 dark:text-green-500" />
  ) : status === 'failed' ? (
    <GoXCircleFill className="text-red-600 dark:text-red-500" />
  ) : (
    <GoClockFill className="text-yellow-600 dark:text-yellow-500" />
  )

export interface IInstallEvent {
  event: TInstallEvent
}

export const InstallEvent: FC<IInstallEvent> = ({ event }) => {
  const { install } = useInstallContext()

  const payload = JSON.parse(atob(event?.payload))
  const eventUrl =
    (event?.operation === 'deploy' &&
      `/dashboard/${install.org_id}/${install.id}/components/${payload?.install_component_id}/deploys/${payload?.id}`) ||
    ((event?.operation === 'provision' ||
      event?.operation === 'reprovision' ||
      event?.operation === 'deprovision') &&
      `/dashboard/${install.org_id}/${install.id}/runs/${payload?.id}`) ||
    null
  return (
    <div className="flex flex-wrap items-center gap-6 py-4">
      <InstallEventStatus status={event?.operation_status} />
      <div className="flex flex-col flex-auto gap-2">
        <Text variant="overline">
          {DateTime.fromISO(event.created_at as string).toRelative()}
        </Text>
        <Heading variant="subheading">{`${event?.operation_name} ${event?.operation_status}`}</Heading>
        {event?.operation === 'deploy' ? (
          <Text variant="status">
            <small>{payload?.component_name}</small>
          </Text>
        ) : null}

        {eventUrl && (
          <Text variant="caption">
            <Link href={eventUrl}>Details</Link>
          </Text>
        )}
      </div>
    </div>
  )
}

export interface IInstallEvents {
  initEvents?: Array<TInstallEvent>
  shouldPoll?: boolean
}

export const InstallEvents: FC<IInstallEvents> = ({
  initEvents = [],
  shouldPoll = false,
}) => {
  const { install } = useInstallContext()
  const [events, setInstallEvents] = useState(initEvents)

  useEffect(() => {
    const fetchInstallEvents = () => {
      fetch(`/api/${install.org_id}/installs/${install.id}/events`)
        .then((res) => res.json().then((e) => setInstallEvents(e)))
        .catch(console.error)
    }

    if (shouldPoll) {
      const pollInstallEvents = setInterval(
        fetchInstallEvents,
        SHORT_POLL_DURATION
      )
      return () => clearInterval(pollInstallEvents)
    }
  }, [events, install, shouldPoll])

  return (
    <div className="flex flex-col divide-y">
      {events.length ? (
        events.map(
          (e) =>
            e?.operation !== 'deploy_components' && (
              <InstallEvent key={e.id} event={e} />
            )
        )
      ) : (
        <>No install events</>
      )}
    </div>
  )
}

export const InstallEventsCard: FC<IInstallEvents> = (props) => {
  return (
    <Card className="max-h-[40rem]">
      <InstallEvents {...props} />
    </Card>
  )
}
