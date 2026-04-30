export default {
  title: 'Workflows/WorkflowSteps',
}

import { WorkflowSteps, WorkflowStepsSkeleton } from './WorkflowSteps'
import { WorkflowContext } from '@/providers/workflow-provider'
import type { TWorkflowStep } from '@/types'

const mockWorkflow = {
  id: 'wf-1',
  owner_id: 'inst-1',
  type: 'deploy',
  status: { status: 'in-progress' },
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

const mockStep: TWorkflowStep = {
  id: 'step-1',
  name: 'deploy component',
  execution_type: 'system',
  step_target_type: 'install_deploys',
  step_target_id: 'deploy-1',
  install_workflow_id: 'wf-1',
  owner_id: 'inst-1',
  finished: false,
  started_at: '2024-01-01T00:00:00Z',
  execution_time: 0,
  status: { status: 'in-progress', history: [] },
} as TWorkflowStep

const completedStep: TWorkflowStep = {
  ...mockStep,
  id: 'step-2',
  name: 'provision runner',
  step_target_type: 'runners',
  finished: true,
  execution_time: 60000000000,
  status: { status: 'success', history: [] },
} as TWorkflowStep

export const Default = () => (
  <WorkflowContext.Provider value={mockWorkflowContext}>
    <WorkflowSteps workflowSteps={[mockStep, completedStep]} />
  </WorkflowContext.Provider>
)

export const Empty = () => (
  <WorkflowContext.Provider value={mockWorkflowContext}>
    <WorkflowSteps workflowSteps={[]} />
  </WorkflowContext.Provider>
)

export const Loading = () => <WorkflowStepsSkeleton />

const provisionSteps: TWorkflowStep[] = [
  {
    id: 'step-provision-1',
    name: 'provision runner service account',
    execution_type: 'system',
    step_target_type: 'runners',
    step_target_id: 'runner-1',
    install_workflow_id: 'wf-provision-1',
    owner_id: 'inst-1',
    finished: true,
    started_at: '2024-01-01T00:00:00Z',
    execution_time: 2000000000,
    status: { status: 'success', history: [] },
  } as TWorkflowStep,
  {
    id: 'step-provision-2',
    name: 'generate install stack',
    execution_type: 'system',
    step_target_type: 'install_stacks',
    step_target_id: 'stack-1',
    install_workflow_id: 'wf-provision-1',
    owner_id: 'inst-1',
    finished: true,
    started_at: '2024-01-01T00:00:02Z',
    execution_time: 4000000000,
    status: { status: 'success', history: [] },
  } as TWorkflowStep,
  {
    id: 'step-provision-3',
    name: 'await install stack',
    execution_type: 'system',
    step_target_type: 'install_stacks',
    step_target_id: 'stack-1',
    install_workflow_id: 'wf-provision-1',
    owner_id: 'inst-1',
    finished: false,
    started_at: '2024-01-01T00:00:06Z',
    execution_time: 0,
    status: { status: 'in-progress', history: [] },
  } as TWorkflowStep,
  {
    id: 'step-provision-4',
    name: 'update install stack outputs',
    execution_type: 'system',
    step_target_type: undefined,
    step_target_id: undefined,
    install_workflow_id: 'wf-provision-1',
    owner_id: 'inst-1',
    finished: false,
    started_at: undefined,
    execution_time: 0,
    status: { status: 'pending', history: [] },
  } as TWorkflowStep,
  {
    id: 'step-provision-5',
    name: 'runner healthy',
    execution_type: 'system',
    step_target_type: undefined,
    step_target_id: undefined,
    install_workflow_id: 'wf-provision-1',
    owner_id: 'inst-1',
    finished: false,
    started_at: undefined,
    execution_time: 0,
    status: { status: 'pending', history: [] },
  } as TWorkflowStep,
]

export const ProvisionInProgress = () => (
  <WorkflowContext.Provider value={mockWorkflowContext}>
    <WorkflowSteps
      workflowSteps={provisionSteps}
      eagerStepsLoaded
      allStepsLoaded={false}
    />
  </WorkflowContext.Provider>
)

export const ProvisionAllStepsLoaded = () => (
  <WorkflowContext.Provider value={mockWorkflowContext}>
    <WorkflowSteps
      workflowSteps={provisionSteps}
      eagerStepsLoaded
      allStepsLoaded
    />
  </WorkflowContext.Provider>
)
