import type { Meta, StoryObj } from '@ladle/react'
import { ResourcesDetails } from './ResourcesDetails'

export default {
  title: 'Deploys/HelmOutputs/ResourcesDetails',
} satisfies Meta

const mockResources = {
  'default/my-configmap': {
    Kind: 'ConfigMap',
    Content: `apiVersion: v1
kind: ConfigMap
metadata:
  name: my-configmap
  namespace: default
data:
  DATABASE_HOST: db.example.com
  LOG_LEVEL: info`,
  },
  'default/my-secret': {
    Kind: 'Secret',
    Content: `apiVersion: v1
kind: Secret
metadata:
  name: my-secret
  namespace: default
type: Opaque`,
  },
  'default/my-hpa': {
    Kind: 'HorizontalPodAutoscaler',
    Content: `apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: my-hpa
spec:
  minReplicas: 2
  maxReplicas: 10`,
  },
}

export const Default: StoryObj = {
  render: () => <ResourcesDetails resources={mockResources} />,
}
