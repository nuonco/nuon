import { Outlet, useParams } from 'react-router-dom'
import { usePolling } from '@/hooks/use-polling'
import { useQuery } from '@/hooks/use-query'
import { useOrg } from '@/hooks/use-org'
import { InstallActionRunHeader } from '@/components/actions/InstallActionRunHeader'
import { BackToTop } from '@/components/common/BackToTop'
import { PageSection } from '@/components/layout/PageSection'
import { TabNav } from '@/components/navigation/TabNav'
import { InstallActionRunProvider } from '@/providers/install-action-run-provider'
import type { TInstallActionRun, TInstallAction, TWorkflow } from '@/types'

export default function InstallActionRunLayout() {
  const { orgId, installId, actionId, runId } = useParams()
  const { org } = useOrg()

  const { data: installActionRun, isLoading: isLoadingRun } = usePolling<TInstallActionRun>({
    path: `/api/ctl-api/v1/installs/${installId}/actions/${actionId}/runs/${runId}`,
    shouldPoll: true,
    pollInterval: 30000,
  })

  const { data: installAction, isLoading: isLoadingAction } = useQuery<TInstallAction>({
    path: `/api/ctl-api/v1/installs/${installId}/actions/${actionId}`,
  })

  const { data: workflow } = useQuery<TWorkflow>({
    path: `/api/ctl-api/v1/workflows/${installActionRun?.install_workflow_id}`,
    enabled: !!installActionRun?.install_workflow_id,
    dependencies: [installActionRun?.install_workflow_id],
  })

  if (isLoadingRun || isLoadingAction) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-gray-900" />
      </div>
    )
  }

  const containerId = 'action-run-page'
  return (
    <InstallActionRunProvider
      initInstallActionRun={installActionRun || null}
      shouldPoll
    >
      <PageSection id={containerId} isScrollable>
        <InstallActionRunHeader
          actionId={actionId || ''}
          actionName={installAction?.action_workflow?.name}
          workflow={workflow || null}
        />
        <TabNav
          basePath={`/${orgId}/installs/${installId}/actions/${actionId}/${runId}`}
          tabs={[
            { text: 'Summary', path: '/' },
            { text: 'Logs', path: '/logs' },
          ]}
        />
        <Outlet />
        <BackToTop containerId={containerId} />
      </PageSection>
    </InstallActionRunProvider>
  )
}
