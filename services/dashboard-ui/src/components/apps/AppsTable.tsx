'use client'

import { useSearchParams } from 'next/navigation'
import type { ColumnDef } from '@tanstack/react-table'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Table } from '@/components/common/Table'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import { type IPagination } from '@/components/common/Pagination'
import { useOrg } from '@/hooks/use-org'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import { useQueryParams } from '@/hooks/use-query-params'
import type { TApp } from '@/types'

export type TAppRow = {
  actionHref: string
  appId: string
  defaultBranch: string
  name: string
  nameHref: string
  platform: string
  sandboxHref: string
  sandboxName: string
}

function parseAppsToTableData(apps: TApp[], orgId: string): TAppRow[] {
  return apps.map((app) => ({
    actionHref: `/${orgId}/apps/${app.id}`,
    appId: app.id,
    defaultBranch: app?.config_repo || 'main',
    name: app.name,
    nameHref: `/${orgId}/apps/${app.id}`,
    platform: app?.runner_config?.cloud_platform || 'aws',
    sandboxHref: `https://${app?.sandbox_config?.public_git_vcs_config?.repo}`,
    sandboxName: app.name,
  }))
}

const columns: ColumnDef<TAppRow>[] = [
  {
    accessorKey: 'name',
    header: 'App name',
    cell: (info) => (
      <Link href={info.row.original.nameHref}>{info.getValue() as string}</Link>
    ),
    enableSorting: true,
  },
  {
    accessorKey: 'appId',
    header: 'App ID',
    cell: (info) => <ID>{info.getValue() as string}</ID>,
    enableSorting: true,
  },
  {
    accessorKey: 'defaultBranch',
    header: 'Default branch',
    cell: (info) => (
      <Text family="mono" theme="neutral">
        {info.getValue() as string}
      </Text>
    ),
    enableSorting: true,
  },
  {
    accessorKey: 'sandboxName',
    header: 'Sandbox',
    cell: (info) => (
      <Link href={info.row.original.sandboxHref}>
        {info.getValue() as string}
      </Link>
    ),
  },
  {
    accessorKey: 'platform',
    header: 'Platform',
    cell: (info) => (
      <Text className="flex items-center gap-1">
        {info.getValue() as string}
      </Text>
    ),
    enableSorting: true,
  },
  {
    enableSorting: false,
    accessorKey: 'actionHref',
    header: 'Action',
    cell: (info) => (
      <Link className="text-left" href={info.getValue() as string}>
        View
      </Link>
    ),
  },
]

export const AppsTable = ({
  apps: initApps,
  pagination,
  pollInterval = 20000,
  shouldPoll = false,
}: { apps: TApp[]; pagination: IPagination } & IPollingProps) => {
  const searchParams = useSearchParams()
  const { org } = useOrg()
  const queryParams = useQueryParams({
    offset: pagination?.offset,
    limit: pagination?.limit,
    q: searchParams.get('q'),
  })
  const { data: apps } = usePolling({
    initData: initApps,
    path: `/api/orgs/${org.id}/apps${queryParams}`,
    pollInterval,
    shouldPoll,
  })

  return (
    <Table<TAppRow>
      data={parseAppsToTableData(apps, org.id)}
      columns={columns}
      emptyMessage="No applications found"
      pagination={pagination}
      searchPlaceholder="Search app name..."
    />
  )
}

export const AppsTableSkeleton = () => {
  return <TableSkeleton columns={columns} skeletonRows={3} />
}
