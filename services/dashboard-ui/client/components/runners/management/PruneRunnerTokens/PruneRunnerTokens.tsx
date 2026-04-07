import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface IPruneRunnerTokensModal extends IModal {
  isLoading: boolean
  onConfirm: () => void
}

export const PruneRunnerTokensModal = ({
  isLoading,
  onConfirm,
  ...props
}: IPruneRunnerTokensModal) => {
  return (
    <Modal
      heading="Prune old runner tokens?"
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            Pruning...
          </span>
        ) : (
          'Prune tokens'
        ),
        disabled: isLoading,
        onClick: onConfirm,
        variant: 'primary' as const,
      }}
      {...props}
    >
      <div className="flex flex-col gap-4">
        <Text variant="body" theme="neutral">
          This will remove old authentication tokens for this runner, keeping
          only the most recent token.
        </Text>
        <Text variant="body" theme="neutral">
          Use this to clean up accumulated tokens without disrupting the active
          runner.
        </Text>
      </div>
    </Modal>
  )
}
