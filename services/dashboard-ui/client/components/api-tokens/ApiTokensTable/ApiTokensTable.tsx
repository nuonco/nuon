import { useMemo } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TToken } from '@/types'
import { DeleteApiTokenButton } from '@/components/api-tokens/DeleteApiToken'

export const API_TOKENS_TABLE_LIMIT = 20

const ActionCell = ({ token }: { token: TToken }) => (
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
  isLoading,
  pagination,
}: {
  data: TToken[]
  isLoading: boolean
  pagination: { hasNext: boolean; offset: number; limit: number }
}) => {
  const columns: ColumnDef<TToken>[] = useMemo(
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
        id: 'account',
        header: 'Account',
        cell: (props) => {
          const token = props.row.original
          return (
            <div className="flex flex-col">
              <Text variant="body">{token.account_name || '—'}</Text>
              {token.account_email ? (
                <Text variant="subtext" theme="neutral">
                  {token.account_email}
                </Text>
              ) : null}
            </div>
          )
        },
      },
      {
        header: 'Role',
        accessorKey: 'account_role',
        cell: (props) => <Text variant="body">{props.getValue<string>() || '—'}</Text>,
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
    []
  )

  return (
    <Table<TToken>
      columns={columns}
      data={data}
      isLoading={isLoading}
      pagination={pagination}
      searchPlaceholder="Search tokens"
      emptyStateProps={{
        emptyTitle: 'No API tokens yet',
        emptyMessage: 'Create a token to access the Nuon API for this org.',
      }}
    />
  )
}
