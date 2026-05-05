export default {
  title: 'Workflows/StepBanner',
}

import { StepBanner } from './StepBanner'
import type { TWorkflowStep } from '@/types'

const baseStep = {
  id: 'step-1',
  name: 'deploy component',
  execution_type: 'system',
  status: { status: 'in-progress', history: [] },
} as TWorkflowStep

const errorStep = {
  ...baseStep,
  status: { status: 'error', history: [] },
} as TWorkflowStep

const approvalStep = {
  ...baseStep,
  approval: { id: 'apr-1', type: 'terraform_plan', response: null },
} as TWorkflowStep

export const Default = () => <StepBanner step={baseStep} />

export const ErrorState = () => <StepBanner step={errorStep} />

export const WithApproval = () => <StepBanner step={approvalStep} />

export const PlanOnly = () => <StepBanner step={approvalStep} planOnly />

const autoRetriedStep = {
  ...baseStep,
  status: {
    status: 'error',
    status_human_description: 'failed to poll for build: component build is in an error state',
    history: [],
    metadata: { auto_retried: true, retry_idx: 1, max_retries: 15 },
  },
} as TWorkflowStep

export const AutoRetriedDismissable = () => (
  <StepBanner
    step={autoRetriedStep}
    planOnly
    onDismiss={() => alert('dismissed')}
    onViewDetails={() => alert('view details')}
  />
)

export const ErrorWithViewDetails = () => (
  <StepBanner
    step={errorStep}
    planOnly
    onViewDetails={() => alert('view details')}
  />
)
