import { describe, expect, test } from 'bun:test'
import { layoutResourceKinds, type TResourceBoardItem } from './resource-layout'

const group = (
  kind: string,
  api_group = '',
  count = 1,
  provider = 'kubernetes'
): [string, TResourceBoardItem[]] => [
  `${provider}/${api_group}/${kind}`,
  Array.from({ length: count }, (_, index) => ({
    id: `${provider}/${api_group}/${kind}/${index}`,
    summary: 'Reported resource',
    resource: { provider, api_group, kind, name: `resource-${index}` },
  })),
]

describe('resource canvas layout', () => {
  test('orders observed networking, workload and pod zones independently of group sizes', () => {
    const input = [
      group('Pod', '', 8),
      group('Deployment', 'apps', 3),
      group('Service', '', 2),
      group('HTTPRoute', 'gateway.networking.k8s.io'),
      group('Gateway', 'gateway.networking.k8s.io'),
      group('Ingress', 'networking.k8s.io'),
    ]
    const layout = layoutResourceKinds(input)
    expect(
      layout.sections.map(({ title, x, y, count }) => [title, x, y, count])
    ).toEqual([
      ['Entry points', 0, 0, 2],
      ['Routes', 288, 0, 1],
      ['Services', 576, 0, 2],
      ['Workloads', 864, 0, 3],
      ['Pods', 1152, 0, 8],
    ])
    expect(layout.groups.map(({ key, x }) => [key, x])).toEqual([
      ['kubernetes/networking.k8s.io/Ingress', 0],
      ['kubernetes/gateway.networking.k8s.io/Gateway', 0],
      ['kubernetes/gateway.networking.k8s.io/HTTPRoute', 288],
      ['kubernetes//Service', 576],
      ['kubernetes/apps/Deployment', 864],
      ['kubernetes//Pod', 1152],
    ])
    expect(layout.groups[0].y).toBe(40)
    expect(
      layoutResourceKinds(
        input.map(([key, resources]) => [key, resources.slice(0, 1)])
      ).sections.map(({ x }) => x)
    ).toEqual([0, 288, 576, 864, 1152])
  })

  test('keeps Ingress first and Pods after Deployment when filtering to these three kinds', () => {
    const layout = layoutResourceKinds([
      group('Pod', '', 6),
      group('Deployment', 'apps', 3),
      group('Ingress', 'networking.k8s.io'),
    ])
    expect(
      layout.groups.map(({ resources, x, y }) => [
        resources[0].resource.kind,
        x,
        y,
      ])
    ).toEqual([
      ['Ingress', 0, 40],
      ['Deployment', 288, 40],
      ['Pod', 576, 40],
    ])
  })

  test('recognizes reported classes, Gateway API routes, service endpoints and controllers', () => {
    const layout = layoutResourceKinds([
      group('IngressClass', 'networking.k8s.io'),
      group('GatewayClass', 'gateway.networking.k8s.io'),
      ...['GRPCRoute', 'TCPRoute', 'TLSRoute', 'UDPRoute'].map((kind) =>
        group(kind, 'gateway.networking.k8s.io')
      ),
      group('EndpointSlice', 'discovery.k8s.io'),
      group('Endpoints'),
      ...['DaemonSet', 'ReplicaSet', 'StatefulSet'].map((kind) =>
        group(kind, 'apps')
      ),
      group('ReplicationController'),
      group('Job', 'batch'),
      group('CronJob', 'batch'),
      group('ReferenceGrant', 'gateway.networking.k8s.io'),
      group('BackendTLSPolicy', 'gateway.networking.k8s.io'),
    ])
    expect(layout.sections.map(({ title, count }) => [title, count])).toEqual([
      ['Entry points', 2],
      ['Routes', 4],
      ['Services', 2],
      ['Workloads', 6],
      ['Supporting resources', 2],
    ])
    expect(
      layout.groups
        .filter(({ x }) => x === 864)
        .map(({ resources }) => resources[0].resource.kind)
    ).toEqual([
      'ReplicaSet',
      'StatefulSet',
      'DaemonSet',
      'ReplicationController',
      'CronJob',
      'Job',
    ])
    expect(layout.groups.every(({ additional }) => !additional)).toBe(true)
  })

  test('keeps unfamiliar API groups and non-Kubernetes providers in Additional kinds', () => {
    const layout = layoutResourceKinds([
      group('Deployment', 'example.com'),
      group('HTTPRoute', 'example.com'),
      group('Pod', '', 1, 'aws'),
      group('Database', 'databases.example.com'),
      group('Database', 'storage.example.com'),
      group('Certificate', 'cert-manager.io'),
      group('Gateway', 'gateway.networking.k8s.io'),
    ])
    expect(layout.sections.map(({ title, count }) => [title, count])).toEqual([
      ['Entry points', 1],
      ['Additional kinds', 6],
    ])
    expect(
      layout.groups.filter(({ additional }) => additional).map(({ key }) => key)
    ).toEqual([
      'aws//Pod',
      'kubernetes/cert-manager.io/Certificate',
      'kubernetes/databases.example.com/Database',
      'kubernetes/example.com/Deployment',
      'kubernetes/example.com/HTTPRoute',
      'kubernetes/storage.example.com/Database',
    ])
  })

  test('places supporting resources and Additional kinds below main zones without overlaps across rows', () => {
    const layout = layoutResourceKinds([
      group('Deployment', 'apps', 4),
      group('Pod', '', 7),
      group('HorizontalPodAutoscaler', 'autoscaling', 3),
      group('PersistentVolumeClaim', '', 2),
      ...[
        'ConfigMap',
        'Secret',
        'ServiceAccount',
        'ResourceQuota',
        'LimitRange',
      ].map((kind, index) => group(kind, '', (index % 3) + 1)),
      ...Array.from({ length: 7 }, (_, index) =>
        group('Database', `operator-${index}.example.com`, (index % 4) + 1)
      ),
    ])
    const supporting = layout.sections.find(({ id }) => id === 'supporting')!
    const additional = layout.sections.find(({ id }) => id === 'additional')!
    const cards = layout.groups.flatMap(({ resources, x, y, headerHeight }) =>
      resources.map((item, index) => ({
        id: item.id,
        x,
        y: y + headerHeight + 16 + index * 232,
      }))
    )
    const mainCards = cards.filter(({ id }) => /\/(Deployment|Pod)\//.test(id))
    expect(supporting.y).toBeGreaterThan(
      Math.max(...mainCards.map(({ y }) => y + 208))
    )
    expect(additional.y).toBeGreaterThan(
      Math.max(
        ...cards
          .filter(({ id }) => !id.includes('operator-'))
          .map(({ y }) => y + 208)
      )
    )
    for (let a = 0; a < cards.length; a++) {
      for (let b = a + 1; b < cards.length; b++) {
        const left = cards[a]
        const right = cards[b]
        expect(
          left.x + 256 <= right.x ||
            right.x + 256 <= left.x ||
            left.y + 208 <= right.y ||
            right.y + 208 <= left.y
        ).toBe(true)
      }
    }
    expect(
      layout.groups.filter(({ additional }) => additional).at(-1)?.y
    ).toBeGreaterThan(layout.groups.filter(({ additional }) => additional)[0].y)
  })

  test('sorts by namespace, name and stable ID without mutating inputs', () => {
    const pods = group('Pod', '', 4)
    pods[1] = [
      {
        ...pods[1][0],
        id: 'z',
        resource: { ...pods[1][0].resource, namespace: 'jobs', name: 'a' },
      },
      {
        ...pods[1][1],
        id: 'b',
        resource: { ...pods[1][1].resource, namespace: 'acme', name: 'z' },
      },
      {
        ...pods[1][2],
        id: 'a',
        resource: { ...pods[1][2].resource, namespace: 'acme', name: 'z' },
      },
      {
        ...pods[1][3],
        id: 'c',
        resource: { ...pods[1][3].resource, namespace: 'acme', name: 'a' },
      },
    ]
    const input = [
      pods,
      group('Service'),
      group('Gateway', 'gateway.networking.k8s.io'),
    ]
    const before = structuredClone(input)
    const layout = layoutResourceKinds(input)
    expect(
      layout.groups
        .find(({ key }) => key === 'kubernetes//Pod')
        ?.resources.map(({ id }) => id)
    ).toEqual(['c', 'a', 'b', 'z'])
    expect(
      layoutResourceKinds(
        [...input]
          .reverse()
          .map(([key, resources]) => [key, [...resources].reverse()])
      )
    ).toEqual(layout)
    expect(input).toEqual(before)
  })

  test('omits unobserved zones and compacts filtered and custom-only inventories', () => {
    expect(layoutResourceKinds([])).toEqual({ sections: [], groups: [] })
    expect(layoutResourceKinds([['empty', []]])).toEqual({
      sections: [],
      groups: [],
    })
    const pods = layoutResourceKinds([group('Pod')])
    expect(
      pods.sections.map(({ title, x, y, width }) => [title, x, y, width])
    ).toEqual([['Pods', 0, 0, 256]])
    const custom = layoutResourceKinds([group('Widget', 'example.com')])
    expect(
      custom.sections.map(({ title, x, y, width }) => [title, x, y, width])
    ).toEqual([['Additional kinds', 0, 0, 256]])
    expect(custom.groups[0].y).toBe(64)
  })

  test('ignores names, selectors, ownership and health when assigning zones', () => {
    const input = [group('Service'), group('Deployment', 'apps'), group('Pod')]
    const changed = input.map(
      ([key, resources]) =>
        [
          key,
          resources.map((item) => ({
            ...item,
            resource: {
              ...item.resource,
              name: 'unrelated-name',
              health: 'unhealthy',
              details: JSON.stringify({
                metadata: {
                  ownerReferences: [{ kind: 'Widget', name: 'other' }],
                  labels: { app: 'other' },
                },
                spec: { selector: { app: 'other' } },
                status: { phase: 'Failed' },
              }),
            },
          })),
        ] as [string, TResourceBoardItem[]]
    )
    const positions = (layout: ReturnType<typeof layoutResourceKinds>) => ({
      sections: layout.sections,
      groups: layout.groups.map(
        ({ resources: _resources, ...position }) => position
      ),
    })
    expect(positions(layoutResourceKinds(changed))).toEqual(
      positions(layoutResourceKinds(input))
    )
  })
})
