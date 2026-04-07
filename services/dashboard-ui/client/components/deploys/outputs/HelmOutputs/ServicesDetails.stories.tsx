import type { Meta, StoryObj } from '@ladle/react'
import { ServicesDetails } from './ServicesDetails'

export default {
  title: 'Deploys/HelmOutputs/ServicesDetails',
} satisfies Meta

const mockServices = {
  default: {
    'my-app': {
      metadata: {
        creationTimestamp: '2024-01-15T10:30:00Z',
        uid: 'svc-abc-123',
        resourceVersion: '11111',
      },
      spec: {
        type: 'ClusterIP',
        clusterIP: '10.100.0.1',
        sessionAffinity: 'None',
      },
    },
    'my-app-lb': {
      metadata: {
        creationTimestamp: '2024-01-15T10:30:00Z',
        uid: 'svc-def-456',
        resourceVersion: '22222',
      },
      spec: {
        type: 'LoadBalancer',
        clusterIP: '10.100.0.2',
        sessionAffinity: 'None',
      },
    },
  },
}

export const Default: StoryObj = {
  render: () => <ServicesDetails services={mockServices} />,
}

export const Empty: StoryObj = {
  render: () => <ServicesDetails services={{}} />,
}
