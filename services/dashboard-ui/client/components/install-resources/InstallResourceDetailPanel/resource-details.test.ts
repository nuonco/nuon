import { describe, expect, test } from 'bun:test'
import type { TInstallResource } from '@/types'
import {
  describeResource,
  parseResourceDetails,
  resourceSections,
  resourceStatusFields,
} from './resource-details'

const output = (kind: string, details?: unknown) =>
  describeResource({
    name: 'api',
    namespace: 'acme',
    provider: kind.endsWith('Probe') ? 'probe' : 'kubernetes',
    kind,
    details: details === undefined ? undefined : JSON.stringify(details),
  } as TInstallResource)

describe('resource descriptions', () => {
  test('keeps desired, current and ready replica counts distinct, including zero', () => {
    const text = output('Deployment', {
      spec: {
        replicas: 5,
        selector: { matchLabels: { app: 'api', tier: 'backend' } },
      },
      status: { replicas: 4, readyReplicas: 0, availableReplicas: 2 },
    })
    expect(text).toMatch(/Desired replicas:\s+5/)
    expect(text).toMatch(/Current replicas:\s+4/)
    expect(text).toMatch(/Ready replicas:\s+0/)
    expect(text).toMatch(/Available replicas:\s+2/)
    expect(text).toContain('app=api, tier=backend')
    expect(text).not.toContain('Updated replicas')
    expect(
      output('Deployment', { status: { readyReplicas: 2 } })
    ).not.toContain('Desired replicas')
  })

  test('reports both ready and blocked containers, plus init container failures', () => {
    const text = output('Pod', {
      status: {
        phase: 'Running',
        initContainerStatuses: [
          {
            name: 'setup',
            ready: false,
            restartCount: 0,
            state: { waiting: { reason: 'ImagePullBackOff' } },
          },
        ],
        containerStatuses: [
          {
            name: 'api',
            ready: false,
            restartCount: 7,
            state: { waiting: { reason: 'CrashLoopBackOff' } },
          },
          {
            name: 'proxy',
            ready: true,
            restartCount: 0,
            state: { running: {} },
          },
        ],
      },
    })
    expect(text).toMatch(/setup\s+false\s+0\s+ImagePullBackOff/)
    expect(text).toMatch(/api\s+false\s+7\s+CrashLoopBackOff/)
    expect(text).toMatch(/proxy\s+true\s+0\s+Running/)
  })

  test('uses service target ports, not the externally exposed port', () => {
    const text = output('Service', {
      spec: {
        type: 'ClusterIP',
        clusterIP: '10.0.0.42',
        selector: { app: 'api' },
        ports: [{ name: 'http', port: 80, targetPort: 8080, protocol: 'TCP' }],
      },
    })
    expect(text).toContain('10.0.0.42')
    expect(text).toMatch(/http\s+80\s+8080\s+TCP/)
  })

  test('preserves named and numeric ingress backends, including the default', () => {
    const text = output('Ingress', {
      spec: {
        defaultBackend: {
          service: { name: 'fallback', port: { number: 9000 } },
        },
        rules: [
          {
            host: 'api.example.com',
            http: {
              paths: [
                {
                  path: '/v1',
                  backend: { service: { name: 'api', port: { name: 'http' } } },
                },
              ],
            },
          },
        ],
      },
    })
    expect(text).toMatch(/api.example.com\s+\/v1\s+api\s+http/)
    expect(text).toMatch(/\*\s+\(default\)\s+fallback\s+9000/)
  })

  test('does not drop custom status fields, conditions or diagnosis', () => {
    const text = output('Widget', {
      status: {
        phase: 'Reconciling',
        conditions: [
          {
            type: 'Ready',
            status: 'False',
            reason: 'Waiting',
            message: 'Database unavailable',
          },
        ],
      },
      diagnosis: { warning: 'Connection refused' },
    })
    expect(text).toMatch(/Ready\s+False\s+Waiting/)
    expect(text).toContain('Ready: Database unavailable')
    expect(text).toMatch(/phase:\s+Reconciling/)
    expect(text).toMatch(/warning:\s+Connection refused/)
  })

  test('distinguishes current and desired HPA replicas and preserves its target kind', () => {
    const text = output('HorizontalPodAutoscaler', {
      spec: {
        scaleTargetRef: { kind: 'StatefulSet', name: 'queue' },
        minReplicas: 0,
        maxReplicas: 8,
      },
      status: { currentReplicas: 2, desiredReplicas: 5 },
    })
    expect(text).toContain('StatefulSet/queue')
    expect(text).toMatch(/Minimum replicas:\s+0/)
    expect(text).toMatch(/Current replicas:\s+2/)
    expect(text).toMatch(/Desired replicas:\s+5/)
  })

  test('renders Gateway listeners and cross-namespace non-Service route backends', () => {
    expect(
      output('Gateway', {
        spec: {
          listeners: [
            {
              name: 'https',
              protocol: 'HTTPS',
              port: 443,
              hostname: 'api.example.com',
            },
          ],
        },
      })
    ).toMatch(/https\s+HTTPS\s+443\s+api.example.com/)
    const text = output('HTTPRoute', {
      spec: {
        parentRefs: [{ name: 'public' }],
        rules: [
          {
            backendRefs: [
              {
                kind: 'Bucket',
                name: 'assets',
                namespace: 'media',
                port: 9000,
                weight: 0,
              },
              { name: 'api', port: 80 },
            ],
          },
        ],
      },
      status: {
        parents: [
          {
            conditions: [
              {
                type: 'ResolvedRefs',
                status: 'False',
                reason: 'RefNotPermitted',
              },
            ],
          },
        ],
      },
    })
    expect(text).toMatch(/Gateway\s+public\s+acme/)
    expect(text).toMatch(/Bucket\s+assets\s+media\s+9000\s+0/)
    expect(text).toMatch(/Service\s+api\s+acme\s+80\s+1/)
    expect(text).toContain('RefNotPermitted')
  })

  test('renders parent conditions separately with listener, controller and native values intact', () => {
    const resource = {
      provider: 'kubernetes',
      kind: 'HTTPRoute',
      namespace: 'acme',
      details: JSON.stringify({
        status: {
          parents: [
            {
              parentRef: { name: 'public', sectionName: 'https', port: 443 },
              controllerName: 'example.com/public',
              conditions: [
                { type: 'Accepted', status: 'True', reason: 'Accepted' },
              ],
            },
            {
              parentRef: {
                kind: 'Service',
                group: '',
                name: 'internal',
                namespace: 'edge',
              },
              controllerName: 'example.com/internal',
              conditions: [
                {
                  type: 'Accepted',
                  status: 'False',
                  reason: 'NotAllowedByListeners',
                  message: 'Route is not allowed by this listener.',
                  lastTransitionTime: '2026-10-08T12:40:00Z',
                  observedGeneration: 0,
                },
                { type: 'ResolvedRefs', status: 'Unknown' },
              ],
            },
          ],
        },
      }),
    }
    const parents = resourceSections(resource)
    expect(parents).toHaveLength(2)
    expect(parents[0].fields).toEqual([
      ['Kind', 'Gateway'],
      ['Name', 'public'],
      ['Namespace', 'acme'],
      ['API group', 'gateway.networking.k8s.io'],
      ['Section', 'https'],
      ['Port', '443'],
      ['Controller', 'example.com/public'],
    ])
    expect(parents[0].table?.rows).toEqual([['Accepted', 'True', 'Accepted']])
    expect(parents[1].fields).toContainEqual(['Kind', 'Service'])
    expect(parents[1].fields).toContainEqual(['Namespace', 'edge'])
    expect(parents[1].fields).toContainEqual(['API group', ''])
    expect(parents[1].fields).toContainEqual([
      'Controller',
      'example.com/internal',
    ])
    expect(parents[1].table?.headers).toEqual([
      'TYPE',
      'STATUS',
      'REASON',
      'LAST TRANSITION',
      'OBSERVED GENERATION',
    ])
    expect(parents[1].table?.rows).toEqual([
      [
        'Accepted',
        'False',
        'NotAllowedByListeners',
        '2026-10-08T12:40:00Z',
        '0',
      ],
      [
        'ResolvedRefs',
        'Unknown',
        'Not reported',
        'Not reported',
        'Not reported',
      ],
    ])
    expect(parents[1].text).toBe(
      'Accepted: Route is not allowed by this listener.'
    )
    const text = describeResource(resource)
    expect(text).not.toContain('parentRef:')
    expect(text).toMatch(/Accepted\s+False\s+NotAllowedByListeners/)
  })

  test('does not fabricate conditions when a parent has not reported them', () => {
    const parents = resourceSections({
      kind: 'HTTPRoute',
      namespace: 'acme',
      details: JSON.stringify({
        status: { parents: [{ parentRef: { name: 'public' } }] },
      }),
    })
    expect(parents).toHaveLength(1)
    expect(parents[0].fields).toContainEqual(['Name', 'public'])
    expect(parents[0].table).toBeUndefined()
    expect(parents[0].text).toBe('No conditions reported.')
    expect(
      resourceSections({
        kind: 'HTTPRoute',
        details: JSON.stringify({ status: { parents: [] } }),
      })
    ).toEqual([])
  })

  test('describes probe outcomes without inventing HTTP fields for TCP', () => {
    const http = output('HTTPProbe', {
      probe: {
        target: 'https://api.example.com/ready',
        status_code: 503,
        latency_ms: 0,
      },
    })
    expect(http).toMatch(/HTTP status:\s+503/)
    expect(http).toMatch(/Latency \(ms\):\s+0/)
    const tcp = output('TCPProbe', {
      probe: { target: 'db:5432', latency_ms: 12 },
    })
    expect(tcp).toMatch(/Target:\s+db:5432/)
    expect(tcp).not.toContain('HTTP status')
  })

  test('does not treat a custom Kubernetes kind called HTTPProbe as a runner probe', () => {
    const text = describeResource({
      provider: 'kubernetes',
      kind: 'HTTPProbe',
      api_group: 'checks.example.com',
      details: JSON.stringify({
        spec: { endpoint: 'https://api.example.com/ready' },
        status: { phase: 'Reconciling' },
      }),
    })
    expect(text).toContain('https://api.example.com/ready')
    expect(text).toContain('Reconciling')
    expect(text).not.toContain('HTTP status')
  })

  test('keeps exec probe arguments distinct and preserves a successful zero exit code', () => {
    const text = output('ExecProbe', {
      probe: {
        command: ['check', '--name', 'primary database'],
        exit_code: 0,
        latency_ms: 12,
      },
    })
    expect(text).toContain('["check","--name","primary database"]')
    expect(text).toMatch(/Exit code:\s+0/)
    expect(text).not.toContain('HTTP status')
  })

  test('keeps malformed and non-Kubernetes details readable without fabricating fields', () => {
    expect(parseResourceDetails('not json')).toBeUndefined()
    expect(
      describeResource({ name: 'check', details: 'Connection timed out' })
    ).toContain('Connection timed out')
    expect(output('Deployment')).not.toContain('replicas')
    expect(output('Check', { code: 503 })).toMatch(/code:\s+503/)
  })

  test('preserves scalar CR fields, nested data and conditions as independent sections', () => {
    const sections = resourceSections({
      kind: 'Database',
      health: 'unknown',
      details: JSON.stringify({
        spec: { engine: 'postgres', replicas: 0, paused: false },
        status: {
          phase: 'Reconciling',
          endpoints: [{ port: 5432, name: 'primary' }],
          conditions: [{ type: 'Ready', status: 'False' }],
        },
      }),
    })
    expect(sections[0]).toEqual({
      title: 'Spec',
      fields: [
        ['engine', 'postgres'],
        ['replicas', '0'],
        ['paused', 'false'],
      ],
    })
    expect(sections[1].text).toContain('port: 5432')
    expect(sections[1].text).not.toContain('conditions:')
    expect(sections[2].table?.rows).toEqual([
      ['Ready', 'False', 'Not reported'],
    ])
    expect(
      sections
        .flatMap((section) => section.fields || [])
        .map(([label]) => label)
    ).not.toContain('Health')
  })
})

describe('native snapshot summaries', () => {
  const fields = (kind: string, details: unknown) =>
    resourceStatusFields({
      provider: 'kubernetes',
      kind,
      health: 'healthy',
      native_status: 'Degraded',
      message: 'Normalized verdict must not override native fields',
      details: JSON.stringify(details),
    })

  test('reports readiness and total restarts independently of Running phase and derived health', () => {
    expect(
      fields('Pod', {
        status: {
          phase: 'Running',
          containerStatuses: [
            { ready: true, restartCount: 4, state: { running: {} } },
            {
              ready: false,
              restartCount: 7,
              state: { waiting: { reason: 'CrashLoopBackOff' } },
            },
          ],
        },
      })
    ).toEqual([
      ['Ready', '1 / 2'],
      ['Status', 'CrashLoopBackOff'],
      ['Restarts', '11'],
    ])
    expect(
      resourceStatusFields({
        kind: 'Pod',
        health: 'healthy',
        native_status: 'Healthy',
      })
    ).toEqual([])
  })

  test('surfaces init failures ahead of app state, counts restarts, and skips completed init containers', () => {
    const app = { ready: true, restartCount: 5, state: { running: {} } }
    expect(
      fields('Pod', {
        status: {
          phase: 'Pending',
          initContainerStatuses: [
            {
              ready: false,
              restartCount: 2,
              state: { waiting: { reason: 'ImagePullBackOff' } },
            },
          ],
          containerStatuses: [app],
        },
      })
    ).toEqual([
      ['Ready', '1 / 1'],
      ['Status', 'Init:ImagePullBackOff'],
      ['Restarts', '7'],
    ])
    expect(
      fields('Pod', {
        status: {
          phase: 'Running',
          initContainerStatuses: [
            { restartCount: 1, state: { terminated: { exitCode: 0 } } },
          ],
          containerStatuses: [app],
        },
      })
    ).toEqual([
      ['Ready', '1 / 1'],
      ['Status', 'Running'],
      ['Restarts', '6'],
    ])
  })

  test('does not invent readiness or restart counts for partial snapshots; shows termination exit code', () => {
    expect(
      fields('Pod', {
        status: {
          phase: 'Running',
          containerStatuses: [{ state: { running: {} } }],
        },
      })
    ).toEqual([['Status', 'Running']])
    expect(
      fields('Pod', {
        status: {
          phase: 'Failed',
          containerStatuses: [
            {
              ready: false,
              restartCount: 0,
              state: { terminated: { exitCode: 9 } },
            },
          ],
        },
      })
    ).toEqual([
      ['Ready', '0 / 1'],
      ['Status', 'ExitCode:9'],
      ['Restarts', '0'],
    ])
  })

  test('preserves custom condition types and polarity without a generic assessment', () => {
    expect(
      fields('Database', {
        status: {
          phase: 'Reconciling',
          conditions: [
            { type: 'Ready', status: 'False', reason: 'WaitingForDatabase' },
            { type: 'ReplicaFailure', status: 'True' },
          ],
        },
      })
    ).toEqual([
      ['phase', 'Reconciling'],
      ['Ready', 'False'],
      ['ReplicaFailure', 'True'],
    ])
  })

  test('shows reported gateway class references and distinct class controller fields without creating resources', () => {
    expect(
      fields('Gateway', {
        spec: { gatewayClassName: 'external-edge' },
        status: { conditions: [{ type: 'Accepted', status: 'False' }] },
      })
    ).toEqual([
      ['Class', 'external-edge'],
      ['Accepted', 'False'],
    ])
    expect(
      fields('GatewayClass', {
        spec: { controllerName: 'example.com/gateway' },
      })
    ).toEqual([['Controller', 'example.com/gateway']])
    expect(
      fields('IngressClass', { spec: { controller: 'example.com/ingress' } })
    ).toEqual([['Controller', 'example.com/ingress']])
    for (const kind of ['Gateway', 'GatewayClass', 'IngressClass'])
      expect(fields(kind, {})).toEqual([])
  })

  test('keeps HTTPRoute condition scope instead of collapsing opposing parent conditions', () => {
    expect(
      fields('HTTPRoute', {
        status: {
          parents: [
            {
              parentRef: { name: 'public', namespace: 'edge' },
              conditions: [{ type: 'Accepted', status: 'False' }],
            },
            {
              parentRef: { name: 'internal', namespace: 'acme' },
              conditions: [{ type: 'Accepted', status: 'True' }],
            },
          ],
        },
      })
    ).toEqual([
      ['edge/public: Accepted', 'False'],
      ['acme/internal: Accepted', 'True'],
    ])
  })

  test('uses explicit probe outcomes and zero latency without treating similarly named Kubernetes kinds as probes', () => {
    expect(
      resourceStatusFields({
        provider: 'probe',
        kind: 'HTTPProbe',
        health: 'healthy',
        details: JSON.stringify({ probe: { status_code: 503, latency_ms: 0 } }),
      })
    ).toEqual([
      ['Result', 'HTTP 503'],
      ['Latency', '0ms'],
    ])
    expect(
      resourceStatusFields({
        provider: 'probe',
        kind: 'ExecProbe',
        details: JSON.stringify({ probe: { exit_code: 0 } }),
      })
    ).toEqual([['Result', 'Exit code 0']])
    expect(
      resourceStatusFields({
        provider: 'probe',
        kind: 'TCPProbe',
        health: 'healthy',
      })
    ).toEqual([['Result', 'Connected']])
    expect(
      resourceStatusFields({
        provider: 'probe',
        kind: 'TCPProbe',
        health: 'unknown',
      })
    ).toEqual([])
    expect(fields('HTTPProbe', { status: { phase: 'Reconciling' } })).toEqual([
      ['phase', 'Reconciling'],
    ])
  })

  test('retains condition messages, timestamps and observed generations in descriptions but omits normalized labels', () => {
    const resource = {
      provider: 'kubernetes',
      kind: 'Deployment',
      health: 'unhealthy',
      native_status: 'Degraded',
      details: JSON.stringify({
        status: {
          conditions: [
            {
              type: 'Available',
              status: 'False',
              reason: 'MinimumReplicasUnavailable',
              message: 'No ready replicas',
              lastTransitionTime: '2026-10-08T12:40:00Z',
              observedGeneration: 0,
            },
          ],
        },
      }),
    }
    const conditions = resourceSections(resource).find(
      (section) => section.title === 'Conditions'
    )
    expect(conditions?.table?.rows).toEqual([
      [
        'Available',
        'False',
        'MinimumReplicasUnavailable',
        '2026-10-08T12:40:00Z',
        '0',
      ],
    ])
    const description = describeResource(resource)
    expect(description).toContain('Available: No ready replicas')
    expect(description).not.toContain('unhealthy')
    expect(description).not.toContain('Degraded')
  })
})
