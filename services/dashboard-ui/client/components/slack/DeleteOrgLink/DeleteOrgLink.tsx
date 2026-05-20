import { Banner } from '@/components/common/Banner'
import { Code } from '@/components/common/Code'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

export const DeleteOrgLinkModal = ({
  teamId,
  isPending,
  error,
  onSubmit,
  ...props
}: {
  teamId: string
  isPending: boolean
  error: TAPIError | null
  onSubmit: () => void
} & IModal) => (
  <Modal
    heading={
      <Text flex className="gap-4" variant="h3" weight="strong" theme="error">
        <Icon variant="WarningIcon" size="24" />
        Unlink Slack workspace?
      </Text>
    }
    primaryActionTrigger={{
      children: isPending ? (
        <span className="flex items-center gap-2">
          <Icon variant="Loading" /> Unlinking…
        </span>
      ) : (
        'Unlink workspace'
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
          {error?.error || 'Unable to unlink workspace'}
        </Banner>
      ) : null}

      <div className="flex flex-col gap-3">
        <Text variant="base" weight="strong">
          This workspace will stop receiving lifecycle events for this org.
        </Text>
        <Code variant="inline" className="!px-2 !py-1">
          {teamId}
        </Code>
        <Text variant="body" theme="neutral">
          Channel subscriptions tied to this link are also removed. To re-link,
          run the install flow again from this workspace.
        </Text>
      </div>
    </div>
  </Modal>
)
