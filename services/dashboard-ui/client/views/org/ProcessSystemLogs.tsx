import { useParams } from 'react-router'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { PageLayout } from '@/components/layout/PageLayout'
import { ProcessSystemLogsPanel } from '@/components/runners/ProcessSystemLogsPanel'
import { useOrg } from '@/hooks/use-org'

export const ProcessSystemLogs = () => {
  const { org } = useOrg()
  const { processId } = useParams<{ processId: string }>()
  const runnerId = org?.runner_group?.runners?.[0]?.id

  return (
    <PageLayout>
      <PageTitle title="System logs" />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/runner`, text: 'Build runner' },
          {
            path: `/${org?.id}/runner/processes/${processId}/logs`,
            text: 'System logs',
          },
        ]}
      />
      <ProcessSystemLogsPanel runnerId={runnerId} processId={processId} />
    </PageLayout>
  )
}
