'use client'

import { useSearchParams } from 'next/navigation'
import type { ColumnDef } from '@tanstack/react-table'
import { CloudPlatform } from '@/components/common/CloudPlatform'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Table } from '@/components/common/Table'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import { type IPagination } from '@/components/common/Pagination'
import { useOrg } from '@/hooks/use-org'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import { useQueryParams } from '@/hooks/use-query-params'
import type { TApp, TCloudPlatform } from '@/types'

export type TAppRow = {
  actionHref: string
  appId: string
  configVersion: number
  defaultBranch: string
  name: string
  nameHref: string
  platform: string
  sandboxHref: string
  sandboxName: string
}

function parseAppsToTableData(apps: TApp[], orgId: string): TAppRow[] {
  return apps.map((app) => {
    const sandbox = app?.sandbox_config?.public_git_vcs_config ||
      app?.sandbox_config?.connected_github_vcs_config || { repo: undefined }
    return {
      actionHref: `/${orgId}/apps/${app.id}`,
      appId: app.id,
      configVersion: app?.app_configs?.length,
      defaultBranch: app?.config_repo || 'main',
      name: app.name,
      nameHref: `/${orgId}/apps/${app.id}`,
      platform: app?.runner_config?.cloud_platform || 'unknown',
      sandboxHref: `https://${sandbox?.repo}`,
      sandboxName: sandbox?.repo,
    }
  })
}

const columns: ColumnDef<TAppRow>[] = [
  {
    accessorKey: 'name',
    header: 'App name',
    cell: (info) => (
      <span>
        <Text variant="body">
          <Link href={info.row.original.nameHref}>
            {info.getValue() as string}
          </Link>
        </Text>
        <ID>{info?.row?.original?.appId}</ID>
      </span>
    ),
    enableSorting: true,
  },
  /* {
   *   accessorKey: 'defaultBranch',
   *   header: 'Default branch',
   *   cell: (info) => (
   *     <Text family="mono" theme="neutral">
   *       {info.getValue() as string}
   *     </Text>
   *   ),
   *   enableSorting: true,
   * }, */
  {
    accessorKey: 'configVersion',
    header: 'Config version',
    cell: (info) => (
      <Text family="mono" theme="neutral">
        {info.getValue() as number}
      </Text>
    ),
    enableSorting: true,
  },
  {
    accessorKey: 'sandboxName',
    header: 'Sandbox',
    cell: (info) => (
      <Text>
        <Link href={info.row.original.sandboxHref} isExternal>
          {info.getValue() as string}
          <Icon variant="ArrowSquareOutIcon" />
        </Link>
      </Text>
    ),
  },
  {
    accessorKey: 'platform',
    header: 'Platform',
    cell: (info) => (
      <CloudPlatform
        displayVariant="icon-only"
        iconSize="24"
        platform={info?.getValue() as TCloudPlatform}
      />
    ),
    enableSorting: true,
  },
  {
    enableSorting: false,
    accessorKey: 'actionHref',
    id: 'action',
    header: '',
    cell: (info) => (
      <Text>
        <Link className="text-left" href={info.getValue() as string}>
          View <Icon variant="CaretRightIcon" />
        </Link>
      </Text>
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
