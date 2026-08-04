import { useSurfaces } from '@/hooks/use-surfaces'
import type { TApp } from '@/types'
import { CreateInstallModal } from '../CreateInstallModal'
import { CreateInstallButton } from './CreateInstallButton'
import type { IButtonAsButton } from '@/components/common/Button'

export const CreateInstallButtonContainer = ({
  initialApp,
  ...props
}: { initialApp?: TApp } & IButtonAsButton) => {
  const { addModal } = useSurfaces()

  const handleOpen = () => {
    const modal = <CreateInstallModal initialApp={initialApp} />
    addModal(modal)
  }

  return <CreateInstallButton onOpen={handleOpen} {...props} />
}
