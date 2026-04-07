export default {
  title: 'Workflows/StepTitle',
}

import { StepTitle } from './StepTitle'
import type { TWorkflowStep } from '@/types'

const inProgressStep = {
  id: 'step-1',
  name: 'deploy component',
  retried: false,
  status: { status: 'in-progress' },
} as TWorkflowStep

const retriedStep = {
  ...inProgressStep,
  retried: true,
} as TWorkflowStep

const succeededStep = {
  ...inProgressStep,
  status: { status: 'success' },
} as TWorkflowStep

export const InProgress = () => <StepTitle step={inProgressStep} />

export const Retried = () => <StepTitle step={retriedStep} />

export const Succeeded = () => <StepTitle step={succeededStep} />
