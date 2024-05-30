'use client'

import { DateTime } from 'luxon'
import React, { type FC } from 'react'
import { Card, Heading, Link, Status, Text, type ICard } from '@/components'
import {
  useInstallContext,
  useInstallComponentContext,
  InstallComponentProvider,
} from '@/context'
import type { TInstallComponent, TInstallDeploy } from '@/types'

export const InstallComponentsListCard: FC<ICard> = (props) => {
  return (
    <Card {...props}>
      <InstallComponents />
    </Card>
  )
}

export const InstallComponents: FC = () => {
  const { install } = useInstallContext()

  return (
    <div className="flex flex-col divide-y">
      {install?.install_components?.length ? (
        install.install_components.map((c) => (
          <InstallComponentProvider key={c.id} initInstallComponent={c as TInstallComponent}>
            <InstallComponent />
          </InstallComponentProvider>
        ))
      ) : (
        <Text variant="caption">No install components to show</Text>
      )}
    </div>
  )
}

export const InstallComponent: FC = () => {
  const {
    installComponent: {
      component,
      id,
      install_deploys,
      install_id,
      org_id,
      updated_at,
    },
  } = useInstallComponentContext()
  const activeDeploy = install_deploys?.[0]

  return (
    <div className="flex flex-col flex-auto gap-2 py-4">
      <InstallComponentStatus showDescription />

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
  isCompact?: boolean
  isStatusTextHidden?: boolean
  showDescription?: boolean
}

export const InstallComponentStatus: FC<IInstallComponentStatus> = ({
  isCompact = false,
  isStatusTextHidden = false,
  showDescription = false,
}) => {
  const {
    installComponent: { component, install_deploys },
  } = useInstallComponentContext()

  const status = install_deploys?.[0] || {
    status: 'waiting',
    status_description: 'Checking install component status',
  }

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

export const InstallDeploys: FC = () => {
  const { install } = useInstallContext()
  const {
    installComponent: { install_deploys },
  } = useInstallComponentContext()

  return (
    <div className="flex flex-col divide-y">
      {install_deploys?.map((d) => (
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
            href={`/dashboard/${d['org_id']}/${install.id}/components/${d.install_component_id}/deploys/${d.id}`}
          >
            Details
          </Link>
        </div>
      ))}
    </div>
  )
}
