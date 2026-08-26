import { Outlet, useParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { InstallActionRunHeader } from '@/components/actions/InstallActionRunHeader'
import { CompositeError } from '@/components/common/CompositeError'
import { DetailPage } from '@/components/layout/DetailPage'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { useInstallActionRun } from '@/hooks/use-install-action-run'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { InstallActionRunProvider } from '@/providers/install-action-run-provider'
import { getWorkflow, getInstallAction } from '@/lib'

const ActionRunLayoutInner = () => {
  const { actionId, actionRunId } = useParams()
  const { org } = useOrg()
  const { install } = useInstall()
  const { installActionRun } = useInstallActionRun()

  const basePath = `/${org?.id}/installs/${install?.id}/actions/${actionId}/runs/${actionRunId}`
  const { data: action } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['action', org?.id, actionId],
    queryFn: () =>
      getInstallAction({
        orgId: org!.id,
        installId: install!.id,
        actionId: actionId!,
      }),
    enabled: !!org?.id && !!install?.id && !!actionId,
  })

  const { data: workflow } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['workflow', org?.id, installActionRun?.install_workflow_id],
    queryFn: () =>
      getWorkflow({
        orgId: org.id,
        workflowId: installActionRun.install_workflow_id,
      }),
    enabled: !!org?.id && !!installActionRun?.install_workflow_id,
  })

  const actionName = action?.action_workflow?.name

  return (
    <>
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/actions`,
            text: 'Actions',
          },
          {
            path: `/${org?.id}/installs/${install?.id}/actions/${actionId}`,
            text: actionName || 'Action',
          },
          {
            path: basePath,
            text: installActionRun?.trigger_type
              ? `${installActionRun.trigger_type} run`
              : 'Run',
          },
        ]}
      />
      <DetailPage
        header={
          <InstallActionRunHeader
            action={action}
            actionId={actionId!}
            actionName={actionName ?? ''}
            workflow={workflow}
          />
        }
        banners={
          installActionRun?.composite_error ? (
            <CompositeError error={installActionRun.composite_error} />
          ) : null
        }
        tabNav={{
          basePath,
          tabs: [
            { path: '/', text: 'Summary' },
            { path: '/logs', text: 'Logs' },
            ...(org?.features?.['trace-view']
              ? [{ path: '/trace', text: 'Trace' }]
              : []),
          ],
        }}
      >
        <Outlet />
      </DetailPage>
    </>
  )
}

export const ActionRunLayout = () => {
  const { actionRunId } = useParams()

  return (
    <InstallActionRunProvider runId={actionRunId!} shouldPoll>
      <ActionRunLayoutInner />
    </InstallActionRunProvider>
  )
}
