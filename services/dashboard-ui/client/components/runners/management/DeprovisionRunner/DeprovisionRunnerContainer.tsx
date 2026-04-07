import { useSurfaces } from '@/hooks/use-surfaces'
import { DeprovisionRunnerButton, DeprovisionRunnerModal } from './DeprovisionRunner'
import type { IButtonAsButton } from '@/components/common/Button'

interface IDeprovisionRunnerButtonContainer extends IButtonAsButton {
  buttonText?: string
  headingText?: string
}

export const DeprovisionRunnerButtonContainer = ({
  buttonText = 'Deprovision runner',
  headingText = 'Deprovision runner information',
  ...props
}: IDeprovisionRunnerButtonContainer) => {
  const { addModal, removeModal } = useSurfaces()

  const handleOpen = () => {
    let modalId: string
    const modal = (
      <DeprovisionRunnerModal
        headingText={headingText}
        onClose={() => removeModal(modalId)}
      />
    )
    modalId = addModal(modal)
  }

  return (
    <DeprovisionRunnerButton
      buttonText={buttonText}
      onOpen={handleOpen}
      {...props}
    />
  )
}
