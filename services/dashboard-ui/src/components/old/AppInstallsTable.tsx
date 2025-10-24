'use client'

import React, { type FC, useEffect, useMemo, useState } from 'react'
import { type ColumnDef } from '@tanstack/react-table'
import { CaretRightIcon } from '@phosphor-icons/react'
import { InstallPlatform } from '@/components/old/InstallCloudPlatform'
import { Link } from '@/components/old/Link'
import { StatusBadge } from '@/components/old/Status'
import { Table } from '@/components/old/DataTable'
import { DebouncedSearchInput } from '@/components/old/DebouncedSearchInput'
import { ID, Text } from '@/components/old/Typography'
import { AWS_REGIONS, AZURE_REGIONS } from '@/configs/cloud-regions'
import type { TInstall } from '@/types'
import { getFlagEmoji } from '@/utils'

type TDataStatuses = {
  composite_component_status: string
  composite_component_status_description: string
  runner_status: string
  runner_status_description: string
  sandbox_status: string
  sandbox_status_description: string
}

type TData = {
  name: string
  installId: string
  statuses: TDataStatuses
  app: string
  appId: string
  platform: string
  region: string | undefined
}

function parseInstallsToTableData(installs: Array<TInstall>): Array<TData> {
  return installs.map((install) => ({
    name: install.name,
    installId: install.id,
    statuses: {
      composite_component_status: install.composite_component_status,
      composite_component_status_description:
        install.composite_component_status_description,
      runner_status: install.runner_status,
      runner_status_description: install.runner_status_description,
      sandbox_status: install.sandbox_status,
      sandbox_status_description: install.sandbox_status_description,
    },
    app: install?.app?.name,
    appId: install?.app?.id,
    platform: install?.app_sandbox_config?.cloud_platform,
    region: install?.aws_account
      ? install?.aws_account?.region
      : install?.azure_account
        ? install.azure_account?.location
        : undefined,
  }))
}

export interface IAppInstallsTable {
  installs: Array<TInstall>
  orgId: string
}

export const AppInstallsTable: FC<IAppInstallsTable> = ({
  installs,
  orgId,
}) => {
  const [data, updateData] = useState(parseInstallsToTableData(installs))
  const [columnFilters, setColumnFilters] = useState([])
  const [globalFilter, setGlobalFilter] = useState('')

  useEffect(() => {
    updateData(parseInstallsToTableData(installs))
  }, [installs])

  const columns: Array<ColumnDef<TData>> = useMemo(
    () => [
      {
        header: 'Name',
        accessorKey: 'name',
        cell: (props) => (
          <div className="flex flex-col gap-2">
            <Link href={`/${orgId}/installs/${props.row.original.installId}`}>
              <Text variant="med-14">{props.getValue<string>()}</Text>
            </Link>
            <ID id={props.row.original.installId} />
          </div>
        ),
      },
      {
        header: 'Statuses',
        accessorKey: 'statuses',
        enableSorting: false,
        enableColumnFilter: true,
        cell: (props) => (
          <div className="flex flex-col gap-2">
            <StatusBadge
              status={props.getValue<TDataStatuses>().runner_status}
              description={
                props?.getValue<TDataStatuses>()?.runner_status_description
              }
              descriptionAlignment="right"
              label="Runner"
              isLabelStatusText
            />
            <StatusBadge
              status={props.getValue<TDataStatuses>().sandbox_status}
              description={
                props?.getValue<TDataStatuses>()?.sandbox_status_description
              }
              descriptionAlignment="right"
              label="Sandbox"
              isLabelStatusText
            />
            <StatusBadge
              status={
                props.getValue<TDataStatuses>().composite_component_status
              }
              description={
                props?.getValue<TDataStatuses>()
                  ?.composite_component_status_description
              }
              descriptionAlignment="right"
              label="Components"
              isLabelStatusText
            />
          </div>
        ),
        filterFn: (row, columnId, filterValue) => {
          const statuses = row.getValue<TDataStatuses>(columnId)
          return (
            statuses.sandbox_status.includes(filterValue) ||
            statuses.runner_status.includes(filterValue) ||
            statuses.composite_component_status.includes(filterValue)
          )
        },
      },
      {
        header: 'App',
        accessorKey: 'app',
        cell: (props) => (
          <Link href={`/${orgId}/apps/${props.row.original.appId}`}>
            <Text className="break-all">{props.getValue<string>()}</Text>
          </Link>
        ),
      },
      {
        header: 'Region',
        accessorKey: 'region',
        cell: (props) => {
          const region = props.row.original.region
            ? props.row.original.platform === 'azure'
              ? AZURE_REGIONS.find((r) => r.value === props.row.original.region)
              : AWS_REGIONS.find((r) => r.value === props.row.original.region)
            : null
          return (
            <Text className="break-all">
              {region ? (
                <>
                  {getFlagEmoji(region.iconVariant?.substring(5))}{' '}
                  {region?.text}
                </>
              ) : (
                'Unknown'
              )}
            </Text>
          )
        },
      },
      {
        header: 'Platform',
        accessorKey: 'platform',
        cell: (props) => (
          <Text className="gap-4">
            <InstallPlatform platform={props.getValue<'aws' | 'azure'>()} />
          </Text>
        ),
      },
      {
        id: 'test',
        enableSorting: false,
        cell: (props) => (
          <Link
            href={`/${orgId}/installs/${props.row.original.installId}`}
            variant="ghost"
          >
            <CaretRightIcon />
          </Link>
        ),
      },
    ],
    []
  )

  const handleStatusFilter = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value
    setColumnFilters(() => [{ id: 'statuses', value: value }])
  }

  const clearStatusFilter = () => {
    setColumnFilters(() => [])
  }

  const handleGlobleFilter = (e: React.ChangeEvent<HTMLInputElement>) => {
    setGlobalFilter(e.target.value || '')
  }

  return (
    <Table
      header={
        <>
          <DebouncedSearchInput placeholder="Search install name" />
        </>
      }
      data={data}
      columns={columns}
      columnFilters={columnFilters}
      emptyMessage="Reset your search or clear your filters and try again."
      emptyTitle="No installs found"
      globalFilter={globalFilter}
    />
  )
}
