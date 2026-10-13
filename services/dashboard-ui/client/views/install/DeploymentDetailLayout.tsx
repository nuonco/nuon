import { useState } from 'react'
import { Outlet, useLocation, useParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { DeploymentDetail } from '@/components/installs/DeploymentDetail'
import { DeploymentAlerts } from '@/components/installs/DeploymentDetail/DeploymentDetailContent'
import { useDeploymentRunBranch } from '@/components/installs/DeploymentDetail/DeploymentMeta'
import {
  DEPLOYMENT_TABS,
  deploymentTabOrder,
} from '@/components/installs/DeploymentDetail/deployment-progress'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { useInstallPage } from '@/hooks/use-install-path'
import { useWorkflow } from '@/hooks/use-workflow'
import { getInstallDeployment } from '@/lib'
import { WorkflowProvider } from '@/providers/workflow-provider'
import { humanize } from '@/utils/string-utils'

const DeploymentDetailLayoutContent = () => {
  const { workflowId } = useParams()
  const { pathname, search } = useLocation()
  const { org, install, href } = useInstallPage()
  const { workflow } = useWorkflow()
  const [tabOrder] = useState(() => deploymentTabOrder(workflow.type))

  const {
    data: deployment,
    isFetched: deploymentFetched,
    isError: deploymentError,
    refetch: refetchDeployment,
  } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['install-deployment', org?.id, install?.id, workflowId],
    queryFn: () =>
      getInstallDeployment({
        orgId: org!.id,
        installId: install!.id,
        workflowId: workflowId!,
      }),
    enabled: !!org?.id && !!install?.id && !!workflowId,
  })

  const title =
    deployment?.title ||
    workflow?.name ||
    humanize(workflow?.type) ||
    'Deployment'
  const installRoot = `/${org?.id}/installs/${install?.id}`
  const page = `${installRoot}/deployments/${workflowId}`
  const branch = useDeploymentRunBranch({
    orgId: org?.id,
    appId: install?.app_id,
    type: deployment?.type || workflow?.type,
    deployment,
    workflow,
  })
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
        branch={branch}
        deployment={deployment ?? undefined}
        search={search}
        tabOrder={tabOrder}
        workflow={workflow}
        banners={<DeploymentAlerts />}
      >
        <Outlet
          context={{
            deployment: deployment ?? undefined,
            deploymentFetched,
            deploymentError,
            refetchDeployment,
          }}
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
