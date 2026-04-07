export default {
  title: 'Workflows/StepDetails/StandardActionSteps',
}

import { StandardActionSteps, StandardActionStepsSkeleton } from './StandardActionSteps'
import type { TInstallActionRun } from '@/types'

const mockActionRun: TInstallActionRun = {
  id: 'run-1',
  steps: [
    { id: 'step-1', status: 'in-progress', execution_duration: 5400000000 },
    { id: 'step-2', status: 'finished', execution_duration: 12800000000 },
    { id: 'step-3', status: 'pending', execution_duration: 0 },
  ],
  config: {
    steps: [
      { idx: 0, name: 'pre_deploy_check' },
      { idx: 1, name: 'run_migrations' },
      { idx: 2, name: 'health_check' },
    ],
  },
} as TInstallActionRun

export const Default = () => <StandardActionSteps actionRun={mockActionRun} />

export const AllSucceeded = () => (
  <StandardActionSteps
    actionRun={{
      ...mockActionRun,
      steps: [
        { id: 'step-1', status: 'finished', execution_duration: 5400000000 },
        { id: 'step-2', status: 'finished', execution_duration: 12800000000 },
        { id: 'step-3', status: 'finished', execution_duration: 3200000000 },
      ],
    } as TInstallActionRun}
  />
)

export const Loading = () => <StandardActionStepsSkeleton />
