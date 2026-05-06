import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Code } from '@/components/common/Code'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Label } from '@/components/common/form/Label'
import { RadioInput } from '@/components/common/form/RadioInput'
import {
  InterestsPicker,
  allEvents,
  type Interests,
} from '@/components/interests'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError, TWebhook } from '@/types'

// EditWebhookModal — backed by the new PATCH /v1/orgs/current/webhooks/{id}
// endpoint added in Phase D. URL is read-only (it's part of the unique index;
// rename = delete + recreate). Secret rotation is opt-in via a checkbox; if
// untouched, the field stays nil on the wire which the backend treats as
// "leave unchanged". A "clear secret" toggle lets the user remove an existing
// secret without sending a replacement.
export type EditWebhookFormInput = {
  // undefined → leave unchanged (do not include in the PATCH body)
  // empty string → clear
  // string → rotate to this value
  webhookSecret: string | undefined
  interests: Interests
}

export const EditWebhookModal = ({
  webhook,
  isPending,
  error,
  onSubmit,
  ...props
}: {
  webhook: TWebhook
  isPending: boolean
  error: TAPIError | null
  onSubmit: (input: EditWebhookFormInput) => void
} & Omit<IModal, 'onSubmit'>) => {
  const [secretMode, setSecretMode] = useState<'keep' | 'rotate' | 'clear'>(
    'keep'
  )
  const [webhookSecret, setWebhookSecret] = useState('')
  const [interests, setInterests] = useState<Interests>(
    () => webhook.interests ?? allEvents()
  )

  const computedSecret: string | undefined =
    secretMode === 'keep'
      ? undefined
      : secretMode === 'clear'
        ? ''
        : webhookSecret.trim()

  const canSubmit = !isPending && (secretMode !== 'rotate' || webhookSecret.trim().length > 0)

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="WebhooksLogo" size="24" />
          Edit webhook
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Saving...
          </span>
        ) : (
          'Save changes'
        ),
        disabled: !canSubmit,
        onClick: () =>
          onSubmit({
            webhookSecret: computedSecret,
            interests,
          }),
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to update webhook'}
          </Banner>
        ) : null}

        <div className="flex flex-col gap-2">
          <Label>URL</Label>
          <Code variant="inline" className="!px-2 !py-1">
            {webhook.webhook_url}
          </Code>
          <Text variant="subtext" theme="neutral">
            URLs are unique per org and cannot be changed. Delete + recreate
            the webhook to rename.
          </Text>
        </div>

        <div className="flex flex-col gap-2">
          <Label>Signing secret</Label>
          <Text variant="subtext" theme="neutral">
            {webhook.has_secret
              ? 'A signing secret is currently configured. Existing secrets cannot be retrieved — rotate or clear it from here.'
              : 'No signing secret is configured. Set one to start signing delivered payloads.'}
          </Text>
          <div className="flex flex-col gap-1">
            <RadioInput
              id="secret-keep"
              name="secret-mode"
              value="keep"
              checked={secretMode === 'keep'}
              onChange={() => setSecretMode('keep')}
              labelProps={{
                labelText: webhook.has_secret
                  ? 'Leave existing secret unchanged'
                  : 'Do not set a secret',
                labelTextProps: { variant: 'subtext' },
              }}
            />
            <RadioInput
              id="secret-rotate"
              name="secret-mode"
              value="rotate"
              checked={secretMode === 'rotate'}
              onChange={() => setSecretMode('rotate')}
              labelProps={{
                labelText: webhook.has_secret
                  ? 'Rotate to a new secret'
                  : 'Set a new secret',
                labelTextProps: { variant: 'subtext' },
              }}
            />
            {webhook.has_secret ? (
              <RadioInput
                id="secret-clear"
                name="secret-mode"
                value="clear"
                checked={secretMode === 'clear'}
                onChange={() => setSecretMode('clear')}
                labelProps={{
                  labelText: 'Clear the existing secret',
                  labelTextProps: { variant: 'subtext' },
                }}
              />
            ) : null}
          </div>
          {secretMode === 'rotate' ? (
            <Input
              id="webhook-secret"
              placeholder="New signing secret"
              type="password"
              value={webhookSecret}
              onChange={(e) => setWebhookSecret(e.target.value)}
              autoComplete="off"
            />
          ) : null}
        </div>

        <div className="flex flex-col gap-2">
          <Label>Events</Label>
          <InterestsPicker
            variant="webhook"
            value={interests}
            onChange={setInterests}
          />
        </div>
      </div>
    </Modal>
  )
}
