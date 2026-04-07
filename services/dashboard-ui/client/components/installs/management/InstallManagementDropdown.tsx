import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { RunnerProvider } from '@/providers/runner-provider'
import { useInstall } from '@/hooks/use-install'
import { ShutdownRunnerControl } from '@/components/runners/management/ShutdownRunnerControl'
import { AuditHistoryButton } from './AuditHistory'
import { DeprovisionButton } from './Deprovision'
import { DeprovisionStackButton } from './DeprovisionStack'
import { EditInputsButton } from './EditInputs'
import { ViewCurrentInputsButton } from './ViewCurrentInputs'
import { EnableAutoApproveButton } from './EnableAutoApprove'
import { EnableConfigSyncButton } from './EnableConfigSync'
import { ForgetButton } from './Forget'
import { GenerateInstallConfigButton } from './GenerateInstallConfig'
import { ReprovisionButton } from './Reprovision'
import { ReprovisionSandboxButton } from '@/components/sandbox/management/ReprovisionSandbox'
import { RunAdhocActionButton } from './RunAdhocAction'
import { SyncSecretsButton } from './SyncSecrets'
import { ViewStateButton } from './ViewState'

const InstallManagementDropdownContent = () => {
  const { install } = useInstall()
  const isMobile = !window.matchMedia('(min-width: 768px)').matches
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
      alignment={isMobile ? 'left' : 'right'}
    >
      <Menu className="min-w-56">
        <Text variant="label" theme="neutral">
          Settings
        </Text>
        <EditInputsButton isMenuButton />
        <ViewCurrentInputsButton isMenuButton />
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
        <ReprovisionSandboxButton isMenuButton />
        {install?.runner_id ? (
          <ShutdownRunnerControl
            isMenuButton
            showRunnerLabel
            runnerId={install.runner_id}
          />
        ) : null}
        <SyncSecretsButton isMenuButton />
        <RunAdhocActionButton isMenuButton />
        <hr />
        <Text variant="label" theme="neutral">
          Danger
        </Text>
        <DeprovisionButton isMenuButton />
        <DeprovisionStackButton isMenuButton />
        <span>
          <ForgetButton />
        </span>
      </Menu>
    </Dropdown>
  )
}

export const InstallManagementDropdown = () => {
  const { install } = useInstall()
  if (!install?.runner_id) return <InstallManagementDropdownContent />
  return (
    <RunnerProvider runnerId={install.runner_id}>
      <InstallManagementDropdownContent />
    </RunnerProvider>
  )
}
