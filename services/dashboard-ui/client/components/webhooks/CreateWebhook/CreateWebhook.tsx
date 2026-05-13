import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Label } from '@/components/common/form/Label'
import {
  InterestsPicker,
  allEvents,
  type Interests,
} from '@/components/interests'
import { MatchPicker } from '@/components/match/MatchPicker'
import type { SubscriptionMatch } from '@/components/match/types'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

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
  const [webhookUrl, setWebhookUrl] = useState('')
  const [webhookSecret, setWebhookSecret] = useState('')
  const [match, setMatch] = useState<SubscriptionMatch | undefined>(undefined)
  const [interests, setInterests] = useState<Interests>(() => allEvents())

  const trimmedUrl = webhookUrl.trim()
  const isValidUrl = /^https?:\/\/.+/i.test(trimmedUrl)

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
        disabled: !isValidUrl || isPending,
        onClick: () =>
          onSubmit({
            webhookUrl: trimmedUrl,
            webhookSecret: webhookSecret.trim(),
            match,
            interests,
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
            The secret cannot be retrieved later. Edit the webhook to rotate
            it.
          </Text>
        </div>

        <div className="flex flex-col gap-2">
          <Label>Scope</Label>
          <Text variant="subtext" theme="neutral">
            Filter which resources fire deliveries to this webhook.
          </Text>
          <MatchPicker value={match} onChange={setMatch} />
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
