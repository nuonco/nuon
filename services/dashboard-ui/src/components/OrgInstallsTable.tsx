'use client'

import React, { type FC, useMemo, useState } from 'react'
import { type ColumnDef } from '@tanstack/react-table'
import { CaretRight } from '@phosphor-icons/react'
import { InstallPlatform } from '@/components/InstallCloudPlatform'
import { InstallsTableStatusFilter } from '@/components/InstallsTableStatusFilter'
import { Link } from '@/components/Link'
import { StatusBadge } from '@/components/Status'
import { DataTableSearch, Table } from '@/components/DataTable'
import { ID, Text } from '@/components/Typography'
// eslint-disable-next-line import/no-cycle
import type { TInstall } from '@/types'

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
    app: install.app.name,
    appId: install.app.id,
    platform: install?.app_sandbox_config?.cloud_platform,
  }))
}

export interface IOrgInstallsTable {
  installs: Array<TInstall>
  orgId: string
}

export const OrgInstallsTable: FC<IOrgInstallsTable> = ({
  installs,
  orgId,
}) => {
  const [data, _] = useState(parseInstallsToTableData(installs))
  const [columnFilters, setColumnFilters] = useState([])
  const [globalFilter, setGlobalFilter] = useState('')

  const columns: Array<ColumnDef<TData>> = useMemo(
    () => [
      {
        header: 'Name',
        accessorKey: 'name',
        cell: (props) => (
          <div className="flex flex-col gap-2">
            <Link
              href={`/${orgId}/installs/${props.row.original.installId}`}
              variant="default"
            >
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
              status={props.getValue<TDataStatuses>().sandbox_status}
              description={
                props?.getValue<TDataStatuses>()?.sandbox_status_description
              }
              descriptionAlignment="right"
              label="Sandbox"
              isLabelStatusText
            />
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
            <CaretRight />
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
          <DataTableSearch
            handleOnChange={handleGlobleFilter}
            value={globalFilter}
          />

          <InstallsTableStatusFilter
            handleStatusFilter={handleStatusFilter}
            clearStatusFilter={clearStatusFilter}
          />
        </>
      }
      data={data}
      columns={columns}
      columnFilters={columnFilters}
      emptyMessage="Reset your search or clear your filter and try again."
      emptyTitle="No installs found"
      globalFilter={globalFilter}
    />
  )
}
