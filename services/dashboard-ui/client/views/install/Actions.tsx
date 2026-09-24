import { InstallActionsTable } from '@/components/actions/InstallActionsTable'
import { InstallCronOfflineBanner } from '@/components/installs/InstallCronOfflineBanner'
import { RunAdhocActionButton } from '@/components/installs/management/RunAdhocAction'
import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useInstallAppConfig } from '@/hooks/use-install-app-config'
import { useOrg } from '@/hooks/use-org'

export const Actions = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { appConfig } = useInstallAppConfig()

  const hasCronSchedule = !!appConfig?.action_workflow_configs?.some((config) =>
    config.triggers?.some((trigger) => trigger.type === 'cron')
  )

  return (
    <PageSection>
      <PageTitle segments={['Actions', install?.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/actions`,
            text: 'Actions',
          },
        ]}
      />
      <SectionHeader
        title="Actions"
        description="View and manage all actions for this install."
        actions={<RunAdhocActionButton />}
      />

      <InstallCronOfflineBanner
        runnerStatus={install?.runner_status}
        kind="action"
        hasCronSchedule={hasCronSchedule}
      />

      <InstallActionsTable shouldPoll />
    </PageSection>
  )
}
