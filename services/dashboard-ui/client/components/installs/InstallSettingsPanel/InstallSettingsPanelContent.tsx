import type { ReactNode } from 'react'
import { Card } from '@/components/common/Card'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { ShutdownRunnerControl } from '@/components/runners/management/ShutdownRunnerControl'
import { ReprovisionSandboxButton } from '@/components/sandbox/management/ReprovisionSandbox'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { RunnerProvider } from '@/providers/runner-provider'
import { RunAdhocActionButton } from '@/components/installs/management/RunAdhocAction/RunAdhocActionContainer'
import { AuditHistoryButton } from '@/components/installs/management/AuditHistory'
import { DeprovisionButton } from '@/components/installs/management/Deprovision'
import { DeprovisionStackButton } from '@/components/installs/management/DeprovisionStack'
import { EditInputsButton } from '@/components/installs/management/EditInputs'
import { EditLabelsButton } from '@/components/installs/management/EditLabels'
import { EnableConfigSyncButton } from '@/components/installs/management/EnableConfigSync'
import { ForgetButton } from '@/components/installs/management/Forget'
import { GenerateInstallConfigButton } from '@/components/installs/management/GenerateInstallConfig'
import { ReprovisionButton } from '@/components/installs/management/Reprovision'
import { ReprovisionStackButton } from '@/components/installs/management/ReprovisionStack'
import { SyncSecretsButton } from '@/components/installs/management/SyncSecrets'

const Section = ({
  label,
  children,
}: {
  label: string
  children: ReactNode
}) => (
  <section className="flex flex-col gap-3">
    <Text theme="neutral" weight="strong">
      {label}
    </Text>
    <div className="grid grid-cols-1 @lg:grid-cols-2 @4xl:grid-cols-3 @6xl:grid-cols-4 gap-3">
      {children}
    </div>
  </section>
)

const ActionCard = ({
  title,
  description,
  children,
}: {
  title: string
  description: string
  children: ReactNode
}) => (
  <Card className="!p-4 !gap-4">
    <HeadingGroup className="gap-1">
      <Text weight="strong">{title}</Text>
      <Text variant="subtext" theme="neutral">
        {description}
      </Text>
    </HeadingGroup>
    <div className="mt-auto">{children}</div>
  </Card>
)

const InstallSettingsPanelContentInner = () => {
  const { install } = useInstall()
  const { org } = useOrg()
  const canRenameInstall = !!org?.features?.['install-rename']

  return (
    <div className="@container flex flex-col gap-6">
      <Section label="Configuration">
        <ActionCard
          title="Install"
          description="Edit this install's inputs and settings."
        >
          <EditInputsButton showNameField={canRenameInstall} />
        </ActionCard>
        <ActionCard
          title="Labels"
          description="Add or update key-value labels to organize and target this install."
        >
          <EditLabelsButton />
        </ActionCard>
        <ActionCard
          title="Config sync"
          description="Sync settings from the install config file on every deploy."
        >
          <EnableConfigSyncButton />
        </ActionCard>
        <ActionCard
          title="Install config"
          description="Generate a config file from this install's current settings."
        >
          <GenerateInstallConfigButton />
        </ActionCard>
      </Section>

      <Section label="Controls">
        <ActionCard
          title="Adhoc action"
          description="Run a one-off command or script against this install."
        >
          <RunAdhocActionButton />
        </ActionCard>
        <ActionCard
          title="Audit history"
          description="See a complete record of activity on this install."
        >
          <AuditHistoryButton />
        </ActionCard>
        <ActionCard
          title="Reprovision"
          description="Recreate all resources and redeploy every component. This causes downtime."
        >
          <ReprovisionButton />
        </ActionCard>
        <ActionCard
          title="Stack"
          description="Recreate the stack and runner for this install."
        >
          <ReprovisionStackButton />
        </ActionCard>
        <ActionCard
          title="Sandbox"
          description="Recreate all sandbox resources for this install."
        >
          <ReprovisionSandboxButton />
        </ActionCard>
        {install?.runner_id ? (
          <ActionCard
            title="Runner process"
            description="Restart the runner process. Queued jobs finish first where possible."
          >
            <ShutdownRunnerControl
              showRunnerLabel
              isManaged={false}
              runnerId={install.runner_id}
            />
          </ActionCard>
        ) : null}
        <ActionCard
          title="Secrets"
          description="Sync all secrets from the app config to this install's environment."
        >
          <SyncSecretsButton />
        </ActionCard>
      </Section>

      <Section label="Danger">
        <ActionCard
          title="Deprovision"
          description="Remove this install from the cloud account, tearing down its components and sandbox."
        >
          <DeprovisionButton />
        </ActionCard>
        <ActionCard
          title="Deprovision stack"
          description="After deprovisioning, destroy the install's stack in the cloud console."
        >
          <DeprovisionStackButton />
        </ActionCard>
        <ActionCard
          title="Forget"
          description="Remove this install from Nuon after deprovisioning. This cannot be undone."
        >
          <ForgetButton isMenuButton={false} />
        </ActionCard>
      </Section>
    </div>
  )
}

export const InstallSettingsPanelContent = () => {
  const { install } = useInstall()
  if (!install?.runner_id) return <InstallSettingsPanelContentInner />
  return (
    <RunnerProvider runnerId={install.runner_id}>
      <InstallSettingsPanelContentInner />
    </RunnerProvider>
  )
}
