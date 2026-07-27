import { useSurfaces } from '@/hooks/use-surfaces'
import type { TInstallResource } from '@/types'
import type { IButtonAsButton } from '@/components/common/Button'
import {
  InstallResourceDetailPanel,
  InstallResourceDetailPanelButton,
} from './InstallResourceDetailPanel'

interface IInstallResourceDetailPanelButtonContainer extends IButtonAsButton {
  installResource: TInstallResource
}

export const InstallResourceDetailPanelButtonContainer = ({
  installResource,
  ...props
}: IInstallResourceDetailPanelButtonContainer) => {
  const { addPanel } = useSurfaces()

  const handleOpen = () => {
    addPanel(<InstallResourceDetailPanel installResource={installResource} />)
  }

  return <InstallResourceDetailPanelButton onOpen={handleOpen} {...props} />
}
