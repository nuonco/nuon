import type { TInstallResource } from '@/types'

export type TResourceBoardItem = {
  id: string
  resource: TInstallResource
  summary: string
}

const zones = [
  {
    id: 'entry-points',
    title: 'Entry points',
    kinds: [
      'networking.k8s.io/Ingress',
      'networking.k8s.io/IngressClass',
      'gateway.networking.k8s.io/Gateway',
      'gateway.networking.k8s.io/GatewayClass',
    ],
  },
  {
    id: 'routes',
    title: 'Routes',
    kinds: [
      'gateway.networking.k8s.io/HTTPRoute',
      'gateway.networking.k8s.io/GRPCRoute',
      'gateway.networking.k8s.io/TCPRoute',
      'gateway.networking.k8s.io/TLSRoute',
      'gateway.networking.k8s.io/UDPRoute',
    ],
  },
  {
    id: 'services',
    title: 'Services',
    kinds: ['/Service', '/Endpoints', 'discovery.k8s.io/EndpointSlice'],
  },
  {
    id: 'workloads',
    title: 'Workloads',
    kinds: [
      'apps/Deployment',
      'apps/ReplicaSet',
      'apps/StatefulSet',
      'apps/DaemonSet',
      '/ReplicationController',
      'batch/CronJob',
      'batch/Job',
    ],
  },
  { id: 'pods', title: 'Pods', kinds: ['/Pod'] },
  {
    id: 'supporting',
    title: 'Supporting resources',
    kinds: [
      'autoscaling/HorizontalPodAutoscaler',
      '/PersistentVolumeClaim',
      '/PersistentVolume',
      '/ConfigMap',
      '/Secret',
      '/ServiceAccount',
      '/ResourceQuota',
      '/LimitRange',
      'storage.k8s.io/StorageClass',
      'networking.k8s.io/NetworkPolicy',
      'gateway.networking.k8s.io/ReferenceGrant',
      'gateway.networking.k8s.io/BackendTLSPolicy',
      'policy/PodDisruptionBudget',
      'rbac.authorization.k8s.io/Role',
      'rbac.authorization.k8s.io/RoleBinding',
      'rbac.authorization.k8s.io/ClusterRole',
      'rbac.authorization.k8s.io/ClusterRoleBinding',
    ],
  },
  { id: 'additional', title: 'Additional kinds', kinds: [] },
]

export function layoutResourceKinds(
  kinds: ReadonlyArray<readonly [string, TResourceBoardItem[]]>
) {
  const grouped = zones.map(() => [] as (typeof kinds)[number][])
  for (const [key, resources] of kinds) {
    if (!resources.length) continue
    const resource = resources[0].resource
    const zone =
      resource.provider === 'kubernetes'
        ? zones.findIndex(({ kinds }) =>
            kinds.includes(`${resource.api_group || ''}/${resource.kind || ''}`)
          )
        : -1
    grouped[zone < 0 ? zones.length - 1 : zone].push([
      key,
      [...resources].sort(
        (a, b) =>
          (a.resource.namespace || '').localeCompare(
            b.resource.namespace || ''
          ) ||
          (a.resource.name || '').localeCompare(b.resource.name || '') ||
          a.id.localeCompare(b.id)
      ),
    ])
  }
  grouped.forEach((groups, index) =>
    groups.sort(
      ([aKey, a], [bKey, b]) =>
        zones[index].kinds.indexOf(
          `${a[0].resource.api_group || ''}/${a[0].resource.kind || ''}`
        ) -
          zones[index].kinds.indexOf(
            `${b[0].resource.api_group || ''}/${b[0].resource.kind || ''}`
          ) || aKey.localeCompare(bKey)
    )
  )

  const sections: {
    id: string
    title: string
    count: number
    x: number
    y: number
    width: number
    secondary: boolean
  }[] = []
  const groups: {
    key: string
    resources: TResourceBoardItem[]
    x: number
    y: number
    headerHeight: number
    additional: boolean
  }[] = []
  const place = (
    [key, resources]: (typeof kinds)[number],
    x: number,
    y: number,
    additional = false
  ) => {
    const headerHeight = additional ? 48 : 24
    groups.push({ key, resources, x, y, headerHeight, additional })
    return y + headerHeight + 16 + resources.length * 232
  }
  const section = (
    index: number,
    x: number,
    y: number,
    width: number,
    secondary: boolean
  ) =>
    sections.push({
      id: zones[index].id,
      title: zones[index].title,
      count: grouped[index].reduce(
        (total, [, resources]) => total + resources.length,
        0
      ),
      x,
      y,
      width,
      secondary,
    })

  let column = 0
  let bottom = 0
  for (let index = 0; index < 5; index++) {
    if (!grouped[index].length) continue
    const x = column++ * 288
    section(index, x, 0, 256, false)
    let y = 40
    for (const group of grouped[index]) y = place(group, x, y) + 24
    bottom = Math.max(bottom, y)
  }
  const width =
    Math.max(
      column,
      ...grouped.slice(5).map((groups) => Math.min(5, groups.length))
    ) *
      288 -
    32
  for (let index = 5; index < zones.length; index++) {
    const entries = grouped[index]
    if (!entries.length) continue
    const top = bottom ? bottom + 40 : 0
    section(index, 0, top, width, true)
    let y = top + 64
    for (let offset = 0; offset < entries.length; offset += 5) {
      let rowBottom = y
      entries.slice(offset, offset + 5).forEach((group, column) => {
        rowBottom = Math.max(
          rowBottom,
          place(group, column * 288, y, index === 6)
        )
      })
      y = rowBottom + 40
    }
    bottom = y
  }
  return { sections, groups }
}
