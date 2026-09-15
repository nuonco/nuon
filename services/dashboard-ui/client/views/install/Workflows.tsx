import { useMemo } from 'react'
import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { AutoApproveToggle } from '@/components/installs/management/EnableAutoApprove'
import { ActiveWorkflows } from '@/components/workflows/ActiveWorkflows'
import { WorkflowTimeline } from '@/components/workflows/WorkflowTimeline'
import { WorkflowFilters } from '@/components/workflows/filters/WorkflowFilters'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useSimpleIA } from '@/hooks/use-simple-ia'
import { getInstallWorkflows } from '@/lib'
import {
  datePresetQueryParameter,
  readWorkflowFilters,
} from '@/utils/workflow-filters'

const POLL_INTERVAL = 20000

export const Workflows = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const [searchParams] = useSearchParams()
  const filters = readWorkflowFilters(searchParams, 'install')
  const since = searchParams.get('since')
  const createdAtGte = useMemo(() => datePresetQueryParameter(since), [since])
  const hasSimpleIA = useSimpleIA()
  const pageName = hasSimpleIA ? 'Activity' : 'Workflows'
  const pagePath = hasSimpleIA ? 'activity' : 'workflows'

  const { data } = useQuery({
    placeholderData: keepPreviousData,
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
      <PageTitle segments={[pageName, install?.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/${pagePath}`,
            text: pageName,
          },
        ]}
      />

      <ActiveWorkflows workflows={activeWorkflows} install={install} />

      <SectionHeader
        title={hasSimpleIA ? 'Activity' : 'Workflow history'}
        description="View past and active workflows for this install."
      />

      <div className="flex items-center justify-between gap-4">
        <WorkflowFilters owner="install" />
        <div className="shrink-0">
          <AutoApproveToggle />
        </div>
      </div>

      <WorkflowTimeline
        installId={install?.id}
        shouldPoll
        planonly
        type={filters.api.type}
        status={filters.api.status}
        search={filters.api.search}
        createdAtGte={createdAtGte}
        isFiltered={filters.filtered}
      />
    </PageSection>
  )
}
