import { useState } from 'react'
import { Outlet, useLocation, useParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { DeploymentDetail } from '@/components/installs/DeploymentDetail'
import { DeploymentAlerts } from '@/components/installs/DeploymentDetail/DeploymentDetailContent'
import {
  DEPLOYMENT_TABS,
  deploymentTabOrder,
} from '@/components/installs/DeploymentDetail/deployment-progress'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { useInstallPage } from '@/hooks/use-install-path'
import { useWorkflow } from '@/hooks/use-workflow'
import { getInstallDeployments } from '@/lib'
import { WorkflowProvider } from '@/providers/workflow-provider'
import { humanize } from '@/utils/string-utils'

const DeploymentDetailLayoutContent = () => {
  const { workflowId } = useParams()
  const { pathname, search } = useLocation()
  const { org, install, href } = useInstallPage()
  const { workflow } = useWorkflow()
  const [tabOrder] = useState(() => deploymentTabOrder(workflow.status?.status))

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
      return response.deployments.find((item) => item.id === workflowId) ?? null
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
  const basePath = href(`/deployments/${workflowId}`)
  const activeTabIndex = tabOrder.findIndex(
    (key) =>
      pathname ===
      `${basePath}${DEPLOYMENT_TABS[key].path === '/' ? '' : DEPLOYMENT_TABS[key].path}`
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
        activeTabIndex={activeTabIndex}
        basePath={basePath}
        branchHref={branchHref}
        deployment={deployment ?? undefined}
        search={search}
        tabOrder={tabOrder}
        workflow={workflow}
        banners={<DeploymentAlerts />}
      >
        <Outlet
          context={{ deployment: deployment ?? undefined, deploymentFetched }}
        />
      </DeploymentDetail>
    </>
  )
}

export const DeploymentDetailLayout = () => {
  const { workflowId } = useParams()

  return (
    <WorkflowProvider workflowId={workflowId!} shouldPoll>
      <DeploymentDetailLayoutContent key={workflowId} />
    </WorkflowProvider>
  )
}
