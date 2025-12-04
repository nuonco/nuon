'use client'

import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { useSurfaces } from '@/hooks/use-surfaces'
import { CreateInstallModal } from './CreateInstallModal'

interface ICreateInstall {}

export const CreateInstallButton = ({
  ...props
}: ICreateInstall & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <CreateInstallModal />

  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      <Icon variant="Cube" />
      Create install
    </Button>
  )
}