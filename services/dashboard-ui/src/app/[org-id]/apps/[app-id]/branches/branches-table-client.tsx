'use client'

import { useParams } from 'next/navigation'
import type { ColumnDef } from '@tanstack/react-table'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { type IPagination } from '@/components/common/Pagination'
import type { TAppBranch } from '@/types'

type TBranchRow = {
  branchId: string
  branchName: string
  lastCommit: string | null
  workflowCount: number
  createdAt: string
  href: string
}

function parseBranchesToTableData(
  branches: TAppBranch[],
  orgId: string,
  appId: string
): TBranchRow[] {
  return branches.map((branch) => ({
    branchId: branch.id || '',
    branchName: branch.name || '',
    lastCommit: branch.last_synced_commit || null,
    workflowCount: branch.workflows?.length || 0,
    createdAt: branch.created_at || '',
    href: `/${orgId}/apps/${appId}/branches/${branch.id}`,
  }))
}

const columns: ColumnDef<TBranchRow>[] = [
  {
    accessorKey: 'branchName',
    header: 'Branch Name',
    cell: (info) => (
      <span>
        <Text variant="body">
          <Link href={info.row.original.href}>{info.getValue() as string}</Link>
        </Text>
        <ID>{info.row.original.branchId}</ID>
      </span>
    ),
    enableSorting: true,
  },
  {
    accessorKey: 'lastCommit',
    header: 'Last Synced Commit',
    cell: (info) => {
      const commit = info.getValue() as string | null
      return commit ? (
        <code className="text-xs bg-gray-100 px-2 py-1 rounded">
          {commit.slice(0, 7)}
        </code>
      ) : (
        <Text variant="subtext">Not synced yet</Text>
      )
    },
  },
  {
    accessorKey: 'workflowCount',
    header: 'Workflows',
    cell: (info) => {
      const count = info.getValue() as number
      return (
        <Text variant="body">
          {count} workflow{count !== 1 ? 's' : ''}
        </Text>
      )
    },
  },
  {
    accessorKey: 'createdAt',
    header: 'Created',
    cell: (info) => <Time time={info.getValue() as string} format="relative" />,
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

interface IBranchesTableClient {
  branches: TAppBranch[]
  pagination: IPagination
}

export const BranchesTableClient = ({
  branches,
  pagination,
}: IBranchesTableClient) => {
  const params = useParams()
  const orgId = params['org-id'] as string
  const appId = params['app-id'] as string

  return (
    <Table<TBranchRow>
      columns={columns}
      data={parseBranchesToTableData(branches, orgId, appId)}
      emptyStateProps={{
        emptyMessage:
          'Create your first app branch to get started with version management.',
        emptyTitle: 'No branches yet',
      }}
      pagination={pagination}
      searchPlaceholder="Search branch name..."
    />
  )
}