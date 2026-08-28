import { useMemo } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Code } from '@/components/common/Code'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TAppRelease } from '@/types'

export const BundlesTable = ({
  data,
  isLoading,
  orgId,
  appId,
}: {
  data: TAppRelease[]
  isLoading: boolean
  orgId?: string
  appId?: string
}) => {
  const columns: ColumnDef<TAppRelease>[] = useMemo(
    () => [
      {
        header: 'Release',
        accessorKey: 'id',
        cell: (props) => (
          <div className="flex flex-col gap-1">
            <Link
              href={`/${orgId}/apps/${appId}/releases/${props.getValue<string>()}`}
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
        header: 'Contents',
        id: 'contents',
        cell: ({ row }) => (
          <Text variant="subtext">
            {row.original.members?.length ?? 0} items
          </Text>
        ),
      },
      {
        header: 'Platforms',
        id: 'packages',
        cell: ({ row }) => {
          const packages = [...(row.original.packages ?? [])].sort((a, b) =>
            (a.target_platform ?? '').localeCompare(b.target_platform ?? '')
          )
          if (packages.length === 0) {
            return (
              <Text variant="subtext" theme="neutral">
                —
              </Text>
            )
          }

          return (
            <div className="flex flex-col gap-1">
              {packages.map((pkg) => (
                <div key={pkg.id} className="flex items-center gap-2">
                  <Status
                    status={pkg.status ?? 'unknown'}
                    variant="timeline"
                    isWithoutText
                    iconSize={12}
                    title={pkg.status}
                  />
                  <Text variant="subtext">{pkg.target_platform}</Text>
                </div>
              ))}
            </div>
          )
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
    ],
    [orgId, appId]
  )

  if (isLoading) {
    return <BundlesTableSkeleton />
  }

  return (
    <Table<TAppRelease>
      columns={columns}
      data={data}
      enableSearch={false}
      emptyStateProps={{
        emptyTitle: 'No releases yet',
        emptyMessage:
          'Publish an immutable release from the latest app config.',
      }}
    />
  )
}

const skeletonColumns: ColumnDef<TAppRelease>[] = [
  { header: 'Release', accessorKey: 'id' },
  { header: 'Status', accessorKey: 'status' },
  { header: 'Contents', id: 'contents' },
  { header: 'Platforms', id: 'packages' },
  { header: 'Created', accessorKey: 'created_at' },
]

export const BundlesTableSkeleton = () => (
  <TableSkeleton<TAppRelease> columns={skeletonColumns} skeletonRows={3} />
)
