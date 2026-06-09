import type { ReactNode } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Table } from '@/components/common/Table'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TNotebook } from '@/lib'

export type TNotebookRow = {
  id: string
  name: string
  description: ReactNode
  cellCount: number
  lastRun: ReactNode
  lastUpdated: ReactNode
  href: string
}

export function parseNotebooksToTableData(
  notebooks: TNotebook[],
  orgId: string,
  installId: string
): TNotebookRow[] {
  const basePath = `/${orgId}/installs/${installId}`
  return notebooks.map((nb) => ({
    id: nb.id,
    name: nb.name || 'Untitled notebook',
    description: nb.description ? (
      <div className="max-w-[300px] line-clamp-2">
        <Text variant="subtext" theme="neutral">
          {nb.description}
        </Text>
      </div>
    ) : null,
    cellCount: nb.cell_count ?? 0,
    lastRun: nb.latest_run_at ? (
      <Time time={nb.latest_run_at} format="relative" variant="subtext" nowrap />
    ) : null,
    lastUpdated: nb.updated_at ? (
      <Time time={nb.updated_at} format="relative" variant="subtext" nowrap />
    ) : null,
    href: `${basePath}/notebooks/${nb.id}`,
  }))
}

const columns: ColumnDef<TNotebookRow>[] = [
  {
    accessorKey: 'name',
    header: 'Name',
    cell: (info) => (
      <span>
        <Text variant="body">
          <Link href={info.row.original.href}>
            {info.getValue() as string}
          </Link>
        </Text>
        <ID>{info.row.original.id}</ID>
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
    accessorKey: 'cellCount',
    header: 'Cells',
    cell: (info) => (
      <Text variant="subtext">{info.getValue() as number}</Text>
    ),
    enableSorting: true,
  },
  {
    accessorKey: 'lastRun',
    header: 'Last run',
    cell: (info) => info.getValue() as ReactNode,
    enableSorting: false,
  },
  {
    accessorKey: 'lastUpdated',
    header: 'Last updated',
    cell: (info) => info.getValue() as ReactNode,
    enableSorting: false,
  },
]

interface INotebooksTable {
  data: TNotebookRow[]
  isLoading?: boolean
  pagination: { hasNext?: boolean; offset: number; limit: number }
}

export const NotebooksTable = ({
  data,
  isLoading,
  pagination,
}: INotebooksTable) => (
  <Table<TNotebookRow>
    columns={columns}
    data={data}
    isLoading={isLoading}
    emptyStateProps={{
      variant: 'table',
      emptyTitle: 'No notebooks yet',
      emptyMessage:
        'Create a notebook to start running commands on this install.',
    }}
    pagination={pagination}
    searchPlaceholder="Search by name or ID..."
  />
)

export const NotebooksTableSkeleton = () => (
  <TableSkeleton columns={columns} skeletonRows={5} />
)
