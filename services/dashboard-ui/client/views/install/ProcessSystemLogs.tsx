import { useParams } from 'react-router'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { ProcessSystemLogsPanel } from '@/components/runners/ProcessSystemLogsPanel'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'

export const ProcessSystemLogs = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { processId } = useParams<{ processId: string }>()

  return (
    <PageSection flush>
      <PageTitle segments={['System logs', install?.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/runner`,
            text: 'Install runner',
          },
          { path: '', text: 'System logs' },
        ]}
      />
      <ProcessSystemLogsPanel
        runnerId={install?.runner_id}
        processId={processId}
      />
    </PageSection>
  )
}
