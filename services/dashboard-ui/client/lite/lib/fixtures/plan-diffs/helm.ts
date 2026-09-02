import type { THelmPlan } from '@/types'

interface IResource {
  api: string
  kind: string
  name: string
  namespace: string
  before: string
  after: string
}

const deployment = ({
  name,
  namespace,
  image,
  replicas = 2,
  extra = '',
}: {
  name: string
  namespace: string
  image: string
  replicas?: number
  extra?: string
}) => `apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${name}
  namespace: ${namespace}
spec:
  replicas: ${replicas}
  template:
    spec:
      containers:
        - name: ${name}
          image: ${image}
${extra}`

const configMap = (
  name: string,
  namespace: string,
  values: Record<string, string>
) => `apiVersion: v1
kind: ConfigMap
metadata:
  name: ${name}
  namespace: ${namespace}
data:
${Object.entries(values)
  .map(([key, value]) => `  ${key}: ${JSON.stringify(value)}`)
  .join('\n')}`

const service = (name: string, namespace: string, port = 80) => `apiVersion: v1
kind: Service
metadata:
  name: ${name}
  namespace: ${namespace}
spec:
  type: ClusterIP
  ports:
    - port: ${port}
      targetPort: 8080`

const line = (
  resource: Pick<IResource, 'namespace' | 'name' | 'kind' | 'api'>,
  action: 'added' | 'changed' | 'destroyed'
) =>
  `${resource.namespace}, ${resource.name}, ${resource.kind} (${resource.api}) to be ${action}`

const plan = (
  op: string,
  resources: Array<IResource & { action: 'added' | 'changed' | 'destroyed' }>
): THelmPlan => {
  const counts = resources.reduce(
    (total, resource) => {
      total[resource.action] += 1
      return total
    },
    { added: 0, changed: 0, destroyed: 0 }
  )

  return {
    op,
    plan: [
      ...resources.map((resource) => line(resource, resource.action)),
      `Plan: ${counts.added} to add, ${counts.changed} to change, ${counts.destroyed} to destroy`,
    ].join('\n'),
    helm_content_diff: resources.map(
      ({ action: _action, ...resource }) => resource
    ),
  }
}

export const mixedHelmPlan = plan('upgrade', [
  {
    api: 'apps/v1',
    kind: 'Deployment',
    name: 'payments',
    namespace: 'production',
    action: 'changed',
    before: deployment({
      name: 'payments',
      namespace: 'production',
      image: 'example.com/payments:1.2.0',
      replicas: 2,
    }),
    after: deployment({
      name: 'payments',
      namespace: 'production',
      image: 'example.com/payments:1.3.0',
      replicas: 3,
    }),
  },
  {
    api: 'v1',
    kind: 'Service',
    name: 'payments-api',
    namespace: 'production',
    action: 'added',
    before: '',
    after: service('payments-api', 'production'),
  },
  {
    api: 'v1',
    kind: 'ConfigMap',
    name: 'payments-legacy',
    namespace: 'production',
    action: 'destroyed',
    before: configMap('payments-legacy', 'production', { MODE: 'legacy' }),
    after: '',
  },
])

export const nginxIngressUpgradePlan = plan('upgrade', [
  {
    api: 'apps/v1',
    kind: 'Deployment',
    name: 'ingress-controller',
    namespace: 'ingress-system',
    action: 'changed',
    before: deployment({
      name: 'ingress-controller',
      namespace: 'ingress-system',
      image: 'example.com/ingress-controller:1.9.4',
      replicas: 2,
    }),
    after: deployment({
      name: 'ingress-controller',
      namespace: 'ingress-system',
      image: 'example.com/ingress-controller:1.10.1',
      replicas: 3,
      extra: '          args:\n            - --enable-metrics=true',
    }),
  },
  {
    api: 'v1',
    kind: 'ConfigMap',
    name: 'ingress-controller',
    namespace: 'ingress-system',
    action: 'changed',
    before: configMap('ingress-controller', 'ingress-system', {
      'proxy-body-size': '8m',
      'proxy-read-timeout': '60',
    }),
    after: configMap('ingress-controller', 'ingress-system', {
      'proxy-body-size': '16m',
      'proxy-read-timeout': '120',
      'enable-real-ip': 'true',
    }),
  },
  {
    api: 'v1',
    kind: 'Service',
    name: 'ingress-controller',
    namespace: 'ingress-system',
    action: 'changed',
    before: service('ingress-controller', 'ingress-system', 80),
    after: service('ingress-controller', 'ingress-system', 8080),
  },
])

export const certManagerInstallPlan = plan(
  'install',
  [
    ['Deployment', 'certificate-controller', 'apps/v1'],
    ['ServiceAccount', 'certificate-controller', 'v1'],
    ['ClusterRole', 'certificate-controller', 'rbac.authorization.k8s.io/v1'],
    [
      'ClusterRoleBinding',
      'certificate-controller',
      'rbac.authorization.k8s.io/v1',
    ],
    ['Deployment', 'certificate-webhook', 'apps/v1'],
    ['Service', 'certificate-webhook', 'v1'],
  ].map(([kind, name, api]) => ({
    api,
    kind,
    name,
    namespace: 'certificate-system',
    action: 'added' as const,
    before: '',
    after: `apiVersion: ${api}
kind: ${kind}
metadata:
  name: ${name}
  namespace: certificate-system
  labels:
    app.kubernetes.io/name: ${name}`,
  }))
)

export const postgresOperatorUpgradePlan = plan('upgrade', [
  {
    api: 'apps/v1',
    kind: 'Deployment',
    name: 'database-operator',
    namespace: 'database-system',
    action: 'changed',
    before: deployment({
      name: 'database-operator',
      namespace: 'database-system',
      image: 'example.com/database-operator:1.22.1',
    }),
    after: deployment({
      name: 'database-operator',
      namespace: 'database-system',
      image: 'example.com/database-operator:1.23.0',
      extra: '          args:\n            - --leader-elect',
    }),
  },
  {
    api: 'v1',
    kind: 'ConfigMap',
    name: 'database-operator',
    namespace: 'database-system',
    action: 'changed',
    before: configMap('database-operator', 'database-system', {
      CREATE_ANY_SERVICE: 'false',
    }),
    after: configMap('database-operator', 'database-system', {
      CREATE_ANY_SERVICE: 'true',
      ENABLE_INPLACE_UPDATES: 'true',
    }),
  },
  {
    api: 'batch/v1',
    kind: 'CronJob',
    name: 'database-backup',
    namespace: 'database-system',
    action: 'added',
    before: '',
    after: `apiVersion: batch/v1
kind: CronJob
metadata:
  name: database-backup
  namespace: database-system
spec:
  schedule: "0 2 * * *"`,
  },
])

export const prometheusStackChangePlan = plan('upgrade', [
  {
    api: 'apps/v1',
    kind: 'Deployment',
    name: 'metrics-server',
    namespace: 'monitoring',
    action: 'changed',
    before: deployment({
      name: 'metrics-server',
      namespace: 'monitoring',
      image: 'example.com/metrics-server:2.50.1',
    }),
    after: deployment({
      name: 'metrics-server',
      namespace: 'monitoring',
      image: 'example.com/metrics-server:2.51.2',
      extra: '          args:\n            - --enable-remote-write',
    }),
  },
  {
    api: 'v1',
    kind: 'ConfigMap',
    name: 'metrics-config',
    namespace: 'monitoring',
    action: 'changed',
    before: configMap('metrics-config', 'monitoring', {
      interval: '30s',
      retention: '15d',
    }),
    after: configMap('metrics-config', 'monitoring', {
      interval: '15s',
      retention: '30d',
      remote_write: 'enabled',
    }),
  },
  {
    api: 'v1',
    kind: 'Secret',
    name: 'metrics-credentials',
    namespace: 'monitoring',
    action: 'changed',
    before: `apiVersion: v1
kind: Secret
metadata:
  name: metrics-credentials
  namespace: monitoring
stringData:
  endpoint: https://metrics.example.com/v1`,
    after: `apiVersion: v1
kind: Secret
metadata:
  name: metrics-credentials
  namespace: monitoring
stringData:
  endpoint: https://metrics.example.com/v2`,
  },
  {
    api: 'monitoring.coreos.com/v1',
    kind: 'ServiceMonitor',
    name: 'metrics-server',
    namespace: 'monitoring',
    action: 'changed',
    before: `apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: metrics-server
  namespace: monitoring
spec:
  endpoints:
    - interval: 30s`,
    after: `apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: metrics-server
  namespace: monitoring
spec:
  endpoints:
    - interval: 15s
    - interval: 30s
      port: grpc`,
  },
])

export const redisClusterRollbackPlan = plan('rollback', [
  {
    api: 'apps/v1',
    kind: 'Deployment',
    name: 'cache',
    namespace: 'cache-system',
    action: 'changed',
    before: deployment({
      name: 'cache',
      namespace: 'cache-system',
      image: 'example.com/cache:7.2.4',
      replicas: 3,
    }),
    after: deployment({
      name: 'cache',
      namespace: 'cache-system',
      image: 'example.com/cache:7.2.3',
      replicas: 3,
    }),
  },
  {
    api: 'v1',
    kind: 'Service',
    name: 'cache-sentinel',
    namespace: 'cache-system',
    action: 'destroyed',
    before: service('cache-sentinel', 'cache-system', 26379),
    after: '',
  },
  {
    api: 'v1',
    kind: 'ConfigMap',
    name: 'cache',
    namespace: 'cache-system',
    action: 'changed',
    before: configMap('cache', 'cache-system', {
      appendonly: 'yes',
      sentinel: 'enabled',
    }),
    after: configMap('cache', 'cache-system', { appendonly: 'yes' }),
  },
])

export const vmagentSingleRemovalPlan = {
  op: 'upgrade',
  plan: [
    '\u001b[0;33mobservability, metrics-agent, Deployment (apps/v1) to be changed.\u001b[0m',
    'Plan: 0 to add, 1 to change, 0 to destroy',
  ].join('\n'),
  helm_content_diff: [
    {
      api: 'apps/v1',
      kind: 'Deployment',
      name: 'metrics-agent',
      namespace: 'observability',
      entries: [
        { type: 0, payload: 'apiVersion: apps/v1' },
        { type: 0, payload: 'kind: Deployment' },
        { type: 0, payload: 'metadata:' },
        { type: 0, payload: '  name: metrics-agent' },
        { type: 0, payload: 'spec:' },
        { type: 0, payload: '  template:' },
        { type: 0, payload: '    spec:' },
        { type: 0, payload: '      containers:' },
        { type: 0, payload: '        - name: metrics-agent' },
        { type: 0, payload: '          args:' },
        { type: 1, payload: '            - --legacy-endpoint=:8429' },
        { type: 0, payload: '            - --logger-format=json' },
      ],
    },
  ],
} as unknown as THelmPlan

const longValue =
  'https://api.example.com/v1/events?environment=production&region=us-west-2&include=deployments,services,configmaps'

export const longAnnotationsAndEnvVarsPlan = plan('upgrade', [
  {
    api: 'apps/v1',
    kind: 'Deployment',
    name: 'api',
    namespace: 'production',
    action: 'changed',
    before: deployment({
      name: 'api',
      namespace: 'production',
      image: 'example.com/api:2.14.0-amd64-sha256-a3f8c2e1d09b4f7e',
      extra: `          env:
            - name: EVENTS_ENDPOINT
              value: ${longValue}`,
    }),
    after: deployment({
      name: 'api',
      namespace: 'production',
      image: 'example.com/api:2.15.0-amd64-sha256-b4c9d3f2e1a8b5d2',
      extra: `          env:
            - name: EVENTS_ENDPOINT
              value: ${longValue}&include=secrets`,
    }),
  },
  {
    api: 'v1',
    kind: 'ConfigMap',
    name: 'api-config',
    namespace: 'production',
    action: 'changed',
    before: configMap('api-config', 'production', {
      CALLBACKS: 'https://app.example.com/auth/callback',
    }),
    after: configMap('api-config', 'production', {
      CALLBACKS:
        'https://app.example.com/auth/callback,https://admin.example.com/auth/callback',
    }),
  },
])

export const singleImageTagChangePlan = plan('upgrade', [
  {
    api: 'apps/v1',
    kind: 'Deployment',
    name: 'api-auth',
    namespace: 'apps',
    action: 'changed',
    before: deployment({
      name: 'api-auth',
      namespace: 'apps',
      image: 'example.com/api-auth:0.19.1129',
    }),
    after: deployment({
      name: 'api-auth',
      namespace: 'apps',
      image: 'example.com/api-auth:0.19.1204',
    }),
  },
])

const envBlock = (count: number) =>
  Array.from(
    { length: count },
    (_, index) =>
      `            - name: FEATURE_${String(index).padStart(2, '0')}\n              value: "${index % 2 ? 'disabled' : 'enabled'}"`
  ).join('\n')

const largeDeployment = (image: string, replicas: number, memory: string) =>
  deployment({
    name: 'api-auth',
    namespace: 'apps',
    image,
    replicas,
    extra: `          env:
${envBlock(60)}
          resources:
            limits:
              memory: ${memory}`,
  })

export const largeDeploymentSingleChangePlan = plan('upgrade', [
  {
    api: 'apps/v1',
    kind: 'Deployment',
    name: 'api-auth',
    namespace: 'apps',
    action: 'changed',
    before: largeDeployment('example.com/api-auth:0.19.1129', 2, '512Mi'),
    after: largeDeployment('example.com/api-auth:0.19.1204', 2, '512Mi'),
  },
])

export const largeDeploymentScatteredChangesPlan = plan('upgrade', [
  {
    api: 'apps/v1',
    kind: 'Deployment',
    name: 'api-auth',
    namespace: 'apps',
    action: 'changed',
    before: largeDeployment('example.com/api-auth:0.19.1129', 2, '512Mi'),
    after: largeDeployment('example.com/api-auth:0.19.1204', 4, '1Gi'),
  },
])
