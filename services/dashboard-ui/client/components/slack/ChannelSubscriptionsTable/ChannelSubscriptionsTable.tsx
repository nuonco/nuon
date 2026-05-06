import { useMemo } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/common/Badge'
import { Code } from '@/components/common/Code'
import { Table } from '@/components/common/Table'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import {
  ALL_RESOURCES,
  OUTCOME_LABELS,
  RESOURCE_LABELS,
  type Interests,
} from '@/components/interests'
import { describeMatch } from '@/components/match/types'
import { DeleteChannelSubscriptionButton } from '@/components/slack/DeleteChannelSubscription'
import { EditChannelSubscriptionButton } from '@/components/slack/EditChannelSubscription'
import type { TSlackChannelSubscription, TSlackOrgLink } from '@/types'

const teamLabel = (
  sub: TSlackChannelSubscription,
  links: TSlackOrgLink[]
): string => {
  if (!sub.team_id) return '—'
  const _ = links // included so callers passing links can later swap to a name lookup
  return sub.team_id
}

export const ChannelSubscriptionsTable = ({
  data,
  links,
  isLoading,
}: {
  data: TSlackChannelSubscription[]
  links: TSlackOrgLink[]
  isLoading: boolean
}) => {
  const columns: ColumnDef<TSlackChannelSubscription>[] = useMemo(
    () => [
      {
        header: 'Channel',
        accessorKey: 'channel_name',
        cell: (props) => {
          const name = props.getValue<string | undefined>()
          const id = props.row.original.channel_id
          // Scope subtitle mirrors the Slack subscribe modal's
          // describeMatch vocabulary so the dashboard and slash-command
          // surfaces describe the same row identically.
          const scope = describeMatch(props.row.original.match)
          return (
            <div className="flex flex-col gap-1">
              <Text variant="base" weight="strong">
                {name ? `#${name}` : id || '—'}
              </Text>
              {id ? (
                <Code variant="inline" className="!px-2 !py-0.5 w-fit">
                  {id}
                </Code>
              ) : null}
              <Text variant="subtext" theme="neutral">
                {scope}
              </Text>
            </div>
          )
        },
      },
      {
        header: 'Workspace',
        id: 'workspace',
        cell: (props) => (
          <Text variant="subtext" theme="neutral">
            {teamLabel(props.row.original, links)}
          </Text>
        ),
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
        cell: (props) => (
          <div className="flex justify-end gap-1">
            <EditChannelSubscriptionButton
              subscription={props.row.original}
              size="sm"
            />
            <DeleteChannelSubscriptionButton
              subscription={props.row.original}
              size="sm"
            />
          </div>
        ),
      },
    ],
    [links]
  )

  if (isLoading) return <ChannelSubscriptionsTableSkeleton />

  return (
    <Table<TSlackChannelSubscription>
      columns={columns}
      data={data}
      enableSearch={false}
      emptyStateProps={{
        emptyTitle: 'No channel subscriptions',
        emptyMessage:
          'Subscribe a Slack channel to start receiving lifecycle events for this org.',
      }}
    />
  )
}

const skeletonColumns: ColumnDef<TSlackChannelSubscription>[] = [
  { header: 'Channel', accessorKey: 'channel_name' },
  { header: 'Workspace', id: 'workspace' },
  { header: 'Interests', accessorKey: 'interests' },
  { header: 'Subscribed', accessorKey: 'created_at' },
  { header: '', id: 'action' },
]

export const ChannelSubscriptionsTableSkeleton = () => (
  <TableSkeleton<TSlackChannelSubscription>
    columns={skeletonColumns}
    skeletonRows={3}
  />
)

// Compact summary of an Interests config for the table cell. Mirrors the
// picker semantics AND the backend matcher (interests.Matches):
//   - all_events=true                   → "All events"
//   - empty / missing / no resources    → "No events" (matcher returns false
//     for every event in this case — the row was created with explicit
//     opt-out, or the JSONB column was persisted as NULL because the
//     Interests struct was IsZero).
//   - otherwise                         → one badge per enabled resource
//     with the configured outcome appended (omitted when it's the implicit
//     "all activity").
//
// IMPORTANT: do NOT treat `interests === undefined` as "all events". The
// backend serializes a zero-valued Interests with `omitzero`, so an empty
// subscription comes back over the wire with no `interests` key — falsely
// showing "All events" here misleads users into thinking they're still
// subscribed to everything.
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
