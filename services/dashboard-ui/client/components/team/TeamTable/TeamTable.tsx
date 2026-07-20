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
import { ChangeRoleButton } from '@/components/team/ChangeRole'

const ROLE_LABELS: Record<string, string> = {
  org_admin: 'Admin',
  org_support: 'Support',
  org_read_only: 'Read-only',
}

export type TTeamMemberRow = {
  id: string
  name: string
  email: string
  role: string
  status: string
  account: TAccount
}

export function parseAccountToTableData(members: TAccount[]): TTeamMemberRow[] {
  return members.map((member) => {
    const roleType = member.roles?.[0]?.role_type || ''
    return {
      id: member.id || '',
      name: member.email?.split('@')[0] || 'Unknown',
      email: member.email || '',
      role: ROLE_LABELS[roleType] ?? roleType ?? '—',
      status: 'active',
      account: member,
    }
  })
}

const ActionCell = ({
  account,
  isSelf,
}: {
  account: TAccount
  isSelf: boolean
}) => (
  <Dropdown
    id={`action-${account.id}`}
    buttonText={<Icon variant="DotsThreeIcon" size={20} weight="bold" />}
    hideIcon
    variant="ghost"
    buttonClassName="!p-1"
    alignment="right"
  >
    <Menu>
      {!isSelf ? (
        <span>
          <ChangeRoleButton account={account} isMenuButton />
        </span>
      ) : null}
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
  currentAccountId,
}: {
  data: TAccount[]
  isLoading: boolean
  pagination: { hasNext: boolean; offset: number; limit: number }
  currentAccountId?: string
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
        cell: (props) => (
          <ActionCell
            account={props.row.original.account}
            isSelf={props.row.original.account.id === currentAccountId}
          />
        ),
      },
    ],
    [currentAccountId]
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
