import type { ReactNode } from 'react'
import { ID } from '@/components/common/ID'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import {
  DeploymentAlerts,
  DeploymentWorkflowContent,
} from '@/components/installs/DeploymentDetail/DeploymentDetailContent'
import { DeploymentRunStatus } from '@/components/installs/DeploymentDetail/DeploymentProgress'
import { deploymentSteps } from '@/components/installs/DeploymentDetail/deployment-progress'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { WorkflowStepsSkeleton } from '@/components/workflows/WorkflowSteps'
import { useWorkflow } from '@/hooks/use-workflow'

export const InstallWorkflowPanel = ({
  children,
  ...props
}: Partial<IPanel> & { children: ReactNode }) => (
  <Panel heading="Workflow details" size="3/4" {...props}>
    {children}
  </Panel>
)

export const InstallWorkflowPanelContent = () => {
  const { workflow } = useWorkflow()
  const status = workflow?.status?.status ?? ''
  const activity = workflow?.status?.status_human_description

  return (
    <>
      <div className="flex flex-wrap items-center gap-3">
        <DeploymentRunStatus
          run={{ status, steps: deploymentSteps(workflow) }}
        />
        <ID>{workflow.id}</ID>
        <Time time={workflow.created_at} format="relative" variant="subtext" />
      </div>
      {activity && status !== 'error' ? (
        <Text variant="subtext" theme="neutral">
          {activity}
        </Text>
      ) : null}
      <DeploymentAlerts />
      <div className="mt-4">
        <DeploymentWorkflowContent />
      </div>
    </>
  )
}

export const InstallWorkflowPanelLoading = ({
  status,
  createdAt,
}: {
  status?: string
  createdAt?: string
}) => (
  <>
    <div className="flex flex-wrap items-center gap-3">
      {status ? (
        <Status status={status} variant="badge" />
      ) : (
        <Status loading variant="badge" loadingWidth={8} />
      )}
      <ID loading loadingWidth={16} />
      {createdAt ? (
        <Time time={createdAt} format="relative" variant="subtext" />
      ) : (
        <Time loading format="relative" variant="subtext" loadingWidth={10} />
      )}
    </div>
    <div className="mt-4 flex flex-col gap-4">
      <SectionHeader title="Workflow steps" />
      <WorkflowStepsSkeleton />
    </div>
  </>
)
