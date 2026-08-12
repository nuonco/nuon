import { useMemo } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Table } from '@/components/common/Table'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { roleAssignmentID } from '@/hooks/use-roles'
import type { TAccount } from '@/types'
import { ChangeServiceAccountRoleButton } from '@/components/service-accounts/ChangeServiceAccountRole'
import { RenameServiceAccountButton } from '@/components/service-accounts/RenameServiceAccount'
import { CreateServiceAccountTokenButton } from '@/components/service-accounts/ServiceAccountToken'
import { DeleteServiceAccountButton } from '@/components/service-accounts/DeleteServiceAccount'

export type TServiceAccountRow = {
  id: string
  name: string
  identity: string
  role: string
  createdAt: string
  account: TAccount
}

export function parseServiceAccountsToTableData(
  accounts: TAccount[],
  roleTitles: (roleType: string | undefined) => string
): TServiceAccountRow[] {
  return accounts.map((account) => {
    const roleType = roleAssignmentID(account.roles?.[0])
    return {
      id: account.id || '',
      name: account.name || account.email || account.id || '',
      identity: account.email || account.id || '',
      role: roleTitles(roleType),
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
        <RenameServiceAccountButton account={account} isMenuButton />
      </span>
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
  roleTitles: (roleType: string | undefined) => string
  isLoading: boolean
  pagination: { hasNext: boolean; offset: number; limit: number }
}) => {
  const columns: ColumnDef<TServiceAccountRow>[] = useMemo(
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
        header: 'Identity',
        accessorKey: 'identity',
        cell: (props) => <Text variant="body">{props.getValue<string>()}</Text>,
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
  { header: 'Name', accessorKey: 'name' },
  { header: 'Identity', accessorKey: 'identity' },
  { header: 'Role', accessorKey: 'role' },
  { header: 'Created', accessorKey: 'createdAt' },
  { header: 'Action', id: 'action' },
]

export const ServiceAccountsTableSkeleton = () => (
  <TableSkeleton<TServiceAccountRow> columns={skeletonColumns} skeletonRows={5} />
)
