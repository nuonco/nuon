import { useMemo } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { scopeLabels, type TInstallShare } from '../types'

export interface ISharedWithList {
  shares: TInstallShare[]
  isLoading?: boolean
  isError?: boolean
  onRevoke?: (shareId: string) => void
  onResend?: (shareId: string) => void
}

const ActionCell = ({
  share,
  onRevoke,
  onResend,
}: {
  share: TInstallShare
  onRevoke?: (shareId: string) => void
  onResend?: (shareId: string) => void
}) => {
  if (!share.id) return null

  return (
    <Dropdown
      id={`share-action-${share.id}`}
      buttonText={<Icon variant="DotsThreeIcon" size={20} weight="bold" />}
      hideIcon
      variant="ghost"
      buttonClassName="!p-1"
      alignment="right"
    >
      <Menu>
        {share.status === 'pending' && onResend ? (
          <Button onClick={() => onResend(share.id)}>Resend invite</Button>
        ) : null}
        {onRevoke && share.status !== 'revoked' ? (
          <Button variant="danger" onClick={() => onRevoke(share.id)}>
            Revoke access
          </Button>
        ) : null}
      </Menu>
    </Dropdown>
  )
}

export const SharedWithList = ({
  shares,
  isLoading = false,
  isError = false,
  onRevoke,
  onResend,
}: ISharedWithList) => {
  const columns = useMemo<ColumnDef<TInstallShare, unknown>[]>(
    () => [
      {
        accessorKey: 'email',
        header: 'Email',
        cell: ({ row }) => (
          <Text variant="body">{row.original.email}</Text>
        ),
      },
      {
        id: 'scopes',
        header: 'Access',
        enableSorting: false,
        cell: ({ row }) => (
          <div className="flex flex-wrap gap-1">
            {scopeLabels(row.original.scopes).map((label) => (
              <Badge key={label} theme="neutral">
                {label}
              </Badge>
            ))}
          </div>
        ),
      },
      {
        accessorKey: 'status',
        header: 'Status',
        enableSorting: false,
        cell: ({ row }) => (
          <Status variant="badge" status={row.original.status} />
        ),
      },
      {
        accessorKey: 'shared_at',
        header: 'Shared',
        cell: ({ row }) => (
          <Time time={row.original.shared_at} format="relative" />
        ),
      },
      {
        id: 'more-options',
        header: '',
        enableSorting: false,
        cell: ({ row }) => (
          <ActionCell
            share={row.original}
            onRevoke={onRevoke}
            onResend={onResend}
          />
        ),
      },
    ],
    [onRevoke, onResend]
  )

  if (isError) {
    return (
      <EmptyState
        variant="table"
        emptyTitle="Unable to load shares"
        emptyMessage="Try refreshing the page."
      />
    )
  }

  return (
    <Table<TInstallShare>
      columns={columns}
      data={shares}
      enableSearch={false}
      isLoading={isLoading}
      skeletonRows={2}
      emptyStateProps={{
        variant: 'table',
        emptyTitle: 'Not shared yet',
        emptyMessage: 'Share this install to give a customer access.',
      }}
    />
  )
}
