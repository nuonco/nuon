import { useMemo } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TInstallStack } from '@/types'
import { StackVersionActions } from '../StackVersionActions'
import { StackVersionDetails } from '../StackVersionDetails'

export type TStackVersion = TInstallStack['versions'][number]

const panelLinkClass =
  '!p-0 !h-auto !border-none !rounded-none !bg-transparent hover:!bg-transparent active:!bg-transparent focus:!shadow-none text-primary-600 dark:text-primary-500 hover:text-primary-800 hover:dark:text-primary-400 active:text-primary-900 active:dark:text-primary-600'

interface IInstallStacksTable {
  versions: TStackVersion[]
  orgId?: string
  appId?: string
  isLoading?: boolean
}

export const InstallStacksTable = ({
  versions,
  orgId,
  appId,
  isLoading,
}: IInstallStacksTable) => {
  const columns = useMemo<ColumnDef<TStackVersion, unknown>[]>(
    () => [
      {
        accessorKey: 'id',
        header: 'Version',
        cell: ({ row }) => (
          <StackVersionDetails
            version={row.original}
            panelKey={`stack-version-${row.original?.id}`}
            triggerButton={{
              variant: 'ghost',
              className: panelLinkClass,
              children: <ID>{row.original?.id}</ID>,
            }}
          />
        ),
      },
      {
        accessorKey: 'composite_status.status',
        header: 'Status',
        enableSorting: false,
        cell: ({ row }) => (
          <Status
            variant="badge"
            status={row.original?.composite_status?.status}
          />
        ),
      },
      {
        id: 'runs',
        accessorFn: (version) => version?.runs?.length ?? 0,
        header: 'Runs',
        cell: ({ row }) => (
          <Text variant="subtext" family="mono">
            {row.original?.runs?.length ?? 0}
          </Text>
        ),
      },
      {
        accessorKey: 'app_config_id',
        header: 'App config',
        cell: ({ row }) =>
          orgId && appId ? (
            <Link href={`/${orgId}/apps/${appId}`} variant="inline">
              {row.original?.app_config_id}
            </Link>
          ) : (
            <ID>{row.original?.app_config_id}</ID>
          ),
      },
      {
        accessorKey: 'created_at',
        header: 'Created',
        cell: ({ row }) => (
          <Time
            variant="subtext"
            time={row.original?.created_at}
            format="relative"
          />
        ),
      },
      {
        id: 'more-options',
        header: '',
        enableSorting: false,
        cell: ({ row }) => <StackVersionActions version={row.original} />,
      },
    ],
    [orgId, appId]
  )

  return (
    <Table<TStackVersion>
      columns={columns}
      data={versions ?? []}
      enableSearch={false}
      initialSorting={[{ id: 'created_at', desc: true }]}
      isLoading={isLoading}
      emptyStateProps={{
        emptyTitle: 'No stack versions yet',
        emptyMessage:
          'Versions appear here each time this install applies a new stack config.',
        variant: 'table',
      }}
    />
  )
}
