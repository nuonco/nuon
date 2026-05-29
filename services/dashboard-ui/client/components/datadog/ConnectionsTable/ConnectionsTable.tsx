import { useMemo } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/common/Badge'
import { Code } from '@/components/common/Code'
import { Table } from '@/components/common/Table'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { DeleteConnectionButton } from '@/components/datadog/DeleteConnection'
import { EditConnectionButton } from '@/components/datadog/EditConnection'
import { KNOWN_SITES } from '@/components/datadog/SiteInput'
import { TestConnectionButton } from '@/components/datadog/TestConnection'
import type { TDatadogConnection } from '@/types'

// siteLabel translates a stored `site` value back into the human label
// users picked. Known keys map back to the dropdown row. Custom URLs
// surface verbatim because they ARE the identity — there's nothing
// friendlier to show.
const siteLabel = (site: string | undefined): string => {
  if (!site) return '—'
  const known = KNOWN_SITES.find((s) => s.key === site)
  return known ? known.label.split(' (')[0] + ` — ${known.host}` : site
}

export const ConnectionsTable = ({
  data,
  isLoading,
}: {
  data: TDatadogConnection[]
  isLoading: boolean
}) => {
  const columns: ColumnDef<TDatadogConnection>[] = useMemo(
    () => [
      {
        header: 'Name',
        accessorKey: 'name',
        cell: (props) => {
          const c = props.row.original
          return (
            <div className="flex flex-col gap-1">
              <div className="flex items-center gap-2">
                <Text variant="base" weight="strong">
                  {c.name || '(unnamed)'}
                </Text>
                {c.purpose ? (
                  <Badge theme="neutral">{c.purpose}</Badge>
                ) : null}
              </div>
              {c.id ? (
                <Code variant="inline" className="!px-2 !py-0.5 w-fit">
                  {c.id}
                </Code>
              ) : null}
            </div>
          )
        },
      },
      {
        header: 'Site',
        id: 'site',
        cell: (props) => (
          <Text variant="subtext" theme="neutral">
            {siteLabel(props.row.original.site)}
          </Text>
        ),
      },
      {
        header: 'Status',
        accessorKey: 'status',
        cell: (props) => {
          const status = props.getValue<string | undefined>()
          // 'revoked' is the auto-flipped state the lifecycle hook sets
          // after DD returns 401/403. Surface it as `warn` (not error)
          // because the row is still useful — Test re-validates the key,
          // and a successful re-test flips it back.
          if (status === 'revoked') {
            return <Badge theme="warn">Revoked — re-add keys</Badge>
          }
          return <Badge theme="success">Verified</Badge>
        },
      },
      {
        header: 'Connected',
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
        cell: (props) => (
          <div className="flex justify-end gap-1">
            <TestConnectionButton connection={props.row.original} size="sm" />
            <EditConnectionButton connection={props.row.original} size="sm" />
            <DeleteConnectionButton connection={props.row.original} size="sm" />
          </div>
        ),
      },
    ],
    []
  )

  if (isLoading) return <ConnectionsTableSkeleton />

  return (
    <Table<TDatadogConnection>
      columns={columns}
      data={data}
      enableSearch={false}
      emptyStateProps={{
        emptyTitle: 'No Datadog connections',
        emptyMessage:
          'Connect your or a customer\'s Datadog tenant to stream Nuon events into their event stream.',
      }}
    />
  )
}

const skeletonColumns: ColumnDef<TDatadogConnection>[] = [
  { header: 'Name', accessorKey: 'name' },
  { header: 'Site', id: 'site' },
  { header: 'Status', accessorKey: 'status' },
  { header: 'Connected', accessorKey: 'created_at' },
  { header: '', id: 'action' },
]

export const ConnectionsTableSkeleton = () => (
  <TableSkeleton<TDatadogConnection>
    columns={skeletonColumns}
    skeletonRows={3}
  />
)
