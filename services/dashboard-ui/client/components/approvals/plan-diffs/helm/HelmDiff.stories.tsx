export default {
  title: 'Approvals/PlanDiffs/HelmDiff',
}

import { HelmDiff } from './HelmDiff'

const mockPlan = {
  plan: `default, my-app, Deployment (apps/v1) to be changed
default, my-app-svc, Service (v1) to be added
default, my-app-cache, Deployment (apps/v1) to be destroyed
Plan: 1 to add, 1 to change, 1 to destroy`,
  op: 'upgrade',
  helm_content_diff: [
    {
      api: 'apps/v1',
      kind: 'Deployment',
      name: 'my-app',
      namespace: 'default',
      before: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
  namespace: default
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: app
        image: my-app:1.2.0
        resources:
          limits:
            memory: 256Mi`,
      after: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
  namespace: default
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: app
        image: my-app:1.3.0
        resources:
          limits:
            memory: 512Mi`,
    },
    {
      api: 'v1',
      kind: 'Service',
      name: 'my-app-svc',
      namespace: 'default',
      before: '',
      after: `apiVersion: v1
kind: Service
metadata:
  name: my-app-svc
  namespace: default
spec:
  type: ClusterIP
  ports:
  - port: 80
    targetPort: 8080`,
    },
    {
      api: 'apps/v1',
      kind: 'Deployment',
      name: 'my-app-cache',
      namespace: 'default',
      before: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app-cache
  namespace: default
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: redis
        image: redis:6.2`,
      after: '',
    },
  ],
} as any

export const Default = () => <HelmDiff plan={mockPlan} />
