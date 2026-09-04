import { terraformDiff } from '../../lib/diffs'

export const deploymentBefore = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: api
  namespace: apps
spec:
  replicas: 2
  template:
    spec:
      containers:
        - name: api
          image: ghcr.io/example/api:v1.8.0
          env:
            - name: LOG_LEVEL
              value: info`

export const deploymentAfter = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: api
  namespace: apps
spec:
  replicas: 4
  template:
    spec:
      containers:
        - name: api
          image: ghcr.io/example/api:v2.1.0
          env:
            - name: LOG_LEVEL
              value: warn`

export const terraformResourceDiff = terraformDiff({
  before: {
    endpoint: null,
    name: 'platform',
    tags: {
      environment: 'production',
      team: 'platform',
    },
    token: 'old-token',
    version: '1.30',
    vpc_config: {
      endpoint_private_access: false,
      subnet_ids: ['subnet-a', 'subnet-b'],
    },
  },
  after: {
    endpoint: null,
    name: 'platform',
    tags: {
      environment: 'production',
      team: 'platform',
    },
    token: 'new-token',
    version: '1.31',
    vpc_config: {
      endpoint_private_access: true,
      subnet_ids: ['subnet-a', 'subnet-b'],
    },
  },
  afterUnknown: { endpoint: true },
  beforeSensitive: { token: true },
  afterSensitive: { token: true },
  filename: 'aws_eks_cluster.this.tf',
})

export const longManifestDiff = () => {
  const shared = Array.from(
    { length: 420 },
    (_, index) =>
      `  setting_${String(index + 1).padStart(3, '0')}: value-${index + 1}`
  )
  const before = ['apiVersion: v1', 'kind: ConfigMap', 'data:', ...shared]
  const after = [...before]
  after[3] = '  setting_001: updated-at-the-start'
  after[212] = '  setting_210: searchable-change'
  after[after.length - 1] = '  setting_420: updated-at-the-end'
  return { before: before.join('\n'), after: after.join('\n') }
}
