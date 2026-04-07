export default {
  title: 'Deploys/HelmOutputs/Overview',
}

import { Overview } from './Overview'

const mockOutputs = {
  deployments: {
    default: {
      'my-app': {
        metadata: {
          annotations: {
            'meta.helm.sh/release-name': 'my-app',
            'meta.helm.sh/release-namespace': 'default',
          },
        },
        status: {
          replicas: 3,
          readyReplicas: 3,
          availableReplicas: 3,
        },
      },
    },
  },
  services: {
    default: { 'my-app': {} },
  },
  ingresses: {
    default: { 'my-app-ingress': {} },
  },
  resources: {
    'my-app-configmap': { Kind: 'ConfigMap' },
    'my-app-secret': { Kind: 'Secret' },
  },
}

export const Default = () => (
  <Overview createdAt="2024-01-15T10:30:00Z" outputs={mockOutputs} />
)
