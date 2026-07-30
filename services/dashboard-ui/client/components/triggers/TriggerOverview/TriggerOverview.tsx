import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { ClickToCopy } from '@/components/common/ClickToCopy'
import { Code } from '@/components/common/Code'
import { CodeBlock } from '@/components/common/CodeBlock'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TEventFieldSelector, TTrigger, TTriggerSecret } from '@/types'
import { buildTriggerCurl, buildTriggerSetupSteps } from '@/utils/trigger-utils'

const selector = (value?: TEventFieldSelector) =>
  value?.header
    ? `Header: ${value.header}`
    : value?.payload
      ? `Payload: ${value.payload}`
      : '—'

const usesManagedSecret = (trigger: TTrigger) =>
  trigger?.auth_type === 'hmac' ||
  trigger?.auth_type === 'api_key' ||
  trigger?.auth_type === 'basic'

const setupIntro: Record<string, string> = {
  'aws-eventbridge':
    'Run these commands with the AWS CLI to forward events from EventBridge. Replace the example event pattern with the events you want, and the invocation role ARN with a role that allows EventBridge to invoke API destinations.',
  'aws-sns':
    'Subscribe the ingress URL to your SNS topic as an HTTPS endpoint. Nuon verifies the SNS signature and confirms the subscription automatically.',
  github:
    'Add the ingress URL as a repository or organization webhook, signed with the webhook secret. Replace the owner, repo, and events before running.',
  'slack-events':
    'Set the Events API Request URL in your Slack app to this ingress URL, then subscribe to the bot or workspace events you want Nuon to receive. Nuon verifies each callback with the signing secret you supplied.',
  datadog:
    'Create a Datadog Webhooks integration, then add its recipient to the monitors that should send events to this trigger.',
  'google-pubsub':
    'Create a Pub/Sub push subscription pointing at the ingress URL, authenticated with a Google-signed OIDC token for the configured service account. Replace the topic before running.',
  'azure-event-grid':
    'Create an Event Grid event subscription pointing at the ingress URL. Replace the trigger retrigger ID with the Azure retrigger whose events you want to receive. Nuon authenticates deliveries and confirms the subscription automatically.',
}

const setupNeedsSecret = (preset?: string) =>
  preset === 'aws-eventbridge' ||
  preset === 'github' ||
  preset === 'azure-event-grid' ||
  preset === 'datadog'

const canRotateSecret = (preset?: string) => preset !== 'slack-events'

export type TRevealedSecret = {
  secret?: string
}

export const TriggerOverview = ({
  exampleEventType,
  ingressUrl,
  ingressUrlForbidden,
  onHideSecret,
  onRevealSecret,
  onRevokeSecret,
  onRotateIngressURL,
  onRotateSecret,
  revealError,
  revealPendingSecretId,
  revealedSecrets,
  trigger,
}: {
  exampleEventType?: string
  ingressUrl?: string
  ingressUrlForbidden?: boolean
  onHideSecret?: (secretId: string) => void
  onRevealSecret?: (secretId: string) => void
  onRevokeSecret?: (secretId: string) => void
  onRotateIngressURL?: () => void
  onRotateSecret?: () => void
  revealError?: string
  revealPendingSecretId?: string
  revealedSecrets?: Record<string, TRevealedSecret>
  trigger: TTrigger
}) => {
  const activeSecret = trigger?.secrets?.find((secret) => !secret?.revoked_at)
  const activeSecretValue = activeSecret?.id
    ? revealedSecrets?.[activeSecret.id]?.secret
    : undefined
  const curl = buildTriggerCurl(trigger, ingressUrl, exampleEventType)
  const setupSteps = buildTriggerSetupSteps(trigger, {
    ingressUrl,
    secret: activeSecretValue,
  })
  const latestSecret = trigger?.secrets?.at(0)
  const slackSource = trigger?.preset === 'slack-events'
  const secretRow = (secret: TTriggerSecret) => {
    const revealed = secret?.id
      ? revealedSecrets?.[secret.id]?.secret
      : undefined
    const pending = !!secret?.id && revealPendingSecretId === secret.id
    if (secret?.revoked_at)
      return (
        <Text variant="subtext" theme="neutral">
          {slackSource
            ? 'The signing secret was revoked. Recreate this trigger to reconnect Slack.'
            : 'The secret was revoked. Rotate to create a new one.'}
        </Text>
      )
    return (
      <div className="flex flex-col gap-2">
        <div className="flex items-center justify-between gap-4">
          <Text variant="subtext">
            {secret?.created_at ? (
              <>
                Created <Time time={secret.created_at} format="relative" />
              </>
            ) : (
              'Secret'
            )}
            {secret?.expires_at ? (
              <>
                {' '}
                · Expires <Time time={secret.expires_at} format="relative" />
              </>
            ) : null}
            {secret?.last_used_at ? (
              <>
                {' '}
                · Last used{' '}
                <Time time={secret.last_used_at} format="relative" />
              </>
            ) : null}
          </Text>
          {secret?.id && !slackSource ? (
            <div className="flex items-center gap-2">
              {onRevealSecret ? (
                <Button
                  size="sm"
                  disabled={pending}
                  onClick={() =>
                    revealed
                      ? onHideSecret?.(secret.id!)
                      : onRevealSecret(secret.id!)
                  }
                >
                  {revealed
                    ? 'Hide secret'
                    : pending
                      ? 'Revealing...'
                      : 'Reveal secret'}
                </Button>
              ) : null}
              <Button
                size="sm"
                variant="danger"
                onClick={() => onRevokeSecret?.(secret.id!)}
              >
                Revoke
              </Button>
            </div>
          ) : null}
        </div>
        {revealed ? (
          <LabeledValue label="Secret">
            <ClickToCopy>
              <Code>{revealed}</Code>
            </ClickToCopy>
          </LabeledValue>
        ) : null}
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      {setupSteps.length && !ingressUrlForbidden ? (
        <Card className="!p-4 !gap-4">
          <div className="flex flex-col gap-1">
            <Text variant="base" weight="stronger">
              Finish setup
            </Text>
            <Text variant="subtext" theme="neutral">
              {setupIntro[trigger?.preset || ''] || ''}
            </Text>
          </div>
          {setupNeedsSecret(trigger?.preset) && !activeSecretValue ? (
            <div className="flex items-center justify-between gap-4">
              <Text variant="subtext" theme="neutral">
                Commands use a {'<SECRET>'} placeholder until you reveal the
                secret.
              </Text>
              {onRevealSecret && activeSecret?.id ? (
                <Button
                  size="sm"
                  disabled={revealPendingSecretId === activeSecret.id}
                  onClick={() => onRevealSecret(activeSecret.id!)}
                >
                  {revealPendingSecretId === activeSecret.id
                    ? 'Revealing...'
                    : 'Reveal secret'}
                </Button>
              ) : null}
            </div>
          ) : null}
          {setupSteps.map((step, index) => (
            <div className="flex flex-col gap-2" key={step.title}>
              <Text variant="subtext">
                {setupSteps.length > 1 ? `${index + 1}. ` : ''}
                {step.title}
              </Text>
              <CodeBlock language="bash" showCopy>
                {step.command}
              </CodeBlock>
            </div>
          ))}
        </Card>
      ) : null}

      <Card className="!p-4 !gap-4">
        <div className="flex items-center justify-between gap-4">
          <div className="flex flex-col gap-1">
            <Text variant="base" weight="stronger">
              Send an event
            </Text>
            <Text variant="subtext" theme="neutral">
              Send an HTTP request to this trigger to test its routing rules.
            </Text>
          </div>
          <Button size="sm" onClick={onRotateIngressURL}>
            {ingressUrl ? 'Replace ingress URL' : 'Generate ingress URL'}
          </Button>
        </div>
        {ingressUrlForbidden ? (
          <Text variant="subtext" theme="neutral">
            You need update permissions to view the ingress URL.
          </Text>
        ) : ingressUrl ? (
          <LabeledValue label="Ingress URL">
            <ClickToCopy>
              <Code>{ingressUrl}</Code>
            </ClickToCopy>
          </LabeledValue>
        ) : (
          <Banner theme="warn">
            This trigger predates retrievable ingress URLs. Generate a new URL
            to use the trigger example. Any previous ingress URL will stop
            working.
          </Banner>
        )}
        {trigger?.auth_type === 'none' ? (
          <Banner theme="warn">
            Anyone with this URL can send events to this trigger.
          </Banner>
        ) : null}
        {ingressUrlForbidden ? null : trigger?.envelope === 'sns' ? (
          <Banner theme="info">
            Configure the SNS topic shown in the authentication settings to
            deliver notifications to the ingress URL. SNS signs each request, so
            a manual curl request cannot reproduce a valid notification.
          </Banner>
        ) : curl ? (
          <div className="flex flex-col gap-2">
            <Text variant="subtext" theme="neutral">
              Example request
            </Text>
            <CodeBlock language="bash" showCopy>
              {curl}
            </CodeBlock>
          </div>
        ) : null}
      </Card>

      <div className="grid gap-6 md:grid-cols-2">
        <LabeledValue label="Auth type">
          <Code variant="inline">{trigger?.auth_type || '—'}</Code>
        </LabeledValue>
        <LabeledValue label="Envelope">
          <Code variant="inline">{trigger?.envelope || '—'}</Code>
        </LabeledValue>
        <LabeledValue label="Event type selector">
          <Text variant="subtext">{selector(trigger?.type_from)}</Text>
        </LabeledValue>
        <LabeledValue label="External ID selector">
          <Text variant="subtext">{selector(trigger?.id_from)}</Text>
        </LabeledValue>
        <LabeledValue label="Last event">
          {trigger?.last_event_at ? (
            <Time time={trigger.last_event_at} format="relative" shouldTick />
          ) : (
            <Text variant="subtext" theme="neutral">
              No events yet
            </Text>
          )}
        </LabeledValue>
      </div>

      {usesManagedSecret(trigger) ? (
        <Card className="!p-4 !gap-4">
          <div className="flex items-center justify-between gap-4">
            <div className="flex flex-col gap-1">
              <Text variant="base" weight="stronger">
                Credentials
              </Text>
              <Text variant="subtext" theme="neutral">
                {slackSource
                  ? 'Slack signing secrets are write-only. Recreate this trigger to replace the secret.'
                  : 'Revealing a secret value requires update permissions.'}
              </Text>
            </div>
            {canRotateSecret(trigger?.preset) ? (
              <Button size="sm" onClick={onRotateSecret}>
                Rotate secret
              </Button>
            ) : null}
          </div>
          {revealError ? <Banner theme="error">{revealError}</Banner> : null}
          {latestSecret ? (
            secretRow(latestSecret)
          ) : (
            <Text variant="subtext" theme="neutral">
              No secret configured
            </Text>
          )}
        </Card>
      ) : null}
    </div>
  )
}
