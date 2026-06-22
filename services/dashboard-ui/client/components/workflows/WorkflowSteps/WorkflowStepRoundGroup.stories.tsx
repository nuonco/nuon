export default {
  title: 'Workflows/WorkflowStepRoundGroup',
}

import { WorkflowStepRoundGroup } from './WorkflowStepRoundGroup'
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
  id: 'step',
  name: 'step',
  execution_type: 'system',
  step_target_type: 'install_deploys',
  step_target_id: 'deploy-1',
  install_workflow_id: 'wf-1',
  owner_id: 'inst-1',
  group_idx: 7,
  started_at: '2024-01-01T00:00:00Z',
  execution_time: 0,
  status: { status: 'pending', history: [] },
} as TWorkflowStep

const plan = (round: number, overrides: Partial<TWorkflowStep>) =>
  ({
    ...base,
    id: `plan-${round}`,
    name: 'sync and plan whoami',
    execution_type: 'approval',
    group_retry_idx: round,
    ...overrides,
  }) as TWorkflowStep

const apply = (round: number, overrides: Partial<TWorkflowStep>) =>
  ({
    ...base,
    id: `apply-${round}`,
    name: 'apply whoami',
    group_retry_idx: round,
    ...overrides,
  }) as TWorkflowStep

const wrap = (steps: TWorkflowStep[]) => (
  <WorkflowContext.Provider value={mockWorkflowContext}>
    <WorkflowStepRoundGroup steps={steps} />
  </WorkflowContext.Provider>
)

export const TwoPriorRounds = () =>
  wrap([
    plan(0, {
      retried: true,
      finished: true,
      execution_time: 28000000000,
      status: { status: 'success', history: [], metadata: { status: 'approved' } },
    }),
    apply(0, {
      finished: true,
      execution_time: 7000000000,
      status: {
        status: 'error',
        history: [],
        metadata: { auto_retried: true, retry_idx: 1, max_retries: 15 },
      },
    }),
    plan(1, {
      retried: true,
      finished: true,
      execution_time: 25000000000,
      status: { status: 'success', history: [], metadata: { status: 'approved' } },
    }),
    apply(1, {
      finished: true,
      execution_time: 7000000000,
      status: {
        status: 'error',
        history: [],
        metadata: { auto_retried: true, retry_idx: 2, max_retries: 15 },
      },
    }),
    plan(2, {
      status: {
        status: 'approval-awaiting',
        history: [],
        metadata: { is_retry: true, retry_idx: 2 },
      },
      approval: { id: 'apr-1', type: 'terraform_plan' },
    }),
    apply(2, {
      started_at: undefined,
      status: {
        status: 'pending',
        history: [],
        metadata: { is_retry: true, retry_idx: 2 },
      },
    }),
  ])

export const OnePriorRound = () =>
  wrap([
    plan(0, {
      retried: true,
      finished: true,
      execution_time: 28000000000,
      status: { status: 'success', history: [], metadata: { status: 'approved' } },
    }),
    apply(0, {
      finished: true,
      execution_time: 7000000000,
      status: {
        status: 'error',
        history: [],
        metadata: { auto_retried: true, retry_idx: 1, max_retries: 15 },
      },
    }),
    plan(1, {
      finished: true,
      execution_time: 25000000000,
      status: { status: 'success', history: [] },
    }),
    apply(1, {
      started_at: undefined,
      status: { status: 'pending', history: [], metadata: { is_retry: true, retry_idx: 1 } },
    }),
  ])
