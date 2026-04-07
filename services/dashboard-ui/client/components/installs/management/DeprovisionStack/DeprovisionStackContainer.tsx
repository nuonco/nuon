import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import type { IModal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useSurfaces } from '@/hooks/use-surfaces'
import { DeprovisionStackModal } from './DeprovisionStack'

interface IDeprovisionStack {}

export const DeprovisionStackModalContainer = ({ ...props }: IDeprovisionStack & IModal) => {
  const { removeModal } = useSurfaces()
  const { install } = useInstall()

  return (
    <DeprovisionStackModal
      installName={install.name}
      onDismiss={() => removeModal(props.modalId)}
      {...props}
    />
  )
}

export const DeprovisionStackButton = ({
  ...props
}: IDeprovisionStack & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <DeprovisionStackModalContainer />

  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      Deprovision stack
      <Icon variant="StackMinus" />
    </Button>
  )
}
