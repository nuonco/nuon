export default { title: 'Triggers/Trigger events' }
import { TriggerEvents } from './TriggerEvents'
const props = {
  eventTypes: [{ event_type: 'push', count: 3 }],
  filters: { q: '', eventType: '', outcome: '', order: 'desc' },
  hasError: false,
  isLoading: false,
  isLoadingMore: false,
  onFilterChange: () => undefined,
  onRetry: () => undefined,
  orgId: 'org-1',
  triggerId: 'trigger-1',
}
export const Default = () => (
  <TriggerEvents
    {...props}
    data={[
      {
        id: 'event-1',
        trigger_name: 'GitHub',
        event_type: 'push',
        routing_status: 'matched',
        received_at: '2026-07-22T12:00:00Z',
      },
    ]}
  />
)
export const Empty = () => <TriggerEvents {...props} data={[]} />
export const Loading = () => <TriggerEvents {...props} data={[]} isLoading />
