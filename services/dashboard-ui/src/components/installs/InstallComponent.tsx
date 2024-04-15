'use client'

import { DateTime } from 'luxon'
import React, { type FC, useEffect, useState } from 'react'
import { FaAws, FaDocker, FaGitAlt, FaGithub } from 'react-icons/fa'
import {
  GoGitCommit,
  GoArrowLeft,
  GoKebabHorizontal,
  GoCheckCircleFill,
  GoClockFill,
  GoContainer,
  GoXCircleFill,
  GoInfo,
} from 'react-icons/go'
import {
  SiAwslambda,
  SiOpencontainersinitiative,
  SiHelm,
  SiTerraform,
} from 'react-icons/si'
import { VscAzure } from 'react-icons/vsc'
import {
  Card,
  Code,
  ComponentType,
  Heading,
  Link,
  Status,
  Text,
  type THeadingVariant,
} from '@/components'
import type {
  TBuild,
  TComponent,
  TComponentConfig,
  TInstall,
  TInstallAwsAccount,
  TInstallAzureAccount,
  TInstallComponent,
  TInstallDeploy,
  TSandboxConfig,
} from '@/types'
import { getFullInstallStatus } from '@/utils'

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
  config: TComponentConfig
  install: TInstall
  installComponent: TInstallComponent
  build: TBuild
}> = ({ build, component, config, install, installComponent }) => {
  return (
    <div className="flex flex-wrap gap-8 items-end border-b pb-6">
      <div className="flex flex-col flex-auto gap-2">
        <Text variant="overline">{component?.id}</Text>
        <Heading level={1} variant="title">
          {component?.name}
        </Heading>
        <Text className="gap-4" variant="caption">
          <Text variant="status">{install?.app?.name}</Text>
          <ComponentType config={config} />
        </Text>
      </div>

      <div className="flex flex-col flex-auto gap-6">
        <div>
          <InstallComponentStatus component={installComponent} />
          <LatestDeploy
            {...{
              ...(installComponent?.install_deploys?.[0] as TInstallDeploy),
              install_id: install?.id,
            }}
          />
        </div>

        <div className="flex items-center gap-4">
          <GoGitCommit className="text-xl" />{' '}
          <span className="flex flex-col">
            <Text className="truncate" variant="caption">{build?.vcs_connection_commit?.message}</Text>
            <Text variant="overline">
              {build?.vcs_connection_commit?.sha?.slice(0, 7)}           
            </Text>
          </span>
        </div>
      </div>
    </div>
  )
}

export const LatestDeploy: FC<TInstallDeploy> = ({
  build_id,
  id,
  install_id,
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
