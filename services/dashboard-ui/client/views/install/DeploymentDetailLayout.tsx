import { Outlet, useParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { ApprovalBanner } from '@/components/approvals/ApprovalBanner'
import { DeploymentDetail } from '@/components/installs/DeploymentDetail'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { WorkflowAlertBanners } from '@/components/workflows/WorkflowDetails'
import { useInstallPage } from '@/hooks/use-install-path'
import { useRespondedApprovals } from '@/hooks/use-responded-approvals'
import { useWorkflow } from '@/hooks/use-workflow'
import { getInstallDeployments } from '@/lib'
import { WorkflowProvider } from '@/providers/workflow-provider'
import type { TWorkflowStep } from '@/types'
import { humanize } from '@/utils/string-utils'

const awaitingApproval = (
  step: TWorkflowStep,
  hasResponded: (id: string) => boolean
) => {
  const status = step?.status?.status
  const terminal =
    status === 'error' ||
    status === 'cancelled' ||
    status === 'discarded' ||
    status === 'approval-expired' ||
    status === 'approval-denied'
  return (
    !!step?.approval?.type &&
    step.approval.type !== 'approve-all' &&
    step.approval.type !== 'noop' &&
    !step.approval.response &&
    status !== 'auto-skipped' &&
    !terminal &&
    !!step.id &&
    !hasResponded(step.id)
  )
}

const DeploymentDetailLayoutContent = () => {
  const { workflowId } = useParams()
  const { org, install, href } = useInstallPage()
  const { workflow, failedSteps, pendingApprovals } = useWorkflow()
  const { hasResponded } = useRespondedApprovals()

  const { data: deployment, isFetched: deploymentFetched } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['install-deployment', org?.id, install?.id, workflowId],
    queryFn: async () => {
      const response = await getInstallDeployments({
        orgId: org!.id,
        installId: install!.id,
        search: workflowId,
        limit: 20,
      })
      return response.deployments.find((item) => item.id === workflowId)
    },
    enabled: !!org?.id && !!install?.id && !!workflowId,
  })

  const title =
    deployment?.title ||
    workflow?.name ||
    humanize(workflow?.type) ||
    'Deployment'
  const installRoot = `/${org?.id}/installs/${install?.id}`
  const page = `${installRoot}/deployments/${workflowId}`
  const branch = deployment?.app_branch
  const branchHref = branch
    ? `/${org?.id}/apps/${install?.app_id}/branches/${branch.id}`
    : undefined
  const approvals = (pendingApprovals as TWorkflowStep[]).filter((step) =>
    awaitingApproval(step, hasResponded)
  )

  return (
    <>
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: installRoot, text: install?.name },
          { path: `${installRoot}/deployments`, text: 'Deployments' },
          { path: page, text: title },
        ]}
      />
      <DeploymentDetail
        basePath={href(`/deployments/${workflowId}`)}
        branchHref={branchHref}
        deployment={deployment}
        workflow={workflow}
        banners={
          <>
            <WorkflowAlertBanners
              workflow={workflow}
              failedSteps={failedSteps}
            />
            {approvals.map((step) => (
              <ApprovalBanner key={step.id} step={step} />
            ))}
          </>
        }
      >
        <Outlet context={{ deployment, deploymentFetched }} />
      </DeploymentDetail>
    </>
  )
}

export const DeploymentDetailLayout = () => {
  const { workflowId } = useParams()

  return (
    <WorkflowProvider workflowId={workflowId!} shouldPoll>
      <DeploymentDetailLayoutContent />
    </WorkflowProvider>
  )
}
