import type { Meta, StoryObj } from '@ladle/react'
import { HelmManifest } from './HelmManifest'

export default {
  title: 'Deploys/HelmOutputs/HelmManifest',
} satisfies Meta

export const Default: StoryObj = {
  render: () => (
    <HelmManifest
      manifest={`---
# Source: my-app/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: my-app
  template:
    spec:
      containers:
      - name: my-app
        image: my-registry/my-app:1.2.3
        ports:
        - containerPort: 8080
---
# Source: my-app/templates/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: my-app
  namespace: default
spec:
  type: ClusterIP
  ports:
  - port: 80
    targetPort: 8080`}
    />
  ),
}
