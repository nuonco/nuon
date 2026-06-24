import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { DeployComponentButton } from '@/components/install-components/management/DeployComponent'
import { DriftScanComponentButton } from '@/components/install-components/management/DriftScanComponent'
import { ForgetComponentButton } from '@/components/install-components/management/Forget'
import { TeardownComponentButton } from '@/components/install-components/management/TeardownComponent'
import { ToggleComponentButton } from '@/components/install-components/management/ToggleComponent'
import { UnlockTerraformWorkspaceButton } from '@/components/terraform-workspace/UnlockTerraformWorkspace'
import type { TComponent, TComponentConfig, TInstallComponent } from '@/types'

export const ManagementDropdown = ({
  component,
  componentConfig,
  currentBuildId,
  currentDeployStatus,
  installComponent,
}: {
  component: TComponent
  componentConfig?: TComponentConfig
  currentBuildId?: string
  currentDeployStatus?: string
  installComponent?: TInstallComponent
}) => {
  const workspaceId = installComponent?.terraform_workspace?.id
  const isToggleable = componentConfig?.toggleable === true
  const isDisabled = currentDeployStatus === 'disabled'

  return (
    <Dropdown
      id={`component-${component.id}-mgmt`}
      variant="secondary"
      buttonText={
        <>
          <Icon variant="SlidersHorizontalIcon" /> Component controls
        </>
      }
      alignment="right"
    >
      <Menu>
        <Text>Controls</Text>
        {isToggleable ? (
          <ToggleComponentButton
            component={component}
            enabling={isDisabled}
            isMenuButton
          />
        ) : null}
        {!isDisabled ? (
          <>
            <DriftScanComponentButton
              component={component}
              currentBuildId={currentBuildId}
              isMenuButton
            />
            <DeployComponentButton
              component={component}
              currentBuildId={currentBuildId}
              currentDeployStatus={currentDeployStatus}
              isMenuButton
            />
          </>
        ) : null}
        {(component?.type === 'terraform_module' || component?.type === 'pulumi') && workspaceId ? (
          <UnlockTerraformWorkspaceButton
            workspaceId={workspaceId}
            description={component.name}
            isMenuButton
          />
        ) : null}
        <hr />
        <Text>Remove</Text>
        {currentDeployStatus === 'inactive' ? (
          <ForgetComponentButton component={component} isMenuButton />
        ) : (
          <TeardownComponentButton component={component} isMenuButton />
        )}
      </Menu>
    </Dropdown>
  )
}
