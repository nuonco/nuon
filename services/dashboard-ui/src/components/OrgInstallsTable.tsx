'use client'

import React, { type FC, useMemo, useState } from 'react'
import { type ColumnDef } from '@tanstack/react-table'
import { DotsThreeVertical } from '@phosphor-icons/react'
import {
  DataTableSearch,
  Dropdown,
  Heading,
  InstallPlatform,
  Link,
  Status,
  Table,
  Text,
} from '@/components'
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
            <Heading variant="subheading">{props.getValue<string>()}</Heading>
            <Text variant="id">{props.row.original.installId}</Text>
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
            <Status
              status={props.getValue<TDataStatues>().sandbox_status}
              label="Sandbox"
              isLabelStatusText
            />
            <Status
              status={props.getValue<TDataStatues>().runner_status}
              label="Runner"
              isLabelStatusText
            />
            <Status
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
          <Text className="break-all">{props.getValue<string>()}</Text>
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
            className="text-xl"
            href={`/beta/${orgId}/installs/${props.row.original.installId}`}
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
              <label className="flex gap-4 items-center w-full px-4 py-2 cursor-pointer">
                <input
                  className="accent-primary-600"
                  name="status-filter"
                  onChange={handleStatusFilter}
                  value="error"
                  type="radio"
                />
                <span>error</span>
              </label>
              <label className="flex gap-4 items-center w-full px-4 py-2 cursor-pointer">
                <input
                  className="accent-primary-600"
                  name="status-filter"
                  onChange={handleStatusFilter}
                  value="processing"
                  type="radio"
                />
                <span>processing</span>
              </label>
              <label className="flex gap-4 items-center w-full px-4 py-2 cursor-pointer">
                <input
                  className="accent-primary-600"
                  name="status-filter"
                  onChange={handleStatusFilter}
                  value="noop"
                  type="radio"
                />
                <span>noop</span>
              </label>
              <label className="flex gap-4 items-center w-full px-4 py-2 cursor-pointer">
                <input
                  className="accent-primary-600"
                  name="status-filter"
                  onChange={handleStatusFilter}
                  value="active"
                  type="radio"
                />
                <span>active</span>
              </label>
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
