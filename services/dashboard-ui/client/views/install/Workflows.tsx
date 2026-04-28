import { useSearchParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { AutoApproveToggle } from '@/components/installs/management/EnableAutoApprove'
import { PageHeader } from '@/components/layout/PageHeader'
import { PageHeadingGroup } from '@/components/layout/PageHeadingGroup'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { ActiveWorkflows } from '@/components/workflows/ActiveWorkflows'
import { WorkflowTimeline } from '@/components/workflows/WorkflowTimeline'
import { ShowDriftScanContainer as ShowDriftScan } from '@/components/workflows/filters/ShowDriftScans'
import { WorkflowTypeFilter } from '@/components/workflows/filters/WorkflowTypeFilter'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallWorkflows } from '@/lib'

const POLL_INTERVAL = 20000

export const Workflows = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const [searchParams] = useSearchParams()

  const type = searchParams.get('type') || ''
  const showDrifts = searchParams.get('drifts') !== 'false'

  const { data } = useQuery({
    queryKey: ['install-active-workflows', org?.id, install?.id],
    queryFn: () =>
      getInstallWorkflows({
        orgId: org.id,
        installId: install!.id,
        finished: false,
        limit: 50,
        offset: 0,
      }),
    refetchInterval: POLL_INTERVAL,
    enabled: !!org?.id && !!install?.id,
  })

  const activeWorkflows = (data?.data ?? []).filter(
    (w) =>
      w.status?.status &&
      w.status.status !== 'pending' &&
      w.status.status !== 'queued'
  )

  return (
    <PageSection>
      <PageTitle title={`Workflows | ${install?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/workflows`,
            text: 'Workflows',
          },
        ]}
      />

      <ActiveWorkflows
        workflows={activeWorkflows}
        install={install}
      />

      <PageHeader flush>
        <PageHeadingGroup
          title="Workflow history"
          subtitle="View past and active workflows for this install."
          titleProps={{ variant: 'base', weight: 'strong' }}
          headingLevel={2}
        />
        <div className="flex items-center gap-4">
          <AutoApproveToggle />
          <ShowDriftScan />
          <WorkflowTypeFilter />
        </div>
      </PageHeader>

      <WorkflowTimeline
        installId={install?.id}
        shouldPoll
        planonly={showDrifts}
        type={type}
      />
    </PageSection>
  )
}
