import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { ApprovalBanner } from '@/components/approvals/ApprovalBanner'
import { Plan } from '@/components/approvals/Plan'
import { DeployHeader } from '@/components/deploys/DeployHeader'
import { SSELogs, LogsSkeleton } from '@/components/log-stream/SSELogs'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { DeployProvider } from '@/providers/deploy-provider'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import { LogViewerProvider } from '@/providers/log-viewer-provider'
import { useDeploy } from '@/hooks/use-deploy'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useRespondedApprovals } from '@/hooks/use-responded-approvals'
import { getComponent, getWorkflow } from '@/lib'

export const DeployDetail = () => {
  const { componentId, deployId, installId } = useParams()

  return (
    <DeployProvider deployId={deployId!} installId={installId!} shouldPoll>
      <DeployDetailContent componentId={componentId!} />
    </DeployProvider>
  )
}

const DeployDetailContent = ({ componentId }: { componentId: string }) => {
  const { deploy } = useDeploy()
  const { install } = useInstall()
  const { org } = useOrg()

  const { data: component } = useQuery({
    queryKey: ['component', org?.id, componentId],
    queryFn: () => getComponent({ orgId: org.id, componentId }),
    enabled: !!org?.id && !!componentId,
  })

  const { data: workflow } = useQuery({
    queryKey: ['workflow', org?.id, deploy?.install_workflow_id],
    queryFn: () => getWorkflow({ orgId: org.id, workflowId: deploy!.install_workflow_id }),
    enabled: !!org?.id && !!deploy?.install_workflow_id,
  })

  const { hasResponded } = useRespondedApprovals()

  if (!deploy || !component) return null

  const step = workflow?.steps
    ?.filter(
      (s) => s?.step_target_id === deploy?.id && s?.execution_type === 'approval'
    )
    ?.at(-1) ?? null
  const responded = step ? hasResponded(step.id) : false
  const logStream = deploy?.log_stream
  const stepStatus = step?.status?.status
  const isTerminal = stepStatus === 'error' || stepStatus === 'cancelled' || stepStatus === 'discarded'
  const isAutoApprove =
    step?.approval?.type === 'approve-all' ||
    step?.approval?.response?.type === 'auto-approve'
  const pendingApproval =
    step?.approval && !step?.approval?.response && !responded && !isTerminal && stepStatus !== 'auto-skipped'
  const completedApproval =
    step?.approval && (!!step?.approval?.response || responded) && !isTerminal && stepStatus !== 'auto-skipped'
  const showPlanBelow = completedApproval || isAutoApprove

  return (
    <PageSection>
      <PageTitle title={`Deploy | ${install?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          { path: `/${org?.id}/installs/${install?.id}/components`, text: 'Components' },
          {
            path: `/${org?.id}/installs/${install?.id}/components/${componentId}`,
            text: deploy?.component_name,
          },
          {
            path: `/${org?.id}/installs/${install?.id}/components/${componentId}/deploys/${deploy?.id}`,
            text: 'Deploy',
          },
        ]}
      />

      <DeployHeader component={component} workflow={workflow} stepId={step?.id} />

      {pendingApproval && !isAutoApprove ? (
        <div className="flex flex-col gap-4">
          <ApprovalBanner step={step} />
          <Plan step={step} />
        </div>
      ) : null}

      {logStream ? (
        <LogStreamProvider logStreamId={logStream.id}>
          <LogViewerProvider>
            <SSELogs />
          </LogViewerProvider>
        </LogStreamProvider>
      ) : (
        <LogsSkeleton />
      )}

      {showPlanBelow && step ? (
        <div className="flex flex-col gap-4">
          {!isAutoApprove && <ApprovalBanner step={step} />}
          <Plan step={step} />
        </div>
      ) : null}
    </PageSection>
  )
}
