import { useQuery } from '@tanstack/react-query'
import { Dropdown, type IDropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getTerraformWorkspaceLock } from '@/lib'
import { DriftScanSandboxButton } from './DriftScanSandbox'
import { ReprovisionSandboxButton } from './ReprovisionSandbox'
import { DeprovisionSandboxButton } from './DeprovisionSandbox'
import { UnlockSandboxTerraformStateButton } from './UnlockSandboxTerraformState'

export const ManagementDropdown = ({
  alignment = 'right',
  ...props
}: Omit<IDropdown, 'id' | 'buttonText' | 'children'>) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const workspaceId = install?.sandbox?.terraform_workspace?.id

  const { data: lock } = useQuery({
    queryKey: ['terraform-workspace-lock', org?.id, workspaceId],
    queryFn: () =>
      getTerraformWorkspaceLock({
        orgId: org.id,
        workspaceId: workspaceId!,
      }),
    enabled: !!org?.id && !!workspaceId,
  })

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
        {lock ? (
          <UnlockSandboxTerraformStateButton isMenuButton />
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
