import type { TTriggerEvent, TEventTypeFacet } from '@/types'
import { Input } from '@/components/common/form/Input'
import { Select } from '@/components/common/form/Select'
import { EventsTableComponent } from '@/components/triggers/event-ledger/EventsTable'

export const TriggerEvents = ({
  data,
  eventTypes,
  hasError,
  isLoading,
  isLoadingMore,
  onLoadMore,
  onRetry,
  onFilterChange,
  filters,
  orgId,
  triggerId,
}: {
  data: TTriggerEvent[]
  eventTypes: TEventTypeFacet[]
  hasError: boolean
  isLoading: boolean
  isLoadingMore: boolean
  onLoadMore?: () => void
  onRetry: () => void
  onFilterChange: (key: string, value: string) => void
  filters: { q: string; eventType: string; outcome: string; order: string }
  orgId: string
  triggerId: string
}) => (
  <EventsTableComponent
    data={data}
    enableSearch={false}
    receivedOrder={filters.order === 'asc' ? 'asc' : 'desc'}
    onReceivedOrderChange={(order) => onFilterChange('order', order)}
    hasError={hasError}
    hasLoadedData={!isLoading}
    isLoading={isLoading}
    isLoadingMore={isLoadingMore}
    onLoadMore={onLoadMore}
    onRetry={onRetry}
    orgId={orgId}
    eventHref={(event) => `/${orgId}/triggers/${triggerId}/events/${event?.id}`}
    filters={
      <div className="grid gap-3 md:grid-cols-3">
        <Input
          id="event-trigger-event-search"
          type="search"
          labelProps={{ labelText: 'Search by event or external ID' }}
          placeholder="Search by event or external ID"
          value={filters.q}
          onChange={(e) => onFilterChange('q', e.target.value)}
        />
        <Select
          id="event-trigger-event-type"
          labelProps={{ labelText: 'Event type' }}
          value={filters.eventType}
          onChange={(e) => onFilterChange('event_type', e.target.value)}
          options={[
            { value: '', label: 'All event types' },
            ...eventTypes.flatMap((facet) =>
              facet?.event_type
                ? [
                    {
                      value: facet.event_type,
                      label: `${facet.event_type} (${facet?.count ?? 0})`,
                    },
                  ]
                : []
            ),
          ]}
        />
        <Select
          id="event-trigger-event-outcome"
          labelProps={{ labelText: 'Outcome' }}
          value={filters.outcome}
          onChange={(e) => onFilterChange('outcome', e.target.value)}
          options={[
            { value: '', label: 'All outcomes' },
            { value: 'ok', label: 'Ok' },
            { value: 'ignored', label: 'Ignored' },
            { value: 'rejected', label: 'Rejected' },
            { value: 'processing', label: 'Processing' },
            { value: 'failed', label: 'Failed' },
          ]}
        />
      </div>
    }
  />
)
