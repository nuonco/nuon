import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'

export const CreateInstallButton = ({
  onClick,
  ...props
}: { onClick: () => void } & Omit<IButtonAsButton, 'onClick'>) => (
  <Button onClick={onClick} {...props}>
    <Icon variant="PlusIcon" size={16} />
    Create install
  </Button>
)
