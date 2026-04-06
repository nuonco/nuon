export default {
  title: 'Actions/InstallActionRunOutputs',
}

import { InstallActionRunOutputs } from './InstallActionRunOutputs'

const mockRunWithOutputs = {
  steps: [
    { id: 'step-run-1', action_config_step_id: 'step-1', status: 'succeeded', execution_duration: 45000000000 },
    { id: 'step-run-2', action_config_step_id: 'step-2', status: 'succeeded', execution_duration: 30000000000 },
  ],
  config: {
    steps: [
      { id: 'step-1', name: 'terraform-apply', idx: 0 },
      { id: 'step-2', name: 'verify', idx: 1 },
    ],
  },
  outputs: {
    steps: {
      'terraform-apply': {
        cluster_endpoint: 'https://eks.us-west-2.amazonaws.com/cluster-1',
        cluster_name: 'prod-cluster',
        vpc_id: 'vpc-abc123',
        status: 'active',
      },
      'verify': null,
    },
  },
} as any

export const Default = () => (
  <InstallActionRunOutputs installActionRun={mockRunWithOutputs} />
)

const mockRunNoOutputs = {
  steps: [
    { id: 'step-run-1', action_config_step_id: 'step-1', status: 'succeeded', execution_duration: 10000000000 },
  ],
  config: {
    steps: [
      { id: 'step-1', name: 'run-script', idx: 0 },
    ],
  },
  outputs: { steps: {} },
} as any

export const NoOutputs = () => (
  <InstallActionRunOutputs installActionRun={mockRunNoOutputs} />
)
