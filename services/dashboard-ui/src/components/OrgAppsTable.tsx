'use client'

import React, { type FC, useMemo, useState } from 'react'
import { type ColumnDef } from '@tanstack/react-table'
import { DotsThreeVertical } from '@phosphor-icons/react'
import { ClickToCopy } from '@/components/ClickToCopy'
import { DataTableSearch, Table } from '@/components/DataTable'
import { InstallPlatform } from '@/components/InstallCloudPlatform'
import { Link } from '@/components/Link'
import { Heading, Text } from '@/components/Typography'
import type { TApp } from '@/types'

type TData = {
  appId: string
  name: string
  platform: string
  runner_type: string
  sandbox_version: string
}

function parseAppsToTableData(apps: Array<TApp>): Array<TData> {
  return apps.map((app) => ({
    appId: app.id,
    name: app.name,
    platform: app.runner_config?.cloud_platform,
    runner_type: app.runner_config.app_runner_type,
    sandbox_version: app.sandbox_config?.terraform_version,
  }))
}

export interface IOrgAppsTable {
  apps: Array<TApp>
  orgId: string
}

export const OrgAppsTable: FC<IOrgAppsTable> = ({ apps, orgId }) => {
  const [data, _] = useState(parseAppsToTableData(apps))
  const [columnFilters, __] = useState([])
  const [globalFilter, setGlobalFilter] = useState('')

  const columns: Array<ColumnDef<TData>> = useMemo(
    () => [
      {
        header: 'Name',
        accessorKey: 'name',
        cell: (props) => (
          <div className="flex flex-col gap-2">
            <Link href={`/${orgId}/apps/${props.row.original.appId}`}>
              <Heading variant="subheading">{props.getValue<string>()}</Heading>
            </Link>

            <Text variant="id">
              <ClickToCopy>{props.row.original.appId}</ClickToCopy>
            </Text>
          </div>
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
        header: 'Sandbox',
        accessorKey: 'sandbox_version',
        cell: (props) => <Text>{props.getValue<string>()}</Text>,
      },
      {
        header: 'Runner',
        accessorKey: 'runner_type',
        cell: (props) => <Text>{props.getValue<string>()}</Text>,
      },
      {
        id: 'test',
        enableSorting: false,
        cell: (props) => (
          <Link
            href={`/${orgId}/apps/${props.row.original.appId}`}
            variant="ghost"
          >
            <DotsThreeVertical />
          </Link>
        ),
      },
    ],
    []
  )

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
        </>
      }
      data={data}
      columns={columns}
      columnFilters={columnFilters}
      globalFilter={globalFilter}
    />
  )
}
