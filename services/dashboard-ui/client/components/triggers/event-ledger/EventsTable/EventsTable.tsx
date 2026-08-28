import { useMemo, type ReactNode } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TTriggerEvent } from '@/types'
import { eventOutcome } from '../events'

export const EventsTable = ({
  data,
  filters,
  hasError = false,
  hasLoadedData = false,
  isLoading,
  isLoadingMore = false,
  isRetrying = false,
  onRetry,
  onLoadMore,
  orgId,
  eventHref,
  enableSearch = true,
  receivedOrder,
  onReceivedOrderChange,
}: {
  data: TTriggerEvent[]
  filters?: ReactNode
  hasError?: boolean
  hasLoadedData?: boolean
  isLoading: boolean
  isLoadingMore?: boolean
  isRetrying?: boolean
  onRetry?: () => void
  onLoadMore?: () => void
  orgId: string
  eventHref?: (event: TTriggerEvent) => string
  enableSearch?: boolean
  receivedOrder?: 'asc' | 'desc'
  onReceivedOrderChange?: (order: 'asc' | 'desc') => void
}) => {
  const columns: ColumnDef<TTriggerEvent>[] = useMemo(
    () => [
      {
        header: 'Trigger',
        accessorFn: (event) =>
          event?.trigger_name || event?.trigger_id || 'Unknown',
        cell: ({ row, getValue }) => (
          <div className="flex flex-col gap-1">
            <Link
              href={
                eventHref?.(row.original) ??
                `/${orgId}/settings/triggers/${row.original?.trigger_id}/events/${row.original?.id}`
              }
            >
              {getValue<string>()}
            </Link>
            <Text variant="subtext" theme="neutral" family="mono">
              {row.original?.trigger_id || '—'}
            </Text>
          </div>
        ),
      },
      {
        header: 'Type',
        accessorKey: 'event_type',
        cell: ({ getValue }) => (
          <Text variant="subtext" family="mono">
            {getValue<string>() || '—'}
          </Text>
        ),
      },
      {
        header: 'Outcome',
        accessorFn: eventOutcome,
        cell: ({ getValue }) => {
          const outcome = getValue<string>()
          const status =
            outcome === 'ok'
              ? 'success'
              : outcome === 'rejected' || outcome === 'failed'
                ? 'error'
                : outcome === 'processing'
                  ? 'info'
                  : 'neutral'
          return (
            <Status status={status} variant="badge">
              {outcome}
            </Status>
          )
        },
      },
      {
        header: 'Matches',
        accessorKey: 'match_count',
        cell: ({ row }) => (
          <div className="flex flex-col gap-1">
            <Text variant="subtext">
              {row.original?.match_count ?? 0} rule
              {(row.original?.match_count ?? 0) === 1 ? '' : 's'}
            </Text>
            <Text variant="subtext" theme="neutral">
              {row.original?.waiter_match_count ?? 0} waiting step
              {(row.original?.waiter_match_count ?? 0) === 1 ? '' : 's'}
            </Text>
          </div>
        ),
      },
      {
        header:
          receivedOrder && onReceivedOrderChange
            ? () => (
                <button
                  type="button"
                  className="flex cursor-pointer items-center gap-1"
                  aria-label={`Received, ${receivedOrder === 'desc' ? 'newest first' : 'oldest first'}. Sort ${receivedOrder === 'desc' ? 'oldest first' : 'newest first'}`}
                  onClick={() =>
                    onReceivedOrderChange(
                      receivedOrder === 'desc' ? 'asc' : 'desc'
                    )
                  }
                >
                  Received
                  <Icon
                    variant={
                      receivedOrder === 'desc' ? 'ArrowDownIcon' : 'ArrowUpIcon'
                    }
                  />
                </button>
              )
            : 'Received',
        accessorKey: 'received_at',
        cell: ({ getValue }) => {
          const receivedAt = getValue<string | undefined>()
          return receivedAt ? (
            <Time
              variant="subtext"
              time={receivedAt}
              format="relative"
              shouldTick
            />
          ) : (
            <Text variant="subtext" theme="neutral">
              —
            </Text>
          )
        },
      },
    ],
    [eventHref, onReceivedOrderChange, orgId, receivedOrder]
  )

  if (hasError && !hasLoadedData) {
    return (
      <div className="flex flex-col items-start gap-3">
        <Text theme="error">Events loading failed.</Text>
        <Button variant="secondary" disabled={isRetrying} onClick={onRetry}>
          <Icon variant={isRetrying ? 'Loading' : 'ArrowClockwiseIcon'} />
          {isRetrying ? 'Retrying events' : 'Retry loading events'}
        </Button>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-4">
      {hasError ? (
        <Banner theme="warn">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div className="flex flex-col gap-1">
              <Text weight="strong">Event refresh failed</Text>
              <Text variant="subtext">
                Showing the most recently loaded events.
              </Text>
            </div>
            <Button variant="secondary" disabled={isRetrying} onClick={onRetry}>
              <Icon variant={isRetrying ? 'Loading' : 'ArrowClockwiseIcon'} />
              {isRetrying ? 'Refreshing events' : 'Refresh events'}
            </Button>
          </div>
        </Banner>
      ) : null}
      {filters}
      <Table
        columns={columns}
        data={data}
        isLoading={isLoading}
        enableSorting={false}
        enableSearch={enableSearch}
        searchPlaceholder="Search events"
        emptyStateProps={{
          emptyTitle: 'No events yet',
          emptyMessage:
            'Events will appear here after a trigger receives a request.',
        }}
      />
      {onLoadMore ? (
        <div className="flex justify-center">
          <Button
            variant="secondary"
            disabled={isLoadingMore}
            onClick={onLoadMore}
          >
            {isLoadingMore ? 'Loading more events' : 'Load more events'}
          </Button>
        </div>
      ) : null}
    </div>
  )
}
