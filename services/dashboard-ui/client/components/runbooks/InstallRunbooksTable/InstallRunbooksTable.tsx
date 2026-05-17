import type { ReactNode } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Table } from '@/components/common/Table'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import type { TInstallRunbook } from '@/lib/ctl-api/installs/runbooks'

export type TInstallRunbookRow = {
  runbookId: string
  runbookName: string
  description: ReactNode
  stepCount: ReactNode
  href: string
}

export function parseInstallRunbooksToTableData(
  runbooks: TInstallRunbook[],
  orgId: string,
  installId: string
): TInstallRunbookRow[] {
  return runbooks.map((runbook) => {
    const basePath = `/${orgId}/installs/${installId}`
    return {
      runbookId: runbook.id,
      runbookName: runbook.name,
      description: runbook.description ? (
        <Text variant="subtext" theme="neutral">
          {runbook.description}
        </Text>
      ) : (
        <Icon variant="MinusIcon" />
      ),
      stepCount: (
        <Text variant="subtext">
          {runbook.steps?.length ?? 0} step{(runbook.steps?.length ?? 0) === 1 ? '' : 's'}
        </Text>
      ),
      href: `${basePath}/runbooks/${runbook.id}`,
    }
  })
}

const columns: ColumnDef<TInstallRunbookRow>[] = [
  {
    accessorKey: 'runbookName',
    header: 'Runbook',
    cell: (info) => (
      <span>
        <Text variant="body">
          <Link href={info.row.original.href}>{info.getValue() as string}</Link>
        </Text>
        <ID>{info.row.original.runbookId}</ID>
      </span>
    ),
    enableSorting: true,
  },
  {
    accessorKey: 'description',
    header: 'Description',
    cell: (info) => info.getValue() as ReactNode,
    enableSorting: false,
  },
  {
    accessorKey: 'stepCount',
    header: 'Steps',
    cell: (info) => info.getValue() as ReactNode,
    enableSorting: false,
  },
  {
    enableSorting: false,
    accessorKey: 'href',
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

interface IInstallRunbooksTable {
  data: TInstallRunbookRow[]
  isLoading?: boolean
  pagination: { hasNext?: boolean; offset: number; limit: number }
}

export const InstallRunbooksTable = ({ data, isLoading, pagination }: IInstallRunbooksTable) => {
  return (
    <Table<TInstallRunbookRow>
      columns={columns}
      data={data}
      isLoading={isLoading}
      emptyStateProps={{
        variant: 'actions',
        emptyTitle: 'No runbooks yet',
        emptyMessage: 'Runbooks let you run operational procedures on this install.',
      }}
      pagination={pagination}
      searchPlaceholder="Search runbook name..."
    />
  )
}

export const InstallRunbooksTableSkeleton = () => {
  return <TableSkeleton columns={columns} skeletonRows={5} />
}
