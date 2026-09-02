import { DriftedBanner } from '@/components/install-components/DriftedBanner'
import { SandboxRunsTimeline } from '@/components/sandbox/SandboxRunsTimeline'
import { ManagementDropdown } from '@/components/sandbox/management/ManagementDropdown'
import { SandboxConfigCard } from '@/components/sandbox/SandboxConfigCard'
import { TerraformWorkspaceCard } from '@/components/terraform-workspace/TerraformWorkspaceCard'
import { DetailHeader } from '@/components/layout/DetailHeader'
import { DetailPage } from '@/components/layout/DetailPage'
import {
  HistoryPanelButton,
  HistoryRail,
} from '@/components/layout/HistoryRail'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useInstallAppConfig } from '@/hooks/use-install-app-config'
import { useOrg } from '@/hooks/use-org'

export const Sandbox = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  const { appConfig: configResult } = useInstallAppConfig()

  const sandboxConfig = configResult?.sandbox
  const isPulumi = sandboxConfig?.type === 'pulumi'

  const latestSandboxRunId = install?.install_sandbox_runs?.at(0)?.id
  const driftedObject = install?.drifted_objects?.find(
    (d) =>
      d?.target_type === 'install_sandbox_run' &&
      d?.target_id === latestSandboxRunId
  )

  const history = <SandboxRunsTimeline shouldPoll />

  return (
    <>
      <PageTitle segments={['Sandbox', install?.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/sandbox`,
            text: 'Sandbox',
          },
        ]}
      />

      <DetailPage
        header={
          <DetailHeader
            backLink={false}
            title="Sandbox details"
            id={install?.sandbox?.id}
            actions={
              <>
                <HistoryPanelButton title="Sandbox history" history={history} />
                <ManagementDropdown />
              </>
            }
          />
        }
      >
        <HistoryRail title="Sandbox history" history={history}>
          {driftedObject ? <DriftedBanner drifted={driftedObject} /> : null}

          <SandboxConfigCard config={sandboxConfig} loading={!sandboxConfig} />

          <TerraformWorkspaceCard
            componentType={isPulumi ? 'pulumi' : 'terraform_module'}
          />
        </HistoryRail>
      </DetailPage>
    </>
  )
}
