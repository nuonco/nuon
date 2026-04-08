export default {
  title: 'Deploys/HelmOutputs/HelmOutputs',
}

import { HelmOutputs, HelmOutputsSkeleton } from './HelmOutputs'

const mockOutputs = {
  deployments: {
    default: {
      'my-app': {
        metadata: {
          annotations: {
            'meta.helm.sh/release-name': 'my-app',
            'meta.helm.sh/release-namespace': 'default',
          },
          creationTimestamp: '2024-01-15T10:30:00Z',
          generation: 3,
          resourceVersion: '12345',
          uid: 'dep-abc-123',
        },
        status: {
          replicas: 3,
          readyReplicas: 3,
          availableReplicas: 3,
          updatedReplicas: 3,
          conditions: [
            { type: 'Available', status: 'True', reason: 'MinimumReplicasAvailable' },
          ],
        },
      },
    },
  },
  services: {
    default: {
      'my-app': {
        metadata: {
          creationTimestamp: '2024-01-15T10:30:00Z',
          uid: 'svc-abc-123',
          resourceVersion: '11111',
        },
        spec: { type: 'ClusterIP', clusterIP: '10.100.0.1', sessionAffinity: 'None' },
      },
    },
  },
  ingresses: {},
  resources: {
    'my-app-configmap': { Kind: 'ConfigMap', Content: 'apiVersion: v1\nkind: ConfigMap' },
  },
  manifest: `---
# Source: my-app/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app`,
}

export const Default = () => <HelmOutputs createdAt="2024-01-15T10:30:00Z" outputs={mockOutputs} />

export const Loading = () => <HelmOutputsSkeleton />
