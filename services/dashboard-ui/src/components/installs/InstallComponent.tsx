'use client'

import { DateTime } from 'luxon'
import React, { type FC, useEffect, useState } from 'react'
import { ComponentType, Heading, Link, Status, Text } from '@/components'
import type {
  TComponent,
  TComponentConfig,
  TInstallComponent,
  TInstallDeploy,
} from '@/types'

export const InstallComponents: FC<{ components: TInstallComponent[] }> = ({
  components,
}) => {
  return (
    <div className="flex flex-col divide-y">
      {components?.length ? (
        components?.map((c) => (
          <InstallComponent key={c?.id} install_component={c} />
        ))
      ) : (
        <Text variant="caption">No install components to show</Text>
      )}
    </div>
  )
}

export const InstallComponent: FC<{ install_component: TInstallComponent }> = ({
  install_component: {
    component,
    component_id,
    id,
    install_id,
    install_deploys,
    org_id,
    updated_at,
  },
}) => {
  const activeDeploy = install_deploys?.[0]

  return (
    <div className="flex flex-col flex-auto gap-2 py-4">
      <InstallComponentStatus
        component={{
          component,
          component_id,
          id,
          org_id,
          install_id,
          install_deploys,
        }}
        showDescription
      />

      <span>
        <Text variant="overline">{component?.id}</Text>
        <Heading variant="subheading">{component?.name}</Heading>

        {activeDeploy ? (
          <Text className="flex flex-wrap gap-4 items-center" variant="caption">
            Last deployed{' '}
            {DateTime.fromISO(activeDeploy?.created_at as string).toRelative()}
          </Text>
        ) : (
          <Text className="flex flex-wrap gap-4 items-center" variant="caption">
            As of {DateTime.fromISO(updated_at as string).toRelative()}
          </Text>
        )}
      </span>

      <Link
        className="text-xs"
        href={`/dashboard/${org_id}/${install_id}/components/${id}`}
      >
        Details
      </Link>
    </div>
  )
}

export interface IInstallComponentStatus {
  component: TInstallComponent
  isCompact?: boolean
  isStatusTextHidden?: boolean
  showDescription?: boolean
}

export const InstallComponentStatus: FC<IInstallComponentStatus> = ({
  component: { component_id, component, install_id, install_deploys, org_id },
  isCompact = false,
  isStatusTextHidden = false,
  showDescription = false,
}) => {
  const [status, setStatus] = useState(
    install_deploys?.[0] || {
      status: 'waiting',
      status_description: 'Checking install component status',
    }
  )

  const fetchStatus = () => {
    fetch(`/api/${org_id}/${install_id}/components/${component_id}/status`)
      .then((res) => res.json().then((s) => setStatus(s)))
      .catch(console.error)
  }

  useEffect(() => {
    fetchStatus()
  }, [])

  let pollStatus: NodeJS.Timeout
  useEffect(() => {
    pollStatus = setInterval(fetchStatus, 5000)
    return () => clearInterval(pollStatus)
  }, [status])

  return (
    <Status
      status={status?.status}
      description={showDescription && status?.status_description}
      label={isCompact && component?.name}
      isLabelStatusText={isCompact}
      isStatusTextHidden={isStatusTextHidden}
    />
  )
}

export const InstallComponentHeading: FC<{
  component: TComponent
}> = ({ component }) => {
  return (
    <span className="flex flex-col gap-0">
      <Text variant="overline">{component?.id}</Text>
      <Heading level={1} variant="title">
        {component?.name}
      </Heading>
    </span>
  )
}

export const InstallComponentSummary: FC<{
  appName: React.ReactNode
  config: TComponentConfig
}> = ({ appName, config }) => {
  return (
    <Text className="gap-4" variant="caption">
      <Text variant="status">{appName}</Text>
      <ComponentType config={config} />
    </Text>
  )
}

export const LatestDeploy2: FC<TInstallComponent> = (installComponent) => {
  return (
    <div>
      <InstallComponentStatus component={installComponent} />
      <LatestDeploy
        {...{
          ...(installComponent?.install_deploys?.[0] as TInstallDeploy),
        }}
      />
    </div>
  )
}

export const LatestDeploy: FC<TInstallDeploy> = ({
  build_id,
  id,
  install_deploy_type,
  release_id,
}) => {
  return (
    <span className="flex flex-col gap-2">
      <Text variant="overline">
        <b>Deploy ID:</b> {id}
      </Text>

      <Text variant="overline">
        <b>Build ID:</b> {build_id}
      </Text>

      {install_deploy_type === 'release' && (
        <Text variant="overline">
          <b>Release ID:</b> {release_id}
        </Text>
      )}
    </span>
  )
}

export const InstallDeploys: FC<{
  deploys: Array<TInstallDeploy>
  installId: string
}> = ({ deploys, installId }) => {
  // TODO: realtime feed of deploys

  return (
    <div className="flex flex-col divide-y">
      {deploys?.map((d) => (
        <div key={d?.id} className="flex flex-col py-4">
          <span className="flex flex-wrap items-center gap-6">
            <Status status={d?.status} />
            <Text variant="overline">
              {DateTime.fromISO(d?.created_at as string).toRelative()}
            </Text>
          </span>
          <Text variant="label">{d?.id}</Text>
          <Text variant="caption">
            {DateTime.fromISO(d?.created_at as string).toLocaleString(
              DateTime.DATETIME_SHORT_WITH_SECONDS
            )}{' '}
          </Text>
          <Link
            className="text-xs"
            href={`/dashboard/${d?.org_id}/${installId}/components/${d?.install_component_id}/deploys/${d?.id}`}
          >
            Details
          </Link>
        </div>
      ))}
    </div>
  )
}
