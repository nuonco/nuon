export default {
  title: 'Workflows/StackStepDetails',
}

import { StackStepDetails } from './StackStepDetails'
import type { TWorkflowStep } from '@/types'

const mockStep = {
  id: 'step-1',
  name: 'generate install stack',
  step_target_type: 'install_stack_versions',
  step_target_id: 'stack-1',
  owner_id: 'inst-1',
  status: { status: 'in-progress', history: [] },
} as TWorkflowStep

export const GenerateStackLoading = () => (
  <StackStepDetails step={mockStep} isLoading={true} />
)

export const AwaitStackLoading = () => (
  <StackStepDetails
    step={{ ...mockStep, name: 'await install stack' } as TWorkflowStep}
    isLoading={true}
  />
)
