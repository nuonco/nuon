export default { title: 'Triggers/Event details' }

import { EventDetails } from './EventDetails'

const event = {
  id: 'event-1',
  trigger_id: 'trigger-1',
  trigger_name: 'GitHub',
  external_id: 'delivery-1',
  event_type: 'push',
  received_at: '2026-07-22T12:00:00Z',
  routing_status: 'matched',
  match_count: 1,
  dispatch_count: 1,
  payload: { ref: 'refs/heads/main', repository: { full_name: 'nuonco/nuon' } },
  headers: { 'X-GitHub-Event': ['push'] },
  match_explanations: [
    {
      rule_id: 'rule-1',
      rule_name: 'deploy-main',
      app_id: 'app-1',
      event_type_matched: true,
      matched: true,
      filters: [
        {
          from: 'payload',
          path: '$.ref',
          op: 'eq',
          expected: 'refs/heads/main',
          selected: ['refs/heads/main'],
          matched: true,
        },
      ],
    },
  ],
  dispatches: [
    {
      id: 'dispatch-1',
      trigger_rule_id: 'rule-1',
      target_type: 'app_branch_run',
      status: 'triggered',
      attempts: 1,
      app_id: 'app-1',
      target_id: 'branch-1',
      result_resource_type: 'app_branch_run',
      result_resource_id: 'run-1',
      workflow_id: 'workflow-1',
    },
  ],
}

export const Default = () => (
  <EventDetails
    event={event}
    orgId="org-1"
    isRawLoading={false}
    isReplaying={false}
    onRevealRaw={() => undefined}
    onReplay={() => undefined}
    onRetryDispatch={() => undefined}
  />
)
export const Rejected = () => (
  <EventDetails
    event={{
      ...event,
      routing_status: 'rejected',
      routing_error: 'authentication failed with HTTP 401',
      match_explanations: [],
    }}
    orgId="org-1"
    isRawLoading={false}
    isReplaying={false}
    onRevealRaw={() => undefined}
    onReplay={() => undefined}
    onRetryDispatch={() => undefined}
  />
)
export const LoadFailed = () => (
  <EventDetails
    orgId="org-1"
    isRawLoading={false}
    isReplaying={false}
    onRevealRaw={() => undefined}
    onReplay={() => undefined}
    onRetry={() => undefined}
    onRetryDispatch={() => undefined}
  />
)
export const RawRequestFailed = () => (
  <EventDetails
    event={event}
    orgId="org-1"
    hasRawError
    isRawLoading={false}
    isReplaying={false}
    onRevealRaw={() => undefined}
    onReplay={() => undefined}
    onRetryDispatch={() => undefined}
  />
)
export const RefreshFailed = () => (
  <EventDetails
    event={event}
    orgId="org-1"
    hasError
    isRawLoading={false}
    isReplaying={false}
    onRevealRaw={() => undefined}
    onReplay={() => undefined}
    onRetry={() => undefined}
    onRetryDispatch={() => undefined}
  />
)
