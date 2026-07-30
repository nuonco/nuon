export default { title: 'Triggers/Trigger overview' }
import { TriggerOverview } from './TriggerOverview'
export const Default = () => (
  <TriggerOverview
    ingressUrl="https://api.example.com/v1/event-ingress/example-key"
    trigger={{
      id: 'evs-1',
      name: 'release-approval-http',
      auth_type: 'hmac',
      envelope: 'none',
      type_from: { header: 'X-Event-Type' },
      id_from: { header: 'X-Event-ID' },
      secrets: [{ id: 'sec-1', key_id: 'key-1' }],
    }}
    onRotateIngressURL={() => undefined}
    onRotateSecret={() => undefined}
    onRevokeSecret={() => undefined}
  />
)

export const NoAuth = () => (
  <TriggerOverview
    ingressUrl="https://api.example.com/v1/event-ingress/example-key"
    trigger={{
      id: 'evs-2',
      name: 'release-approval-http',
      auth_type: 'none',
      envelope: 'none',
      type_from: { payload: '$.type' },
      id_from: { payload: '$.id' },
    }}
    onRotateIngressURL={() => undefined}
  />
)

export const GitHubSetup = () => (
  <TriggerOverview
    ingressUrl="https://api.example.com/v1/event-ingress/example-key"
    trigger={{
      id: 'evs-4',
      name: 'github-events',
      auth_type: 'hmac',
      envelope: 'none',
      preset: 'github',
      secrets: [{ id: 'sec-1', key_id: 'key-1' }],
    }}
    onRevealSecret={() => undefined}
    onHideSecret={() => undefined}
    onRotateIngressURL={() => undefined}
    onRotateSecret={() => undefined}
    onRevokeSecret={() => undefined}
  />
)

export const EventBridgeSetupRevealed = () => (
  <TriggerOverview
    ingressUrl="https://api.example.com/v1/event-ingress/example-key"
    trigger={{
      id: 'evs-5',
      name: 'aws-events',
      auth_type: 'api_key',
      envelope: 'none',
      preset: 'aws-eventbridge',
      secrets: [{ id: 'sec-1', key_id: 'key-1' }],
    }}
    revealedSecrets={{ 'sec-1': { secret: 'example-secret' } }}
    onRevealSecret={() => undefined}
    onHideSecret={() => undefined}
    onRotateIngressURL={() => undefined}
    onRotateSecret={() => undefined}
    onRevokeSecret={() => undefined}
  />
)

export const Forbidden = () => (
  <TriggerOverview
    ingressUrlForbidden
    trigger={{
      id: 'evs-3',
      name: 'release-approval-http',
      auth_type: 'hmac',
      envelope: 'none',
    }}
  />
)
