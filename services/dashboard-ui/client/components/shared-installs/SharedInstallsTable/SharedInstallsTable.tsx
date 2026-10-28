import { useMemo } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/common/Badge'
import { EmptyState } from '@/components/common/EmptyState'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { scopeLabels, type TInstallShare } from '../types'

export type TSharedInstallRow = {
  installId: string
  installName: string
  installHref: string
  appName: string
  shares: TInstallShare[]
  sharedAt: string
}

export interface ISharedInstallsTable {
  rows: TSharedInstallRow[]
  isLoading?: boolean
  isError?: boolean
}

export const SharedInstallsTable = ({
  rows,
  isLoading = false,
  isError = false,
}: ISharedInstallsTable) => {
  const columns = useMemo<ColumnDef<TSharedInstallRow, unknown>[]>(
    () => [
      {
        accessorKey: 'installName',
        header: 'Install',
        cell: ({ row }) => (
          <div className="flex flex-col gap-0.5">
            <Link href={row.original.installHref}>
              {row.original.installName}
            </Link>
            <Text variant="subtext" theme="neutral">
              {row.original.appName}
            </Text>
          </div>
        ),
      },
      {
        id: 'shared-with',
        header: 'Shared with',
        enableSorting: false,
        cell: ({ row }) => {
          const emails = row.original.shares.map((share) => share.email)
          if (emails.length === 0) {
            return (
              <Text variant="body" theme="neutral">
                —
              </Text>
            )
          }
          if (emails.length <= 2) {
            return <Text variant="body">{emails.join(', ')}</Text>
          }
          return (
            <Text variant="body">
              {emails.slice(0, 2).join(', ')} +{emails.length - 2} more
            </Text>
          )
        },
      },
      {
        id: 'scopes',
        header: 'Access',
        enableSorting: false,
        cell: ({ row }) => {
          const scopes = [
            ...new Set(row.original.shares.flatMap((share) => share.scopes)),
          ]
          return (
            <div className="flex flex-wrap gap-1">
              {scopeLabels(scopes).map((label) => (
                <Badge key={label} theme="neutral">
                  {label}
                </Badge>
              ))}
            </div>
          )
        },
      },
      {
        id: 'status',
        header: 'Status',
        enableSorting: false,
        cell: ({ row }) => {
          const hasAccepted = row.original.shares.some(
            (share) => share.status === 'accepted'
          )
          const hasPending = row.original.shares.some(
            (share) => share.status === 'pending'
          )
          const status = hasAccepted
            ? 'accepted'
            : hasPending
              ? 'pending'
              : 'revoked'
          return <Status variant="badge" status={status} />
        },
      },
      {
        accessorKey: 'sharedAt',
        header: 'Last shared',
        cell: ({ row }) => (
          <Time time={row.original.sharedAt} format="relative" />
        ),
      },
    ],
    []
  )

  if (isError) {
    return (
      <EmptyState
        variant="table"
        emptyTitle="Unable to load shared installs"
        emptyMessage="Try refreshing the page."
      />
    )
  }

  return (
    <Table<TSharedInstallRow>
      columns={columns}
      data={rows}
      enableSearch
      searchPlaceholder="Search shared installs"
      isLoading={isLoading}
      skeletonRows={3}
      emptyStateProps={{
        variant: 'table',
        emptyTitle: 'No shared installs',
        emptyMessage:
          'Share an install from its settings to give a customer access.',
      }}
    />
  )
}
