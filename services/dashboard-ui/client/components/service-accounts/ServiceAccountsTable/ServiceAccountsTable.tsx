import { useMemo } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Table } from '@/components/common/Table'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TAccount, TRoleInfo } from '@/types'
import { ChangeServiceAccountRoleButton } from '@/components/service-accounts/ChangeServiceAccountRole'
import { CreateServiceAccountTokenButton } from '@/components/service-accounts/ServiceAccountToken'
import { DeleteServiceAccountButton } from '@/components/service-accounts/DeleteServiceAccount'

export type TServiceAccountRow = {
  id: string
  identity: string
  role: string
  createdAt: string
  account: TAccount
}

export function roleTitleLookup(roles: TRoleInfo[]): Record<string, string> {
  return roles.reduce<Record<string, string>>((acc, role) => {
    acc[role.role_type] = role.title
    return acc
  }, {})
}

export function parseServiceAccountsToTableData(
  accounts: TAccount[],
  roleTitles: Record<string, string>
): TServiceAccountRow[] {
  return accounts.map((account) => {
    const roleType = account.roles?.[0]?.role_type || ''
    return {
      id: account.id || '',
      identity: account.email || account.id || '',
      role: roleTitles[roleType] ?? roleType ?? '—',
      createdAt: account.created_at || '',
      account,
    }
  })
}

const ActionCell = ({ account }: { account: TAccount }) => (
  <Dropdown
    id={`action-${account.id}`}
    buttonText={<Icon variant="DotsThreeIcon" size={20} weight="bold" />}
    hideIcon
    variant="ghost"
    buttonClassName="!p-1"
    alignment="right"
  >
    <Menu>
      <span>
        <ChangeServiceAccountRoleButton account={account} isMenuButton />
      </span>
      <span>
        <CreateServiceAccountTokenButton account={account} isMenuButton />
      </span>
      <span>
        <DeleteServiceAccountButton account={account} isMenuButton />
      </span>
    </Menu>
  </Dropdown>
)

export const SERVICE_ACCOUNTS_TABLE_LIMIT = 20

export const ServiceAccountsTable = ({
  data,
  roleTitles,
  isLoading,
  pagination,
}: {
  data: TAccount[]
  roleTitles: Record<string, string>
  isLoading: boolean
  pagination: { hasNext: boolean; offset: number; limit: number }
}) => {
  const columns: ColumnDef<TServiceAccountRow>[] = useMemo(
    () => [
      {
        header: 'Identity',
        accessorKey: 'identity',
        cell: (props) => (
          <Text variant="body" weight="strong">
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
        header: 'Created',
        accessorKey: 'createdAt',
        cell: (props) => {
          const value = props.getValue<string>()
          return value ? <Time time={value} format="relative" /> : <Text>—</Text>
        },
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
    return <ServiceAccountsTableSkeleton />
  }

  return (
    <Table<TServiceAccountRow>
      columns={columns}
      data={parseServiceAccountsToTableData(data, roleTitles)}
      pagination={pagination}
      enableSearch={false}
      emptyStateProps={{
        emptyTitle: 'No service accounts yet',
        emptyMessage: 'Create a service account to automate access to the Nuon API.',
      }}
    />
  )
}

const skeletonColumns: ColumnDef<TServiceAccountRow>[] = [
  { header: 'Identity', accessorKey: 'identity' },
  { header: 'Role', accessorKey: 'role' },
  { header: 'Created', accessorKey: 'createdAt' },
  { header: 'Action', id: 'action' },
]

export const ServiceAccountsTableSkeleton = () => (
  <TableSkeleton<TServiceAccountRow> columns={skeletonColumns} skeletonRows={5} />
)
