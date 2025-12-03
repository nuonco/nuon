import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { AuditHistoryButton } from './AuditHistory'
import { DeprovisionButton } from './Deprovision'
import { DeprovisionStackButton } from './DeprovisionStack'
import { EnableAutoApproveButton } from './EnableAutoApprove'
import { EnableConfigSyncButton } from './EnableConfigSync'
import { ForgetButton } from './Forget'
import { GenerateInstallConfigButton } from './GenerateInstallConfig'
import { ReprovisionButton } from './Reprovision'
import { SyncSecretsButton } from './SyncSecrets'
import { ViewStateButton } from './ViewState'

export const InstallManagementDropdown = () => {
  return (
    <Dropdown
      buttonText={
        <>
          <Icon variant="SlidersHorizontalIcon" />
          Manage
        </>
      }
      id="install-mgmt"
      variant="primary"
      alignment="right"
    >
      <Menu className="min-w-56">
        <Text variant="label" theme="neutral">
          Settings
        </Text>
        <Button>
          Edit inputs <Icon variant="PencilSimpleLine" />
        </Button>
        <AuditHistoryButton isMenuButton />
        <ViewStateButton isMenuButton />
        <EnableAutoApproveButton isMenuButton />
        <EnableConfigSyncButton isMenuButton />
        <GenerateInstallConfigButton isMenuButton />
        <hr />
        <Text variant="label" theme="neutral">
          Controls
        </Text>
        <ReprovisionButton isMenuButton />
        <SyncSecretsButton isMenuButton />
        <DeprovisionButton isMenuButton />
        <DeprovisionStackButton isMenuButton />
        <hr />
        <Text variant="label" theme="neutral">
          Danger
        </Text>
        <span>
          <ForgetButton />
        </span>
      </Menu>
    </Dropdown>
  )
}
