import { useMemo } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/common/Badge'
import { Table } from '@/components/common/Table'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { DeleteEventSubscriptionButton } from '@/components/datadog/DeleteEventSubscription'
import { EditEventSubscriptionButton } from '@/components/datadog/EditEventSubscription'
import {
  ALL_RESOURCES,
  OUTCOME_LABELS,
  RESOURCE_LABELS,
  type Interests,
} from '@/components/interests'
import { describeMatch } from '@/components/match/types'
import type {
  TDatadogConnection,
  TDatadogEventSubscription,
} from '@/types'

// connectionLabel resolves the parent connection's display name so the
// row reads "Internal monitoring" instead of an opaque `ddc...` ID.
const connectionLabel = (
  sub: TDatadogEventSubscription,
  connections: TDatadogConnection[]
): string => {
  const conn = connections.find((c) => c.id === sub.connection_id)
  return conn?.name || sub.connection_id || '—'
}

export const EventSubscriptionsTable = ({
  data,
  connections,
  isLoading,
}: {
  data: TDatadogEventSubscription[]
  connections: TDatadogConnection[]
  isLoading: boolean
}) => {
  const columns: ColumnDef<TDatadogEventSubscription>[] = useMemo(
    () => [
      {
        header: 'Scope',
        id: 'scope',
        cell: (props) => {
          const sub = props.row.original
          return (
            <div className="flex flex-col gap-1">
              <Text variant="base" weight="strong">
                {connectionLabel(sub, connections)}
              </Text>
              <Text variant="subtext" theme="neutral">
                {describeMatch(sub.match ?? undefined)}
              </Text>
            </div>
          )
        },
      },
      {
        header: 'Interests',
        accessorKey: 'interests',
        cell: (props) => {
          const interests = props.getValue<Interests | undefined>()
          return <InterestsSummary interests={interests} />
        },
      },
      {
        header: 'Tags',
        accessorKey: 'additional_tags',
        cell: (props) => {
          const tags = props.getValue<string[] | undefined>() ?? []
          if (tags.length === 0) {
            return (
              <Text variant="subtext" theme="neutral">
                —
              </Text>
            )
          }
          return (
            <div className="flex flex-wrap gap-1">
              {tags.map((t) => (
                <Badge key={t} theme="neutral">
                  {t}
                </Badge>
              ))}
            </div>
          )
        },
      },
      {
        header: 'Subscribed',
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
          const sub = props.row.original
          const conn = connections.find((c) => c.id === sub.connection_id)
          return (
            <div className="flex justify-end gap-1">
              <EditEventSubscriptionButton
                subscription={sub}
                connection={conn}
                size="sm"
              />
              <DeleteEventSubscriptionButton subscription={sub} size="sm" />
            </div>
          )
        },
      },
    ],
    [connections]
  )

  if (isLoading) return <EventSubscriptionsTableSkeleton />

  return (
    <Table<TDatadogEventSubscription>
      columns={columns}
      data={data}
      enableSearch={false}
      emptyStateProps={{
        emptyTitle: 'No event subscriptions',
        emptyMessage:
          'Subscribe a scope to start streaming Nuon events into a Datadog connection.',
      }}
    />
  )
}

const skeletonColumns: ColumnDef<TDatadogEventSubscription>[] = [
  { header: 'Scope', id: 'scope' },
  { header: 'Interests', accessorKey: 'interests' },
  { header: 'Tags', accessorKey: 'additional_tags' },
  { header: 'Subscribed', accessorKey: 'created_at' },
  { header: '', id: 'action' },
]

export const EventSubscriptionsTableSkeleton = () => (
  <TableSkeleton<TDatadogEventSubscription>
    columns={skeletonColumns}
    skeletonRows={3}
  />
)

// Same shape as the Slack subscription table's InterestsSummary —
// keeping the rendering identical so users see the SAME glance-view of a
// subscription regardless of which integration it routes to.
const InterestsSummary = ({
  interests,
}: {
  interests: Interests | undefined
}) => {
  if (interests?.all_events) {
    return <Badge theme="neutral">All events</Badge>
  }
  const resources = interests?.resources ?? {}
  const enabled = ALL_RESOURCES.filter((kind) =>
    Object.prototype.hasOwnProperty.call(resources, kind)
  )
  if (enabled.length === 0) {
    return <Badge theme="warn">No events</Badge>
  }
  return (
    <div className="flex flex-wrap gap-1">
      {enabled.map((kind) => {
        const cfg = resources[kind] ?? {}
        const outcome = cfg.outcome
        const suffix =
          outcome && outcome !== 'all' ? ` · ${OUTCOME_LABELS[outcome]}` : ''
        return (
          <Badge key={kind} theme="neutral">
            {RESOURCE_LABELS[kind]}
            {suffix}
          </Badge>
        )
      })}
    </div>
  )
}
