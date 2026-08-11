import { useForm, useStore } from '@tanstack/react-form'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { FormErrorBanner } from '@/components/common/form/FormErrorBanner'
import { FormInput } from '@/components/common/form/FormInput'
import { allEvents, type Interests } from '@/components/interests'
import { FormInterestsPicker } from '@/components/interests/FormInterestsPicker'
import { Label } from '@/components/common/form/Label'
import { FormMatchPicker } from '@/components/match/FormMatchPicker'
import type { SubscriptionMatch } from '@/components/match/types'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'
import { createWebhookSchema, type CreateWebhookValues } from './schema'

export type CreateWebhookInput = {
  webhookUrl: string
  webhookSecret: string
  match: SubscriptionMatch | undefined
  interests: Interests
}

export const CreateWebhookModal = ({
  isPending,
  error,
  onSubmit,
  ...props
}: {
  isPending: boolean
  error: TAPIError | null
  onSubmit: (input: CreateWebhookInput) => void
} & Omit<IModal, 'onSubmit'>) => {
  const form = useForm({
    defaultValues: {
      webhookUrl: '',
      webhookSecret: '',
      match: undefined,
      interests: allEvents(),
    } as CreateWebhookValues,
    validators: {
      onMount: createWebhookSchema,
      onChange: createWebhookSchema,
    },
    onSubmit: ({ value }) =>
      onSubmit({
        webhookUrl: value.webhookUrl.trim(),
        webhookSecret: value.webhookSecret.trim(),
        match: value.match,
        interests: value.interests,
      }),
  })

  const canSubmit = useStore(form.store, (s) => s.canSubmit)

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="WebhooksLogoIcon" size="24" />
          Create webhook
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Creating...
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="PlusIcon" />
            Create webhook
          </span>
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
        <FormErrorBanner error={error} fallback="Unable to create webhook" />

        <Text variant="body" theme="neutral">
          Receive workflow and workflow step lifecycle events for this org as
          CloudEvents v1.0 payloads. When a signing secret is set, requests are
          signed with HMAC-SHA256 in the{' '}
          <span className="font-mono">X-Nuon-Signature</span> header.
        </Text>

        <form.Field name="webhookUrl">
          {(field) => (
            <FormInput
              field={field}
              id="webhook-url"
              placeholder="https://example.com/webhooks/nuon"
              type="url"
              labelProps={{ labelText: 'Webhook URL' }}
              helperText="Must be an absolute http or https URL."
            />
          )}
        </form.Field>

        <form.Field name="webhookSecret">
          {(field) => (
            <FormInput
              field={field}
              id="webhook-secret"
              placeholder="Used to sign delivered payloads"
              type="password"
              autoComplete="off"
              labelProps={{ labelText: 'Signing secret (optional)' }}
              helperText="The secret cannot be retrieved later. Edit the webhook to rotate it."
            />
          )}
        </form.Field>

        <div className="flex flex-col gap-2">
          <Label>Scope</Label>
          <Text variant="subtext" theme="neutral">
            Filter which resources fire deliveries to this webhook.
          </Text>
          <form.Field name="match">
            {(field) => <FormMatchPicker field={field} />}
          </form.Field>
        </div>

        <div className="flex flex-col gap-2">
          <Label>Events</Label>
          <Text variant="subtext" theme="neutral">
            Pick which events fire deliveries to this webhook.
          </Text>
          <form.Field name="interests">
            {(field) => <FormInterestsPicker field={field} />}
          </form.Field>
        </div>
      </form>
    </Modal>
  )
}
