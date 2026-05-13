import { Banner } from '@/components/common/Banner'
import { Code } from '@/components/common/Code'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

export const DeleteChannelSubscriptionModal = ({
  channelLabel,
  isPending,
  error,
  onSubmit,
  ...props
}: {
  channelLabel: string
  isPending: boolean
  error: TAPIError | null
  onSubmit: () => void
} & IModal) => (
  <Modal
    heading={
      <Text flex className="gap-4" variant="h3" weight="strong" theme="error">
        <Icon variant="WarningIcon" size="24" />
        Remove channel subscription?
      </Text>
    }
    primaryActionTrigger={{
      children: isPending ? (
        <span className="flex items-center gap-2">
          <Icon variant="Loading" /> Removing…
        </span>
      ) : (
        'Remove subscription'
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
          {error?.error || 'Unable to remove subscription'}
        </Banner>
      ) : null}

      <div className="flex flex-col gap-3">
        <Text variant="base" weight="strong">
          Lifecycle events will stop posting to this channel.
        </Text>
        <Code variant="inline" className="!px-2 !py-1">
          {channelLabel}
        </Code>
        <Text variant="body" theme="neutral">
          Re-subscribe later from the dashboard or via the{' '}
          <span className="font-mono">/nuon subscribe</span> slash command in
          Slack.
        </Text>
      </div>
    </div>
  </Modal>
)
