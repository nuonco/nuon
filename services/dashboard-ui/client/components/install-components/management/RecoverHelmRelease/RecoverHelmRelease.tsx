import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

interface IRecoverHelmReleaseModal extends Omit<IModal, 'onSubmit'> {
  componentName: string
  status?: string
  isPending: boolean
  error?: TAPIError | null
  onSubmit: () => void
  onClose: () => void
}

export const RecoverHelmReleaseModal = ({
  componentName,
  status,
  isPending,
  error,
  onSubmit,
  onClose,
  ...props
}: IRecoverHelmReleaseModal) => {
  return (
    <Modal
      heading={
        <div className="flex flex-col gap-2">
          <Text flex className="gap-4" variant="h3" weight="strong">
            <Icon variant="ArrowCounterClockwiseIcon" size="24" />
            Recover Helm release?
          </Text>
          <Text variant="body" className="text-cool-grey-600 dark:text-cool-grey-400">
            Unstick the Helm release for {componentName}
          </Text>
        </div>
      }
      primaryActionTrigger={{
        children: isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            Recovering release
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="ArrowCounterClockwiseIcon" />
            Recover release
          </span>
        ),
        disabled: isPending,
        onClick: onSubmit,
        variant: 'danger' as const,
      }}
      onClose={onClose}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {(error as any)?.error ? (
          <Banner theme="error">
            {(error as any)?.error || 'Unable to recover the Helm release'}
          </Banner>
        ) : null}

        <div className="flex flex-col gap-4">
          <Text variant="body" theme="neutral">
            Recovering should only be done when a previous operation left the release
            {status ? ` in ${status}` : ' part-way through'}. Nothing is deployed — deploy{' '}
            {componentName} afterwards.
          </Text>

          <Banner theme="warn">
            <Text variant="body">
              <strong>Warning:</strong> the release is rolled back to the last revision that
              rolled out. If none ever did, it is removed instead.
            </Text>
          </Banner>
        </div>
      </div>
    </Modal>
  )
}
