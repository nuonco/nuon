import { DeploymentDetail } from '@/components/installs/DeploymentDetail'
import {
  DeploymentAlerts,
  DeploymentWorkflowContent,
} from '@/components/installs/DeploymentDetail/DeploymentDetailContent'
import type { TDeploymentBranchMeta } from '@/components/installs/DeploymentDetail/DeploymentMeta'
import {
  DEPLOYMENT_TABS,
  deploymentTabOrder,
} from '@/components/installs/DeploymentDetail/deployment-progress'
import { DetailHeader } from '@/components/layout/DetailHeader'
import { DetailPage } from '@/components/layout/DetailPage'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { WorkflowStepsSkeleton } from '@/components/workflows/WorkflowSteps'
import { useWorkflowMetrics } from '@/hooks/use-workflow-metrics'
import { WorkflowContext } from '@/providers/workflow-provider'
import type { TInstallDeploymentRecord, TWorkflow } from '@/types'
import { DEPLOYMENT_PATH } from './fixtures'

export interface IDeploymentDetailV2 {
  branch?: TDeploymentBranchMeta
  deployment?: TInstallDeploymentRecord
  isLoading?: boolean
  workflow?: TWorkflow
}

const WORKFLOW_TAB = 'workflow'

const Loading = () => {
  const tabs = deploymentTabOrder('app_branch_update')
  return (
    <DetailPage
      header={
        <DetailHeader
          backLink={false}
          loading
          loadingWidth={32}
          title="Deployment"
        />
      }
      tabNav={{
        activeIndex: tabs.indexOf(WORKFLOW_TAB),
        basePath: DEPLOYMENT_PATH,
        tabs: tabs.map((key) => DEPLOYMENT_TABS[key]),
      }}
    >
      <div className="flex flex-col gap-4">
        <SectionHeader title="Workflow steps" />
        <WorkflowStepsSkeleton />
      </div>
    </DetailPage>
  )
}

const Loaded = ({
  branch,
  deployment,
  workflow,
}: {
  branch?: TDeploymentBranchMeta
  deployment?: TInstallDeploymentRecord
  workflow: TWorkflow
}) => {
  const metrics = useWorkflowMetrics(workflow)
  return (
    <WorkflowContext.Provider
      value={{ workflow, ...metrics, stopPolling: () => {} }}
    >
      <DeploymentDetail
        activeTabIndex={deploymentTabOrder(workflow.type).indexOf(WORKFLOW_TAB)}
        banners={<DeploymentAlerts />}
        basePath={DEPLOYMENT_PATH}
        branch={branch}
        deployment={deployment}
        workflow={workflow}
      >
        <DeploymentWorkflowContent />
      </DeploymentDetail>
    </WorkflowContext.Provider>
  )
}

export const DeploymentDetailV2 = ({
  branch,
  deployment,
  isLoading,
  workflow,
}: IDeploymentDetailV2) => (
  <div className="mx-auto w-full max-w-6xl p-4 md:p-6">
    {isLoading || !workflow ? (
      <Loading />
    ) : (
      <Loaded branch={branch} deployment={deployment} workflow={workflow} />
    )}
  </div>
)
