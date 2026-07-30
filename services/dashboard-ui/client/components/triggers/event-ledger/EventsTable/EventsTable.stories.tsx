export default { title: 'Triggers/Events table' }

import { EventsTable } from './EventsTable'

const events = [
  {
    id: 'event-1',
    trigger_id: 'trigger-1',
    trigger_name: 'GitHub',
    event_type: 'push',
    routing_status: 'matched',
    match_count: 2,
    received_at: '2026-07-22T12:00:00Z',
  },
  {
    id: 'event-2',
    trigger_id: 'trigger-2',
    trigger_name: 'Pub/Sub',
    event_type: 'image.pushed',
    routing_status: 'ignored',
    match_count: 0,
    received_at: '2026-07-22T11:00:00Z',
  },
  {
    id: 'event-3',
    trigger_id: 'trigger-3',
    trigger_name: 'Custom',
    routing_status: 'rejected',
    match_count: 0,
    received_at: '2026-07-22T10:00:00Z',
  },
]

export const Default = () => (
  <EventsTable data={events} isLoading={false} orgId="org-1" />
)
export const Empty = () => (
  <EventsTable data={[]} isLoading={false} orgId="org-1" />
)
export const Loading = () => <EventsTable data={[]} isLoading orgId="org-1" />
export const RefreshFailed = () => (
  <EventsTable
    data={events}
    hasError
    hasLoadedData
    isLoading={false}
    onRetry={() => undefined}
    orgId="org-1"
  />
)
export const InitialLoadFailed = () => (
  <EventsTable
    data={[]}
    hasError
    isLoading={false}
    onRetry={() => undefined}
    orgId="org-1"
  />
)
