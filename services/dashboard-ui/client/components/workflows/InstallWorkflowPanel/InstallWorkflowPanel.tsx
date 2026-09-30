import type { ReactNode } from 'react'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { WorkflowDetails } from '@/components/workflows/WorkflowDetails'
import { WorkflowSteps } from '@/components/workflows/WorkflowSteps'
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

  return (
    <div className="flex flex-col gap-6">
      <WorkflowDetails />
      <div className="flex flex-col gap-4">
        <SectionHeader title="Workflow steps" />
        <WorkflowSteps
          approvalPrompt={workflow.approval_option === 'prompt'}
          planOnly={workflow.plan_only}
        />
      </div>
    </div>
  )
}
