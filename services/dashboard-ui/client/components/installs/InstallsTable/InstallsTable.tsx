import type { ReactNode } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { CloudPlatform } from '@/components/common/CloudPlatform'
import { CloudRegion } from '@/components/common/CloudRegion'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { InstallStatuses } from '@/components/installs/InstallStatuses'
import { QuickManagementDropdown } from '@/components/installs/management/QuickManagementDropdown'
import type { TCloudPlatform, TInstall } from '@/types'

const InstallNameSkeleton = () => (
  <span className="block my-1">
    <div className="mb-1">
      <Skeleton height="16px" width="140px" />
    </div>
    <Skeleton height="12px" width="200px" />
  </span>
)

const AppNameSkeleton = () => <Skeleton height="16px" width="100px" />

const StatusesSkeleton = () => (
  <div className="flex items-center gap-2">
    <Skeleton height="20px" width="50px" className="rounded-full" />
    <Skeleton height="20px" width="60px" className="rounded-full" />
    <Skeleton height="20px" width="75px" className="rounded-full" />
  </div>
)

const RegionSkeleton = () => (
  <div className="flex items-center gap-1">
    <Skeleton height="16px" width="16px" />
    <Skeleton height="14px" width="120px" />
  </div>
)

const PlatformSkeleton = () => (
  <div className="flex items-center gap-1">
    <Skeleton height="16px" width="16px" />
    <Skeleton height="14px" width="40px" />
  </div>
)

const ActionSkeleton = () => (
  <div className="flex items-center gap-1">
    <Skeleton height="14px" width="30px" />
    <Skeleton height="12px" width="12px" />
  </div>
)

export type InstallRow = {
  action: ReactNode
  appHref: string
  appName: string
  created_at: string
  installId: string
  name: string
  nameHref: string
  region?: ReactNode
  statuses: ReactNode
  platform: ReactNode
  updated_at: string
}

export function parseInstallsToTableData(
  installs: TInstall[],
  orgId: string
): InstallRow[] {
  return installs.map((install) => ({
    appHref: `/${install.org_id}/apps/${install.app_id}`,
    appName: install?.app?.name,
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
    statuses: (
      <InstallStatuses install={install} isLabelHidden tooltipPosition="top" />
    ),
    platform: (
      <CloudPlatform
        platform={(install?.cloud_platform as TCloudPlatform) || 'unknown'}
        variant="subtext"
        colorVariant="color"
        displayVariant="icon-only"
      />
    ),
    created_at: install?.created_at,
    updated_at: install?.updated_at,
    action: (
      <div className="hidden md:block">
        <QuickManagementDropdown install={install} />
      </div>
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
    accessorKey: 'created_at',
    header: 'Created',
    cell: (info) => <Time time={info.getValue<string>()} variant="subtext" format="relative" />,
  },
  {
    accessorKey: 'updated_at',
    header: 'Updated',
    cell: (info) => (
      <Time
        time={info.getValue<string>()}
        variant="subtext"
        format="relative"
      />
    ),
  },
  {
    enableSorting: false,
    accessorKey: 'action',
    id: 'action',
    header: '',
    cell: (info) => info.getValue<ReactNode>(),
  },
]

interface IInstallsTable {
  data: InstallRow[]
  isLoading: boolean
  emptyStateAction?: ReactNode
  filterActions?: ReactNode
  pagination: { hasNext?: boolean; offset: number; limit: number }
}

export const InstallsTable = ({
  data,
  isLoading,
  emptyStateAction,
  filterActions,
  pagination,
}: IInstallsTable) => {
  if (isLoading) {
    return <InstallsTableSkeleton />
  }

  return (
    <Table<InstallRow>
      columns={columns}
      data={data}
      emptyStateProps={{
        emptyMessage:
          'An install is an instance of an application running in a customer cloud account.',
        emptyTitle: 'No installs created',
        action: emptyStateAction,
      }}
      filterActions={filterActions}
      pagination={pagination}
      searchPlaceholder="Search install name..."
    />
  )
}

export const InstallsTableSkeleton = () => {
  const skeletonData = Array.from({ length: 5 }, (_, i) => ({
    appHref: '',
    appName: '',
    installId: '',
    name: '',
    nameHref: '',
    region: <RegionSkeleton />,
    statuses: <StatusesSkeleton />,
    platform: <PlatformSkeleton />,
    created_at: '',
    updated_at: '',
    action: '',
  }))

  const skeletonColumns: ColumnDef<InstallRow>[] = [
    {
      accessorKey: 'name',
      header: 'Install name',
      cell: () => <InstallNameSkeleton />,
      enableSorting: true,
    },
    {
      accessorKey: 'appName',
      header: 'App',
      cell: () => <AppNameSkeleton />,
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
      cell: (info) => info.getValue() as ReactNode,
    },
    {
      accessorKey: 'platform',
      header: 'Platform',
      cell: (info) => info.getValue() as ReactNode,
      enableSorting: true,
    },
    {
      accessorKey: 'created_at',
      header: 'Created',
      cell: () => <Skeleton height="16px" width="100px" />,
      enableSorting: true,
    },
    {
      accessorKey: 'updated_at',
      header: 'Updated',
      cell: () => <Skeleton height="16px" width="100px" />,
      enableSorting: true,
    },
    {
      enableSorting: false,
      accessorKey: 'action',
      id: 'action',
      header: '',
      cell: () => <ActionSkeleton />,
    },
  ]

  return (
    <Table<InstallRow>
      columns={skeletonColumns}
      data={skeletonData}
      filterActions={<Skeleton height="32px" width="130px" />}
      pagination={{ limit: 5, offset: 0 }}
      isLoading={false}
      enableSorting={false}
    />
  )
}
