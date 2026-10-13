import type { TInstallResource } from '@/types'
import type { TResourceBoardItem } from '@/components/install-resources/InstallResourceBoard/InstallResourceBoard'

export type THealthPreviewState =
  | 'healthy'
  | 'progressing'
  | 'degraded'
  | 'unhealthy'
  | 'access-error'
  | 'no-observations'
  | 'empty'
  | 'loading'
  | 'stale'
  | 'stale-checks'

export function healthPreviewFixture(state: THealthPreviewState) {
  const now = Date.now()
  const observedAt = new Date(
    now - (state === 'stale' ? 12 * 60_000 : 42_000)
  ).toISOString()
  const problem = state === 'degraded' || state === 'unhealthy'
  const ready =
    state === 'unhealthy'
      ? 0
      : state === 'degraded' || state === 'progressing'
        ? 2
        : 3
  const item = (
    kind: string,
    name: string,
    summary: string,
    details: unknown,
    health = 'healthy',
    apiGroup = ''
  ): TResourceBoardItem => ({
    id: `${apiGroup}/${kind}/${name}`,
    summary,
    resource: {
      kind,
      name,
      namespace: 'acme',
      provider: 'kubernetes',
      api_group: apiGroup,
      health,
      observed_at: observedAt,
      details: JSON.stringify(details),
    },
  })
  const workload = (
    name: string,
    count = 3,
    health = 'healthy',
    kind = 'Deployment'
  ) =>
    item(
      kind,
      name,
      `${count} / 3 ready`,
      {
        spec: { replicas: 3, strategy: { type: 'RollingUpdate' } },
        status: {
          replicas: 3,
          readyReplicas: count,
          availableReplicas: count,
          updatedReplicas: 3,
          conditions: [
            {
              type: 'Available',
              status: count > 0 ? 'True' : 'False',
              reason:
                count > 0
                  ? 'MinimumReplicasAvailable'
                  : 'MinimumReplicasUnavailable',
            },
          ],
        },
      },
      health,
      'apps'
    )
  const pod = (name: string, failed = false) =>
    item(
      'Pod',
      name,
      failed ? 'CrashLoopBackOff · 17 restarts' : '1 / 1 ready',
      {
        spec: { nodeName: 'worker-node-02' },
        status: {
          phase: 'Running',
          podIP: '10.0.2.18',
          containerStatuses: [
            {
              name: 'app',
              ready: !failed,
              restartCount: failed ? 17 : 0,
              state: failed
                ? { waiting: { reason: 'CrashLoopBackOff' } }
                : { running: { startedAt: observedAt } },
            },
          ],
          conditions: [
            {
              type: 'Ready',
              status: failed ? 'False' : 'True',
              reason: failed ? 'ContainersNotReady' : 'ContainersReady',
            },
          ],
        },
        ...(failed
          ? {
              diagnosis: {
                reason: 'CrashLoopBackOff',
                message: 'The app container has restarted 17 times.',
              },
            }
          : {}),
      },
      failed ? 'unhealthy' : 'healthy'
    )
  const pvc = (name: string) =>
    item('PersistentVolumeClaim', name, 'Bound · 20Gi', {
      spec: { storageClassName: 'gp3', volumeName: `pvc-${name}` },
      status: { phase: 'Bound', capacity: { storage: '20Gi' } },
    })
  const database = (name: string, apiGroup: string) =>
    item(
      'Database',
      name,
      'No health assessment available',
      {
        spec: { engine: 'postgres', replicas: 2 },
        status: { phase: 'Reconciling' },
      },
      'unknown',
      apiGroup
    )
  const gateway = item(
    'Gateway',
    'public-gateway',
    'HTTPS · :443',
    {
      spec: {
        gatewayClassName: 'example-gateway',
        listeners: [{ name: 'https', protocol: 'HTTPS', port: 443 }],
      },
      status: {
        conditions: [
          { type: 'Accepted', status: 'True' },
          { type: 'Programmed', status: 'True' },
        ],
      },
    },
    'healthy',
    'gateway.networking.k8s.io'
  )
  const route = item(
    'HTTPRoute',
    'api-route',
    'api.example.com',
    {
      spec: {
        hostnames: ['api.example.com'],
        parentRefs: [{ name: 'public-gateway' }],
        rules: [{ backendRefs: [{ name: 'api', port: 80 }] }],
      },
      status: {
        parents: [
          {
            parentRef: { name: 'public-gateway' },
            conditions: [
              { type: 'Accepted', status: 'True' },
              { type: 'ResolvedRefs', status: 'True' },
            ],
          },
        ],
      },
    },
    'unknown',
    'gateway.networking.k8s.io'
  )

  const groups = [
    {
      id: 'api',
      name: 'api',
      items: [
        workload(
          'api',
          ready,
          problem || state === 'progressing' ? state : 'healthy'
        ),
        pod('api-a1', state === 'unhealthy'),
        pod('api-b2', state === 'unhealthy'),
        pod('api-c3', problem),
        item(
          'Service',
          'api',
          '80 → 8080 / TCP',
          {
            spec: {
              type: 'ClusterIP',
              clusterIP: '10.0.0.42',
              selector: { app: 'api' },
              ports: [
                { name: 'http', port: 80, targetPort: 8080, protocol: 'TCP' },
              ],
            },
          },
          'not-applicable'
        ),
        item(
          'Ingress',
          'api-public',
          'api.example.com',
          {
            spec: {
              ingressClassName: 'nginx',
              rules: [
                {
                  host: 'api.example.com',
                  http: {
                    paths: [
                      {
                        path: '/',
                        backend: {
                          service: { name: 'api', port: { number: 80 } },
                        },
                      },
                    ],
                  },
                },
              ],
            },
            status: {
              loadBalancer: { ingress: [{ hostname: 'lb.example.com' }] },
            },
          },
          'healthy',
          'networking.k8s.io'
        ),
        pvc('api-data'),
        item(
          'HorizontalPodAutoscaler',
          'api-autoscaler',
          '3 replicas',
          {
            spec: {
              scaleTargetRef: {
                apiVersion: 'apps/v1',
                kind: 'Deployment',
                name: 'api',
              },
              minReplicas: 2,
              maxReplicas: 8,
            },
            status: {
              currentReplicas: 3,
              desiredReplicas: 3,
              conditions: [
                {
                  type: 'AbleToScale',
                  status: 'True',
                  reason: 'SucceededGetScale',
                },
                {
                  type: 'ScalingActive',
                  status: 'True',
                  reason: 'ValidMetricFound',
                },
              ],
            },
          },
          'healthy',
          'autoscaling'
        ),
        item(
          'CronJob',
          'api-cleanup',
          'Every hour',
          {
            spec: {
              schedule: '0 * * * *',
              suspend: false,
              concurrencyPolicy: 'Forbid',
            },
            status: { lastScheduleTime: observedAt },
          },
          'healthy',
          'batch'
        ),
        gateway,
        route,
        database('api-db', 'db.example.com'),
        item(
          'Certificate',
          'api-tls',
          'No health assessment available',
          {
            spec: {
              dnsNames: ['api.example.com'],
              issuerRef: { name: 'example-issuer' },
            },
            status: {
              conditions: [
                { type: 'Ready', status: 'False', reason: 'Issuing' },
              ],
            },
          },
          'unknown',
          'cert-manager.io'
        ),
      ],
    },
    {
      id: 'worker',
      name: 'worker',
      items: [
        workload('worker'),
        pod('worker-a1'),
        pod('worker-b2'),
        item(
          'CronJob',
          'queue-cleanup',
          'Every 15 minutes',
          { spec: { schedule: '*/15 * * * *', suspend: false } },
          'healthy',
          'batch'
        ),
        item(
          'Job',
          'queue-cleanup-293842',
          '1 / 1 completed',
          {
            spec: { completions: 1 },
            status: {
              succeeded: 1,
              conditions: [
                {
                  type: 'Complete',
                  status: 'True',
                  reason: 'CompletionsReached',
                },
              ],
            },
          },
          'healthy',
          'batch'
        ),
      ],
    },
    {
      id: 'ledger',
      name: 'ledger',
      items: [
        database('ledger-db', 'databases.example.com'),
        workload('ledger-db', 3, 'healthy', 'StatefulSet'),
        pod('ledger-db-0'),
        pvc('data-ledger-db-0'),
      ],
    },
  ].map((group) => ({
    ...group,
    items: group.items.map((entry) => ({
      ...entry,
      resource: {
        ...entry.resource,
        source: 'component',
        component_id: `cmp-${group.id}`,
        install_component_id: `instcmp-${group.id}`,
        namespace: group.id === 'worker' ? 'jobs' : 'acme',
      },
    })),
  }))

  const sandboxResources: TInstallResource[] = [
    item(
      'Deployment',
      'cert-manager',
      '1 / 1 ready',
      {
        spec: { replicas: 1 },
        status: { readyReplicas: 1, availableReplicas: 1 },
      },
      'healthy',
      'apps'
    ),
    item(
      'Certificate',
      'webhook-tls',
      'Ready=True',
      { status: { conditions: [{ type: 'Ready', status: 'True' }] } },
      'unknown',
      'cert-manager.io'
    ),
  ].map(({ resource }) => ({
    ...resource,
    source: 'sandbox',
    owner_name: 'cert-manager',
    namespace: 'cert-manager',
  }))

  const checks: TInstallResource[] = [
    {
      name: 'api-http',
      source: 'component',
      install_component_id: 'instcmp-api',
      kind: 'HTTPProbe',
      provider: 'probe',
      health: problem ? 'unhealthy' : 'healthy',
      message: problem
        ? 'HTTP endpoint returned 503.'
        : 'HTTP endpoint returned 200.',
      observed_at: observedAt,
      details: JSON.stringify({
        probe: {
          type: 'http',
          target: 'https://api.example.com/healthz',
          status_code: problem ? 503 : 200,
          latency_ms: 28,
        },
      }),
    },
    {
      name: 'ledger-tcp',
      source: 'component',
      install_component_id: 'instcmp-ledger',
      kind: 'TCPProbe',
      provider: 'probe',
      health: 'healthy',
      message: 'TCP connection established.',
      observed_at: new Date(
        now - (state === 'stale' ? 12 * 60_000 : 78_000)
      ).toISOString(),
      details: JSON.stringify({
        probe: {
          type: 'tcp',
          target: 'ledger-db.acme.svc:5432',
          latency_ms: 4,
        },
      }),
    },
    {
      name: 'queue-lag',
      source: 'component',
      install_component_id: 'instcmp-worker',
      kind: 'CustomCheck',
      provider: 'custom',
      health: 'healthy',
      message: 'Consumer lag is within the configured threshold.',
      observed_at: new Date(
        now -
          (state === 'stale'
            ? 12 * 60_000
            : state === 'stale-checks'
              ? 180_000
              : 90_000)
      ).toISOString(),
      stale_after_seconds: 120,
      details: JSON.stringify({ lag: 12, threshold: 100 }),
    },
    ...(state === 'stale-checks'
      ? [
          {
            name: 'backup-status',
            source: 'component',
            install_component_id: 'instcmp-worker',
            kind: 'CustomCheck',
            provider: 'custom',
            health: 'healthy',
            message: 'Latest backup completed successfully.',
            observed_at: new Date(now - 180_000).toISOString(),
            stale_after_seconds: 600,
            details: JSON.stringify({ completed: true }),
          },
        ]
      : []),
  ]

  return { groups, sandboxResources, checks, observedAt }
}
