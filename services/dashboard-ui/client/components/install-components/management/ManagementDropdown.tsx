import { Button } from '@/components/common/Button'
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

const DisabledMenuItem = ({
  label,
  icon,
  reason,
}: {
  label: string
  icon: 'CloudArrowDownIcon' | 'CloudArrowUpIcon' | 'TrashIcon'
  reason: string
}) => (
  <Button
    isMenuButton
    disabled
    className="w-full"
    tooltipProps={{
      className: 'block !w-full',
      position: 'left',
      tipContent: reason,
    }}
  >
    {label}
    <Icon variant={icon} />
  </Button>
)

export const ManagementDropdown = ({
  component,
  componentConfig,
  currentBuildId,
  currentDeployStatus,
  installComponent,
  isConfigLoading,
  removed = false,
}: {
  component: TComponent
  componentConfig?: TComponentConfig
  currentBuildId?: string
  currentDeployStatus?: string
  installComponent?: TInstallComponent
  isConfigLoading?: boolean
  removed?: boolean
}) => {
  const workspaceId = installComponent?.terraform_workspace?.id
  const isToggleable = componentConfig?.toggleable === true
  const isDisabled = currentDeployStatus === 'disabled'

  const isTornDown = currentDeployStatus === 'inactive'
  const isInConfig = !!componentConfig

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
        {removed ? (
          <DisabledMenuItem
            label="Deploy component"
            icon="CloudArrowUpIcon"
            reason="This component is no longer in the install's app config version."
          />
        ) : !isDisabled ? (
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
        {(component?.type === 'terraform_module' ||
          component?.type === 'pulumi') &&
        workspaceId ? (
          <UnlockTerraformWorkspaceButton
            workspaceId={workspaceId}
            description={component.name}
            isMenuButton
          />
        ) : null}
        <hr />
        <Text>Remove</Text>
        {isTornDown ? (
          <DisabledMenuItem
            label="Teardown component"
            icon="CloudArrowDownIcon"
            reason="This component is already torn down."
          />
        ) : (
          <TeardownComponentButton
            component={component}
            isMenuButton
            variant="danger"
          />
        )}
        <ForgetComponentButton
          component={component}
          isMenuButton
          isTornDown={isTornDown}
          isInConfig={isInConfig}
          isConfigLoading={isConfigLoading}
        />
      </Menu>
    </Dropdown>
  )
}
