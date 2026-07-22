import { useState } from 'react'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Input } from '@/components/common/form/Input'
import { RadioInput } from '@/components/common/form/RadioInput'
import { Select } from '@/components/common/form/Select'
import { Textarea } from '@/components/common/form/Textarea'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type {
  TAPIError,
  TCreateTriggerBody,
  TTriggerAuthType,
  TTriggerEnvelope,
} from '@/types'

type TProvider =
  | 'github'
  | 'slack'
  | 'datadog'
  | 'aws'
  | 'gcp'
  | 'azure'
  | 'custom'
type TAwsMethod = 'aws-eventbridge' | 'aws-sns'

const providerOptions: {
  value: TProvider
  title: string
  description: string
}[] = [
  {
    value: 'github',
    title: 'GitHub',
    description:
      'Repository or organization webhooks, verified with the webhook secret.',
  },
  {
    value: 'slack',
    title: 'Slack',
    description:
      'Events API callbacks, verified with your Slack app signing secret.',
  },
  {
    value: 'datadog',
    title: 'Datadog',
    description:
      'Monitor and event notifications from a Datadog Webhooks integration.',
  },
  {
    value: 'aws',
    title: 'AWS',
    description:
      'Events from AWS services via EventBridge, or an existing SNS topic.',
  },
  {
    value: 'gcp',
    title: 'Google Cloud',
    description:
      'Pub/Sub push subscriptions, verified with a Google-signed OIDC token.',
  },
  {
    value: 'azure',
    title: 'Azure',
    description:
      'Events from Azure services via an Event Grid webhook with API-key authentication.',
  },
  {
    value: 'custom',
    title: 'Custom',
    description: 'Any other webhook. Configure auth and envelope manually.',
  },
]

const RadioLabel = ({
  title,
  description,
  badge,
}: {
  title: string
  description: string
  badge?: string
}) => (
  <span className="flex flex-col gap-1">
    <span className="flex items-center gap-2">
      <Text variant="body">{title}</Text>
      {badge ? <Badge theme="info">{badge}</Badge> : null}
    </span>
    <Text variant="subtext" theme="neutral">
      {description}
    </Text>
  </span>
)

export const CreateTriggerModal = ({
  error,
  isPending,
  onSubmit,
  ...props
}: {
  error: TAPIError | null
  isPending: boolean
  onSubmit: (body: TCreateTriggerBody) => void
} & Omit<IModal, 'onSubmit'>) => {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [provider, setProvider] = useState<TProvider>('github')
  const [awsMethod, setAwsMethod] = useState<TAwsMethod>('aws-eventbridge')
  const [authType, setAuthType] = useState<TTriggerAuthType>('hmac')
  const [envelope, setEnvelope] = useState<TTriggerEnvelope>('none')
  const [issuer, setIssuer] = useState('https://accounts.google.com')
  const [audience, setAudience] = useState('')
  const [identity, setIdentity] = useState('')
  const [topicArn, setTopicArn] = useState('')
  const [header, setHeader] = useState('')
  const [username, setUsername] = useState('')
  const [slackSigningSecret, setSlackSigningSecret] = useState('')

  const preset =
    provider === 'aws'
      ? awsMethod
      : provider === 'gcp'
        ? 'google-pubsub'
        : provider === 'azure'
          ? 'azure-event-grid'
          : provider === 'slack'
            ? 'slack-events'
            : provider

  const custom = provider === 'custom'
  const submit = () => {
    const auth_config: Record<string, string | string[]> = {}
    if (preset === 'google-pubsub' || (custom && authType === 'bearer_jwt'))
      Object.assign(auth_config, {
        issuer,
        audience: audience
          .split(',')
          .map((v) => v.trim())
          .filter(Boolean),
        expected_email: identity,
      })
    if (preset === 'aws-sns' || (custom && authType === 'sns_signature'))
      auth_config.topic_arn = topicArn.trim()
    if (custom && (authType === 'hmac' || authType === 'api_key'))
      auth_config.header = header.trim()
    if (custom && authType === 'basic') auth_config.username = username.trim()
    onSubmit({
      name: name.trim(),
      description: description.trim(),
      ...(custom ? { auth_type: authType, envelope } : { preset }),
      ...(preset === 'slack-events'
        ? { secret: slackSigningSecret.trim() }
        : {}),
      ...(Object.keys(auth_config).length ? { auth_config } : {}),
    })
  }
  const valid =
    !!name.trim() &&
    ((preset !== 'google-pubsub' && !(custom && authType === 'bearer_jwt')) ||
      (!!issuer.trim() && !!audience.trim() && !!identity.trim())) &&
    ((preset !== 'aws-sns' && !(custom && authType === 'sns_signature')) ||
      !!topicArn.trim()) &&
    (preset !== 'slack-events' || !!slackSigningSecret.trim())
  return (
    <Modal
      heading="Create trigger"
      size="lg"
      primaryActionTrigger={{
        children: isPending ? 'Creating...' : 'Create trigger',
        disabled: !valid || isPending,
        onClick: submit,
        variant: 'primary',
      }}
      {...props}
    >
      {error ? (
        <Banner theme="error">
          {error?.error || 'Unable to create trigger.'}
        </Banner>
      ) : null}
      <Input
        id="event-trigger-name"
        labelProps={{ labelText: 'Name' }}
        value={name}
        onChange={(event) => setName(event.target.value)}
        required
      />
      <Textarea
        id="event-trigger-description"
        labelProps={{ labelText: 'Description (optional)' }}
        value={description}
        onChange={(event) => setDescription(event.target.value)}
      />
      <div className="flex flex-col gap-1">
        <Text variant="subtext">Where do events come from?</Text>
        {providerOptions.map((option) => (
          <RadioInput
            key={option.value}
            name="event-trigger-provider"
            value={option.value}
            checked={provider === option.value}
            onChange={() => setProvider(option.value)}
            labelProps={{
              labelText: (
                <RadioLabel
                  title={option.title}
                  description={option.description}
                />
              ),
            }}
          />
        ))}
      </div>
      {provider === 'aws' ? (
        <div className="flex flex-col gap-1">
          <Text variant="subtext">How will AWS send events?</Text>
          <RadioInput
            name="event-trigger-aws-method"
            value="aws-eventbridge"
            checked={awsMethod === 'aws-eventbridge'}
            onChange={() => setAwsMethod('aws-eventbridge')}
            labelProps={{
              labelText: (
                <RadioLabel
                  title="EventBridge"
                  badge="Recommended"
                  description="Best for events from AWS services such as S3, ECR, CodePipeline, and CloudWatch alarms. An EventBridge rule forwards events here with an API key."
                />
              ),
            }}
          />
          <RadioInput
            name="event-trigger-aws-method"
            value="aws-sns"
            checked={awsMethod === 'aws-sns'}
            onChange={() => setAwsMethod('aws-sns')}
            labelProps={{
              labelText: (
                <RadioLabel
                  title="SNS topic"
                  description="Use only if your events already publish to an SNS topic. Nuon verifies the SNS signature and confirms the subscription automatically."
                />
              ),
            }}
          />
        </div>
      ) : null}
      {preset === 'slack-events' ? (
        <Input
          id="event-trigger-slack-signing-secret"
          labelProps={{ labelText: 'Slack signing secret' }}
          helperText="Find this under Basic Information → App Credentials in your Slack app settings."
          type="password"
          value={slackSigningSecret}
          onChange={(e) => setSlackSigningSecret(e.target.value)}
        />
      ) : null}
      {preset === 'google-pubsub' || (custom && authType === 'bearer_jwt') ? (
        <>
          <Input
            id="event-trigger-issuer"
            labelProps={{ labelText: 'Issuer' }}
            value={issuer}
            onChange={(e) => setIssuer(e.target.value)}
          />
          <Input
            id="event-trigger-audience"
            labelProps={{ labelText: 'Audience' }}
            helperText="Separate multiple audiences with commas."
            value={audience}
            onChange={(e) => setAudience(e.target.value)}
          />
          <Input
            id="event-trigger-identity"
            labelProps={{ labelText: 'Expected identity' }}
            helperText="Service account email expected in the token."
            value={identity}
            onChange={(e) => setIdentity(e.target.value)}
          />
        </>
      ) : null}
      {preset === 'aws-sns' || (custom && authType === 'sns_signature') ? (
        <Input
          id="event-trigger-topic-arn"
          labelProps={{ labelText: 'SNS topic ARN' }}
          value={topicArn}
          onChange={(e) => setTopicArn(e.target.value)}
        />
      ) : null}
      {custom ? (
        <>
          <Select
            id="event-trigger-auth-type"
            labelProps={{ labelText: 'Auth type' }}
            value={authType}
            onChange={(e) => setAuthType(e.target.value as TTriggerAuthType)}
            options={[
              ['none', 'None'],
              ['hmac', 'HMAC'],
              ['api_key', 'API key'],
              ['basic', 'Basic'],
              ['bearer_jwt', 'Bearer JWT'],
              ['sns_signature', 'SNS signature'],
            ].map(([value, label]) => ({ value, label }))}
          />
          <Select
            id="event-trigger-envelope"
            labelProps={{ labelText: 'Envelope' }}
            value={envelope}
            onChange={(e) => setEnvelope(e.target.value as TTriggerEnvelope)}
            options={[
              ['none', 'None'],
              ['cloudevents', 'CloudEvents'],
              ['pubsub_push', 'Pub/Sub push'],
              ['sns', 'SNS'],
            ].map(([value, label]) => ({ value, label }))}
          />
          {authType === 'hmac' || authType === 'api_key' ? (
            <Input
              id="event-trigger-auth-header"
              labelProps={{ labelText: 'Auth header (optional)' }}
              value={header}
              onChange={(e) => setHeader(e.target.value)}
            />
          ) : null}
          {authType === 'basic' ? (
            <Input
              id="event-trigger-username"
              labelProps={{ labelText: 'Username (optional)' }}
              value={username}
              onChange={(e) => setUsername(e.target.value)}
            />
          ) : null}
        </>
      ) : null}
    </Modal>
  )
}
