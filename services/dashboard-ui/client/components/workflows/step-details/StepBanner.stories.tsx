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
