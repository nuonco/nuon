export default {
  title: 'Workflows/WorkflowDetails/WorkflowMetrics',
}

import { WorkflowMetrics } from './WorkflowMetrics'

const mockWorkflow = {
  id: 'wf-123',
  type: 'deploy',
  approval_option: 'prompt',
  plan_only: false,
  execution_time: 120000000000,
} as any

export const Default = () => (
  <div className="max-w-2xl p-4">
    <WorkflowMetrics
      workflow={mockWorkflow}
      pendingApprovalsCount={2}
      policyViolationsCount={0}
      discardedStepsCount={1}
      completedStepsCount={5}
      totalSteps={8}
    />
  </div>
)

export const DriftScan = () => (
  <div className="max-w-2xl p-4">
    <WorkflowMetrics
      workflow={{ ...mockWorkflow, plan_only: true }}
      pendingApprovalsCount={0}
      policyViolationsCount={3}
      discardedStepsCount={0}
      completedStepsCount={4}
      totalSteps={4}
    />
  </div>
)

export const WithViolations = () => (
  <div className="max-w-2xl p-4">
    <WorkflowMetrics
      workflow={{ ...mockWorkflow, approval_option: 'auto' }}
      pendingApprovalsCount={0}
      policyViolationsCount={2}
      discardedStepsCount={0}
      completedStepsCount={3}
      totalSteps={5}
    />
  </div>
)
