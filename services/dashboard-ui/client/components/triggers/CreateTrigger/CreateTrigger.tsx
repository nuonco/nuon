import { useForm, useStore } from '@tanstack/react-form'
import { Badge } from '@/components/common/Badge'
import { Text } from '@/components/common/Text'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormInput } from '@/components/common/form/FormInput'
import { FormRadioGroup } from '@/components/common/form/FormRadioGroup'
import { FormSelect } from '@/components/common/form/FormSelect'
import { FormTextarea } from '@/components/common/form/FormTextarea'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type {
  TAPIError,
  TCreateTriggerBody,
  TTriggerAuthConfig,
} from '@/types'
import {
  createTriggerSchema,
  getTriggerPreset,
  type CreateTriggerValues,
  type TProvider,
} from './schema'

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
  const form = useForm({
    defaultValues: {
      name: '',
      description: '',
      provider: 'github',
      awsMethod: 'aws-eventbridge',
      authType: 'hmac',
      envelope: 'none',
      issuer: 'https://accounts.google.com',
      audience: '',
      identity: '',
      topicArn: '',
      header: '',
      username: '',
      slackSigningSecret: '',
    } as CreateTriggerValues,
    validators: {
      onMount: createTriggerSchema,
      onChange: createTriggerSchema,
    },
    onSubmit: ({ value }) => {
      const preset = getTriggerPreset(value.provider, value.awsMethod)
      const custom = value.provider === 'custom'

      const auth_config: TTriggerAuthConfig = {}
      if (preset === 'google-pubsub' || (custom && value.authType === 'bearer_jwt')) {
        auth_config.issuer = value.issuer
        auth_config.audience = value.audience
          .split(',')
          .map((v) => v.trim())
          .filter(Boolean)
        auth_config.expected_email = value.identity
      }
      if (preset === 'aws-sns' || (custom && value.authType === 'sns_signature')) {
        auth_config.topic_arn = value.topicArn.trim()
      }
      if (custom && (value.authType === 'hmac' || value.authType === 'api_key')) {
        auth_config.header = value.header.trim()
      }
      if (custom && value.authType === 'basic') {
        auth_config.username = value.username.trim()
      }

      onSubmit({
        name: value.name.trim(),
        description: value.description.trim(),
        ...(custom
          ? { auth_type: value.authType, envelope: value.envelope }
          : { preset }),
        ...(preset === 'slack-events'
          ? { secret: value.slackSigningSecret.trim() }
          : {}),
        ...(Object.keys(auth_config).length ? { auth_config } : {}),
      })
    },
  })

  const canSubmit = useStore(form.store, (s) => s.canSubmit)
  const values = useStore(form.store, (s) => s.values)
  const preset = getTriggerPreset(values.provider, values.awsMethod)
  const custom = values.provider === 'custom'

  return (
    <Modal
      heading="Create trigger"
      size="lg"
      primaryActionTrigger={{
        children: isPending ? 'Creating trigger' : 'Create trigger',
        disabled: !canSubmit || isPending,
        onClick: () => form.handleSubmit(),
        variant: 'primary',
      }}
      {...props}
    >
      <form
        autoComplete="off"
        noValidate
        onSubmit={(e) => e.preventDefault()}
        className="flex flex-col gap-4"
      >
        <FormErrorBanner error={error} fallback="Unable to create trigger" />

        <form.Field name="name">
          {(field) => (
            <FormInput
              field={field}
              id="event-trigger-name"
              disabled={isPending}
              labelProps={{ labelText: 'Name' }}
            />
          )}
        </form.Field>

        <form.Field name="description">
          {(field) => (
            <FormTextarea
              field={field}
              id="event-trigger-description"
              disabled={isPending}
              labelProps={{ labelText: 'Description (optional)' }}
            />
          )}
        </form.Field>

        <form.Field name="provider">
          {(field) => (
            <FormRadioGroup
              field={field}
              label="Where do events come from?"
              disabled={isPending}
              options={providerOptions.map((option) => ({
                value: option.value,
                label: (
                  <RadioLabel
                    title={option.title}
                    description={option.description}
                  />
                ),
              }))}
            />
          )}
        </form.Field>

        {values.provider === 'aws' ? (
          <form.Field name="awsMethod">
            {(field) => (
              <FormRadioGroup
                field={field}
                label="How will AWS send events?"
                disabled={isPending}
                options={[
                  {
                    value: 'aws-eventbridge',
                    label: (
                      <RadioLabel
                        title="EventBridge"
                        badge="Recommended"
                        description="Best for events from AWS services such as S3, ECR, CodePipeline, and CloudWatch alarms. An EventBridge rule forwards events here with an API key."
                      />
                    ),
                  },
                  {
                    value: 'aws-sns',
                    label: (
                      <RadioLabel
                        title="SNS topic"
                        description="Use only if your events already publish to an SNS topic. Nuon verifies the SNS signature and confirms the subscription automatically."
                      />
                    ),
                  },
                ]}
              />
            )}
          </form.Field>
        ) : null}

        {preset === 'slack-events' ? (
          <form.Field name="slackSigningSecret">
            {(field) => (
              <FormInput
                field={field}
                id="event-trigger-slack-signing-secret"
                type="password"
                disabled={isPending}
                labelProps={{ labelText: 'Slack signing secret' }}
                helperText="Find this under Basic Information → App Credentials in your Slack app settings."
              />
            )}
          </form.Field>
        ) : null}

        {preset === 'google-pubsub' || (custom && values.authType === 'bearer_jwt') ? (
          <>
            <form.Field name="issuer">
              {(field) => (
                <FormInput
                  field={field}
                  id="event-trigger-issuer"
                  disabled={isPending}
                  labelProps={{ labelText: 'Issuer' }}
                />
              )}
            </form.Field>
            <form.Field name="audience">
              {(field) => (
                <FormInput
                  field={field}
                  id="event-trigger-audience"
                  disabled={isPending}
                  labelProps={{ labelText: 'Audience' }}
                  helperText="Separate multiple audiences with commas."
                />
              )}
            </form.Field>
            <form.Field name="identity">
              {(field) => (
                <FormInput
                  field={field}
                  id="event-trigger-identity"
                  disabled={isPending}
                  labelProps={{ labelText: 'Expected identity' }}
                  helperText="Service account email expected in the token."
                />
              )}
            </form.Field>
          </>
        ) : null}

        {preset === 'aws-sns' || (custom && values.authType === 'sns_signature') ? (
          <form.Field name="topicArn">
            {(field) => (
              <FormInput
                field={field}
                id="event-trigger-topic-arn"
                disabled={isPending}
                labelProps={{ labelText: 'SNS topic ARN' }}
              />
            )}
          </form.Field>
        ) : null}

        {custom ? (
          <>
            <form.Field name="authType">
              {(field) => (
                <FormSelect
                  field={field}
                  id="event-trigger-auth-type"
                  disabled={isPending}
                  labelProps={{ labelText: 'Auth type' }}
                  options={[
                    ['none', 'None'],
                    ['hmac', 'HMAC'],
                    ['api_key', 'API key'],
                    ['basic', 'Basic'],
                    ['bearer_jwt', 'Bearer JWT'],
                    ['sns_signature', 'SNS signature'],
                  ].map(([value, label]) => ({ value, label }))}
                />
              )}
            </form.Field>
            <form.Field name="envelope">
              {(field) => (
                <FormSelect
                  field={field}
                  id="event-trigger-envelope"
                  disabled={isPending}
                  labelProps={{ labelText: 'Envelope' }}
                  options={[
                    ['none', 'None'],
                    ['cloudevents', 'CloudEvents'],
                    ['pubsub_push', 'Pub/Sub push'],
                    ['sns', 'SNS'],
                  ].map(([value, label]) => ({ value, label }))}
                />
              )}
            </form.Field>
            {values.authType === 'hmac' || values.authType === 'api_key' ? (
              <form.Field name="header">
                {(field) => (
                  <FormInput
                    field={field}
                    id="event-trigger-auth-header"
                    disabled={isPending}
                    labelProps={{ labelText: 'Auth header (optional)' }}
                  />
                )}
              </form.Field>
            ) : null}
            {values.authType === 'basic' ? (
              <form.Field name="username">
                {(field) => (
                  <FormInput
                    field={field}
                    id="event-trigger-username"
                    disabled={isPending}
                    labelProps={{ labelText: 'Username (optional)' }}
                  />
                )}
              </form.Field>
            ) : null}
          </>
        ) : null}
      </form>
    </Modal>
  )
}
