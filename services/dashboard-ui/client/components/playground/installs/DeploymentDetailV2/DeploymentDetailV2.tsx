import { useEffect, useState, type ReactNode } from 'react'
import { useLocation, useNavigate } from 'react-router'
import { EmptyState } from '@/components/common/EmptyState'
import { DeploymentDetail } from '@/components/installs/DeploymentDetail'
import {
  DeploymentAlerts,
  DeploymentChangesContent,
  DeploymentWorkflowContent,
} from '@/components/installs/DeploymentDetail/DeploymentDetailContent'
import type { TDeploymentBranchMeta } from '@/components/installs/DeploymentDetail/DeploymentMeta'
import { DetailHeader } from '@/components/layout/DetailHeader'
import { DetailPage } from '@/components/layout/DetailPage'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { WorkflowStepsSkeleton } from '@/components/workflows/WorkflowSteps'
import { useWorkflowMetrics } from '@/hooks/use-workflow-metrics'
import { WorkflowContext } from '@/providers/workflow-provider'
import type { TInstallDeploymentRecord, TWorkflow } from '@/types'
import { DEPLOYMENT_PATH } from './fixtures'
import {
  deploymentV2TabOrder,
  deploymentV2Tabs,
  DEPLOYMENT_V2_TABS,
  type TDeploymentV2Tab,
} from './tabs'

export interface IDeploymentDetailV2 {
  branch?: TDeploymentBranchMeta
  deployment?: TInstallDeploymentRecord
  isLoading?: boolean
  liveView?: ReactNode
  workflow?: TWorkflow
}

const NO_ACTIVE_TAB = -1

const Placeholder = ({
  title,
  message,
}: {
  title: string
  message: string
}) => <EmptyState variant="history" emptyTitle={title} emptyMessage={message} />

const Loading = () => (
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
      activeIndex: NO_ACTIVE_TAB,
      basePath: DEPLOYMENT_PATH,
      tabs: deploymentV2Tabs(),
    }}
  >
    <div className="flex flex-col gap-4">
      <SectionHeader title="Workflow steps" />
      <WorkflowStepsSkeleton />
    </div>
  </DetailPage>
)

const Loaded = ({
  branch,
  deployment,
  liveView,
  workflow,
}: {
  branch?: TDeploymentBranchMeta
  deployment?: TInstallDeploymentRecord
  liveView?: ReactNode
  workflow: TWorkflow
}) => {
  const metrics = useWorkflowMetrics(workflow)
  const { pathname } = useLocation()
  const navigate = useNavigate()
  const [order] = useState(() => deploymentV2TabOrder(workflow))

  useEffect(() => {
    navigate(DEPLOYMENT_PATH, { replace: true })
  }, [])

  const matched = order.findIndex(
    (key) => pathname === `${DEPLOYMENT_PATH}${DEPLOYMENT_V2_TABS[key].path}`
  )
  const activeTabIndex = matched < 0 ? 0 : matched

  const content: Record<TDeploymentV2Tab, ReactNode> = {
    live: liveView ?? (
      <Placeholder
        title="Run view"
        message="Live step output will appear here."
      />
    ),
    template: (
      <Placeholder
        title="Template updates"
        message="Not included in this playground."
      />
    ),
    workflow: <DeploymentWorkflowContent />,
    changes: <DeploymentChangesContent deployment={deployment} />,
  }

  return (
    <WorkflowContext.Provider
      value={{ workflow, ...metrics, stopPolling: () => {} }}
    >
      <DeploymentDetail
        activeTabIndex={activeTabIndex}
        banners={
          <DeploymentAlerts
            showApprovals={!(liveView && order[activeTabIndex] === 'live')}
          />
        }
        basePath={DEPLOYMENT_PATH}
        branch={branch}
        deployment={deployment}
        tabs={order.map((key) => DEPLOYMENT_V2_TABS[key])}
        workflow={workflow}
      >
        {content[order[activeTabIndex]]}
      </DeploymentDetail>
    </WorkflowContext.Provider>
  )
}

export const DeploymentDetailV2 = ({
  branch,
  deployment,
  isLoading,
  liveView,
  workflow,
}: IDeploymentDetailV2) => (
  <div className="w-full">
    {isLoading || !workflow ? (
      <Loading />
    ) : (
      <Loaded
        branch={branch}
        deployment={deployment}
        liveView={liveView}
        workflow={workflow}
      />
    )}
  </div>
)
