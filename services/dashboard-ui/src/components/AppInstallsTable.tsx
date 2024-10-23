'use client'

import React, { type FC, useMemo, useState } from 'react'
import { type ColumnDef } from '@tanstack/react-table'
import { DotsThreeVertical } from '@phosphor-icons/react'
import { ClickToCopy } from '@/components/ClickToCopy'
import { Dropdown } from '@/components/Dropdown'
import { InstallPlatform } from '@/components/InstallCloudPlatform'
import { Link } from '@/components/Link'
import { RadioInput } from '@/components/Input'
import { StatusBadge } from '@/components/Status'
import { DataTableSearch, Table } from '@/components/DataTable'
import { Heading, Text } from '@/components/Typography'
import type { TInstall } from '@/types'

type TDataStatues = {
  composite_component_status: string
  runner_status: string
  sandbox_status: string
}

type TData = {
  name: string
  installId: string
  statues: TDataStatues
  app: string
  appId: string
  platform: string
}

function parseInstallsToTableData(installs: Array<TInstall>): Array<TData> {
  return installs.map((install) => ({
    name: install.name,
    installId: install.id,
    statues: {
      composite_component_status: install.composite_component_status,
      runner_status: install.runner_status,
      sandbox_status: install.sandbox_status,
    },
    app: install?.app?.name,
    appId: install?.app?.id,
    platform: install?.app_sandbox_config?.cloud_platform,
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
            <Link href={`/${orgId}/installs/${props.row.original.installId}`}>
              <Heading variant="subheading">{props.getValue<string>()}</Heading>
            </Link>
            <ClickToCopy>
              <Text variant="id">{props.row.original.installId}</Text>
            </ClickToCopy>
          </div>
        ),
      },
      {
        header: 'Statues',
        accessorKey: 'statues',
        enableSorting: false,
        enableColumnFilter: true,
        cell: (props) => (
          <div className="flex flex-col gap-2">
            <StatusBadge
              status={props.getValue<TDataStatues>().sandbox_status}
              label="Sandbox"
              isLabelStatusText
            />
            <StatusBadge
              status={props.getValue<TDataStatues>().runner_status}
              label="Runner"
              isLabelStatusText
            />
            <StatusBadge
              status={props.getValue<TDataStatues>().composite_component_status}
              label="Components"
              isLabelStatusText
            />
          </div>
        ),
        filterFn: (row, columnId, filterValue) => {
          const statues = row.getValue<TDataStatues>(columnId)
          return (
            statues.sandbox_status.includes(filterValue) ||
            statues.runner_status.includes(filterValue) ||
            statues.composite_component_status.includes(filterValue)
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
            <DotsThreeVertical />
          </Link>
        ),
      },
    ],
    []
  )

  const handleStatusFilter = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value
    setColumnFilters(() => [{ id: 'statues', value: value }])
  }

  const handleGlobleFilter = (e: React.ChangeEvent<HTMLInputElement>) => {
    setGlobalFilter(e.target.value)
  }

  return (
    <Table
      header={
        <>
          <DataTableSearch
            handleOnChange={handleGlobleFilter}
            value={globalFilter}
          />

          <Dropdown
            className="text-sm"
            id="install-filter"
            text="Filter"
            alignment="right"
          >
            <div>
              <RadioInput
                name="status-filter"
                onChange={handleStatusFilter}
                value="error"
                labelText="Error"
              />

              <RadioInput
                name="status-filter"
                onChange={handleStatusFilter}
                value="processing"
                labelText="Processing"
              />

              <RadioInput
                name="status-filter"
                onChange={handleStatusFilter}
                value="noop"
                labelText="NOOP"
              />

              <RadioInput
                name="status-filter"
                onChange={handleStatusFilter}
                value="active"
                labelText="Active"
              />
            </div>
          </Dropdown>
        </>
      }
      data={data}
      columns={columns}
      columnFilters={columnFilters}
      globalFilter={globalFilter}
    />
  )
}
