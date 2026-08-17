import { useMemo } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Code } from '@/components/common/Code'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { DownloadBundleButton } from '@/components/apps/bundles/DownloadBundle'
import { RegisterAirgapInstallButton } from '@/components/apps/bundles/RegisterAirgapInstall'
import type { TAirgapBundle } from '@/types'
import { formatBytes } from '@/utils/string-utils'

export const BundlesTable = ({
  data,
  isLoading,
  orgId,
  appId,
}: {
  data: TAirgapBundle[]
  isLoading: boolean
  orgId?: string
  appId?: string
}) => {
  const columns: ColumnDef<TAirgapBundle>[] = useMemo(
    () => [
      {
        header: 'Bundle',
        accessorKey: 'id',
        cell: (props) => (
          <div className="flex flex-col gap-1">
            <Link
              href={`/${orgId}/apps/${appId}/bundles/${props.getValue<string>()}`}
            >
              <Code variant="inline" className="!px-2 !py-1 w-fit">
                {props.getValue<string>()}
              </Code>
            </Link>
            {props.row.original.status_description ? (
              <Text variant="subtext" theme="neutral">
                {props.row.original.status_description}
              </Text>
            ) : null}
          </div>
        ),
      },
      {
        header: 'Status',
        accessorKey: 'status',
        cell: (props) => {
          const status = props.getValue<string | undefined>()
          return status ? (
            <Status status={status} variant="badge" />
          ) : (
            <Text variant="subtext" theme="neutral">
              —
            </Text>
          )
        },
      },
      {
        header: 'Platform',
        accessorKey: 'target_platform',
        cell: (props) => (
          <Text variant="subtext">{props.getValue<string>() || '—'}</Text>
        ),
      },
      {
        header: 'Size',
        accessorKey: 'size',
        cell: (props) => {
          const size = props.getValue<number | undefined>()
          return <Text variant="subtext">{size ? formatBytes(size) : '—'}</Text>
        },
      },
      {
        header: 'Created',
        accessorKey: 'created_at',
        cell: (props) => {
          const time = props.getValue<string | undefined>()
          return time ? (
            <Time variant="subtext" time={time} format="relative" />
          ) : (
            <Text variant="subtext" theme="neutral">
              —
            </Text>
          )
        },
      },
      {
        id: 'action',
        header: '',
        cell: (props) =>
          props.row.original.status === 'active' ? (
            <div className="flex justify-end gap-1">
              <RegisterAirgapInstallButton bundle={props.row.original} />
              <DownloadBundleButton bundle={props.row.original} />
            </div>
          ) : null,
      },
    ],
    [orgId, appId]
  )

  if (isLoading) {
    return <BundlesTableSkeleton />
  }

  return (
    <Table<TAirgapBundle>
      columns={columns}
      data={data}
      enableSearch={false}
      emptyStateProps={{
        emptyTitle: 'No bundles yet',
        emptyMessage:
          'Create a bundle to package this app for air-gapped installs.',
      }}
    />
  )
}

const skeletonColumns: ColumnDef<TAirgapBundle>[] = [
  { header: 'Bundle', accessorKey: 'id' },
  { header: 'Status', accessorKey: 'status' },
  { header: 'Platform', accessorKey: 'target_platform' },
  { header: 'Size', accessorKey: 'size' },
  { header: 'Created', accessorKey: 'created_at' },
  { header: '', id: 'action' },
]

export const BundlesTableSkeleton = () => (
  <TableSkeleton<TAirgapBundle> columns={skeletonColumns} skeletonRows={3} />
)
