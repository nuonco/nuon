import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { DriftedBanner } from '@/components/install-components/DriftedBanner'
import { SandboxConfigCard } from '@/components/sandbox/SandboxConfigCard'
import { ManagementDropdown } from '@/components/sandbox/management/ManagementDropdown'
import { Panel } from '@/components/surfaces/Panel'
import { TerraformWorkspaceCard } from '@/components/terraform-workspace/TerraformWorkspaceCard'
import { useInstall } from '@/hooks/use-install'
import { useInstallAppConfig } from '@/hooks/use-install-app-config'
import { InstallSandbox } from './InstallSandbox'

export const InstallSandboxContainer = () => {
  const { install } = useInstall()
  const { appConfig, isLoading: configLoading } = useInstallAppConfig()
  const latestRun = install?.install_sandbox_runs?.at(0)
  const driftedObject = install?.drifted_objects?.find(
    (drifted) =>
      drifted?.target_type === 'install_sandbox_run' &&
      drifted?.target_id === latestRun?.id
  )
  const sandboxConfig = appConfig?.sandbox
  const isPulumi = sandboxConfig?.type === 'pulumi'
  const workspaceId = install?.sandbox?.terraform_workspace?.id
  const stateTitle = isPulumi ? 'Pulumi state' : 'Terraform state'

  return (
    <InstallSandbox
      sandbox={install?.sandbox}
      latestRun={latestRun}
      actions={
        <>
          {workspaceId ? (
            <Panel
              heading={stateTitle}
              panelKey="sandbox-workspace-state"
              size="3/4"
              triggerButton={{
                variant: 'secondary',
                children: (
                  <>
                    <Icon variant="StackIcon" size={16} />
                    {stateTitle}
                  </>
                ),
              }}
            >
              <TerraformWorkspaceCard
                componentType={isPulumi ? 'pulumi' : 'terraform_module'}
                description="the sandbox"
                hideHeading
              />
            </Panel>
          ) : null}
          <ManagementDropdown />
        </>
      }
      driftBanner={
        driftedObject ? <DriftedBanner drifted={driftedObject} /> : undefined
      }
      config={
        configLoading || sandboxConfig ? (
          <SandboxConfigCard config={sandboxConfig} loading={configLoading} />
        ) : (
          <EmptyState
            variant="table"
            emptyTitle="No sandbox config"
            emptyMessage="The current app config does not define a sandbox."
          />
        )
      }
    />
  )
}
