import type { ReactNode } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Link } from '@/components/common/Link'
import { Table } from '@/components/common/Table'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TRunbook } from '@/lib/ctl-api/apps/runbooks'

export type TRunbookRow = {
  runbookId: string
  runbookName: string
  description: ReactNode
  labels: ReactNode
  lastUpdated: ReactNode
  href: string
}

export function parseRunbooksToTableData(
  runbooks: TRunbook[],
  orgId: string,
  appId: string
): TRunbookRow[] {
  return runbooks.map((runbook) => {
    const basePath = `/${orgId}/apps/${appId}`
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
      labels:
        runbook.labels && Object.keys(runbook.labels).length > 0 ? (
          <span className="flex flex-wrap gap-1">
            {Object.keys(runbook.labels)
              .sort()
              .map((k) => (
                <LabelBadge key={k} labelKey={k} labelValue={runbook.labels[k]} size="sm" />
              ))}
          </span>
        ) : (
          <Icon variant="MinusIcon" />
        ),
      lastUpdated: runbook.updated_at ? (
        <Text flex className="gap-2">
          <Icon variant="CalendarBlankIcon" />
          <Time time={runbook.updated_at} format="relative" variant="subtext" />
        </Text>
      ) : (
        <Icon variant="MinusIcon" />
      ),
      href: `${basePath}/runbooks/${runbook.id}`,
    }
  })
}

const columns: ColumnDef<TRunbookRow>[] = [
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
    accessorKey: 'lastUpdated',
    header: 'Last updated',
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

interface IRunbooksTable {
  data: TRunbookRow[]
  isLoading: boolean
  pagination: { hasNext?: boolean; offset: number; limit: number }
}

export const RunbooksTable = ({ data, isLoading, pagination }: IRunbooksTable) => {
  return (
    <Table<TRunbookRow>
      columns={columns}
      data={data}
      isLoading={isLoading}
      emptyStateProps={{
        variant: 'actions',
        emptyTitle: 'No runbooks yet',
        emptyMessage: 'Runbooks let you define operational procedures for your installs.',
      }}
      pagination={pagination}
      searchPlaceholder="Search by name or ID..."
    />
  )
}

export const RunbooksTableSkeleton = () => {
  return <TableSkeleton columns={columns} skeletonRows={5} />
}
