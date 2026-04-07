export default {
  title: 'Workflows/ActionRunStepDetails',
}

import { ActionRunStepDetails, ActionRunStepDetailsSkeleton } from './ActionRunStepDetails'
import type { TWorkflowStep, TInstallActionRun } from '@/types'

const mockStep = {
  id: 'step-1',
  name: 'run action',
  step_target_type: 'install_action_workflow_runs',
  step_target_id: 'run-1',
  owner_id: 'inst-1',
  execution_type: 'system',
  created_by_id: 'acc-1',
  created_by: { id: 'acc-1', email: 'user@example.com' },
  status: { status: 'in-progress', history: [] },
} as TWorkflowStep

const mockActionRun = {
  id: 'run-1',
  trigger_type: 'manual',
  created_by_id: 'acc-1',
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:05:00Z',
} as TInstallActionRun

export const Default = () => (
  <ActionRunStepDetails
    step={mockStep}
    actionRun={mockActionRun}
    createdBy={mockStep.created_by}
    error={null}
    isLoading={false}
  />
)

export const Loading = () => (
  <ActionRunStepDetails
    step={mockStep}
    error={null}
    isLoading={true}
  />
)

export const WithError = () => (
  <ActionRunStepDetails
    step={mockStep}
    error={new Error('Failed')}
    isLoading={false}
  />
)

export const Skeleton = () => <ActionRunStepDetailsSkeleton />
