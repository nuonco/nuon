export default {
  title: 'Deploys/HelmOutputs/DeploymentStatuses',
}

import { DeploymentStatuses } from './DeploymentStatuses'

const mockDeployments = {
  default: {
    'my-app': {
      status: {
        replicas: 3,
        readyReplicas: 3,
        availableReplicas: 3,
      },
    },
    'my-worker': {
      status: {
        replicas: 2,
        readyReplicas: 1,
        availableReplicas: 1,
      },
    },
  },
}

export const Default = () => <DeploymentStatuses deployments={mockDeployments} />

export const AllHealthy = () => (
  <DeploymentStatuses
    deployments={{
      production: {
        'api-server': {
          status: { replicas: 3, readyReplicas: 3, availableReplicas: 3 },
        },
        'worker': {
          status: { replicas: 2, readyReplicas: 2, availableReplicas: 2 },
        },
      },
    }}
  />
)
