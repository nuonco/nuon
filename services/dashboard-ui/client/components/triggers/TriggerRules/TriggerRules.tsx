import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import type { TTriggerRule } from '@/types'
import type { ColumnDef } from '@tanstack/react-table'

export const TriggerRules = ({
  data,
  hasError,
  isLoading,
  onRetry,
  orgId,
  triggerId,
}: {
  data: TTriggerRule[]
  hasError: boolean
  isLoading: boolean
  onRetry: () => void
  orgId: string
  triggerId: string
}) => {
  const columns: ColumnDef<TTriggerRule>[] = [
    {
      header: 'Rule',
      accessorKey: 'name',
      cell: ({ row }) => (
        <Link
          href={`/${orgId}/settings/triggers/${triggerId}/rules/${row.original?.id}`}
        >
          {row.original?.name || row.original?.id || 'Unnamed rule'}
        </Link>
      ),
    },
    {
      header: 'App',
      id: 'owner',
      cell: ({ row }) => {
        const rule = row.original
        const appHref = rule?.app_id
          ? `/${orgId}/apps/${rule.app_id}`
          : undefined
        return appHref ? (
          <Link href={appHref}>{rule?.app_name || rule.app_id}</Link>
        ) : (
          <Text variant="subtext">Unknown app</Text>
        )
      },
    },
    {
      header: 'Event types',
      id: 'eventTypes',
      cell: ({ row }) => (
        <Text variant="subtext">
          {row.original?.event_types?.join(', ') || 'All event types'}
        </Text>
      ),
    },
    {
      header: 'Target',
      id: 'target',
      cell: ({ row }) => {
        const rule = row.original
        if (rule?.target_type === 'runbook' || rule?.runbook_id) {
          return (
            <div className="flex flex-col gap-1">
              {rule?.app_id && rule?.runbook_id ? (
                <Link
                  href={`/${orgId}/apps/${rule.app_id}/runbooks/${rule.runbook_id}`}
                >
                  {rule?.runbook_name || rule.runbook_id}
                </Link>
              ) : (
                <Text variant="subtext">
                  {rule?.runbook_name || 'Unknown runbook'}
                </Text>
              )}
              <Text variant="subtext" theme="neutral">
                Install {rule?.install_name || 'not configured'}
              </Text>
            </div>
          )
        }
        return rule?.app_id && rule?.app_branch_id ? (
          <Link
            href={`/${orgId}/apps/${rule.app_id}/branches/${rule.app_branch_id}`}
          >
            {rule?.app_branch_name || rule.app_branch_id}
          </Link>
        ) : (
          <Text variant="subtext" theme="neutral">
            Unknown app branch
          </Text>
        )
      },
    },
    {
      header: 'Matching',
      id: 'matching',
      cell: ({ row }) => (
        <Text variant="subtext">
          {row.original?.filters?.length ?? 0} filter
          {(row.original?.filters?.length ?? 0) === 1 ? '' : 's'}
        </Text>
      ),
    },
    {
      header: 'Status',
      id: 'status',
      cell: ({ row }) => (
        <Status
          status={row.original?.enabled === false ? 'neutral' : 'success'}
          variant="badge"
        >
          {row.original?.enabled === false ? 'Disabled' : 'Enabled'}
        </Status>
      ),
    },
  ]

  if (hasError) {
    return (
      <div className="flex flex-col items-start gap-3">
        <Text theme="error">Rules loading failed.</Text>
        <Button variant="secondary" onClick={onRetry}>
          <Icon variant="ArrowClockwiseIcon" />
          Retry loading rules
        </Button>
      </div>
    )
  }

  return (
    <Table
      columns={columns}
      data={data}
      enableSearch={false}
      enableSorting={false}
      isLoading={isLoading}
      emptyStateProps={{
        emptyTitle: 'No active rules yet',
        emptyMessage:
          'Rules will appear after an app config references this trigger.',
      }}
    />
  )
}
