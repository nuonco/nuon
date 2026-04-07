import { useMemo } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import type { TAccount } from '@/types'
import { RemoveUserButton } from '@/components/team/RemoveUser'

export type TTeamMemberRow = {
  id: string
  name: string
  email: string
  role: string
  status: string
  account: TAccount
}

export function parseAccountToTableData(members: TAccount[]): TTeamMemberRow[] {
  return members.map((member) => ({
    id: member.id || '',
    name: member.email?.split('@')[0] || 'Unknown',
    email: member.email || '',
    role: 'All permissions',
    status: 'active',
    account: member,
  }))
}

const ActionCell = ({ account }: { account: TAccount }) => (
  <Dropdown
    id={`action-${account.id}`}
    buttonText={<Icon variant="DotsThree" size={20} weight="bold" />}
    hideIcon
    variant="ghost"
    buttonClassName="!p-1"
    alignment="right"
  >
    <Menu>
      <span>
        <RemoveUserButton account={account} isMenuButton />
      </span>
    </Menu>
  </Dropdown>
)

export const TEAM_TABLE_LIMIT = 20

export const TeamTable = ({
  data,
  isLoading,
  pagination,
}: {
  data: TAccount[]
  isLoading: boolean
  pagination: { hasNext: boolean; offset: number; limit: number }
}) => {
  const columns: ColumnDef<TTeamMemberRow>[] = useMemo(
    () => [
      {
        header: 'Name',
        accessorKey: 'name',
        cell: (props) => (
          <Text variant="body" weight="strong">
            {props.getValue<string>()}
          </Text>
        ),
      },
      {
        header: 'Email',
        accessorKey: 'email',
        cell: (props) => (
          <Text variant="body" className="text-primary-600 dark:text-primary-400">
            {props.getValue<string>()}
          </Text>
        ),
      },
      {
        header: 'Role',
        accessorKey: 'role',
        cell: (props) => (
          <Text variant="body" className="text-primary-600 dark:text-primary-400">
            {props.getValue<string>()}
          </Text>
        ),
      },
      {
        header: 'Status',
        accessorKey: 'status',
        cell: (props) => (
          <Status status={props.getValue<string>()} variant="badge">
            Joined
          </Status>
        ),
      },
      {
        id: 'action',
        header: 'Action',
        cell: (props) => <ActionCell account={props.row.original.account} />,
      },
    ],
    []
  )

  if (isLoading) {
    return <TeamTableSkeleton />
  }

  return (
    <Table<TTeamMemberRow>
      columns={columns}
      data={parseAccountToTableData(data)}
      pagination={pagination}
      enableSearch={false}
      emptyStateProps={{
        emptyTitle: 'No team members',
        emptyMessage: 'No team members found.',
      }}
    />
  )
}

const skeletonColumns: ColumnDef<TTeamMemberRow>[] = [
  { header: 'Name', accessorKey: 'name' },
  { header: 'Email', accessorKey: 'email' },
  { header: 'Role', accessorKey: 'role' },
  { header: 'Status', accessorKey: 'status' },
  { header: 'Action', id: 'action' },
]

export const TeamTableSkeleton = () => (
  <TableSkeleton<TTeamMemberRow> columns={skeletonColumns} skeletonRows={5} />
)
