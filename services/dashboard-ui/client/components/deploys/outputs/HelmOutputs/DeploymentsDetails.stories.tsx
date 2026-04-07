export default {
  title: 'Deploys/HelmOutputs/DeploymentsDetails',
}

import { DeploymentsDetails } from './DeploymentsDetails'

const mockDeployments = {
  default: {
    'my-app': {
      metadata: {
        creationTimestamp: '2024-01-15T10:30:00Z',
        generation: 5,
        resourceVersion: '12345',
        uid: 'abc-def-123',
      },
      status: {
        replicas: 3,
        readyReplicas: 3,
        availableReplicas: 3,
        updatedReplicas: 3,
        conditions: [
          { type: 'Available', status: 'True', reason: 'MinimumReplicasAvailable' },
          { type: 'Progressing', status: 'True', reason: 'NewReplicaSetAvailable' },
        ],
      },
    },
    'my-worker': {
      metadata: {
        creationTimestamp: '2024-01-14T10:30:00Z',
        generation: 2,
        resourceVersion: '67890',
        uid: 'xyz-ghi-456',
      },
      status: {
        replicas: 2,
        readyReplicas: 1,
        availableReplicas: 1,
        updatedReplicas: 2,
        conditions: [
          { type: 'Available', status: 'True', reason: 'MinimumReplicasAvailable' },
        ],
      },
    },
  },
}

export const Default = () => <DeploymentsDetails deployments={mockDeployments} />

export const Empty = () => <DeploymentsDetails deployments={{}} />
