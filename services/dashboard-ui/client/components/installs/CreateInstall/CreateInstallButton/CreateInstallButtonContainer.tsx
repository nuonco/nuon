import { useSurfaces } from '@/hooks/use-surfaces'
import { CreateInstallModal } from '../CreateInstallModal'
import { CreateInstallButton } from './CreateInstallButton'
import type { IButtonAsButton } from '@/components/common/Button'

export const CreateInstallButtonContainer = (props: IButtonAsButton) => {
  const { addModal } = useSurfaces()

  const handleOpen = () => {
    const modal = <CreateInstallModal />
    addModal(modal)
  }

  return <CreateInstallButton onOpen={handleOpen} {...props} />
}
