import { useSearchParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { AutoApproveToggle } from '@/components/installs/management/EnableAutoApprove'
import { ActiveWorkflows } from '@/components/workflows/ActiveWorkflows'
import { WorkflowTimeline } from '@/components/workflows/WorkflowTimeline'
import { ShowDriftScanContainer as ShowDriftScan } from '@/components/workflows/filters/ShowDriftScans'
import { WorkflowTypeFilter } from '@/components/workflows/filters/WorkflowTypeFilter'
import { WorkflowSearch } from '@/components/workflows/filters/WorkflowSearch'
import { RunAdhocActionButton } from '@/components/installs/management/RunAdhocAction'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallWorkflows } from '@/lib'

const POLL_INTERVAL = 20000

export const Workflows = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const [searchParams] = useSearchParams()

  const type = searchParams.get('type') || ''
  const search = searchParams.get('search') || ''
  const showDrifts = searchParams.get('drifts') !== 'false'

  const { data } = useQuery({
    queryKey: ['install-active-workflows', org?.id, install?.id],
    queryFn: () =>
      getInstallWorkflows({
        orgId: org.id,
        installId: install!.id,
        finished: false,
        planonly: false,
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

      <div className="flex items-start justify-between gap-4">
        <HeadingGroup>
          <Text variant="base" weight="strong">
            Workflow history
          </Text>
          <Text variant="subtext" theme="neutral">
            View past and active workflows for this install.
          </Text>
        </HeadingGroup>
        <div className="shrink-0">
          <RunAdhocActionButton />
        </div>
      </div>

      <div className="flex items-center justify-between gap-4">
        <WorkflowSearch />

        <div className="shrink-0 flex items-center gap-4">
          <AutoApproveToggle />
          <ShowDriftScan />
          <WorkflowTypeFilter />
        </div>
      </div>

      <WorkflowTimeline
        installId={install?.id}
        shouldPoll
        planonly={showDrifts}
        type={type}
        search={search}
      />
    </PageSection>
  )
}
