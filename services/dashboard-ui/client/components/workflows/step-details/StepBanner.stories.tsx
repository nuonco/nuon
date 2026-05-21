export default {
  title: 'Workflows/StepBanner',
}

import { StepBanner } from './StepBanner'
import type { TWorkflowStep } from '@/types'

const baseStep = {
  id: 'step-1',
  name: 'deploy component',
  execution_type: 'system',
  created_by: { email: 'user@example.com' },
  status: { status: 'in-progress', history: [] },
} as TWorkflowStep

// --- No banner (in-progress, no special state) ---

export const InProgress = () => <StepBanner step={baseStep} />

// --- Error states ---

export const Error = () => (
  <StepBanner
    step={{
      ...baseStep,
      status: {
        status: 'error',
        status_human_description: 'failed to apply terraform plan',
        history: [],
      },
    } as TWorkflowStep}
  />
)

export const ErrorRetryable = () => (
  <StepBanner
    step={{
      ...baseStep,
      retryable: true,
      status: {
        status: 'error',
        status_human_description: 'failed to apply terraform plan',
        history: [],
      },
    } as TWorkflowStep}
  />
)

export const ErrorRetryableAndSkippable = () => (
  <StepBanner
    step={{
      ...baseStep,
      retryable: true,
      skippable: true,
      status: {
        status: 'error',
        status_human_description: 'failed to apply terraform plan',
        history: [],
      },
    } as TWorkflowStep}
  />
)

export const ErrorWithRetryInfo = () => (
  <StepBanner
    step={{
      ...baseStep,
      status: {
        status: 'error',
        status_human_description: 'failed to poll for build',
        history: [],
        metadata: { retry_type: 'manual', retry_idx: 2, max_retries: 15 },
      },
    } as TWorkflowStep}
  />
)

export const ErrorAutoRetried = () => (
  <StepBanner
    step={{
      ...baseStep,
      status: {
        status: 'error',
        status_human_description: 'failed to poll for build',
        history: [],
        metadata: { auto_retried: true, retry_idx: 1, max_retries: 15 },
      },
    } as TWorkflowStep}
  />
)

export const ErrorAutoRetriesExhausted = () => (
  <StepBanner
    step={{
      ...baseStep,
      status: {
        status: 'error',
        status_human_description: 'auto-retries exhausted',
        history: [],
        metadata: {
          auto_retries_exhausted: true,
          max_auto_retries: 3,
          retry_index: 5,
          max_retries: 15,
        },
      },
    } as TWorkflowStep}
  />
)

export const ErrorRetriesExhausted = () => (
  <StepBanner
    step={{
      ...baseStep,
      status: {
        status: 'error',
        status_human_description: 'all retries exhausted',
        history: [],
        metadata: { retries_exhausted: true, retry_index: 15, max_retries: 15 },
      },
    } as TWorkflowStep}
  />
)

export const StalePlan = () => (
  <StepBanner
    step={{
      ...baseStep,
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
          check_label_age_minutes: '4380',
          check_label_threshold_minutes: '4320',
        },
      },
    } as TWorkflowStep}
  />
)

export const SupersededPlan = () => (
  <StepBanner
    step={{
      ...baseStep,
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
    } as TWorkflowStep}
  />
)

// --- Failed pending retry ---

export const FailedPendingRetry = () => (
  <StepBanner
    step={{
      ...baseStep,
      retryable: true,
      skippable: true,
      status: {
        status: 'failed-pending-retry',
        status_human_description: 'step failed, awaiting user action',
        history: [],
      },
    } as TWorkflowStep}
  />
)

export const FailedPendingRetryOnlyRetryable = () => (
  <StepBanner
    step={{
      ...baseStep,
      retryable: true,
      status: {
        status: 'failed-pending-retry',
        status_human_description: 'step failed, awaiting user action',
        history: [],
      },
    } as TWorkflowStep}
  />
)

// --- Terminal states ---

export const Cancelled = () => (
  <StepBanner
    step={{
      ...baseStep,
      status: {
        status: 'cancelled',
        status_human_description: 'Workflow was cancelled by user',
        history: [],
      },
    } as TWorkflowStep}
  />
)

export const Discarded = () => (
  <StepBanner
    step={{
      ...baseStep,
      status: {
        status: 'discarded',
        status_human_description: 'The plan step was discarded and retried by the user',
        history: [],
      },
    } as TWorkflowStep}
  />
)

export const UserSkipped = () => (
  <StepBanner
    step={{
      ...baseStep,
      status: {
        status: 'user-skipped',
        status_human_description: 'User chose to skip this step',
        history: [],
      },
    } as TWorkflowStep}
  />
)

export const NotAttempted = () => (
  <StepBanner
    step={{
      ...baseStep,
      status: {
        status: 'not-attempted',
        status_human_description: 'A previous step failed',
        history: [],
      },
    } as TWorkflowStep}
  />
)

export const Skipped = () => (
  <StepBanner
    step={{
      ...baseStep,
      execution_type: 'skipped',
      status: { status: 'success', history: [] },
    } as TWorkflowStep}
  />
)

export const Retried = () => (
  <StepBanner
    step={{
      ...baseStep,
      retryable: true,
      retried: true,
      status: {
        status: 'error',
        status_human_description: 'Step was retried',
        history: [],
        metadata: { retry_type: 'manual' },
      },
    } as TWorkflowStep}
  />
)

// --- Policy banners ---

export const PolicyAllPassed = () => (
  <StepBanner
    step={{
      ...baseStep,
      status: {
        status: 'approved',
        history: [],
        metadata: {
          passed_policy_ids: ['pol-1', 'pol-2'],
        },
      },
    } as TWorkflowStep}
  />
)

export const PolicyAutoApproved = () => (
  <StepBanner
    step={{
      ...baseStep,
      execution_type: 'approval',
      status: {
        status: 'approved',
        history: [],
        metadata: {
          check: 'policy-auto-approve',
          summary: 'Auto-approved: all policies passed',
          auto_approved: true,
          check_label_approval_reason: 'policies_passed',
          passed_policy_ids: ['pol-1', 'pol-2'],
        },
      },
    } as TWorkflowStep}
  />
)

export const PolicyViolations = () => (
  <StepBanner
    step={{
      ...baseStep,
      status: {
        status: 'approval-awaiting',
        history: [],
        metadata: {
          deny_violations: [
            { policy_id: 'pol-1', policy_name: 'cost-limit', message: 'Estimated cost exceeds $500/month' },
          ],
          warn_violations: [
            { policy_id: 'pol-2', policy_name: 'naming-convention', message: 'Resource name does not follow convention' },
          ],
          passed_policy_ids: ['pol-3'],
        },
      },
    } as TWorkflowStep}
  />
)

// --- With callbacks ---

export const ErrorWithViewDetails = () => (
  <StepBanner
    step={{
      ...baseStep,
      status: {
        status: 'error',
        status_human_description: 'failed to apply terraform plan',
        history: [],
      },
    } as TWorkflowStep}
    onViewDetails={() => alert('view details')}
  />
)

export const AutoRetriedDismissable = () => (
  <StepBanner
    step={{
      ...baseStep,
      status: {
        status: 'error',
        status_human_description: 'failed to poll for build',
        history: [],
        metadata: { auto_retried: true, retry_idx: 1, max_retries: 15 },
      },
    } as TWorkflowStep}
    onDismiss={() => alert('dismissed')}
    onViewDetails={() => alert('view details')}
  />
)
