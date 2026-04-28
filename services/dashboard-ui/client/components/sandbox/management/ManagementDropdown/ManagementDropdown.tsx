import { Dropdown, type IDropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { UnlockTerraformWorkspaceButton } from '@/components/terraform-workspace/UnlockTerraformWorkspace'
import { DriftScanSandboxButton } from '../DriftScanSandbox'
import { ReprovisionSandboxButton } from '../ReprovisionSandbox'
import { DeprovisionSandboxButton } from '../DeprovisionSandbox'

interface IManagementDropdown extends Omit<IDropdown, 'id' | 'buttonText' | 'children'> {
  workspaceId?: string
}

export const ManagementDropdown = ({
  alignment = 'right',
  workspaceId,
  ...props
}: IManagementDropdown) => {
  return (
    <Dropdown
      id="sandbox-mgmt"
      variant="secondary"
      buttonText={
        <>
          Sandbox controls
        </>
      }
      alignment={alignment}
      {...props}
    >
      <Menu>
        <div className="px-2 pt-2 pb-1">
          <Text variant="subtext" theme="neutral">
            Controls
          </Text>
        </div>

        <DriftScanSandboxButton isMenuButton />
        <ReprovisionSandboxButton isMenuButton />
        {workspaceId ? (
          <UnlockTerraformWorkspaceButton
            workspaceId={workspaceId}
            description="the sandbox"
            isMenuButton
          />
        ) : null}

        <hr />
        <div className="px-2 pt-2 pb-1">
          <Text variant="subtext" theme="neutral">
            Remove
          </Text>
        </div>

        <DeprovisionSandboxButton isMenuButton />
      </Menu>
    </Dropdown>
  )
}
