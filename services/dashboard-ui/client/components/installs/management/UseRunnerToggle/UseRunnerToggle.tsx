import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface IUseRunnerToggleModal extends Omit<IModal, 'onSubmit'> {
  isPending: boolean
  error: any
  usesRunner: boolean
  isSuccess: boolean
  onSubmit: () => void
}

export const UseRunnerToggleModal = ({
  isPending,
  error,
  usesRunner,
  isSuccess,
  onSubmit,
  ...props
}: IUseRunnerToggleModal) => {
  const buttonText = usesRunner ? 'Switch to runnerless' : 'Switch to a runner'
  const confirmText = usesRunner
    ? "Switching to runnerless removes this install's runner and reprovisions its stack. Jobs will move from the runner to Nuon's control plane."
    : "Switching to a runner reprovisions this install's stack with a new runner. Jobs will move from Nuon's control plane to the runner."

  return (
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="SneakerMoveIcon" size="24" />
          {buttonText}
        </Text>
      }
      primaryActionTrigger={{
        children: isSuccess ? (
          <span className="flex items-center gap-2">
            <Icon variant="CheckCircleIcon" /> Reprovision started
          </span>
        ) : isPending ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Starting reprovision...
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="SneakerMoveIcon" />
            {buttonText}
          </span>
        ),
        onClick: onSubmit,
        disabled: isPending || isSuccess,
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-4">
        {error ? (
          <Banner theme="error">
            {error?.error ||
              `Unable to switch to ${usesRunner ? 'runnerless' : 'a runner'}`}
          </Banner>
        ) : null}

        <Text variant="body">{confirmText}</Text>

        <Banner theme="warn">
          <Text variant="body">
            <strong>Warning:</strong> The stack will be reprovisioned. This may
            cause downtime while jobs move between the runner and the control
            plane.
          </Text>
        </Banner>
      </div>
    </Modal>
  )
}
