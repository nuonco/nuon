import { Button } from '@/components/common/Button'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { ClickToCopy } from '@/components/common/ClickToCopy'
import type { TVCSWebhookSubscription } from '@/types'

interface IWebhookSubscriptionSection {
  webhookSubscription?: TVCSWebhookSubscription
  onCreateSubscription?: () => void
  isCreating?: boolean
}

export const WebhookSubscriptionSection = ({
  webhookSubscription,
  onCreateSubscription,
  isCreating = false,
}: IWebhookSubscriptionSection) => {
  if (!webhookSubscription) {
    return (
      <div className="flex flex-col gap-4">
        <Text variant="body" weight="strong">
          Webhook subscription
        </Text>
        <div className="flex items-center gap-4">
          <Text variant="subtext" theme="neutral">
            No webhook subscription yet
          </Text>
          {onCreateSubscription && (
            <Button
              size="xs"
              variant="secondary"
              onClick={onCreateSubscription}
              disabled={isCreating}
            >
              {isCreating ? 'Creating...' : 'Add webhook subscription'}
            </Button>
          )}
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-4">
      <Text variant="body" weight="strong">
        Webhook subscription
      </Text>
      <div className="grid grid-cols-3 gap-4">
        <LabeledValue label="Webhook URL">
          <ClickToCopy>
            <Text variant="subtext" className="break-all">
              {webhookSubscription.webhook_url || '—'}
            </Text>
          </ClickToCopy>
        </LabeledValue>

        <LabeledValue label="Status">
          <Text variant="subtext">
            {webhookSubscription.status?.status || '—'}
          </Text>
        </LabeledValue>

        <LabeledValue label="Subscription ID">
          <ID theme="default">{webhookSubscription.id}</ID>
        </LabeledValue>
      </div>
      <div className="grid grid-cols-3 gap-4">
        <LabeledValue label="GitHub hook ID">
          <Text variant="subtext">
            {webhookSubscription.github_hook_id || '—'}
          </Text>
        </LabeledValue>

        <LabeledValue label="Created">
          {webhookSubscription.created_at ? (
            <Time
              variant="subtext"
              time={webhookSubscription.created_at}
              format="relative"
            />
          ) : (
            <Text variant="subtext">—</Text>
          )}
        </LabeledValue>
      </div>
    </div>
  )
}
