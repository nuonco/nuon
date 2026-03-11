import { Dropdown, type IDropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { UnlockTerraformWorkspaceButton } from '@/components/terraform-workspace/UnlockTerraformWorkspace'
import { useInstall } from '@/hooks/use-install'
import { DriftScanSandboxButton } from './DriftScanSandbox'
import { ReprovisionSandboxButton } from './ReprovisionSandbox'
import { DeprovisionSandboxButton } from './DeprovisionSandbox'

export const ManagementDropdown = ({
  alignment = 'right',
  ...props
}: Omit<IDropdown, 'id' | 'buttonText' | 'children'>) => {
  const { install } = useInstall()
  const workspaceId = install?.sandbox?.terraform_workspace?.id

  return (
    <Dropdown
      id="sandbox-mgmt"
      buttonText={
        <>
          <Icon variant="SlidersHorizontalIcon" /> Manage sandbox
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
