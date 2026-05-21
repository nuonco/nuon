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

const base: TWorkflowStep = {
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

const wrap = (steps: TWorkflowStep[]) => (
  <WorkflowContext.Provider value={mockWorkflowContext}>
    <WorkflowSteps workflowSteps={steps} />
  </WorkflowContext.Provider>
)

// --- Basic statuses ---

export const InProgress = () => wrap([base])

export const Success = () =>
  wrap([
    {
      ...base,
      id: 'step-success',
      finished: true,
      execution_time: 60000000000,
      status: { status: 'success', history: [] },
    } as TWorkflowStep,
  ])

export const Pending = () =>
  wrap([
    {
      ...base,
      id: 'step-pending',
      started_at: undefined,
      status: { status: 'pending', history: [] },
    } as TWorkflowStep,
  ])

export const Noop = () =>
  wrap([
    {
      ...base,
      id: 'step-noop',
      status: { status: 'noop', history: [] },
    } as TWorkflowStep,
  ])

// --- Error states ---

export const Error = () =>
  wrap([
    {
      ...base,
      id: 'step-error',
      status: {
        status: 'error',
        status_human_description: 'failed to apply terraform plan',
        history: [],
      },
    } as TWorkflowStep,
  ])

export const ErrorRetryable = () =>
  wrap([
    {
      ...base,
      id: 'step-error-retryable',
      retryable: true,
      status: {
        status: 'error',
        status_human_description: 'failed to apply terraform plan',
        history: [],
      },
    } as TWorkflowStep,
  ])

export const ErrorRetryableAndSkippable = () =>
  wrap([
    {
      ...base,
      id: 'step-error-both',
      retryable: true,
      skippable: true,
      status: {
        status: 'error',
        status_human_description: 'failed to apply terraform plan',
        history: [],
      },
    } as TWorkflowStep,
  ])

export const FailedPendingRetry = () =>
  wrap([
    {
      ...base,
      id: 'step-fpr',
      retryable: true,
      skippable: true,
      status: {
        status: 'failed-pending-retry',
        status_human_description: 'step failed, awaiting user action',
        history: [],
      },
    } as TWorkflowStep,
  ])

export const StalePlan = () =>
  wrap([
    {
      ...base,
      id: 'step-stale',
      name: 'approve terraform plan',
      execution_type: 'approval',
      status: {
        status: 'error',
        status_human_description: 'Plan is stale, auto-retrying',
        history: [],
        metadata: {
          check: 'stale-plan',
          summary: 'Plan is stale, auto-retrying',
          detail: 'Approval was submitted 4380m after plan creation (threshold: 4320m)',
          check_label_stale: 'true',
        },
      },
    } as TWorkflowStep,
  ])

export const SupersededPlan = () =>
  wrap([
    {
      ...base,
      id: 'step-superseded',
      name: 'approve terraform plan',
      execution_type: 'approval',
      status: {
        status: 'error',
        status_human_description: 'Plan superseded, auto-retrying',
        history: [],
        metadata: {
          check: 'superseded',
          summary: 'Plan superseded, auto-retrying',
          detail: 'A newer deploy was approved for this component',
          check_label_superseded: 'true',
        },
      },
    } as TWorkflowStep,
  ])

// --- Retry / retried states ---

export const AutoRetried = () =>
  wrap([
    {
      ...base,
      id: 'step-auto-retried',
      status: {
        status: 'error',
        status_human_description: 'failed to poll for build',
        history: [],
        metadata: { auto_retried: true, retry_idx: 1, max_retries: 15 },
      },
    } as TWorkflowStep,
  ])

export const Retried = () =>
  wrap([
    {
      ...base,
      id: 'step-retried',
      retried: true,
      status: {
        status: 'error',
        status_human_description: 'Step was retried',
        history: [],
      },
    } as TWorkflowStep,
  ])

export const IsRetryManual = () =>
  wrap([
    {
      ...base,
      id: 'step-manual-retry',
      status: {
        status: 'in-progress',
        history: [],
        metadata: { is_retry: true, retry_type: 'manual', retry_idx: 2 },
      },
    } as TWorkflowStep,
  ])

export const IsRetryAuto = () =>
  wrap([
    {
      ...base,
      id: 'step-auto-retry',
      status: {
        status: 'in-progress',
        history: [],
        metadata: { is_retry: true, retry_type: 'auto', group_retry_idx: 3 },
      },
    } as TWorkflowStep,
  ])

// --- Approval states ---

export const ApprovalAwaiting = () =>
  wrap([
    {
      ...base,
      id: 'step-awaiting',
      name: 'approve terraform plan',
      execution_type: 'approval',
      status: { status: 'approval-awaiting', history: [] },
      approval: { id: 'apr-1', type: 'terraform_plan' },
    } as TWorkflowStep,
  ])

export const ApprovalApproved = () =>
  wrap([
    {
      ...base,
      id: 'step-approved',
      name: 'approve terraform plan',
      execution_type: 'approval',
      finished: true,
      status: { status: 'approved', history: [] },
    } as TWorkflowStep,
  ])

export const ApprovalPlanCreated = () => (
  <WorkflowContext.Provider value={mockWorkflowContext}>
    <WorkflowSteps
      workflowSteps={[
        {
          ...base,
          id: 'step-plan-created',
          name: 'approve terraform plan',
          execution_type: 'approval',
          finished: true,
          status: { status: 'approved', history: [] },
        } as TWorkflowStep,
      ]}
      planOnly
    />
  </WorkflowContext.Provider>
)

export const ApprovalAutoApproved = () =>
  wrap([
    {
      ...base,
      id: 'step-auto-approved',
      name: 'approve terraform plan',
      execution_type: 'approval',
      finished: true,
      approval: { id: 'apr-1', type: 'approve-all' },
      status: { status: 'success', history: [] },
    } as TWorkflowStep,
  ])

export const ApprovalPolicyAutoApproved = () =>
  wrap([
    {
      ...base,
      id: 'step-policy-auto',
      name: 'approve terraform plan',
      execution_type: 'approval',
      finished: true,
      status: {
        status: 'approved',
        history: [],
        metadata: {
          check: 'policy-auto-approve',
          summary: 'Auto-approved: all policies passed',
          auto_approved: true,
          check_label_approval_reason: 'policies_passed',
        },
      },
    } as TWorkflowStep,
  ])

export const ApprovalDenied = () =>
  wrap([
    {
      ...base,
      id: 'step-denied',
      name: 'approve terraform plan',
      execution_type: 'approval',
      status: { status: 'approval-denied', history: [] },
    } as TWorkflowStep,
  ])

export const ApprovalRetry = () =>
  wrap([
    {
      ...base,
      id: 'step-approval-retry',
      name: 'approve terraform plan',
      execution_type: 'approval',
      status: { status: 'approval-retry', history: [] },
    } as TWorkflowStep,
  ])

// --- Terminal / skip states ---

export const Cancelled = () =>
  wrap([
    {
      ...base,
      id: 'step-cancelled',
      status: {
        status: 'cancelled',
        status_human_description: 'Workflow was cancelled by user',
        history: [],
      },
    } as TWorkflowStep,
  ])

export const Discarded = () =>
  wrap([
    {
      ...base,
      id: 'step-discarded',
      status: {
        status: 'discarded',
        status_human_description: 'The plan step was discarded and retried by the user',
        history: [],
      },
    } as TWorkflowStep,
  ])

export const UserSkipped = () =>
  wrap([
    {
      ...base,
      id: 'step-user-skipped',
      status: {
        status: 'user-skipped',
        status_human_description: 'User chose to skip this step',
        history: [],
      },
    } as TWorkflowStep,
  ])

export const AutoSkipped = () =>
  wrap([
    {
      ...base,
      id: 'step-auto-skipped',
      status: { status: 'auto-skipped', history: [] },
    } as TWorkflowStep,
  ])

export const Skipped = () =>
  wrap([
    {
      ...base,
      id: 'step-skipped',
      execution_type: 'skipped',
      status: { status: 'success', history: [] },
    } as TWorkflowStep,
  ])

export const NotAttempted = () =>
  wrap([
    {
      ...base,
      id: 'step-not-attempted',
      started_at: undefined,
      status: {
        status: 'not-attempted',
        status_human_description: 'A previous step failed',
        history: [],
      },
    } as TWorkflowStep,
  ])

// --- Multi-step workflows ---

export const MixedSteps = () =>
  wrap([
    {
      ...base,
      id: 'step-done',
      name: 'provision runner',
      finished: true,
      execution_time: 2000000000,
      status: { status: 'success', history: [] },
    } as TWorkflowStep,
    {
      ...base,
      id: 'step-approve',
      name: 'approve terraform plan',
      execution_type: 'approval',
      status: { status: 'approval-awaiting', history: [] },
      approval: { id: 'apr-1', type: 'terraform_plan' },
    } as TWorkflowStep,
    {
      ...base,
      id: 'step-pending',
      name: 'apply terraform',
      started_at: undefined,
      status: { status: 'pending', history: [] },
    } as TWorkflowStep,
  ])

// --- Loading ---

export const Loading = () => <WorkflowStepsSkeleton />

export const Empty = () => wrap([])
