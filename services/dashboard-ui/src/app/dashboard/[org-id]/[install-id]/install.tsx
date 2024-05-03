'use client'

import React, { type FC, useEffect, useState } from 'react'
import {
  Card,
  EventsTimeline,
  Heading,
  InstallComponents,
  InstallPageHeader,
  CloudDetails,
  Page,
  SandboxDetails,
} from '@/components'
import type { TInstall, TInstallEvent } from '@/types'
import { POLL_DURATION } from '@/utils'

interface IInstallPage {
  install: TInstall
  events: Array<TInstallEvent>
}

export const InstallPage: FC<IInstallPage> = ({ install, events = [] }) => {
  return (
    <Page
      header={<InstallPageHeader {...install} />}
      links={[{ href: install?.org_id }, { href: install?.id }]}
    >
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 w-full h-fit">
        <div className="flex flex-col gap-6">
          <Heading variant="subtitle">History</Heading>
          <Card className="max-h-[40rem]">
            <EventsTimeline
              feedId={install?.id}
              orgId={install?.org_id}
              initEvents={events}
            />
          </Card>
        </div>

        <div className="flex flex-col gap-6">
          <Heading variant="subtitle">Components</Heading>
          <Card className="max-h-[40rem]">
            <InstallComponents components={install?.install_components} />
          </Card>
        </div>

        <div className="flex flex-col gap-6">
          <Heading variant="subtitle">Details</Heading>

          <Card>
            <SandboxDetails {...install?.app_sandbox_config} />
          </Card>

          <Card>
            <CloudDetails {...install} />
          </Card>
        </div>
      </div>
    </Page>
  )
}

// Same as above just refetchs the install data every 45 seconds
export const ClientRefechInstallPage: FC<IInstallPage> = ({
  install: ssrInstall,
  events: ssrEvents,
}) => {
  const [install, setInstall] = useState(ssrInstall)
  const fetchInstall = () => {
    fetch(`/api/${install?.org_id}/${install?.id}`)
      .then((res) => res.json().then((ins) => setInstall(ins)))
      .catch(console.error)
  }

  useEffect(() => {
    fetchInstall()
  }, [])

  useEffect(() => {
    const pollInstall = setInterval(fetchInstall, POLL_DURATION)
    return () => clearInterval(pollInstall)
  }, [install])

  return <InstallPage install={install} events={ssrEvents} />
}
