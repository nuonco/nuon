import { useMemo } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Table } from '@/components/common/Table'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TStaticToken } from '@/types'
import { DeleteApiTokenButton } from '@/components/api-tokens/DeleteApiToken'

export const API_TOKENS_TABLE_LIMIT = 20

const ActionCell = ({ token }: { token: TStaticToken }) => (
  <Dropdown
    id={`action-${token.id}`}
    buttonText={<Icon variant="DotsThreeIcon" size={20} weight="bold" />}
    hideIcon
    variant="ghost"
    buttonClassName="!p-1"
    alignment="right"
  >
    <Menu>
      <span>
        <DeleteApiTokenButton token={token} isMenuButton />
      </span>
    </Menu>
  </Dropdown>
)

export const ApiTokensTable = ({
  data,
  roleTitles,
  isLoading,
  pagination,
}: {
  data: TStaticToken[]
  roleTitles: (roleType: string | undefined) => string
  isLoading: boolean
  pagination: { hasNext: boolean; offset: number; limit: number }
}) => {
  const columns: ColumnDef<TStaticToken>[] = useMemo(
    () => [
      {
        header: 'Name',
        accessorKey: 'name',
        cell: (props) => (
          <Text variant="body" weight="strong">
            {props.getValue<string>() || 'Unnamed token'}
          </Text>
        ),
      },
      {
        header: 'Role',
        accessorKey: 'role',
        cell: (props) => (
          <Text variant="body">{roleTitles(props.getValue<string>())}</Text>
        ),
      },
      {
        header: 'Created',
        accessorKey: 'created_at',
        cell: (props) => (
          <Time time={props.getValue<string>()} format="relative" />
        ),
      },
      {
        header: 'Expires',
        accessorKey: 'expires_at',
        cell: (props) => (
          <Time time={props.getValue<string>()} format="short-datetime" />
        ),
      },
      {
        id: 'action',
        header: 'Action',
        cell: (props) => <ActionCell token={props.row.original} />,
      },
    ],
    [roleTitles]
  )

  if (isLoading) {
    return <ApiTokensTableSkeleton />
  }

  return (
    <Table<TStaticToken>
      columns={columns}
      data={data}
      pagination={pagination}
      searchPlaceholder="Search tokens"
      emptyStateProps={{
        emptyTitle: 'No API tokens yet',
        emptyMessage: 'Create a token to access the Nuon API for this org.',
      }}
    />
  )
}

const skeletonColumns: ColumnDef<TStaticToken>[] = [
  { header: 'Name', accessorKey: 'name' },
  { header: 'Role', accessorKey: 'role' },
  { header: 'Created', accessorKey: 'created_at' },
  { header: 'Expires', accessorKey: 'expires_at' },
  { header: 'Action', id: 'action' },
]

export const ApiTokensTableSkeleton = () => (
  <TableSkeleton<TStaticToken> columns={skeletonColumns} skeletonRows={5} />
)
