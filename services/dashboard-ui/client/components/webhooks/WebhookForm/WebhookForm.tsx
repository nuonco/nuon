import { useForm, useStore } from '@tanstack/react-form'
import { Code } from '@/components/common/Code'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormInput } from '@/components/common/form/FormInput'
import { FormRadioGroup } from '@/components/common/form/FormRadioGroup'
import { Label } from '@/components/common/form/Label'
import { allEvents } from '@/components/interests'
import { FormInterestsPicker } from '@/components/interests/FormInterestsPicker'
import { FormMatchPicker } from '@/components/match/FormMatchPicker'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError, TWebhook } from '@/types'
import {
  buildWebhookSchema,
  type WebhookFormMode,
  type WebhookFormOutput,
  type WebhookFormValues,
} from './schema'

const computeSecret = (
  mode: WebhookFormMode,
  value: WebhookFormValues
): string | undefined => {
  if (mode === 'create') return value.webhookSecret.trim()
  if (value.secretMode === 'keep') return undefined
  if (value.secretMode === 'clear') return ''
  return value.webhookSecret.trim()
}

export const WebhookFormModal = ({
  mode,
  webhook,
  isPending,
  error,
  onSubmit,
  ...props
}: {
  mode: WebhookFormMode
  webhook?: TWebhook
  isPending: boolean
  error: TAPIError | null
  onSubmit: (output: WebhookFormOutput) => void
} & Omit<IModal, 'onSubmit'>) => {
  const schema = buildWebhookSchema(mode)
  const hasSecret = webhook?.has_secret ?? false

  const form = useForm({
    defaultValues: {
      webhookUrl: webhook?.webhook_url ?? '',
      secretMode: 'keep',
      webhookSecret: '',
      match: webhook?.match,
      interests: webhook?.interests ?? allEvents(),
    } as WebhookFormValues,
    validators: {
      onMount: schema,
      onChange: schema,
    },
    onSubmit: ({ value }) =>
      onSubmit({
        webhookUrl: value.webhookUrl.trim(),
        webhookSecret: computeSecret(mode, value),
        match: value.match,
        interests: value.interests,
      }),
  })

  const canSubmit = useStore(form.store, (s) => s.canSubmit)
  const secretMode = useStore(form.store, (s) => s.values.secretMode)

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="WebhooksLogoIcon" size="24" />
          {mode === 'create' ? 'Create webhook' : 'Edit webhook'}
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />{' '}
            {mode === 'create' ? 'Creating webhook' : 'Saving changes'}
          </span>
        ) : mode === 'create' ? (
          <span className="flex items-center gap-2">
            <Icon variant="PlusIcon" />
            Create webhook
          </span>
        ) : (
          'Save changes'
        ),
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
        className="flex flex-col gap-6"
      >
        <FormErrorBanner
          error={error}
          fallback={
            mode === 'create'
              ? 'Unable to create webhook'
              : 'Unable to update webhook'
          }
        />

        {mode === 'create' ? (
          <Text variant="body" theme="neutral">
            Receive workflow and workflow step lifecycle events for this org as
            CloudEvents v1.0 payloads. When a signing secret is set, requests
            are signed with HMAC-SHA256 in the{' '}
            <span className="font-mono">X-Nuon-Signature</span> header.
          </Text>
        ) : null}

        {mode === 'create' ? (
          <form.Field name="webhookUrl">
            {(field) => (
              <FormInput
                field={field}
                id="webhook-url"
                placeholder="https://example.com/webhooks/nuon"
                type="url"
                disabled={isPending}
                labelProps={{ labelText: 'Webhook URL' }}
                helperText="Must be an absolute http or https URL."
              />
            )}
          </form.Field>
        ) : (
          <div className="flex flex-col gap-2">
            <Label>URL</Label>
            <Code variant="inline" className="!px-2 !py-1">
              {webhook?.webhook_url}
            </Code>
            <Text variant="subtext" theme="neutral">
              URLs are unique per org per scope and cannot be changed in place.
              Delete + recreate the webhook to rename.
            </Text>
          </div>
        )}

        {mode === 'create' ? (
          <form.Field name="webhookSecret">
            {(field) => (
              <FormInput
                field={field}
                id="webhook-secret"
                placeholder="Used to sign delivered payloads"
                type="password"
                autoComplete="off"
                disabled={isPending}
                labelProps={{ labelText: 'Signing secret (optional)' }}
                helperText="The secret cannot be retrieved later. Edit the webhook to rotate it."
              />
            )}
          </form.Field>
        ) : (
          <div className="flex flex-col gap-2">
            <Label>Signing secret</Label>
            <Text variant="subtext" theme="neutral">
              {hasSecret
                ? 'A signing secret is currently configured. Existing secrets cannot be retrieved — rotate or clear it from here.'
                : 'No signing secret is configured. Set one to start signing delivered payloads.'}
            </Text>
            <form.Field name="secretMode">
              {(field) => (
                <FormRadioGroup
                  field={field}
                  disabled={isPending}
                  options={[
                    {
                      value: 'keep',
                      label: hasSecret
                        ? 'Leave existing secret unchanged'
                        : 'Do not set a secret',
                    },
                    {
                      value: 'rotate',
                      label: hasSecret
                        ? 'Rotate to a new secret'
                        : 'Set a new secret',
                    },
                    ...(hasSecret
                      ? [
                          {
                            value: 'clear',
                            label: 'Clear the existing secret',
                          },
                        ]
                      : []),
                  ]}
                />
              )}
            </form.Field>
            {secretMode === 'rotate' ? (
              <form.Field name="webhookSecret">
                {(field) => (
                  <FormInput
                    field={field}
                    id="webhook-secret"
                    placeholder="New signing secret"
                    type="password"
                    autoComplete="off"
                    disabled={isPending}
                  />
                )}
              </form.Field>
            ) : null}
          </div>
        )}

        <div className="flex flex-col gap-2">
          <Label>Scope</Label>
          <Text variant="subtext" theme="neutral">
            Filter which resources fire deliveries to this webhook.
          </Text>
          <form.Field name="match">
            {(field) => <FormMatchPicker field={field} disabled={isPending} />}
          </form.Field>
        </div>

        <div className="flex flex-col gap-2">
          <Label>Events</Label>
          <Text variant="subtext" theme="neutral">
            Pick which events fire deliveries to this webhook.
          </Text>
          <form.Field name="interests">
            {(field) => (
              <FormInterestsPicker field={field} disabled={isPending} />
            )}
          </form.Field>
        </div>
      </form>
    </Modal>
  )
}
