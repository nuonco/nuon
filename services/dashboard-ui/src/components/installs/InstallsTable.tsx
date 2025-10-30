'use client'

import { useSearchParams } from 'next/navigation'
import type { ReactNode } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { CloudPlatform } from '@/components/common/CloudPlatform'
import { CloudRegion } from '@/components/common/CloudRegion'
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
import type { TInstall, TCloudPlatform } from '@/types'
import { InstallStatuses } from './InstallStatuses'

// NOTE: Old stufff
import { CreateInstallModal } from '@/components/old/Installs'

export type InstallRow = {
  actionHref: string
  appHref: string
  appName: string
  installId: string
  name: string
  nameHref: string
  region?: ReactNode
  statuses: ReactNode
  platform: ReactNode
}

function parseInstallsToTableData(
  installs: TInstall[],
  orgId: string
): InstallRow[] {
  return installs.map((install) => ({
    actionHref: `/${orgId}/installs/${install.id}`,
    appHref: `/${install.org_id}/apps/${install.app_id}`,
    appName: install?.app?.name,
    name: install.name,
    nameHref: `/${orgId}/installs/${install.id}`,
    installId: install.id,
    region: (
      <CloudRegion
        variant="subtext"
        platform={install?.aws_account ? 'aws' : 'azure'}
        region={install.aws_account?.region}
        location={install.azure_account?.location}
      />
    ),
    statuses: (
      <InstallStatuses install={install} isLabelHidden tooltipPosition="top" />
    ),
    platform: (
      <CloudPlatform
        platform={(install?.cloud_platform as TCloudPlatform) || 'unknown'}
        variant="subtext"
      />
    ),
  }))
}

const columns: ColumnDef<InstallRow>[] = [
  {
    accessorKey: 'name',
    header: 'Install name',
    cell: (info) => (
      <span>
        <Text variant="body">
          <Link href={info.row.original.nameHref}>
            {info.getValue() as string}
          </Link>
        </Text>
        <ID>{info.row.original.installId as string}</ID>
      </span>
    ),
    enableSorting: true,
  },
  {
    accessorKey: 'appName',
    header: 'App',
    cell: (info) => (
      <Link href={info.row.original.appHref}>{info.getValue() as string}</Link>
    ),
  },
  {
    enableSorting: false,
    accessorKey: 'statuses',
    header: 'Statuses',
    cell: (info) => info.getValue() as ReactNode,
  },
  {
    enableSorting: true,
    accessorKey: 'region',
    header: 'Region',
    cell: (info) => <Text>{info.getValue() as string}</Text>,
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

export const InstallsTable = ({
  installs: initInstalls,
  pagination,
  pollInterval = 20000,
  shouldPoll,
}: {
  installs: TInstall[]
  pagination: IPagination
} & IPollingProps) => {
  const searchParams = useSearchParams()
  const { org } = useOrg()
  const queryParams = useQueryParams({
    offset: pagination?.offset,
    limit: pagination?.limit,
    q: searchParams.get('q'),
  })
  const { data: installs } = usePolling({
    initData: initInstalls,
    path: `/api/orgs/${org.id}/installs${queryParams}`,
    pollInterval,
    shouldPoll,
  })
  return (
    <Table<InstallRow>
      columns={columns}
      data={parseInstallsToTableData(installs, org.id)}
      emptyStateProps={{
        emptyMessage:
          'An install is an instance of an application running in a customer cloud account.',
        emptyTitle: 'No installs created',
        action: <CreateInstallModal />,
      }}
      filterActions={<CreateInstallModal />}
      pagination={pagination}
      searchPlaceholder="Search install name..."
    />
  )
}

export const InstallsTableSkeleton = () => {
  return <TableSkeleton columns={columns} skeletonRows={5} />
}
