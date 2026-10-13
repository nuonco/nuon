import { useEffect } from 'react'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { TInstallResource } from '@/types'
import type { IButtonAsButton } from '@/components/common/Button'
import { resourceIdentity } from '@/components/install-resources/resource-utils'
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
  const { panels, addPanel, updatePanel } = useSurfaces()
  const openPanel = panels.find((panel) => {
    const resource = (
      panel.content.props as { installResource?: TInstallResource }
    ).installResource
    return (
      panel.isVisible &&
      panel.content.type === InstallResourceDetailPanel &&
      resource &&
      resourceIdentity(resource) === resourceIdentity(installResource)
    )
  })
  const panelId = openPanel?.id
  const panelResource = (
    openPanel?.content.props as
      | { installResource?: TInstallResource }
      | undefined
  )?.installResource
  useEffect(() => {
    if (panelId && panelResource !== installResource)
      updatePanel(
        panelId,
        <InstallResourceDetailPanel installResource={installResource} />
      )
  }, [installResource, panelId, panelResource, updatePanel])

  const handleOpen = () => {
    addPanel(<InstallResourceDetailPanel installResource={installResource} />)
  }

  return <InstallResourceDetailPanelButton onOpen={handleOpen} {...props} />
}
