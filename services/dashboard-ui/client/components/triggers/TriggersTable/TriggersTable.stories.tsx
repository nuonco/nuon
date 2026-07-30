export default { title: 'Triggers/Triggers table' }
import { TriggerFilters } from '../TriggerFilters/TriggerFilters'
import { TriggersTable } from './TriggersTable'
const triggers = [
  {
    id: 'evs-1',
    name: 'GitHub',
    status: 'active',
    auth_type: 'hmac' as const,
    envelope: 'none' as const,
    last_event_at: '2026-07-22T12:00:00Z',
  },
  {
    id: 'evs-2',
    name: 'PagerDuty',
    status: 'active',
    auth_type: 'api_key' as const,
    envelope: 'cloudevents' as const,
    last_event_at: '2026-07-24T09:30:00Z',
  },
  {
    id: 'evs-3',
    name: 'SNS alerts',
    status: 'inactive',
    auth_type: 'sns_signature' as const,
    envelope: 'sns' as const,
  },
]
const noop = () => {}
export const Default = () => (
  <TriggersTable data={triggers} isLoading={false} orgId="org-1" />
)
export const WithFilters = () => (
  <TriggersTable
    data={triggers}
    filterActions={
      <TriggerFilters
        trigger=""
        authType=""
        envelope=""
        onSourceChange={noop}
        onAuthTypeChange={noop}
        onEnvelopeChange={noop}
        onClearSource={noop}
        onClearAuthType={noop}
        onClearEnvelope={noop}
      />
    }
    isLoading={false}
    orgId="org-1"
  />
)
export const Empty = () => (
  <TriggersTable data={[]} isLoading={false} orgId="org-1" />
)
export const Loading = () => <TriggersTable data={[]} isLoading orgId="org-1" />
