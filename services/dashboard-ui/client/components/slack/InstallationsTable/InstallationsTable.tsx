import { useMemo } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/common/Badge'
import { Code } from '@/components/common/Code'
import { Table } from '@/components/common/Table'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { DeleteOrgLinkButton } from '@/components/slack/DeleteOrgLink'
import type { TSlackInstallation, TSlackOrgLink } from '@/types'

export const InstallationsTable = ({
  data,
  links,
  isLoading,
}: {
  data: TSlackInstallation[]
  links: TSlackOrgLink[]
  isLoading: boolean
}) => {
  const columns: ColumnDef<TSlackInstallation>[] = useMemo(
    () => [
      {
        header: 'Workspace',
        accessorKey: 'team_name',
        cell: (props) => {
          const teamName = props.getValue<string | undefined>()
          const teamId = props.row.original.team_id
          return (
            <div className="flex flex-col gap-1">
              <Text variant="base" weight="strong">
                {teamName || teamId || '—'}
              </Text>
              {teamId ? (
                <Code variant="inline" className="!px-2 !py-0.5 w-fit">
                  {teamId}
                </Code>
              ) : null}
            </div>
          )
        },
      },
      {
        header: 'Status',
        accessorKey: 'status',
        cell: (props) => {
          const status = props.getValue<string | undefined>()
          const theme =
            status === 'active'
              ? 'success'
              : status === 'uninstalled'
                ? 'error'
                : 'neutral'
          return <Badge theme={theme}>{status || 'unknown'}</Badge>
        },
      },
      {
        header: 'Installed',
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
        cell: (props) => {
          const teamId = props.row.original.team_id
          const link = links.find((l) => l.team_id === teamId)
          if (!link) return null
          return (
            <div className="flex justify-end">
              <DeleteOrgLinkButton link={link} size="sm" />
            </div>
          )
        },
      },
    ],
    [links]
  )

  if (isLoading) return <InstallationsTableSkeleton />

  return (
    <Table<TSlackInstallation>
      columns={columns}
      data={data}
      enableSearch={false}
      emptyStateProps={{
        emptyTitle: 'No Slack workspaces installed',
        emptyMessage:
          'Install the Nuon Slack app in a workspace to start receiving lifecycle events here.',
      }}
    />
  )
}

const skeletonColumns: ColumnDef<TSlackInstallation>[] = [
  { header: 'Workspace', accessorKey: 'team_name' },
  { header: 'Status', accessorKey: 'status' },
  { header: 'Installed', accessorKey: 'created_at' },
  { header: '', id: 'action' },
]

export const InstallationsTableSkeleton = () => (
  <TableSkeleton<TSlackInstallation> columns={skeletonColumns} skeletonRows={2} />
)
