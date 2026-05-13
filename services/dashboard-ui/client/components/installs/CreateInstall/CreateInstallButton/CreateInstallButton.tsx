import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'

interface ICreateInstallButton extends IButtonAsButton {
  onOpen: () => void
}

export const CreateInstallButton = ({ onOpen, ...props }: ICreateInstallButton) => {
  return (
    <Button onClick={onOpen} {...props}>
      <Icon variant="CubeIcon" />
      Create install
    </Button>
  )
}
