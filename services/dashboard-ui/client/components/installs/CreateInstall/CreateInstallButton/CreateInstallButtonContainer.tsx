import { useNavigate } from 'react-router'
import type { IButtonAsButton } from '@/components/common/Button'
import { createInstallWizardPath } from '@/components/installs/CreateInstallWizard/wizard-path'
import { useInstallUIWizards } from '@/hooks/use-install-ui-wizards'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { TApp } from '@/types'
import { CreateInstallModal } from '../CreateInstallModal'
import { CreateInstallButton } from './CreateInstallButton'

export const CreateInstallButtonContainer = ({
  initialApp,
  initialBranchId,
  ...props
}: { initialApp?: TApp; initialBranchId?: string } & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const { org } = useOrg()
  const navigate = useNavigate()
  const wizards = useInstallUIWizards()

  const handleOpen = () => {
    if (wizards && org.id && initialApp?.id && initialBranchId) {
      navigate(createInstallWizardPath(org.id, initialApp.id, initialBranchId))
      return
    }

    addModal(
      <CreateInstallModal
        initialApp={initialApp}
        initialBranchId={initialBranchId}
      />
    )
  }

  return <CreateInstallButton onOpen={handleOpen} {...props} />
}
