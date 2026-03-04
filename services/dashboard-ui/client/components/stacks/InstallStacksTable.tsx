import { useSearchParams } from 'react-router'
import type { ReactNode } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { useQuery } from '@tanstack/react-query'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallStack } from '@/lib'
import type { TInstallStack } from '@/types'
import { StackVersionDetails } from './StackVersionDetails'

const LIMIT = 10

export type TInstallStackRow = {
  versionId: string
  appConfigId: string
  appStackConfigHref: string
  status: ReactNode
  runs: string
  createdAt: string
  more: ReactNode
}

function parseInstallStackSummaryToTableData(
  stack: TInstallStack,
  orgId: string,
  appId: string
): TInstallStackRow[] {
  return stack?.versions.map((version) => {
    return {
      versionId: version?.id,
      appConfigId: version?.app_config_id,
      appStackConfigHref: `/${orgId}/apps/${appId}`,
      status: (
        <Status variant="badge" status={version.composite_status?.status} />
      ),
      runs: version?.runs?.length?.toString() || '-',
      createdAt: version?.created_at,
      more: <StackVersionDetails version={version} />,
    }
  })
}

const columns: ColumnDef<TInstallStackRow>[] = [
  {
    accessorKey: 'versionId',
    header: 'Version',
    cell: (info) => <ID>{info.getValue<string>()}</ID>,
    enableSorting: true,
  },
  {
    accessorKey: 'appConfigId',
    header: 'App version',
    cell: (info) => (
      <Text variant="subtext">
        <Link href={info?.row?.original?.appStackConfigHref}>
          {info.getValue<string>()}
        </Link>
      </Text>
    ),
  },
  {
    enableSorting: false,
    accessorKey: 'status',
    header: 'Status',
    cell: (info) => info.getValue() as ReactNode,
  },
  {
    enableSorting: true,
    accessorKey: 'runs',
    header: 'Runs',
    cell: (info) => info.getValue<string>(),
  },
  {
    accessorKey: 'createdAt',
    header: 'Created',
    cell: (info) => (
      <Time
        time={info.getValue() as string}
        variant="subtext"
        format="relative"
      />
    ),
    enableSorting: true,
  },
  {
    accessorKey: 'more',
    header: '',
    id: 'more-options',
    cell: (info) => info.getValue() as string,
    enableSorting: true,
  },
]

export const InstallStacksTable = ({
  pollInterval = 20000,
  shouldPoll,
}: {
  pollInterval?: number
  shouldPoll?: boolean
}) => {
  const [searchParams] = useSearchParams()
  const { org } = useOrg()
  const { install } = useInstall()

  const { data: response, isLoading } = useQuery({
    queryKey: ['install-stack', org?.id, install?.id, searchParams.get('q')],
    queryFn: () => getInstallStack({ orgId: org.id, installId: install.id }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!install?.id,
  })

  const stack = response
  const pagination = { hasNext: false, offset: 0, limit: LIMIT }

  if (isLoading) return <InstallStacksTableSkeleton />
  if (!stack) return null

  return (
    <Table<TInstallStackRow>
      columns={columns}
      data={parseInstallStackSummaryToTableData(stack, org.id, install.app_id)}
      emptyMessage="No stack found"
      pagination={pagination}
      searchPlaceholder="Search stack version..."
    />
  )
}

export const InstallStacksTableSkeleton = () => {
  return <TableSkeleton columns={columns} skeletonRows={5} />
}
