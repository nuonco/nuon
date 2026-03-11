import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { DeployComponentButton } from '@/components/install-components/management/DeployComponent'
import { DriftScanComponentButton } from '@/components/install-components/management/DriftScanComponent'
import { ForgetComponentButton } from '@/components/install-components/management/Forget'
import { TeardownComponentButton } from '@/components/install-components/management/TeardownComponent'
import { UnlockTerraformWorkspaceButton } from '@/components/terraform-workspace/UnlockTerraformWorkspace'
import type { TComponent, TInstallComponent } from '@/types'

export const ManagementDropdown = ({
  component,
  currentBuildId,
  currentDeployStatus,
  installComponent,
}: {
  component: TComponent
  currentBuildId?: string
  currentDeployStatus?: string
  installComponent?: TInstallComponent
}) => {
  const workspaceId = installComponent?.terraform_workspace?.id

  return (
    <Dropdown
      id={`component-${component.id}-mgmt`}
      variant="primary"
      buttonText={
        <>
          <Icon variant="SlidersHorizontalIcon" /> Manage
        </>
      }
      alignment="right"
    >
      <Menu>
        <Text>Controls</Text>
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
        {component?.type === 'terraform_module' && workspaceId ? (
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
