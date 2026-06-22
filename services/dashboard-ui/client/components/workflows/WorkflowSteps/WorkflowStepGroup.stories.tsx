export default {
  title: 'Workflows/WorkflowStepGroup',
}

import { WorkflowStepGroup } from './WorkflowStepGroup'
import { WorkflowContext } from '@/providers/workflow-provider'
import type { TWorkflowStep } from '@/types'

const mockWorkflowContext = {
  workflow: { id: 'wf-1', status: { status: 'in-progress' } },
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
} as any

const base: TWorkflowStep = {
  id: 'step-cert-1',
  name: 'sync and plan certificate',
  execution_type: 'system',
  step_target_type: 'install_deploys',
  step_target_id: 'deploy-1',
  install_workflow_id: 'wf-1',
  owner_id: 'inst-1',
  finished: true,
  started_at: '2024-01-01T00:00:00Z',
  execution_time: 12000000000,
  status: { status: 'error', history: [] },
} as TWorkflowStep

const attempt = (idx: number, overrides: Partial<TWorkflowStep>) =>
  ({
    ...base,
    id: `step-cert-${idx}`,
    execution_time: 12000000000 + idx * 1000000000,
    ...overrides,
  }) as TWorkflowStep

const wrap = (steps: TWorkflowStep[]) => (
  <WorkflowContext.Provider value={mockWorkflowContext}>
    <WorkflowStepGroup steps={steps} />
  </WorkflowContext.Provider>
)

export const ThreePriorAttempts = () =>
  wrap([
    attempt(1, {
      status: {
        status: 'error',
        history: [],
        metadata: { auto_retried: true, retry_idx: 1, max_retries: 15 },
      },
    }),
    attempt(2, {
      status: {
        status: 'error',
        history: [],
        metadata: { auto_retried: true, retry_idx: 2, max_retries: 15 },
      },
    }),
    attempt(3, {
      status: {
        status: 'error',
        history: [],
        metadata: { auto_retried: true, retry_idx: 3, max_retries: 15 },
      },
    }),
    attempt(4, {
      retryable: true,
      execution_time: 13000000000,
      status: {
        status: 'error',
        history: [],
        metadata: { is_retry: true, retry_idx: 3, max_retries: 15 },
      },
    }),
  ])

export const SinglePriorAttempt = () =>
  wrap([
    attempt(1, {
      status: {
        status: 'error',
        history: [],
        metadata: { auto_retried: true, retry_idx: 1, max_retries: 15 },
      },
    }),
    attempt(2, {
      finished: true,
      execution_time: 9000000000,
      status: { status: 'success', history: [] },
    }),
  ])
