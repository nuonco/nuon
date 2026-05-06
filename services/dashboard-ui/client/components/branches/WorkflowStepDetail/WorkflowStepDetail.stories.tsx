export default {
  title: 'Branches/WorkflowStepDetail',
}

import { WorkflowStepDetail } from './WorkflowStepDetail'

const noop = () => {}

const mockStep = {
  id: 'step-abc123',
  name: 'Deploy to staging',
  status: { status: 'in-progress', status_human_description: 'Waiting for pods to be ready' },
  group_idx: 1,
  idx: 1,
  execution_type: 'deploy',
  retryable: true,
  started_at: '2024-06-15T10:30:00Z',
  install_workflow_id: 'wf-xyz789',
} as any

export const Default = () => (
  <WorkflowStepDetail step={mockStep} onClose={noop} />
)

export const Completed = () => (
  <WorkflowStepDetail
    step={{
      ...mockStep,
      status: { status: 'success', status_human_description: 'Deployment completed' },
      finished_at: '2024-06-15T10:35:00Z',
      execution_time: 300000000000,
    }}
    onClose={noop}
  />
)

export const Failed = () => (
  <WorkflowStepDetail
    step={{
      ...mockStep,
      status: { status: 'error', status_human_description: 'Pod CrashLoopBackOff' },
      finished_at: '2024-06-15T10:32:00Z',
      execution_time: 120000000000,
      retryable: false,
    }}
    onClose={noop}
  />
)

export const Minimal = () => (
  <WorkflowStepDetail
    step={{ id: 'step-min', name: 'Basic step', status: { status: 'pending' } } as any}
    onClose={noop}
  />
)
