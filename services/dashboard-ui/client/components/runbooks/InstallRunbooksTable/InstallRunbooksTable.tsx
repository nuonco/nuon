import type { ReactNode } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/common/Badge'
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
  labels: ReactNode
  href: string
}

export function parseInstallRunbooksToTableData(
  runbooks: TInstallRunbook[],
  orgId: string,
  installId: string
): TInstallRunbookRow[] {
  return runbooks.map((ir) => {
    const basePath = `/${orgId}/installs/${installId}`
    const runbook = ir.runbook
    return {
      runbookId: ir.runbook_id ?? ir.id,
      runbookName: runbook?.name ?? '',
      description: runbook?.description ? (
        <Text variant="subtext" theme="neutral">
          {runbook.description}
        </Text>
      ) : (
        <Icon variant="MinusIcon" />
      ),
      labels:
        runbook?.labels && Object.keys(runbook.labels).length > 0 ? (
          <span className="flex flex-wrap gap-1">
            {Object.keys(runbook.labels)
              .sort()
              .map((k) => (
                <Badge key={k} variant="code" size="sm" theme="neutral">
                  {k}: {runbook.labels[k]}
                </Badge>
              ))}
          </span>
        ) : (
          <Icon variant="MinusIcon" />
        ),
      href: `${basePath}/runbooks/${ir.runbook_id ?? ir.id}`,
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
    accessorKey: 'labels',
    header: 'Labels',
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
      searchPlaceholder="Search by name or ID..."
    />
  )
}

export const InstallRunbooksTableSkeleton = () => {
  return <TableSkeleton columns={columns} skeletonRows={5} />
}
