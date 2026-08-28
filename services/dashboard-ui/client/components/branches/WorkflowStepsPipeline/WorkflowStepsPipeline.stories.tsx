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
  <WorkflowStepsPipeline
    steps={mockSteps}
    selectedStepId="step-2"
    onSelectStep={noop}
  />
)

export const Empty = () => (
  <WorkflowStepsPipeline steps={[]} onSelectStep={noop} />
)

export const AllSuccess = () => (
  <WorkflowStepsPipeline
    steps={mockSteps.map((s) => ({
      ...s,
      status: { status: 'success' },
      execution_time: 30000000000,
    }))}
    onSelectStep={noop}
  />
)

const manySteps = [
  {
    id: 's1',
    name: 'Bundling components and sandbox',
    status: { status: 'success' },
    group_idx: 3,
    execution_time: 1260000000000,
  },
  {
    id: 's2',
    name: 'Plan install group: group-1',
    status: { status: 'success' },
    group_idx: 4,
    execution_time: 480000000000,
  },
  {
    id: 's3',
    name: 'Deploy install group: group-1',
    status: { status: 'in-progress' },
    group_idx: 5,
    execution_time: 11700000000000,
  },
  {
    id: 's4',
    name: 'Plan install group: group-2',
    status: { status: 'pending' },
    group_idx: 6,
  },
  {
    id: 's5',
    name: 'Deploy install group: group-2',
    status: { status: 'pending' },
    group_idx: 7,
  },
  {
    id: 's6',
    name: 'Plan install group: group-3',
    status: { status: 'pending' },
    group_idx: 8,
  },
  {
    id: 's7',
    name: 'Deploy install group: group-3',
    status: { status: 'pending' },
    group_idx: 9,
  },
] as any[]

export const ManySteps = () => (
  <div className="max-w-[760px]">
    <WorkflowStepsPipeline
      steps={manySteps}
      selectedStepId="s3"
      onSelectStep={noop}
    />
  </div>
)

export const WithError = () => (
  <WorkflowStepsPipeline
    steps={[
      { ...mockSteps[0], status: { status: 'success' } },
      {
        ...mockSteps[1],
        status: {
          status: 'error',
          status_human_description: 'Deployment failed',
        },
      },
      mockSteps[2],
    ]}
    onSelectStep={noop}
  />
)
