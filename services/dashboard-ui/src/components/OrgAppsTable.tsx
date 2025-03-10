'use client'

import React, { type FC, useMemo, useState } from 'react'
import { type ColumnDef } from '@tanstack/react-table'
import { CaretRight } from '@phosphor-icons/react'
import { AppSandboxRepoDirLink } from '@/components/AppSandbox'
import { DataTableSearch, Table } from '@/components/DataTable'
import { InstallPlatform } from '@/components/InstallCloudPlatform'
import { Link } from '@/components/Link'
import { ID, Text } from '@/components/Typography'
import type { TApp } from '@/types'

type TData = {
  appId: string
  name: string
  platform: string
  runner_type: string
  sandbox_repo: string | null
  isGithubConnected: boolean
}

function parseAppsToTableData(apps: Array<TApp>): Array<TData> {
  return apps.map((app) => {
    const isGithubConnected = Boolean(
      app.sandbox_config?.connected_github_vcs_config
    )
    const repo =
      app.sandbox_config?.connected_github_vcs_config ||
      app.sandbox_config?.public_git_vcs_config
    const sandbox_repo = repo ? `${repo?.repo}/${repo?.directory}` : null

    return {
      appId: app.id,
      name: app.name,
      platform: app.runner_config?.cloud_platform,
      runner_type: app.runner_config.app_runner_type,
      sandbox_repo,
      isGithubConnected,
    }
  })
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
              <Text variant="med-14">{props.getValue<string>()}</Text>
            </Link>

            <ID id={props.row.original.appId} />
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
        accessorKey: 'sandbox_repo',
        cell: (props) => {
          const repoDirPath = props.getValue<string | null>()
          return repoDirPath ? (
            <AppSandboxRepoDirLink
              repoDirPath={repoDirPath}
              isGithubConnected={props.row.original.isGithubConnected}
            />
          ) : (
            <Text>No sandbox config</Text>
          )
        },
      },
      {
        header: 'Runner',
        accessorKey: 'runner_type',
        cell: (props) => (
          <Text>{props.getValue<string>() || 'No runner config'}</Text>
        ),
      },
      {
        id: 'test',
        enableSorting: false,
        cell: (props) => (
          <Link
            href={`/${orgId}/apps/${props.row.original.appId}`}
            variant="ghost"
          >
            <CaretRight />
          </Link>
        ),
      },
    ],
    []
  )

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
        </>
      }
      data={data}
      columns={columns}
      columnFilters={columnFilters}
      emptyMessage="Reset your search and try again."
      emptyTitle="No apps found"
      globalFilter={globalFilter}
    />
  )
}
