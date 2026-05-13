import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'

interface IDeprovisionRunnerButton extends IButtonAsButton {
  buttonText?: string
  headingText?: string
  onOpen: () => void
}

export const DeprovisionRunnerButton = ({
  buttonText = 'Deprovision runner',
  headingText,
  onOpen,
  ...props
}: IDeprovisionRunnerButton) => {
  return (
    <Button
      onClick={onOpen}
      {...props}
      className="!text-red-600 dark:!text-red-400"
    >
      {props?.isMenuButton ? null : <Icon variant="BoxArrowDownIcon" />}
      {buttonText}
      {props?.isMenuButton ? <Icon variant="BoxArrowDownIcon" /> : null}
    </Button>
  )
}

export const DeprovisionRunnerModal = ({
  headingText = 'Deprovision runner information',
  onClose,
  ...props
}: IModal & {
  headingText?: string
  onClose: () => void
}) => {
  return (
    <Modal
      heading={
        <div className="flex flex-col gap-2">
          <Text
            flex
            className="gap-4"
            variant="h3"
            weight="strong"
            theme="info"
          >
            <Icon variant="BoxArrowDownIcon" size="24" />
            {headingText}
          </Text>
        </div>
      }
      onClose={onClose}
      {...props}
    >
      <div className="flex flex-col gap-4">
        <Text variant="body">
          You can use the shut down button to restart your runner during
          normal operation. If you need to forcefully terminate your
          runner, you can terminate the instance directly from the
          AutoScaling group in your AWS account.
        </Text>
        <Text variant="body">
          Deleting the instance has the chance to lose any state of
          in-flight jobs, but in other cases is a safe operation.
        </Text>
      </div>
    </Modal>
  )
}
