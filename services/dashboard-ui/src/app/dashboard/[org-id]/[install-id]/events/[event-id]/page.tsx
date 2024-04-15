// @ts-nocheck
// TODO: REMOVE THIS ROUTE


import { DateTime } from 'luxon'
import { GoArrowLeft, GoZap } from 'react-icons/go'
import {
  Card,
  Code,
  EventStatus,
  Heading,
  InstallComponent,
  InstallTimeline,
  Link,
  Page,
  Status,
  Sandbox,
  Text,
} from '@/components'
import type { TInstallEvent } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

async function getLogs(orgId, installId, deployId) {
  const res = await fetch(
    `${API_URL}/v1/installs/${installId}/deploys/${deployId}/logs`,
    await getFetchOpts(orgId)
  )

  console.log('logs res?', res)

  if (!res.ok) {
    // This will activate the closest `error.js` Error Boundary
    throw new Error('Failed to fetch data')
  }

  return res.json()
}

async function getPlan(orgId, installId, deployId) {
  const res = await fetch(
    `${API_URL}/v1/installs/${installId}/deploys/${deployId}/plan`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    // This will activate the closest `error.js` Error Boundary
    throw new Error('Failed to fetch data')
  }

  return res.json()
}

const DeployEventDashboard: FC<
  TInstallEvent & { installId: string; orgId: string }
> = async ({ installId, orgId, ...event }) => {
  const payload = JSON.parse(atob(event?.payload))

  const logsReq = getLogs(orgId, installId, payload?.id)
  const planReq = getPlan(orgId, installId, payload?.id)

  const [logs, plan] = await Promise.all([logsReq, planReq])

  return (
    <Page
      heading={
        <div className="flex flex-wrap items-end">
          <div className="flex flex-col flex-auto gap-2">
            <Text variant="overline">
              {DateTime.fromISO(event?.created_at).toRelative()}
            </Text>
            <Heading
              level={1}
              variant="title"
              className="flex gap-1 items-center"
            >
              <span className="text-xl">
                <EventStatus status={event?.operation_status} />
              </span>
              {event?.operation_name} {event?.operation_status}
            </Heading>
          </div>

          <div className="flex flex-col flex-auto gap-1">
            <Text variant="caption">
              <b>Event ID:</b> {event?.id}
            </Text>
            <Text variant="caption">
              <b>Deploy ID:</b> {payload?.id}
            </Text>
            <Text variant="caption">
              <b>Build ID:</b> {payload?.build_id}
            </Text>
            <Text variant="caption">
              <b>Component ID:</b> {payload?.component_id}
            </Text>
            <Text variant="caption">
              <b>Component name:</b> {payload?.component_name}
            </Text>
            <Text variant="caption">
              <b>Created by:</b> {payload?.created_by?.email}
            </Text>
          </div>
        </div>
      }
      links={[
        { href: event?.org_id, text: event?.org_id },
        { href: event?.install_id, text: event?.install_id },
        { href: 'events/' + event?.id, text: event?.id },
      ]}
    >
      <div className="flex flex-col gap-6">
        <div className="flex flex-col gap-3">
          <Heading>Deploy logs</Heading>
          <Code>
            {logs?.length
              ? logs.map((term) => {
                  // handle complete state

                  return term?.Terminal?.events?.length
                    ? term?.Terminal?.events?.map((l, i) => {
                        let line = null

                        if (l?.line) {
                          line = (
                            <span
                              key={`${l?.line?.msg}-${i}`}
                              className="block text-xs"
                            >
                              {l?.line?.msg}
                            </span>
                          )
                        }

                        // raw data

                        if (l?.raw?.data) {
                          line = (
                            <span
                              key={`${l?.raw?.data}-${i}`}
                              className="block text-xs"
                            >
                              {atob(l?.raw?.data)}
                            </span>
                          )
                        }

                        if (l?.step) {
                          line = (
                            <span
                              key={`${l?.step?.msg}-${i}`}
                              className="block text-xs"
                            >
                              {l?.step?.msg}
                            </span>
                          )
                        }

                        // status
                        if (l?.status) {
                          line = (
                            <span
                              key={`${l?.status?.msg}-${i}`}
                              className="block text-xs"
                            >
                              {l?.status?.msg}
                            </span>
                          )
                        }

                        return line
                      })
                    : null
                })
              : 'no logs to show'}
          </Code>
        </div>

        <div className="flex flex-col gap-3">
          <Heading>Deploy plan</Heading>

          <Heading variant="subheading">Rendered variables</Heading>
          <Code>
            {plan.actual?.waypoint_plan?.variables?.variables?.map((v, i) => {
              let variable = null
              if (v?.Actual?.TerraformVariable) {
                variable = (
                  <span className="flex" key={i?.toString()}>
                    <b>{v?.Actual?.TerraformVariable?.name}:</b>{' '}
                    {v?.Actual?.TerraformVariable?.value}
                  </span>
                )
              }

              if (v?.Actual?.HelmValue) {
                variable = (
                  <span className="flex" key={i?.toString()}>
                    <b>{v?.Actual?.HelmValue?.name}:</b>{' '}
                    {v?.Actual?.HelmValue?.value}
                  </span>
                )
              }

              return variable
            })}
          </Code>

          <Heading variant="subheading">Intermediate variables</Heading>
          <Code variant="preformated">
            {JSON.stringify(
              plan.actual?.waypoint_plan?.variables?.intermediate_data,
              null,
              2
            )}
          </Code>

          <Heading variant="subheading">Job config</Heading>
          <Code variant="preformated">
            {plan.actual?.waypoint_plan?.waypoint_job?.hcl_config}
          </Code>
        </div>
      </div>
    </Page>
  )
}

async function getEvent(orgId, installId, eventId): TInstallEvent {
  const res = await fetch(
    `${API_URL}/v1/installs/${installId}/events/${eventId}`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    // This will activate the closest `error.js` Error Boundary
    throw new Error('Failed to fetch data')
  }

  return res.json()
}

export default async function EventDashboard({ params }) {
  const event = await getEvent(
    params?.['org-id'],
    params?.['install-id'],
    params?.['event-id']
  )

  console.log('------- operation -------', event?.operation)
  return (
    (event?.operation === 'deploy' && (
      <DeployEventDashboard
        {...{
          orgId: params?.['org-id'],
          installId: params?.['install-id'],
          ...event,
        }}
      />
    )) || <>some other event</>
  )
}
