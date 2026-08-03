import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import type { IModal } from '@/components/surfaces/Modal'
import { useInstall } from '@/hooks/use-install'
import { useInstallAppConfig } from '@/hooks/use-install-app-config'
import { useSurfaces } from '@/hooks/use-surfaces'
import { DeprovisionStackModal } from './DeprovisionStack'

interface IDeprovisionStack {}

export const DeprovisionStackModalContainer = ({ ...props }: IDeprovisionStack & IModal) => {
  const { removeModal } = useSurfaces()
  const { install } = useInstall()
  const { appConfig } = useInstallAppConfig()

  return (
    <DeprovisionStackModal
      installName={install.name}
      stackType={appConfig?.stack?.type}
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
      variant="danger"
    >
      Deprovision stack
      <Icon variant="StackMinusIcon" />
    </Button>
  )
}
