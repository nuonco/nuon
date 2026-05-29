import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { ShutdownRunnerControl } from '@/components/runners/management/ShutdownRunnerControl'
import { ReprovisionSandboxButton } from '@/components/sandbox/management/ReprovisionSandbox'
import { useInstall } from '@/hooks/use-install'
import { RunnerProvider } from '@/providers/runner-provider'
import { RunAdhocActionButton } from '../RunAdhocAction/RunAdhocActionContainer'
import { AuditHistoryButton } from '../AuditHistory'
import { DeprovisionButton } from '../Deprovision'
import { DeprovisionStackButton } from '../DeprovisionStack'
import { EditLabelsButton } from '../EditLabels'
import { EnableConfigSyncButton } from '../EnableConfigSync'
import { ForgetButton } from '../Forget'
import { GenerateInstallConfigButton } from '../GenerateInstallConfig'
import { ReprovisionButton } from '../Reprovision'
import { SyncSecretsButton } from '../SyncSecrets'

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
        <EditLabelsButton isMenuButton />
        <AuditHistoryButton isMenuButton />
        <EnableConfigSyncButton isMenuButton />
        <GenerateInstallConfigButton isMenuButton />
        <hr />
        <Text variant="label" theme="neutral">
          Controls
        </Text>
        <RunAdhocActionButton isMenuButton />
        <ReprovisionButton isMenuButton />
        <ReprovisionSandboxButton isMenuButton />
        {install?.runner_id ? (
          <ShutdownRunnerControl
            isMenuButton
            showRunnerLabel
            isManaged={false}
            runnerId={install.runner_id}
          />
        ) : null}
        <SyncSecretsButton isMenuButton />
        <hr />
        <Text variant="label" theme="neutral">
          Danger
        </Text>
        <DeprovisionButton isMenuButton />
        <DeprovisionStackButton isMenuButton />
        <ForgetButton />
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
