export default {
  title: 'Branches/WorkflowStepsPipeline',
}

import { WorkflowStepsPipeline } from './WorkflowStepsPipeline'

const noop = () => {}

const mockSteps = [
  {
    id: 'step-1',
    name: 'Build image',
    status: { status: 'success' },
    group_idx: 0,
    execution_time: 45000000000,
    idx: 0,
  },
  {
    id: 'step-2',
    name: 'Deploy to staging',
    status: { status: 'in-progress', status_human_description: 'Deploying...' },
    group_idx: 1,
    idx: 1,
  },
  {
    id: 'step-3',
    name: 'Deploy to production',
    status: { status: 'pending' },
    group_idx: 2,
    idx: 2,
  },
] as any[]

export const Default = () => (
  <WorkflowStepsPipeline steps={mockSteps} onSelectStep={noop} />
)

export const WithSelectedStep = () => (
  <WorkflowStepsPipeline steps={mockSteps} selectedStepId="step-2" onSelectStep={noop} />
)

export const Empty = () => (
  <WorkflowStepsPipeline steps={[]} onSelectStep={noop} />
)

export const AllSuccess = () => (
  <WorkflowStepsPipeline
    steps={mockSteps.map((s) => ({ ...s, status: { status: 'success' }, execution_time: 30000000000 }))}
    onSelectStep={noop}
  />
)

export const WithError = () => (
  <WorkflowStepsPipeline
    steps={[
      { ...mockSteps[0], status: { status: 'success' } },
      { ...mockSteps[1], status: { status: 'error', status_human_description: 'Deployment failed' } },
      mockSteps[2],
    ]}
    onSelectStep={noop}
  />
)
