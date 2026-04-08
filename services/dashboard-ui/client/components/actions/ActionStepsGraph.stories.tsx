export default {
  title: 'Actions/ActionStepGraph',
}

import { ActionStepGraph } from './ActionStepsGraph'

const mockSteps = [
  {
    id: 'step-1',
    name: 'Checkout code',
    idx: 0,
    status: 'success',
    execution_duration: 2000000000,
  },
  {
    id: 'step-2',
    name: 'Run tests',
    idx: 1,
    status: 'success',
    execution_duration: 45000000000,
  },
  {
    id: 'step-3',
    name: 'Build Docker image',
    idx: 2,
    status: 'running',
    execution_duration: 0,
  },
  {
    id: 'step-4',
    name: 'Push to registry',
    idx: 3,
    status: 'pending',
    execution_duration: 0,
  },
] as any[]

export const Default = () => <ActionStepGraph steps={mockSteps} />

export const SingleStep = () => <ActionStepGraph steps={[mockSteps[0]]} />

export const Empty = () => <ActionStepGraph steps={[]} />
