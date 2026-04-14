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
import { SimpleInstallStatuses } from '@/components/installs/InstallStatuses'
import type { TInstall, TCloudPlatform } from '@/types'

export type InstallRow = {
  actionHref: string
  installId: string
  name: string
  nameHref: string
  region?: ReactNode
  statuses: ReactNode
  platform: ReactNode
}

export function parseInstallsToTableData(
  installs: TInstall[],
  orgId: string,
  appId?: string
): InstallRow[] {
  return installs.map((install) => ({
    actionHref: `/${orgId}/installs/${install.id}`,
    appName: appId ? undefined : install?.app?.name,
    name: install.name,
    nameHref: `/${orgId}/installs/${install.id}`,
    installId: install.id,
    region: (
      <CloudRegion
        variant="subtext"
        platform={install?.gcp_account ? 'gcp' : install?.aws_account ? 'aws' : 'azure'}
        region={install.gcp_account?.region || install.aws_account?.region}
        location={install.azure_account?.location}
      />
    ),
    statuses: <SimpleInstallStatuses install={install} isLabelHidden />,
    platform: (
      <CloudPlatform
        platform={(install?.cloud_platform as TCloudPlatform) || 'unknown'}
        variant="subtext"
        colorVariant="color"
        displayVariant="icon-only"
        iconSize="20"
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
    enableSorting: false,
    accessorKey: 'statuses',
    header: 'Statuses',
    cell: (info) => info.getValue() as ReactNode,
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
    enableSorting: true,
    accessorKey: 'region',
    header: 'Region',
    cell: (info) => <Text>{info.getValue() as string}</Text>,
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

interface IAppInstallsTable {
  data: InstallRow[]
  isLoading: boolean
  emptyAction?: ReactNode
  pagination: { hasNext?: boolean; offset: number; limit: number }
}

export const AppInstallsTable = ({
  data,
  isLoading,
  emptyAction,
  pagination,
}: IAppInstallsTable) => {
  if (isLoading) {
    return <AppInstallsTableSkeleton />
  }

  return (
    <Table<InstallRow>
      columns={columns}
      data={data}
      emptyStateProps={{
        emptyMessage:
          'An install is an instance of an application running in a customer cloud account.',
        emptyTitle: 'No installs created',
        action: emptyAction,
      }}
      pagination={pagination}
      searchPlaceholder="Search install name..."
    />
  )
}

export const AppInstallsTableSkeleton = () => {
  return <TableSkeleton columns={columns} skeletonRows={5} />
}
