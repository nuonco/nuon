'use client'

import React, { type FC, useEffect, useState } from 'react'
import { Empty } from '@/components/old/Empty'
import { Timeline } from '@/components/old/Timeline'
import { ToolTip } from '@/components/old/ToolTip'
import type { TInstallEvent } from '@/types'
import { SHORT_POLL_DURATION } from '@/utils'

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
        component_id: payload?.component_id as string,
        component_name: payload?.component_name as string,
        created_at: event.created_at,
        event_id: event.id,
        install_id: event.install_id,
        operation: event.operation,
        operation_name:
          payload?.install_deploy_type === 'teardown'
            ? 'Teardown'
            : event.operation_name,
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
  // const events = parseInstallHistory(initEvents)
  const [events, setInstallEvents] = useState(parseInstallHistory(initEvents))

  useEffect(() => {
    const fetchInstallEvents = () => {
      fetch(`/api/${orgId}/installs/${installId}/events`)
        .then((res) =>
          res.json().then((e) => setInstallEvents(parseInstallHistory(e)))
        )
        .catch(console.error)
      // revalidateInstallData({ installId, orgId })
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
    <Timeline
      emptyContent={
        <Empty
          emptyTitle="No install history yet"
          emptyMessage="Waiting for install provision and deployments start."
          variant="history"
        />
      }
      events={events.map((e) => ({
        id: e.payload_id,
        status: e.operation_status,
        underline: (
          <>
            <span>{e.operation_name}</span>
            {e.operation === 'deploy' && (
              <>
                /{' '}
                <ToolTip tipContent={e.component_name}>
                  <span className="!inline truncate max-w-[100px]">
                    {e.component_name}
                  </span>
                </ToolTip>
              </>
            )}
          </>
        ),
        time: e.updated_at,
        href:
          (e?.operation === 'deploy' &&
            `/${e.org_id}/installs/${e.install_id}/components/${e?.component_id}/deploys/${e.payload_id}`) ||
          ((e?.operation === 'provision' ||
            e?.operation === 'reprovision' ||
            e?.operation === 'deprovision') &&
            `/${e.org_id}/installs/${e.install_id}/sandbox/${e.payload_id}`) ||
          null,
      }))}
    />
  )
}
