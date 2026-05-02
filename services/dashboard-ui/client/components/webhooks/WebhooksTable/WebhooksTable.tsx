import { useMemo } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Code } from '@/components/common/Code'
import { Icon } from '@/components/common/Icon'
import { Table } from '@/components/common/Table'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { DeleteWebhookButton } from '@/components/webhooks/DeleteWebhook'
import type { TWebhook } from '@/types'

export const WebhooksTable = ({
  data,
  isLoading,
}: {
  data: TWebhook[]
  isLoading: boolean
}) => {
  const columns: ColumnDef<TWebhook>[] = useMemo(
    () => [
      {
        header: 'URL',
        accessorKey: 'webhook_url',
        cell: (props) => (
          <Code variant="inline" className="!px-2 !py-1">
            {props.getValue<string>()}
          </Code>
        ),
      },
      {
        header: 'Signing secret',
        accessorKey: 'has_secret',
        cell: (props) =>
          props.getValue<boolean>() ? (
            <Text variant="subtext" flex>
              <Icon variant="LockKey" size={14} /> Configured
            </Text>
          ) : (
            <Text variant="subtext" theme="neutral" flex>
              <Icon variant="LockKeyOpen" size={14} /> None
            </Text>
          ),
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
        cell: (props) => (
          <div className="flex justify-end">
            <DeleteWebhookButton webhook={props.row.original} size="sm" />
          </div>
        ),
      },
    ],
    []
  )

  if (isLoading) {
    return <WebhooksTableSkeleton />
  }

  return (
    <Table<TWebhook>
      columns={columns}
      data={data}
      enableSearch={false}
      emptyStateProps={{
        emptyTitle: 'No webhooks configured',
        emptyMessage:
          'Create a webhook to receive workflow lifecycle events from this org.',
      }}
    />
  )
}

const skeletonColumns: ColumnDef<TWebhook>[] = [
  { header: 'URL', accessorKey: 'webhook_url' },
  { header: 'Secret', accessorKey: 'has_secret' },
  { header: 'Created', accessorKey: 'created_at' },
  { header: '', id: 'action' },
]

export const WebhooksTableSkeleton = () => (
  <TableSkeleton<TWebhook> columns={skeletonColumns} skeletonRows={3} />
)
