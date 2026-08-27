import { useMemo } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Code } from '@/components/common/Code'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { describeMatch } from '@/components/match/types'
import { DeleteWebhookButton } from '@/components/webhooks/DeleteWebhook'
import { EditWebhookButton } from '@/components/webhooks/EditWebhook'
import type { TWebhook } from '@/types'

const ActionCell = ({ webhook }: { webhook: TWebhook }) => (
  <Dropdown
    id={`action-${webhook.id}`}
    buttonText={<Icon variant="DotsThreeIcon" size={20} weight="bold" />}
    hideIcon
    variant="ghost"
    buttonClassName="!p-1"
    alignment="right"
  >
    <Menu>
      <EditWebhookButton webhook={webhook} isMenuButton />
      <span>
        <DeleteWebhookButton webhook={webhook} isMenuButton />
      </span>
    </Menu>
  </Dropdown>
)

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
        cell: (props) => {
          // Scope subtitle mirrors the Slack channel-subscriptions table /
          // CLI describeMatch vocabulary so the dashboard, Slack modal,
          // and CLI describe a row identically.
          const scope = describeMatch(props.row.original.match)
          return (
            <div className="flex flex-col gap-1">
              <Code variant="inline" className="!px-2 !py-1 w-fit">
                {props.getValue<string>()}
              </Code>
              <Text variant="subtext" theme="neutral">
                {scope}
              </Text>
            </div>
          )
        },
      },
      {
        header: 'Signing secret',
        accessorKey: 'has_secret',
        cell: (props) =>
          props.getValue<boolean>() ? (
            <Text variant="subtext" flex>
              <Icon variant="LockKeyIcon" size={14} /> Configured
            </Text>
          ) : (
            <Text variant="subtext" theme="neutral" flex>
              <Icon variant="LockKeyOpenIcon" size={14} /> None
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
        cell: (props) => <ActionCell webhook={props.row.original} />,
      },
    ],
    []
  )

  return (
    <Table<TWebhook>
      columns={columns}
      data={data}
      isLoading={isLoading}
      enableSearch={false}
      emptyStateProps={{
        emptyTitle: 'No webhooks configured',
        emptyMessage:
          'Create a webhook to receive workflow lifecycle events from this org.',
      }}
    />
  )
}
