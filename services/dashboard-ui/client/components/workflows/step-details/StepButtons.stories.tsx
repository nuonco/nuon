export default {
  title: 'Workflows/StepButtons',
}

import { StepButtons } from './StepButtons'
import type { TWorkflowStep } from '@/types'

const baseStep = {
  id: 'step-1',
  name: 'deploy component',
  execution_type: 'system',
  skippable: false,
  status: { status: 'in-progress', history: [] },
} as TWorkflowStep

const retryableStep = {
  ...baseStep,
  status: { status: 'error', history: [] },
  skippable: true,
} as TWorkflowStep

export const Default = () => <StepButtons step={baseStep} />

export const WithRetry = () => <StepButtons step={retryableStep} />

export const ApproveAll = () => <StepButtons step={baseStep} isApproveAll />
