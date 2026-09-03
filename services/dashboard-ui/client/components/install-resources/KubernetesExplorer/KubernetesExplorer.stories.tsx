import type { THelmRelease, TInstallResource } from '@/types'
import { KubernetesExplorer } from './KubernetesExplorer'

export default {
  title: 'Install Resources/KubernetesExplorer',
}

const NOW = '2026-08-03T10:00:00Z'

const res = (
  kind: string,
  apiGroup: string,
  obj: Record<string, any>,
  source: 'component' | 'sandbox' = 'component'
): TInstallResource => ({
  kind,
  api_group: apiGroup,
  name: obj?.metadata?.name,
  namespace: obj?.metadata?.namespace,
  details: JSON.stringify(obj),
  observed_at: NOW,
  source,
})

const meta = (
  name: string,
  namespace: string,
  createdDaysAgo: number,
  labels: Record<string, string> = {},
  owner?: { kind: string; name: string }
) => ({
  name,
  namespace,
  uid: `uid-${name}`,
  creationTimestamp: new Date(
    Date.parse(NOW) - createdDaysAgo * 24 * 60 * 60 * 1000
  ).toISOString(),
  labels,
  ...(owner
    ? {
        ownerReferences: [
          {
            apiVersion: 'apps/v1',
            kind: owner.kind,
            name: owner.name,
            uid: `uid-${owner.name}`,
            controller: true,
            blockOwnerDeletion: true,
          },
        ],
      }
    : {}),
})

const runningPod = ({
  name,
  namespace,
  app,
  image,
  node,
  days,
  restarts = 0,
  owner,
}: {
  name: string
  namespace: string
  app: string
  image: string
  node: string
  days: number
  restarts?: number
  owner?: { kind: string; name: string }
}) =>
  res('Pod', '', {
    apiVersion: 'v1',
    kind: 'Pod',
    metadata: meta(
      name,
      namespace,
      days,
      { app, 'pod-template-hash': name.split('-').at(-2) ?? '' },
      owner
    ),
    spec: {
      nodeName: node,
      containers: [{ name: app, image }],
    },
    status: {
      phase: 'Running',
      podIP: `10.0.${days}.${name.length}`,
      conditions: [
        { type: 'Ready', status: 'True' },
        { type: 'ContainersReady', status: 'True' },
      ],
      containerStatuses: [
        {
          name: app,
          image,
          ready: true,
          restartCount: restarts,
          state: { running: { startedAt: NOW } },
        },
      ],
    },
  })

const mockResources: TInstallResource[] = [
  res('Deployment', 'apps', {
    apiVersion: 'apps/v1',
    kind: 'Deployment',
    metadata: meta('web', 'acme-shop', 42, {
      app: 'web',
      'app.kubernetes.io/part-of': 'acme-shop',
    }),
    spec: {
      replicas: 3,
      selector: { matchLabels: { app: 'web' } },
      template: {
        spec: {
          containers: [
            {
              name: 'web',
              image: 'ghcr.io/acme/shop-web:1.24.3',
              ports: [{ containerPort: 8080 }],
            },
          ],
        },
      },
    },
    status: {
      replicas: 3,
      readyReplicas: 3,
      updatedReplicas: 3,
      availableReplicas: 3,
      conditions: [
        {
          type: 'Available',
          status: 'True',
          reason: 'MinimumReplicasAvailable',
          message: 'Deployment has minimum availability.',
        },
        {
          type: 'Progressing',
          status: 'True',
          reason: 'NewReplicaSetAvailable',
          message: 'ReplicaSet "web-6f9d54c8b7" has successfully progressed.',
        },
      ],
    },
  }),
  res('Deployment', 'apps', {
    apiVersion: 'apps/v1',
    kind: 'Deployment',
    metadata: meta('api', 'acme-shop', 42, {
      app: 'api',
      'app.kubernetes.io/part-of': 'acme-shop',
    }),
    spec: {
      replicas: 3,
      selector: { matchLabels: { app: 'api' } },
      template: {
        spec: {
          containers: [{ name: 'api', image: 'ghcr.io/acme/shop-api:2.1.0' }],
        },
      },
    },
    status: {
      replicas: 3,
      readyReplicas: 2,
      updatedReplicas: 3,
      availableReplicas: 2,
      conditions: [
        {
          type: 'Available',
          status: 'False',
          reason: 'MinimumReplicasUnavailable',
          message: 'Deployment does not have minimum availability.',
        },
        {
          type: 'Progressing',
          status: 'True',
          reason: 'ReplicaSetUpdated',
          message: 'ReplicaSet "api-7c4b9f6d55" is progressing.',
        },
      ],
    },
  }),
  res('Deployment', 'apps', {
    apiVersion: 'apps/v1',
    kind: 'Deployment',
    metadata: meta('worker', 'acme-shop', 42, { app: 'worker' }),
    spec: {
      replicas: 1,
      selector: { matchLabels: { app: 'worker' } },
      template: {
        spec: {
          containers: [
            { name: 'worker', image: 'ghcr.io/acme/shop-worker:2.1.0' },
          ],
        },
      },
    },
    status: {
      replicas: 1,
      readyReplicas: 1,
      updatedReplicas: 1,
      availableReplicas: 1,
      conditions: [
        {
          type: 'Available',
          status: 'True',
          reason: 'MinimumReplicasAvailable',
          message: 'Deployment has minimum availability.',
        },
      ],
    },
  }),

  res('StatefulSet', 'apps', {
    apiVersion: 'apps/v1',
    kind: 'StatefulSet',
    metadata: meta('postgres', 'acme-shop', 42, { app: 'postgres' }),
    spec: {
      replicas: 1,
      serviceName: 'postgres',
      selector: { matchLabels: { app: 'postgres' } },
      template: {
        spec: {
          containers: [{ name: 'postgres', image: 'postgres:16.4' }],
        },
      },
    },
    status: {
      replicas: 1,
      readyReplicas: 1,
      currentReplicas: 1,
    },
  }),

  res(
    'DaemonSet',
    'apps',
    {
      apiVersion: 'apps/v1',
      kind: 'DaemonSet',
      metadata: meta('log-agent', 'observability', 90, { app: 'log-agent' }),
      spec: {
        selector: { matchLabels: { app: 'log-agent' } },
        template: {
          spec: {
            containers: [
              { name: 'agent', image: 'public.ecr.aws/acme/log-agent:0.9.2' },
            ],
          },
        },
      },
      status: {
        desiredNumberScheduled: 4,
        currentNumberScheduled: 4,
        numberReady: 4,
        numberAvailable: 4,
      },
    },
    'sandbox'
  ),

  res('ReplicaSet', 'apps', {
    apiVersion: 'apps/v1',
    kind: 'ReplicaSet',
    metadata: meta(
      'web-6f9d54c8b7',
      'acme-shop',
      3,
      { app: 'web' },
      { kind: 'Deployment', name: 'web' }
    ),
    spec: { replicas: 3 },
    status: { replicas: 3, readyReplicas: 3, availableReplicas: 3 },
  }),
  res('ReplicaSet', 'apps', {
    apiVersion: 'apps/v1',
    kind: 'ReplicaSet',
    metadata: meta(
      'api-7c4b9f6d55',
      'acme-shop',
      1,
      { app: 'api' },
      { kind: 'Deployment', name: 'api' }
    ),
    spec: { replicas: 3 },
    status: { replicas: 3, readyReplicas: 2, availableReplicas: 2 },
  }),

  res('Job', 'batch', {
    apiVersion: 'batch/v1',
    kind: 'Job',
    metadata: meta('db-migrate-2114', 'acme-shop', 1, { app: 'db-migrate' }),
    spec: { completions: 1, parallelism: 1, backoffLimit: 4 },
    status: {
      succeeded: 1,
      startTime: NOW,
      completionTime: NOW,
      conditions: [
        {
          type: 'Complete',
          status: 'True',
          reason: 'CompletionsReached',
          message: 'Reached expected number of succeeded pods',
        },
      ],
    },
  }),

  runningPod({
    name: 'web-6f9d54c8b7-x2lqp',
    namespace: 'acme-shop',
    app: 'web',
    image: 'ghcr.io/acme/shop-web:1.24.3',
    node: 'ip-10-0-12-41.us-west-2.compute.internal',
    days: 3,
    owner: { kind: 'ReplicaSet', name: 'web-6f9d54c8b7' },
  }),
  runningPod({
    name: 'web-6f9d54c8b7-m8dwt',
    namespace: 'acme-shop',
    app: 'web',
    image: 'ghcr.io/acme/shop-web:1.24.3',
    node: 'ip-10-0-14-102.us-west-2.compute.internal',
    days: 3,
    owner: { kind: 'ReplicaSet', name: 'web-6f9d54c8b7' },
  }),
  runningPod({
    name: 'web-6f9d54c8b7-kv4njis',
    namespace: 'acme-shop',
    app: 'web',
    image: 'ghcr.io/acme/shop-web:1.24.3',
    node: 'ip-10-0-15-77.us-west-2.compute.internal',
    days: 3,
    owner: { kind: 'ReplicaSet', name: 'web-6f9d54c8b7' },
  }),
  runningPod({
    name: 'api-7c4b9f6d55-p9qzw',
    namespace: 'acme-shop',
    app: 'api',
    image: 'ghcr.io/acme/shop-api:2.1.0',
    node: 'ip-10-0-12-41.us-west-2.compute.internal',
    days: 1,
    owner: { kind: 'ReplicaSet', name: 'api-7c4b9f6d55' },
  }),
  runningPod({
    name: 'api-7c4b9f6d55-h3kfm',
    namespace: 'acme-shop',
    app: 'api',
    image: 'ghcr.io/acme/shop-api:2.1.0',
    node: 'ip-10-0-14-102.us-west-2.compute.internal',
    days: 1,
    restarts: 2,
    owner: { kind: 'ReplicaSet', name: 'api-7c4b9f6d55' },
  }),
  res('Pod', '', {
    apiVersion: 'v1',
    kind: 'Pod',
    metadata: meta('api-7c4b9f6d55-t6vrb', 'acme-shop', 1, {
      app: 'api',
      'pod-template-hash': '7c4b9f6d55',
    }),
    spec: {
      nodeName: 'ip-10-0-15-77.us-west-2.compute.internal',
      containers: [{ name: 'api', image: 'ghcr.io/acme/shop-api:2.1.0' }],
    },
    status: {
      phase: 'Running',
      podIP: '10.0.15.201',
      conditions: [
        {
          type: 'Ready',
          status: 'False',
          reason: 'ContainersNotReady',
          message: 'containers with unready status: [api]',
        },
      ],
      containerStatuses: [
        {
          name: 'api',
          image: 'ghcr.io/acme/shop-api:2.1.0',
          ready: false,
          restartCount: 17,
          state: {
            waiting: {
              reason: 'CrashLoopBackOff',
              message:
                'back-off 5m0s restarting failed container=api pod=api-7c4b9f6d55-t6vrb_acme-shop',
            },
          },
        },
      ],
    },
  }),
  runningPod({
    name: 'worker-59fd7b8c44-q2mxs',
    namespace: 'acme-shop',
    app: 'worker',
    image: 'ghcr.io/acme/shop-worker:2.1.0',
    node: 'ip-10-0-12-41.us-west-2.compute.internal',
    days: 2,
    owner: { kind: 'ReplicaSet', name: 'worker-59fd7b8c44' },
  }),
  runningPod({
    name: 'postgres-0',
    namespace: 'acme-shop',
    app: 'postgres',
    image: 'postgres:16.4',
    node: 'ip-10-0-14-102.us-west-2.compute.internal',
    days: 42,
    owner: { kind: 'StatefulSet', name: 'postgres' },
  }),
  runningPod({
    name: 'log-agent-8fkzt',
    namespace: 'observability',
    app: 'log-agent',
    image: 'public.ecr.aws/acme/log-agent:0.9.2',
    node: 'ip-10-0-12-41.us-west-2.compute.internal',
    days: 90,
    owner: { kind: 'DaemonSet', name: 'log-agent' },
  }),
  runningPod({
    name: 'log-agent-vw2pj',
    namespace: 'observability',
    app: 'log-agent',
    image: 'public.ecr.aws/acme/log-agent:0.9.2',
    node: 'ip-10-0-14-102.us-west-2.compute.internal',
    days: 90,
    owner: { kind: 'DaemonSet', name: 'log-agent' },
  }),

  res('Service', '', {
    apiVersion: 'v1',
    kind: 'Service',
    metadata: meta('web', 'acme-shop', 42, { app: 'web' }),
    spec: {
      type: 'ClusterIP',
      clusterIP: '172.20.14.101',
      selector: { app: 'web' },
      ports: [{ name: 'http', port: 80, targetPort: 8080, protocol: 'TCP' }],
    },
  }),
  res('Service', '', {
    apiVersion: 'v1',
    kind: 'Service',
    metadata: meta('api', 'acme-shop', 42, { app: 'api' }),
    spec: {
      type: 'ClusterIP',
      clusterIP: '172.20.18.42',
      selector: { app: 'api' },
      ports: [{ name: 'http', port: 80, targetPort: 3000, protocol: 'TCP' }],
    },
  }),
  res('Service', '', {
    apiVersion: 'v1',
    kind: 'Service',
    metadata: meta('postgres', 'acme-shop', 42, { app: 'postgres' }),
    spec: {
      type: 'ClusterIP',
      clusterIP: 'None',
      selector: { app: 'postgres' },
      ports: [{ name: 'pg', port: 5432, targetPort: 5432, protocol: 'TCP' }],
    },
  }),

  res('Ingress', 'networking.k8s.io', {
    apiVersion: 'networking.k8s.io/v1',
    kind: 'Ingress',
    metadata: meta('acme-shop', 'acme-shop', 42, { app: 'web' }),
    spec: {
      ingressClassName: 'alb',
      rules: [
        {
          host: 'shop.acme.dev',
          http: {
            paths: [
              {
                path: '/',
                pathType: 'Prefix',
                backend: { service: { name: 'web', port: { number: 80 } } },
              },
              {
                path: '/api',
                pathType: 'Prefix',
                backend: { service: { name: 'api', port: { number: 80 } } },
              },
            ],
          },
        },
      ],
      tls: [{ hosts: ['shop.acme.dev'], secretName: 'shop-acme-dev-tls' }],
    },
    status: {
      loadBalancer: {
        ingress: [
          { hostname: 'k8s-acmeshop-1f2e3d4c.us-west-2.elb.amazonaws.com' },
        ],
      },
    },
  }),

  res('PersistentVolumeClaim', '', {
    apiVersion: 'v1',
    kind: 'PersistentVolumeClaim',
    metadata: meta('data-postgres-0', 'acme-shop', 42, { app: 'postgres' }),
    spec: {
      accessModes: ['ReadWriteOnce'],
      storageClassName: 'gp3',
      resources: { requests: { storage: '20Gi' } },
      volumeName: 'pvc-9b1c2d3e-4f5a-6b7c-8d9e-0f1a2b3c4d5e',
    },
    status: {
      phase: 'Bound',
      accessModes: ['ReadWriteOnce'],
      capacity: { storage: '20Gi' },
    },
  }),

  res('Certificate', 'cert-manager.io', {
    apiVersion: 'cert-manager.io/v1',
    kind: 'Certificate',
    metadata: meta('shop-acme-dev-tls', 'acme-shop', 42, { app: 'web' }),
    spec: {
      secretName: 'shop-acme-dev-tls',
      dnsNames: ['shop.acme.dev'],
      issuerRef: { name: 'letsencrypt-prod', kind: 'ClusterIssuer' },
    },
    status: {
      notBefore: '2026-07-01T00:00:00Z',
      notAfter: '2026-09-29T00:00:00Z',
      renewalTime: '2026-08-30T00:00:00Z',
      conditions: [
        {
          type: 'Ready',
          status: 'True',
          reason: 'Ready',
          message: 'Certificate is up to date and has not expired',
        },
      ],
    },
  }),

  res('ExternalSecret', 'external-secrets.io', {
    apiVersion: 'external-secrets.io/v1beta1',
    kind: 'ExternalSecret',
    metadata: meta('shop-api-secrets', 'acme-shop', 42, { app: 'api' }),
    spec: {
      refreshInterval: '1h',
      secretStoreRef: {
        name: 'aws-secrets-manager',
        kind: 'ClusterSecretStore',
      },
      target: { name: 'shop-api-secrets' },
      data: [
        {
          secretKey: 'DATABASE_URL',
          remoteRef: { key: 'prod/acme-shop/database-url' },
        },
        {
          secretKey: 'STRIPE_KEY',
          remoteRef: { key: 'prod/acme-shop/stripe-key' },
        },
      ],
    },
    status: {
      refreshTime: NOW,
      conditions: [
        {
          type: 'Ready',
          status: 'True',
          reason: 'SecretSynced',
          message: 'secret synced',
        },
      ],
    },
  }),

  res('ExternalSecret', 'external-secrets.io', {
    apiVersion: 'external-secrets.io/v1beta1',
    kind: 'ExternalSecret',
    metadata: meta('shop-worker-secrets', 'acme-shop', 42, { app: 'worker' }),
    spec: {
      refreshInterval: '1h',
      secretStoreRef: {
        name: 'aws-secrets-manager',
        kind: 'ClusterSecretStore',
      },
      target: { name: 'shop-worker-secrets' },
      data: [
        {
          secretKey: 'QUEUE_URL',
          remoteRef: { key: 'prod/acme-shop/queue-url' },
        },
      ],
    },
    status: {
      conditions: [
        {
          type: 'Ready',
          status: 'False',
          reason: 'SecretSyncedError',
          message:
            'could not get secret data from provider: AccessDeniedException',
        },
      ],
    },
  }),
]

const daysAgo = (days: number) =>
  new Date(Date.parse(NOW) - days * 24 * 60 * 60 * 1000).toISOString()

const release = (
  name: string,
  namespace: string,
  version: number,
  status: string,
  updatedDaysAgo: number
): THelmRelease => ({
  name,
  namespace,
  version,
  status,
  type: 'helm.sh/release.v1',
  key: `sh.helm.release.v1.${name}.v${version}`,
  owner: 'helm',
  labels: {
    name,
    owner: 'helm',
    status,
    version: String(version),
  },
  updated_at: daysAgo(updatedDaysAgo),
})

const mockHelmReleases: THelmRelease[] = [
  release('acme-shop', 'acme-shop', 4, 'deployed', 1),
  release('acme-shop', 'acme-shop', 3, 'superseded', 6),
  release('acme-shop', 'acme-shop', 2, 'superseded', 14),
  release('acme-shop', 'acme-shop', 1, 'superseded', 30),
  release('cert-manager', 'cert-manager', 2, 'deployed', 12),
  release('cert-manager', 'cert-manager', 1, 'superseded', 45),
  release('shop-worker', 'acme-shop', 6, 'pending-upgrade', 0),
  release('shop-worker', 'acme-shop', 5, 'deployed', 9),
  release('legacy-adapter', 'acme-legacy', 3, 'failed', 2),
  release('legacy-adapter', 'acme-legacy', 2, 'deployed', 21),
]

export const Default = () => (
  <div className="h-[760px]">
    <KubernetesExplorer
      resources={mockResources}
      helmReleases={mockHelmReleases}
    />
  </div>
)

export const HelmOnly = () => (
  <div className="h-[560px]">
    <KubernetesExplorer resources={[]} helmReleases={mockHelmReleases} />
  </div>
)

export const CustomResourcesOnly = () => (
  <div className="h-[560px]">
    <KubernetesExplorer
      resources={mockResources.filter((r) =>
        ['Certificate', 'ExternalSecret'].includes(r.kind ?? '')
      )}
    />
  </div>
)

export const Empty = () => (
  <div className="h-[480px]">
    <KubernetesExplorer resources={[]} />
  </div>
)
