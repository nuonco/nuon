import type { ReactNode } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { BranchManagementDropdown } from '@/components/branches/management/BranchManagementDropdown'
import type { TAppBranch } from '@/types'

type TBranchRow = {
  branchId: string
  branchName: string
  workflowCount: number
  createdAt: string
  href: string
  action?: ReactNode
}

export function parseBranchesToTableData(
  branches: TAppBranch[],
  orgId: string,
  appId: string
): TBranchRow[] {
  return branches.map((branch) => ({
    branchId: branch.id || '',
    branchName: branch.name || '',
    workflowCount: branch.workflow_count ?? 0,
    createdAt: branch.created_at || '',
    href: `/${orgId}/apps/${appId}/branches/${branch.id}`,
    action: (
      <div className="hidden md:flex justify-end">
        <BranchManagementDropdown branch={branch} appId={appId} orgId={orgId} />
      </div>
    ),
  }))
}

const columns: ColumnDef<TBranchRow>[] = [
  {
    accessorKey: 'branchName',
    header: 'Branch name',
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
    accessorKey: 'action',
    id: 'action',
    header: '',
    cell: (info) => info.row.original.action,
  },
]

interface IBranchesTable {
  data: TBranchRow[]
  isLoading: boolean
  pagination?: { hasNext: boolean; offset: number; limit: number }
}

export const BranchesTable = ({ data, isLoading, pagination }: IBranchesTable) => {
  return (
    <Table<TBranchRow>
      columns={columns}
      data={data}
      isLoading={isLoading}
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
