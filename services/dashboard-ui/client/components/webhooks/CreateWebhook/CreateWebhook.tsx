import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Label } from '@/components/common/form/Label'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

export const CreateWebhookModal = ({
  isPending,
  error,
  onSubmit,
  ...props
}: {
  isPending: boolean
  error: TAPIError | null
  onSubmit: (params: { webhookUrl: string; webhookSecret: string }) => void
} & Omit<IModal, 'onSubmit'>) => {
  const [webhookUrl, setWebhookUrl] = useState('')
  const [webhookSecret, setWebhookSecret] = useState('')

  const trimmedUrl = webhookUrl.trim()
  const isValidUrl = /^https?:\/\/.+/i.test(trimmedUrl)

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="WebhooksLogo" size="24" />
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
            <Icon variant="Plus" />
            Create webhook
          </span>
        ),
        disabled: !isValidUrl || isPending,
        onClick: () =>
          onSubmit({
            webhookUrl: trimmedUrl,
            webhookSecret: webhookSecret.trim(),
          }),
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to create webhook'}
          </Banner>
        ) : null}

        <Text variant="body" theme="neutral">
          Receive workflow and workflow step lifecycle events for this org as
          CloudEvents v1.0 payloads. When a signing secret is set, requests are
          signed with HMAC-SHA256 in the{' '}
          <span className="font-mono">X-Nuon-Signature</span> header.
        </Text>

        <div className="flex flex-col gap-2">
          <Label htmlFor="webhook-url">Webhook URL</Label>
          <Input
            id="webhook-url"
            placeholder="https://example.com/webhooks/nuon"
            type="url"
            value={webhookUrl}
            onChange={(e) => setWebhookUrl(e.target.value)}
            required
          />
          <Text variant="subtext" theme="neutral">
            Must be an absolute http or https URL.
          </Text>
        </div>

        <div className="flex flex-col gap-2">
          <Label htmlFor="webhook-secret">Signing secret (optional)</Label>
          <Input
            id="webhook-secret"
            placeholder="Used to sign delivered payloads"
            type="password"
            value={webhookSecret}
            onChange={(e) => setWebhookSecret(e.target.value)}
            autoComplete="off"
          />
          <Text variant="subtext" theme="neutral">
            The secret cannot be retrieved later. To rotate it, delete this
            webhook and create a new one.
          </Text>
        </div>
      </div>
    </Modal>
  )
}
