export default {
  title: 'Playground / Table v2 / VCS connections',
}

import type { ColumnDef } from '@tanstack/react-table'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TVCSConnectionRow } from '@/components/vcs-connections/VCSConnectionsTable/VCSConnectionsTable'
import { StoryFrame, hoursAgo } from './story-utils'
import { TableV2 } from '@/components/common/TableV2'

const accountName = (row: TVCSConnectionRow) =>
  row.connection.github_account_name ||
  row.connection.github_account_id ||
  'GitHub'

const mockConnections: TVCSConnectionRow[] = [
  {
    connection: { id: 'vcs-1', github_account_name: 'acme-example' },
    href: '/org-1/settings/vcs/vcs-1',
    status: 'active',
    checkedAt: hoursAgo(1),
  },
  {
    connection: { id: 'vcs-2', github_account_name: 'acme-platform' },
    href: '/org-1/settings/vcs/vcs-2',
    status: 'suspended',
    checkedAt: hoursAgo(30),
  },
  {
    connection: { id: 'vcs-3', github_account_name: 'acme-labs' },
    href: '/org-1/settings/vcs/vcs-3',
    status: 'unknown',
  },
  {
    connection: { id: 'vcs-4', github_account_name: 'acme-data' },
    href: '/org-1/settings/vcs/vcs-4',
    isLoadingStatus: true,
  },
  {
    connection: { id: 'vcs-5', github_account_name: 'acme-security' },
    href: '/org-1/settings/vcs/vcs-5',
    status: 'active',
    checkedAt: hoursAgo(5),
  },
]

const columns: ColumnDef<TVCSConnectionRow>[] = [
  {
    id: 'account',
    header: 'Account',
    accessorFn: accountName,
    enableHiding: false,
    cell: ({ row }) => (
      <span className="flex items-center gap-2">
        <Icon variant="GitHub" size={16} />
        <Link href={row.original.href} variant="inline">
          {accountName(row.original)}
        </Link>
      </span>
    ),
  },
  {
    id: 'status',
    header: 'Status',
    enableSorting: false,
    meta: { card: 'status' },
    cell: ({ row }) => {
      const { status, isLoadingStatus } = row.original
      if (isLoadingStatus) return <Status loading variant="badge" />
      if (!status) return <Text theme="neutral">—</Text>
      return <Status status={status} variant="badge" />
    },
  },
  {
    id: 'checked_at',
    header: 'Last checked',
    enableSorting: false,
    cell: ({ row }) =>
      row.original.checkedAt ? (
        <Time time={row.original.checkedAt} format="relative" />
      ) : (
        <Text theme="neutral">—</Text>
      ),
  },
]

const emptyStateProps = {
  emptyTitle: 'No VCS connections yet',
  emptyMessage:
    'Connect a GitHub account to build components from your repositories.',
  action: (
    <Button variant="primary">
      <Icon variant="PlusIcon" />
      Connect GitHub
    </Button>
  ),
}

export const Default = () => (
  <StoryFrame note="Mirrors the VCS connections table: three columns, no search, no pagination. One row is still checking its status, and one has never been checked.">
    <TableV2
      tableId="playground-vcs-connections"
      columns={columns}
      data={mockConnections}
      enableSearch={false}
      emptyStateProps={emptyStateProps}
    />
  </StoryFrame>
)

export const Loading = () => (
  <StoryFrame>
    <TableV2
      tableId="playground-vcs-connections"
      columns={columns}
      data={[]}
      enableSearch={false}
      isLoading
    />
  </StoryFrame>
)

export const Empty = () => (
  <StoryFrame>
    <TableV2
      tableId="playground-vcs-connections"
      columns={columns}
      data={[]}
      enableSearch={false}
      emptyStateProps={emptyStateProps}
    />
  </StoryFrame>
)
