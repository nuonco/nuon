import { useMemo } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Dropdown } from '@/components/common/Dropdown'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { ResendOrgInviteButton } from '@/components/team/ResendOrgInvite'
import { RevokeOrgInviteButton } from '@/components/team/RevokeOrgInvite'
import type { TOrgInvite } from '@/types'

export interface IInvitedUsers {
  invites: TOrgInvite[]
  roleTitles: (roleType: string | undefined) => string
  isLoading: boolean
  isError: boolean
}

const ActionCell = ({ invite }: { invite: TOrgInvite }) => (
  <Dropdown
    id={`invite-action-${invite?.id}`}
    buttonText={<Icon variant="DotsThreeIcon" size={20} weight="bold" />}
    hideIcon
    variant="ghost"
    buttonClassName="!p-1"
    alignment="right"
  >
    <Menu>
      <span>
        <ResendOrgInviteButton invite={invite} isMenuButton />
      </span>
      <span>
        <RevokeOrgInviteButton invite={invite} isMenuButton />
      </span>
    </Menu>
  </Dropdown>
)

export const InvitedUsers = ({
  invites,
  roleTitles,
  isLoading,
  isError,
}: IInvitedUsers) => {
  const pendingInvites = useMemo(
    () => invites?.filter((invite) => invite?.status !== 'accepted') ?? [],
    [invites]
  )

  const columns = useMemo<ColumnDef<TOrgInvite, unknown>[]>(
    () => [
      {
        accessorKey: 'email',
        header: 'Email',
        cell: ({ row }) => <Text variant="body">{row.original?.email}</Text>,
      },
      {
        id: 'role',
        accessorFn: (invite) => roleTitles(invite?.role_type),
        header: 'Role',
        cell: ({ row }) => (
          <Text variant="body">{roleTitles(row.original?.role_type)}</Text>
        ),
      },
      {
        accessorKey: 'status',
        header: 'Status',
        enableSorting: false,
        cell: ({ row }) => (
          <Status variant="badge" status={row.original?.status} />
        ),
      },
      {
        id: 'more-options',
        header: '',
        enableSorting: false,
        cell: ({ row }) => <ActionCell invite={row.original} />,
      },
    ],
    [roleTitles]
  )

  if (isError) return <InvitedUsersError />

  return (
    <Table<TOrgInvite>
      columns={columns}
      data={pendingInvites}
      enableSearch={false}
      isLoading={isLoading}
      skeletonRows={2}
      emptyStateProps={{
        variant: 'table',
        emptyTitle: 'No active invites',
        emptyMessage: 'No outstanding invites to this org.',
      }}
    />
  )
}

export const InvitedUsersError = ({
  message = 'Unable to load invites. Try refreshing the page.',
  title = 'Unable to load user invites',
}: {
  message?: string
  title?: string
}) => {
  return (
    <EmptyState variant="table" emptyMessage={message} emptyTitle={title} />
  )
}
