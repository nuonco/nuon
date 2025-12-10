'use client'

import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useSurfaces } from '@/hooks/use-surfaces'

export const DeprovisionRunnerButton = ({
  buttonText = 'Deprovision runner',
  headingText = 'Deprovision runner information',
  ...props
}: IButtonAsButton & {
  buttonText?: string
  headingText?: string
}) => {
  const { addModal } = useSurfaces()
  const modal = <DeprovisionRunnerModal headingText={headingText} />
  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
      className="!text-red-600 dark:!text-red-400"
    >
      {props?.isMenuButton ? null : <Icon variant="BoxArrowDown" />}
      {buttonText}
      {props?.isMenuButton ? <Icon variant="BoxArrowDown" /> : null}
    </Button>
  )
}

export const DeprovisionRunnerModal = ({
  headingText = 'Deprovision runner information',
  ...props
}: IModal & {
  headingText?: string
}) => {
  const { removeModal } = useSurfaces()

  const handleClose = () => {
    removeModal(props.modalId)
  }

  return (
    <Modal
      heading={
        <div className="flex flex-col gap-2">
          <Text
            className="inline-flex gap-4 items-center"
            variant="h3"
            weight="strong"
            theme="info"
          >
            <Icon variant="BoxArrowDown" size="24" />
            {headingText}
          </Text>
        </div>
      }
      onClose={handleClose}
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