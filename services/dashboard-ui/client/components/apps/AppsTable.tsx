import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import type { ColumnDef } from '@tanstack/react-table'
import { CloudPlatform } from '@/components/common/CloudPlatform'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Table } from '@/components/common/Table'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import { useOrg } from '@/hooks/use-org'
import { getApps } from '@/lib'
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
      sandboxHref: sandbox?.repo
        ? `https://github.com/${sandbox?.repo}`
        : undefined,
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
  {
    accessorKey: 'configVersion',
    header: 'Config version',
    cell: (info) => (
      <Text family="mono" theme="neutral">
        {info.getValue<number>() === 0 ? "No config" : info.getValue<number>()}
      </Text>
    ),
    enableSorting: true,
  },
  {
    accessorKey: 'sandboxName',
    header: 'Sandbox',
    cell: (info) =>
      info.row.original.sandboxHref ? (
        <Text>
          <Link href={info.row.original.sandboxHref} isExternal>
            {info.getValue() as string}
            <Icon variant="ArrowSquareOutIcon" />
          </Link>
        </Text>
      ) : (
        <Icon variant="MinusIcon" />
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
        colorVariant="color"
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

const LIMIT = 20

export const AppsTable = ({
  pollInterval = 15000,
  shouldPoll = true,
}: {
  pollInterval?: number
  shouldPoll?: boolean
} = {}) => {
  const [searchParams] = useSearchParams()
  const { org } = useOrg()
  const offset = Number(searchParams.get('offset') ?? 0)

  const { data: result, isLoading } = useQuery({
    queryKey: ['apps', org.id, offset, searchParams.get('q')],
    queryFn: () => getApps({
      orgId: org.id,
      offset,
      limit: LIMIT,
      q: searchParams.get('q') || undefined,
    }),
    placeholderData: keepPreviousData,
    refetchInterval: shouldPoll ? pollInterval : false,
  })

  if (isLoading) {
    return <AppsTableSkeleton />
  }

  const apps = result?.data ?? []

  return (
    <Table<TAppRow>
      data={parseAppsToTableData(apps, org.id)}
      columns={columns}
      emptyMessage="No applications found"
      pagination={{ hasNext: result?.pagination?.hasNext ?? false, offset, limit: LIMIT }}
      searchPlaceholder="Search app name..."
    />
  )
}

export const AppsTableSkeleton = () => {
  return <TableSkeleton columns={columns} skeletonRows={3} />
}
