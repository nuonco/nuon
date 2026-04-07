export default {
  title: 'Workflows/WorkflowDetails/WorkflowHeader',
}

import { WorkflowHeader } from './WorkflowHeader'
import { WorkflowContext } from '@/providers/workflow-provider'

const mockWorkflow = {
  id: 'wf-123',
  type: 'deploy',
  name: 'Deploy workflow',
  owner_id: 'owner-123',
  status: 'running',
  metadata: {},
} as any

const mockInstall = {
  app_id: 'app-123',
  app: { name: 'My App' },
  drifted_objects: [],
} as any

const mockWorkflowContext = {
  workflow: mockWorkflow,
  stopPolling: () => {},
  workflowSteps: [],
  hasApprovals: false,
  failedSteps: [],
  pendingApprovals: [],
  discardedSteps: [],
  completedSteps: [],
  stepsWithPolicyViolations: [],
  totalSteps: 0,
  pendingApprovalsCount: 0,
  discardedStepsCount: 0,
  completedStepsCount: 0,
  failedStepsCount: 0,
  policyViolationsCount: 0,
}

const Wrapper = ({ children }: { children: React.ReactNode }) => (
  <WorkflowContext.Provider value={mockWorkflowContext}>
    {children}
  </WorkflowContext.Provider>
)

export const Default = () => (
  <Wrapper>
    <div className="max-w-2xl p-4">
      <WorkflowHeader workflow={mockWorkflow} install={mockInstall} />
    </div>
  </Wrapper>
)

export const WithDrift = () => (
  <Wrapper>
    <div className="max-w-2xl p-4">
      <WorkflowHeader
        workflow={mockWorkflow}
        install={{
          ...mockInstall,
          drifted_objects: [{ install_workflow_id: 'wf-123' }],
        }}
      />
    </div>
  </Wrapper>
)

export const AdhocAction = () => (
  <Wrapper>
    <div className="max-w-2xl p-4">
      <WorkflowHeader
        workflow={{
          ...mockWorkflow,
          type: 'action_workflow_run',
          metadata: {
            adhoc_action: true,
            install_action_workflow_name: 'restart-service',
          },
        }}
        install={mockInstall}
      />
    </div>
  </Wrapper>
)

export const WithApprovals = () => (
  <WorkflowContext.Provider value={{ ...mockWorkflowContext, hasApprovals: true, pendingApprovalsCount: 2 }}>
    <div className="max-w-2xl p-4">
      <WorkflowHeader workflow={mockWorkflow} install={mockInstall} />
    </div>
  </WorkflowContext.Provider>
)
