import { Banner } from '@/components/common/Banner'
import { Code } from '@/components/common/Code'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

export const DeleteWebhookModal = ({
  webhookUrl,
  isPending,
  error,
  onSubmit,
  ...props
}: {
  webhookUrl: string
  isPending: boolean
  error: TAPIError | null
  onSubmit: () => void
} & IModal) => {
  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong" theme="error">
          <Icon variant="Warning" size="24" />
          Delete webhook?
        </Text>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Deleting...
          </span>
        ) : (
          'Delete webhook'
        ),
        disabled: isPending,
        onClick: () => onSubmit(),
        variant: 'danger',
      }}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to delete webhook'}
          </Banner>
        ) : null}

        <div className="flex flex-col gap-3">
          <Text variant="base" weight="strong">
            This webhook will stop receiving operation lifecycle events.
          </Text>
          <Code variant="inline" className="!px-2 !py-1">{webhookUrl}</Code>
          <Text variant="body" theme="neutral">
            If a signing secret was set, it cannot be recovered. To reuse this
            URL with a new secret, delete and recreate the webhook.
          </Text>
        </div>
      </div>
    </Modal>
  )
}
