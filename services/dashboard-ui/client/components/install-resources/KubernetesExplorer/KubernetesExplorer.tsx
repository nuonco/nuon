import { useEffect, useMemo, useState, type ReactNode } from 'react'
import { stringify } from 'yaml'
import { Badge, type TBadgeTheme } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Dropdown } from '@/components/common/Dropdown'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon, type TIconVariant } from '@/components/common/Icon'
import { Input } from '@/components/common/form/Input'
import { Menu } from '@/components/common/Menu'
import { RadioInput } from '@/components/common/form/RadioInput'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { THelmRelease, TInstallResource } from '@/types'
import { cn } from '@/utils/classnames'

export type TK8sEntry = {
  resource: TInstallResource
  kind: string
  apiGroup: string
  name: string
  namespace: string
  obj: any
}

const parseJSON = (value?: string) => {
  if (!value) return undefined
  try {
    return JSON.parse(value)
  } catch {
    return undefined
  }
}

export type THelmEntry = {
  name: string
  namespace: string
  status: string
  revision: number
  latest: THelmRelease
  history: THelmRelease[]
}

export const toHelmEntries = (releases: THelmRelease[]): THelmEntry[] => {
  const groups = new Map<string, THelmRelease[]>()
  for (const release of releases) {
    const key = `${release?.namespace ?? ''}/${release?.name ?? ''}`
    groups.set(key, [...(groups.get(key) ?? []), release])
  }
  return [...groups.values()]
    .map((history) => {
      const sorted = [...history].sort(
        (a, b) => (b?.version ?? 0) - (a?.version ?? 0)
      )
      const latest = sorted[0]
      return {
        name: latest?.name ?? '',
        namespace: latest?.namespace ?? '',
        status: latest?.status ?? 'unknown',
        revision: latest?.version ?? 0,
        latest,
        history: sorted,
      }
    })
    .sort(
      (a, b) =>
        a.namespace.localeCompare(b.namespace) || a.name.localeCompare(b.name)
    )
}

export const toEntries = (resources: TInstallResource[]): TK8sEntry[] =>
  resources.map((resource) => ({
    resource,
    kind: resource?.kind ?? 'Unknown',
    apiGroup: resource?.api_group ?? '',
    name: resource?.name ?? '',
    namespace: resource?.namespace ?? '',
    obj: parseJSON(resource?.details),
  }))

const KIND_SECTIONS: { heading: string; kinds: string[] }[] = [
  {
    heading: 'Workloads',
    kinds: ['Deployment', 'StatefulSet', 'DaemonSet', 'ReplicaSet', 'Job', 'Pod'],
  },
  { heading: 'Network', kinds: ['Service', 'Ingress'] },
  { heading: 'Storage', kinds: ['PersistentVolumeClaim'] },
]

const BUILTIN_KINDS = new Set(KIND_SECTIONS.flatMap((s) => s.kinds))

const KIND_PLURALS: Record<string, string> = {
  Deployment: 'Deployments',
  StatefulSet: 'Stateful sets',
  DaemonSet: 'Daemon sets',
  ReplicaSet: 'Replica sets',
  Job: 'Jobs',
  Pod: 'Pods',
  Service: 'Services',
  Ingress: 'Ingresses',
  PersistentVolumeClaim: 'Persistent volume claims',
  'helm:releases': 'Releases',
}

const pluralize = (kind: string) => KIND_PLURALS[kind] ?? `${kind}s`

const kindKey = (apiGroup: string, kind: string) => `${apiGroup}/${kind}`

const entryKey = (entry: TK8sEntry) =>
  `${kindKey(entry.apiGroup, entry.kind)}/${entry.namespace}/${entry.name}`

const helmEntryKey = (entry: THelmEntry) => `${entry.namespace}/${entry.name}`

const SECTION_ICONS: Record<string, TIconVariant> = {
  Workloads: 'CubeIcon',
  Network: 'GlobeIcon',
  Storage: 'DatabaseIcon',
  Helm: 'Helm',
  'Custom resources': 'PuzzlePieceIcon',
}

const STATUS_THEMES: Record<string, TBadgeTheme> = {
  Running: 'success',
  Available: 'success',
  Ready: 'success',
  Bound: 'success',
  Complete: 'success',
  Completed: 'success',
  Succeeded: 'success',
  Active: 'success',
  Healthy: 'success',
  Progressing: 'info',
  Pending: 'info',
  ContainerCreating: 'info',
  Updating: 'info',
  Running_Job: 'info',
  Terminating: 'warn',
  NotReady: 'warn',
  Suspended: 'warn',
  Degraded: 'error',
  Failed: 'error',
  Error: 'error',
  CrashLoopBackOff: 'error',
  ImagePullBackOff: 'error',
  ErrImagePull: 'error',
  OOMKilled: 'error',
  Evicted: 'error',
  Lost: 'error',
  Unknown: 'neutral',
}

const statusTheme = (label: string): TBadgeTheme =>
  STATUS_THEMES[label] ?? 'neutral'

const HELM_STATUS_THEMES: Record<string, TBadgeTheme> = {
  deployed: 'success',
  superseded: 'neutral',
  failed: 'error',
  'pending-install': 'info',
  'pending-upgrade': 'info',
  'pending-rollback': 'info',
  uninstalling: 'warn',
  uninstalled: 'neutral',
  unknown: 'neutral',
}

const helmStatusTheme = (status: string): TBadgeTheme =>
  HELM_STATUS_THEMES[status] ?? 'neutral'

type TDerivedStatus = { label: string; detail?: string }

const podStatus = (obj: any): TDerivedStatus => {
  const phase = obj?.status?.phase ?? 'Unknown'
  const containers: any[] = obj?.status?.containerStatuses ?? []
  for (const c of containers) {
    const waiting = c?.state?.waiting?.reason
    if (waiting) return { label: waiting }
    const terminated = c?.state?.terminated?.reason
    if (terminated && terminated !== 'Completed') return { label: terminated }
  }
  if (obj?.metadata?.deletionTimestamp) return { label: 'Terminating' }
  return { label: phase }
}

const podReady = (obj: any) => {
  const containers: any[] = obj?.status?.containerStatuses ?? []
  const ready = containers.filter((c) => c?.ready).length
  return `${ready}/${containers.length || obj?.spec?.containers?.length || 0}`
}

const podRestarts = (obj: any) =>
  (obj?.status?.containerStatuses ?? []).reduce(
    (sum: number, c: any) => sum + (c?.restartCount ?? 0),
    0
  )

const findCondition = (obj: any, type: string) =>
  (obj?.status?.conditions ?? []).find((c: any) => c?.type === type)

const deploymentStatus = (obj: any): TDerivedStatus => {
  const available = findCondition(obj, 'Available')
  const progressing = findCondition(obj, 'Progressing')
  if (progressing?.status === 'False')
    return { label: 'Failed', detail: progressing?.message }
  if (available?.status === 'True') return { label: 'Available' }
  return { label: 'Progressing', detail: progressing?.message }
}

const replicaRatio = (ready?: number, desired?: number) =>
  `${ready ?? 0}/${desired ?? 0}`

const jobStatus = (obj: any): TDerivedStatus => {
  if (findCondition(obj, 'Complete')?.status === 'True')
    return { label: 'Complete' }
  const failed = findCondition(obj, 'Failed')
  if (failed?.status === 'True') return { label: 'Failed', detail: failed?.message }
  if (obj?.status?.active) return { label: 'Running' }
  return { label: 'Pending' }
}

const customStatus = (entry: TK8sEntry): TDerivedStatus => {
  const ready = findCondition(entry.obj, 'Ready')
  if (ready) {
    return ready.status === 'True'
      ? { label: 'Ready', detail: ready?.message }
      : { label: 'NotReady', detail: ready?.message }
  }
  const native = parseJSON(entry.resource?.native_status)
  if (native?.status) return { label: native.status, detail: native?.message }
  const phase = entry.obj?.status?.phase
  if (phase) return { label: phase }
  return { label: 'Unknown' }
}

export const deriveStatus = (entry: TK8sEntry): TDerivedStatus => {
  const { kind, obj } = entry
  switch (kind) {
    case 'Pod':
      return podStatus(obj)
    case 'Deployment':
      return deploymentStatus(obj)
    case 'StatefulSet':
      return (obj?.status?.readyReplicas ?? 0) >= (obj?.spec?.replicas ?? 0)
        ? { label: 'Available' }
        : { label: 'Progressing' }
    case 'DaemonSet':
      return (obj?.status?.numberReady ?? 0) >=
        (obj?.status?.desiredNumberScheduled ?? 0)
        ? { label: 'Available' }
        : { label: 'Progressing' }
    case 'ReplicaSet':
      return (obj?.status?.readyReplicas ?? 0) >= (obj?.spec?.replicas ?? 0)
        ? { label: 'Available' }
        : { label: 'Progressing' }
    case 'Job':
      return jobStatus(obj)
    case 'Service':
      return { label: 'Active' }
    case 'Ingress':
      return (obj?.status?.loadBalancer?.ingress ?? []).length > 0
        ? { label: 'Active' }
        : { label: 'Pending' }
    case 'PersistentVolumeClaim':
      return { label: obj?.status?.phase ?? 'Unknown' }
    default:
      return customStatus(entry)
  }
}

const StatusBadge = ({ entry }: { entry: TK8sEntry }) => {
  const { label } = deriveStatus(entry)
  return (
    <Badge size="sm" theme={statusTheme(label)}>
      {label}
    </Badge>
  )
}

const Age = ({ entry }: { entry: TK8sEntry }) => {
  const time =
    entry.obj?.metadata?.creationTimestamp ?? entry.resource?.observed_at
  return time ? (
    <Time variant="subtext" time={time} format="relative" />
  ) : (
    <Text variant="subtext">—</Text>
  )
}

const servicePorts = (obj: any) =>
  (obj?.spec?.ports ?? [])
    .map((p: any) =>
      p?.nodePort
        ? `${p?.port}:${p?.nodePort}/${p?.protocol ?? 'TCP'}`
        : `${p?.port}/${p?.protocol ?? 'TCP'}`
    )
    .join(', ')

const ingressHosts = (obj: any) =>
  (obj?.spec?.rules ?? []).map((r: any) => r?.host).filter(Boolean).join(', ')

const ingressAddress = (obj: any) =>
  (obj?.status?.loadBalancer?.ingress ?? [])
    .map((i: any) => i?.hostname ?? i?.ip)
    .filter(Boolean)
    .join(', ')

type TColumn = {
  header: string
  cell: (entry: TK8sEntry) => ReactNode
}

const nameColumn: TColumn = {
  header: 'Name',
  cell: (e) => (
    <Text variant="body" family="mono" className="truncate">
      {e.name}
    </Text>
  ),
}

const namespaceColumn: TColumn = {
  header: 'Namespace',
  cell: (e) => <Text variant="subtext">{e.namespace || '—'}</Text>,
}

const ageColumn: TColumn = { header: 'Age', cell: (e) => <Age entry={e} /> }

const statusColumn: TColumn = {
  header: 'Status',
  cell: (e) => <StatusBadge entry={e} />,
}

const KIND_COLUMNS: Record<string, TColumn[]> = {
  Deployment: [
    nameColumn,
    namespaceColumn,
    {
      header: 'Ready',
      cell: (e) => (
        <Text variant="subtext" family="mono">
          {replicaRatio(e.obj?.status?.readyReplicas, e.obj?.spec?.replicas)}
        </Text>
      ),
    },
    {
      header: 'Up-to-date',
      cell: (e) => (
        <Text variant="subtext" family="mono">
          {e.obj?.status?.updatedReplicas ?? 0}
        </Text>
      ),
    },
    {
      header: 'Available',
      cell: (e) => (
        <Text variant="subtext" family="mono">
          {e.obj?.status?.availableReplicas ?? 0}
        </Text>
      ),
    },
    ageColumn,
    statusColumn,
  ],
  StatefulSet: [
    nameColumn,
    namespaceColumn,
    {
      header: 'Ready',
      cell: (e) => (
        <Text variant="subtext" family="mono">
          {replicaRatio(e.obj?.status?.readyReplicas, e.obj?.spec?.replicas)}
        </Text>
      ),
    },
    ageColumn,
    statusColumn,
  ],
  DaemonSet: [
    nameColumn,
    namespaceColumn,
    {
      header: 'Desired',
      cell: (e) => (
        <Text variant="subtext" family="mono">
          {e.obj?.status?.desiredNumberScheduled ?? 0}
        </Text>
      ),
    },
    {
      header: 'Ready',
      cell: (e) => (
        <Text variant="subtext" family="mono">
          {e.obj?.status?.numberReady ?? 0}
        </Text>
      ),
    },
    ageColumn,
    statusColumn,
  ],
  ReplicaSet: [
    nameColumn,
    namespaceColumn,
    {
      header: 'Desired',
      cell: (e) => (
        <Text variant="subtext" family="mono">
          {e.obj?.spec?.replicas ?? 0}
        </Text>
      ),
    },
    {
      header: 'Ready',
      cell: (e) => (
        <Text variant="subtext" family="mono">
          {e.obj?.status?.readyReplicas ?? 0}
        </Text>
      ),
    },
    ageColumn,
  ],
  Job: [
    nameColumn,
    namespaceColumn,
    {
      header: 'Completions',
      cell: (e) => (
        <Text variant="subtext" family="mono">
          {`${e.obj?.status?.succeeded ?? 0}/${e.obj?.spec?.completions ?? 1}`}
        </Text>
      ),
    },
    ageColumn,
    statusColumn,
  ],
  Pod: [
    nameColumn,
    namespaceColumn,
    {
      header: 'Ready',
      cell: (e) => (
        <Text variant="subtext" family="mono">
          {podReady(e.obj)}
        </Text>
      ),
    },
    {
      header: 'Restarts',
      cell: (e) => (
        <Text
          variant="subtext"
          family="mono"
          theme={podRestarts(e.obj) > 0 ? 'warn' : undefined}
        >
          {podRestarts(e.obj)}
        </Text>
      ),
    },
    {
      header: 'Node',
      cell: (e) => (
        <Text variant="subtext" family="mono" className="truncate">
          {e.obj?.spec?.nodeName ?? '—'}
        </Text>
      ),
    },
    ageColumn,
    statusColumn,
  ],
  Service: [
    nameColumn,
    namespaceColumn,
    {
      header: 'Type',
      cell: (e) => <Text variant="subtext">{e.obj?.spec?.type ?? '—'}</Text>,
    },
    {
      header: 'Cluster IP',
      cell: (e) => (
        <Text variant="subtext" family="mono">
          {e.obj?.spec?.clusterIP ?? '—'}
        </Text>
      ),
    },
    {
      header: 'Ports',
      cell: (e) => (
        <Text variant="subtext" family="mono">
          {servicePorts(e.obj) || '—'}
        </Text>
      ),
    },
    ageColumn,
  ],
  Ingress: [
    nameColumn,
    namespaceColumn,
    {
      header: 'Class',
      cell: (e) => (
        <Text variant="subtext">{e.obj?.spec?.ingressClassName ?? '—'}</Text>
      ),
    },
    {
      header: 'Hosts',
      cell: (e) => (
        <Text variant="subtext" family="mono">
          {ingressHosts(e.obj) || '—'}
        </Text>
      ),
    },
    {
      header: 'Address',
      cell: (e) => (
        <Text variant="subtext" family="mono" className="truncate">
          {ingressAddress(e.obj) || '—'}
        </Text>
      ),
    },
    ageColumn,
  ],
  PersistentVolumeClaim: [
    nameColumn,
    namespaceColumn,
    statusColumn,
    {
      header: 'Capacity',
      cell: (e) => (
        <Text variant="subtext" family="mono">
          {e.obj?.status?.capacity?.storage ??
            e.obj?.spec?.resources?.requests?.storage ??
            '—'}
        </Text>
      ),
    },
    {
      header: 'Storage class',
      cell: (e) => (
        <Text variant="subtext">{e.obj?.spec?.storageClassName ?? '—'}</Text>
      ),
    },
    ageColumn,
  ],
}

const DEFAULT_COLUMNS: TColumn[] = [
  nameColumn,
  namespaceColumn,
  statusColumn,
  ageColumn,
]

const KindNavButton = ({
  kind,
  apiGroup,
  count,
  isActive,
  onClick,
}: {
  kind: string
  apiGroup?: string
  count: number
  isActive: boolean
  onClick: () => void
}) => (
  <button
    className={cn(
      'flex items-center justify-between gap-2 w-full text-left rounded-md px-2.5 py-1.5 transition-colors',
      isActive
        ? 'bg-cool-grey-100 dark:bg-dark-grey-700'
        : 'hover:bg-cool-grey-50 dark:hover:bg-dark-grey-800'
    )}
    onClick={onClick}
    type="button"
  >
    <span className="flex flex-col min-w-0">
      <Text
        variant="body"
        weight={isActive ? 'strong' : 'normal'}
        className="truncate"
      >
        {pluralize(kind)}
      </Text>
      {apiGroup ? (
        <Text variant="label" family="mono" className="truncate opacity-60">
          {apiGroup}
        </Text>
      ) : null}
    </span>
    <Badge size="sm" theme={isActive ? 'default' : 'neutral'}>
      {count}
    </Badge>
  </button>
)

const OverviewRow = ({ label, children }: { label: string; children: ReactNode }) => (
  <div className="flex flex-col gap-0.5">
    <Text variant="label" className="uppercase opacity-60">
      {label}
    </Text>
    <div className="flex flex-wrap items-center gap-1.5 min-w-0">{children}</div>
  </div>
)

const ConditionsList = ({ entry }: { entry: TK8sEntry }) => {
  const conditions: any[] = entry.obj?.status?.conditions ?? []
  if (!conditions.length) return null
  return (
    <OverviewRow label="Conditions">
      <div className="flex flex-col gap-2 w-full">
        {conditions.map((c: any) => (
          <div
            key={c?.type}
            className="flex flex-col gap-0.5 rounded-md border border-cool-grey-200 dark:border-dark-grey-600 p-2"
          >
            <div className="flex items-center justify-between gap-2">
              <Text variant="subtext" weight="strong" family="mono">
                {c?.type}
              </Text>
              <Badge size="sm" theme={c?.status === 'True' ? 'success' : 'warn'}>
                {c?.status}
              </Badge>
            </div>
            {c?.reason ? (
              <Text variant="subtext" family="mono" className="opacity-70">
                {c?.reason}
              </Text>
            ) : null}
            {c?.message ? (
              <Text variant="subtext" className="opacity-70">
                {c?.message}
              </Text>
            ) : null}
          </div>
        ))}
      </div>
    </OverviewRow>
  )
}

const containers = (obj: any): any[] => {
  const containers =
    obj?.spec?.template?.spec?.containers ?? obj?.spec?.containers ?? []
  return containers.filter((c: any) => c?.image)
}

const labelEntries = (obj: any): [string, string][] =>
  Object.entries(obj?.metadata?.labels ?? {})

const ownerRefs = (obj: any): any[] => obj?.metadata?.ownerReferences ?? []

const controllerRef = (obj: any) =>
  ownerRefs(obj).find((r: any) => r?.controller === true)

const refMatchesEntry = (ref: any, entry: TK8sEntry) => {
  const refUID = ref?.uid
  const entryUID = entry.obj?.metadata?.uid
  return refUID && entryUID
    ? refUID === entryUID
    : ref?.kind === entry.kind && ref?.name === entry.name
}

const resolveRef = (
  ref: any,
  namespace: string,
  entries: TK8sEntry[]
): TK8sEntry | undefined =>
  entries.find(
    (e) => e.namespace === namespace && refMatchesEntry(ref, e)
  )

export const childrenOf = (
  entry: TK8sEntry,
  entries: TK8sEntry[],
  kind: string
): TK8sEntry[] => {
  const byRef = entries.filter(
    (e) =>
      e.kind === kind &&
      e.namespace === entry.namespace &&
      ownerRefs(e.obj).some((r: any) => refMatchesEntry(r, entry))
  )
  if (byRef.length) return byRef
  return entries.filter(
    (e) =>
      e.kind === kind &&
      e.namespace === entry.namespace &&
      e.name.startsWith(`${entry.name}-`)
  )
}

const descendantPods = (
  entry: TK8sEntry,
  entries: TK8sEntry[]
): TK8sEntry[] => {
  if (entry.kind === 'Deployment') {
    const viaReplicaSets = childrenOf(entry, entries, 'ReplicaSet').flatMap(
      (rs) => childrenOf(rs, entries, 'Pod')
    )
    if (viaReplicaSets.length) return viaReplicaSets
  }
  return childrenOf(entry, entries, 'Pod')
}

const ControlledBy = ({
  entry,
  entries,
  onSelect,
}: {
  entry: TK8sEntry
  entries: TK8sEntry[]
  onSelect: (e: TK8sEntry) => void
}) => {
  const ref = controllerRef(entry.obj)
  if (!ref) return null
  const owner = resolveRef(ref, entry.namespace, entries)
  return (
    <OverviewRow label="Controlled by">
      {owner ? (
        <button
          className="flex items-center gap-2 rounded-md px-2 py-1.5 -mx-2 hover:bg-cool-grey-50 dark:hover:bg-dark-grey-800 transition-colors text-left"
          onClick={() => onSelect(owner)}
          type="button"
        >
          <Badge size="sm" variant="code" theme="brand">
            {ref?.kind}
          </Badge>
          <Text variant="subtext" family="mono" className="truncate">
            {ref?.name}
          </Text>
          <Icon variant="CaretRightIcon" size={12} />
        </button>
      ) : (
        <div className="flex items-center gap-2">
          <Badge size="sm" variant="code" theme="neutral">
            {ref?.kind}
          </Badge>
          <Text variant="subtext" family="mono" className="truncate">
            {ref?.name}
          </Text>
          <Text variant="label" className="opacity-60" nowrap>
            not captured
          </Text>
        </div>
      )}
    </OverviewRow>
  )
}

const OwnedResources = ({
  label,
  items,
  onSelect,
  detail,
}: {
  label: string
  items: TK8sEntry[]
  onSelect: (e: TK8sEntry) => void
  detail?: (e: TK8sEntry) => ReactNode
}) => {
  if (!items.length) return null
  return (
    <OverviewRow label={label}>
      <div className="flex flex-col gap-1 w-full">
        {items.map((item) => {
          const { label: statusLabel } = deriveStatus(item)
          return (
            <button
              key={`${item.kind}/${item.name}`}
              className="flex items-center justify-between gap-2 rounded-md px-2 py-1.5 hover:bg-cool-grey-50 dark:hover:bg-dark-grey-800 transition-colors text-left"
              onClick={() => onSelect(item)}
              type="button"
            >
              <Text variant="subtext" family="mono" className="truncate">
                {item.name}
              </Text>
              <span className="flex items-center gap-2 shrink-0">
                {detail ? (
                  <Text variant="subtext" family="mono">
                    {detail(item)}
                  </Text>
                ) : null}
                <Badge size="sm" theme={statusTheme(statusLabel)}>
                  {statusLabel}
                </Badge>
              </span>
            </button>
          )
        })}
      </div>
    </OverviewRow>
  )
}

const DetailOverview = ({
  entry,
  entries,
  onSelect,
}: {
  entry: TK8sEntry
  entries: TK8sEntry[]
  onSelect: (e: TK8sEntry) => void
}) => {
  const { label, detail } = deriveStatus(entry)
  const containerList = containers(entry.obj)
  const labels = labelEntries(entry.obj)
  const isPodOwner = [
    'Deployment',
    'ReplicaSet',
    'StatefulSet',
    'DaemonSet',
    'Job',
  ].includes(entry.kind)

  return (
    <div className="flex flex-col gap-4">
      <OverviewRow label="Status">
        <Badge size="sm" theme={statusTheme(label)}>
          {label}
        </Badge>
        {detail ? (
          <Text variant="subtext" className="opacity-70 w-full">
            {detail}
          </Text>
        ) : null}
      </OverviewRow>

      <ControlledBy entry={entry} entries={entries} onSelect={onSelect} />

      {entry.kind === 'Pod' ? (
        <>
          <OverviewRow label="Ready">
            <Text variant="subtext" family="mono">
              {podReady(entry.obj)}
            </Text>
          </OverviewRow>
          <OverviewRow label="Restarts">
            <Text variant="subtext" family="mono">
              {podRestarts(entry.obj)}
            </Text>
          </OverviewRow>
          {entry.obj?.spec?.nodeName ? (
            <OverviewRow label="Node">
              <Text variant="subtext" family="mono">
                {entry.obj?.spec?.nodeName}
              </Text>
            </OverviewRow>
          ) : null}
          {entry.obj?.status?.podIP ? (
            <OverviewRow label="Pod IP">
              <Text variant="subtext" family="mono">
                {entry.obj?.status?.podIP}
              </Text>
            </OverviewRow>
          ) : null}
        </>
      ) : null}

      {entry.kind === 'Deployment' ? (
        <OverviewRow label="Replicas">
          <Text variant="subtext" family="mono">
            {`${entry.obj?.status?.readyReplicas ?? 0} ready / ${
              entry.obj?.status?.updatedReplicas ?? 0
            } up-to-date / ${entry.obj?.spec?.replicas ?? 0} desired`}
          </Text>
        </OverviewRow>
      ) : null}

      {containerList.length ? (
        <OverviewRow label={containerList.length > 1 ? 'Images' : 'Image'}>
          <div className="flex flex-col gap-1 w-full">
            {containerList.map((container, index) => (
              <Badge
                key={container?.name ?? index}
                size="sm"
                variant="code"
                theme="default"
              >
                <span className="truncate">{container?.image}</span>
              </Badge>
            ))}
          </div>
        </OverviewRow>
      ) : null}

      {labels.length ? (
        <OverviewRow label="Labels">
          {labels.map(([k, v]) => (
            <Badge key={k} size="sm" variant="code" theme="neutral">
              <span className="truncate">{`${k}=${v}`}</span>
            </Badge>
          ))}
        </OverviewRow>
      ) : null}

      <ConditionsList entry={entry} />

      {entry.kind === 'Deployment' ? (
        <OwnedResources
          label="Replica sets"
          items={childrenOf(entry, entries, 'ReplicaSet')}
          onSelect={onSelect}
          detail={(e) =>
            replicaRatio(e.obj?.status?.readyReplicas, e.obj?.spec?.replicas)
          }
        />
      ) : null}

      {isPodOwner ? (
        <OwnedResources
          label="Pods"
          items={descendantPods(entry, entries)}
          onSelect={onSelect}
        />
      ) : null}

      {entry.resource?.observed_at ? (
        <OverviewRow label="Last observed">
          <Time
            variant="subtext"
            time={entry.resource?.observed_at}
            format="relative"
          />
        </OverviewRow>
      ) : null}
    </div>
  )
}

const DetailPane = ({
  entry,
  entries,
  onSelect,
  onClose,
}: {
  entry: TK8sEntry
  entries: TK8sEntry[]
  onSelect: (e: TK8sEntry) => void
  onClose: () => void
}) => {
  const yaml = useMemo(() => {
    if (!entry.obj) return '# no object data captured'
    const doc = {
      apiVersion: entry.obj?.apiVersion,
      kind: entry.obj?.kind ?? entry.kind,
      ...entry.obj,
    }
    return stringify(doc, { indent: 2 })
  }, [entry])

  return (
    <div className="flex flex-col h-full min-h-0">
      <div className="flex items-start justify-between gap-2 p-4 border-b border-cool-grey-200 dark:border-dark-grey-600">
        <div className="flex flex-col gap-1 min-w-0">
          <div className="flex items-center gap-2">
            <Badge size="sm" variant="code" theme="brand">
              {entry.kind}
            </Badge>
            {entry.apiGroup ? (
              <Text variant="label" family="mono" className="opacity-60">
                {entry.apiGroup}
              </Text>
            ) : null}
          </div>
          <Text variant="base" weight="strong" family="mono" className="break-all">
            {entry.name}
          </Text>
          {entry.namespace ? (
            <Text variant="subtext" className="opacity-70">
              namespace: {entry.namespace}
            </Text>
          ) : null}
        </div>
        <Button variant="ghost" size="xs" onClick={onClose} aria-label="Close details">
          <Icon variant="XIcon" size={16} />
        </Button>
      </div>
      <div className="flex-1 min-h-0 overflow-y-auto p-4">
        <Tabs
          key={entryKey(entry)}
          tabs={{
            overview: (
              <DetailOverview entry={entry} entries={entries} onSelect={onSelect} />
            ),
            yaml: (
              <CodeBlock language="yaml" showCopy>
                {yaml}
              </CodeBlock>
            ),
          }}
        />
      </div>
    </div>
  )
}

const helmLabelEntries = (release?: THelmRelease): [string, string][] =>
  Object.entries(release?.labels ?? {}).map(([k, v]) => [k, String(v)])

const HelmDetailPane = ({
  entry,
  onClose,
}: {
  entry: THelmEntry
  onClose: () => void
}) => {
  const record = useMemo(
    () =>
      stringify(
        {
          name: entry.latest?.name,
          namespace: entry.latest?.namespace,
          status: entry.latest?.status,
          revision: entry.latest?.version,
          type: entry.latest?.type,
          key: entry.latest?.key,
          owner: entry.latest?.owner,
          labels: entry.latest?.labels,
          updated_at: entry.latest?.updated_at,
        },
        { indent: 2 }
      ),
    [entry]
  )
  const labels = helmLabelEntries(entry.latest)

  return (
    <div className="flex flex-col h-full min-h-0">
      <div className="flex items-start justify-between gap-2 p-4 border-b border-cool-grey-200 dark:border-dark-grey-600">
        <div className="flex flex-col gap-1 min-w-0">
          <div className="flex items-center gap-2">
            <Badge size="sm" variant="code" theme="brand">
              <Icon variant="Helm" size={12} />
              Release
            </Badge>
          </div>
          <Text variant="base" weight="strong" family="mono" className="break-all">
            {entry.name}
          </Text>
          {entry.namespace ? (
            <Text variant="subtext" className="opacity-70">
              namespace: {entry.namespace}
            </Text>
          ) : null}
        </div>
        <Button variant="ghost" size="xs" onClick={onClose} aria-label="Close details">
          <Icon variant="XIcon" size={16} />
        </Button>
      </div>
      <div className="flex-1 min-h-0 overflow-y-auto p-4">
        <Tabs
          key={helmEntryKey(entry)}
          tabs={{
            overview: (
              <div className="flex flex-col gap-4">
                <OverviewRow label="Status">
                  <Badge size="sm" theme={helmStatusTheme(entry.status)}>
                    {entry.status}
                  </Badge>
                </OverviewRow>
                <OverviewRow label="Revision">
                  <Text variant="subtext" family="mono">
                    {entry.revision}
                  </Text>
                </OverviewRow>
                {entry.latest?.type ? (
                  <OverviewRow label="Storage">
                    <Text variant="subtext" family="mono">
                      {entry.latest?.type}
                    </Text>
                  </OverviewRow>
                ) : null}
                {entry.latest?.key ? (
                  <OverviewRow label="Storage key">
                    <Text variant="subtext" family="mono" className="break-all">
                      {entry.latest?.key}
                    </Text>
                  </OverviewRow>
                ) : null}
                {entry.latest?.owner ? (
                  <OverviewRow label="Owner">
                    <Text variant="subtext" family="mono">
                      {entry.latest?.owner}
                    </Text>
                  </OverviewRow>
                ) : null}
                {labels.length ? (
                  <OverviewRow label="Labels">
                    {labels.map(([k, v]) => (
                      <Badge key={k} size="sm" variant="code" theme="neutral">
                        <span className="truncate">{`${k}=${v}`}</span>
                      </Badge>
                    ))}
                  </OverviewRow>
                ) : null}
                {entry.latest?.updated_at ? (
                  <OverviewRow label="Last updated">
                    <Time
                      variant="subtext"
                      time={entry.latest?.updated_at}
                      format="relative"
                    />
                  </OverviewRow>
                ) : null}
              </div>
            ),
            history: (
              <div className="flex flex-col gap-1">
                {entry.history.map((revision, index) => (
                  <div
                    key={
                      revision?.key ??
                      revision?.version ??
                      revision?.updated_at ??
                      index
                    }
                    className="flex items-center justify-between gap-2 rounded-md border border-cool-grey-200 dark:border-dark-grey-600 px-2.5 py-2"
                  >
                    <div className="flex items-center gap-2 min-w-0">
                      <Text variant="subtext" weight="strong" family="mono" nowrap>
                        rev {revision?.version ?? 0}
                      </Text>
                      {revision?.updated_at ? (
                        <Time
                          variant="subtext"
                          time={revision?.updated_at}
                          format="relative"
                        />
                      ) : null}
                    </div>
                    <Badge
                      size="sm"
                      theme={helmStatusTheme(revision?.status ?? 'unknown')}
                    >
                      {revision?.status ?? 'unknown'}
                    </Badge>
                  </div>
                ))}
              </div>
            ),
            record: (
              <CodeBlock language="yaml" showCopy>
                {record}
              </CodeBlock>
            ),
          }}
        />
      </div>
    </div>
  )
}

const NamespaceFilter = ({
  namespaces,
  value,
  onChange,
}: {
  namespaces: string[]
  value: string
  onChange: (value: string) => void
}) => (
  <Dropdown
    alignment="left"
    id="k8s-explorer-namespace-filter"
    variant="secondary"
    size="sm"
    buttonText={
      <>
        <Icon variant="StackIcon" size={14} />
        {value || 'All namespaces'}
      </>
    }
  >
    <Menu className="!p-0 !w-64">
      <form onReset={() => onChange('')}>
        <div className="flex flex-col gap-0.5 max-h-64 overflow-y-auto w-full p-2">
          <RadioInput
            checked={value === ''}
            labelProps={{ labelText: 'All namespaces' }}
            name="k8s-explorer-namespace"
            onChange={() => onChange('')}
            value=""
          />
          {namespaces.map((ns) => (
            <RadioInput
              key={ns}
              checked={value === ns}
              labelProps={{ labelText: ns }}
              name="k8s-explorer-namespace"
              onChange={() => onChange(ns)}
              value={ns}
            />
          ))}
        </div>
      </form>
    </Menu>
  </Dropdown>
)

const HELM_NAV_KIND = 'helm:releases'

export interface IKubernetesExplorer {
  resources: TInstallResource[]
  helmReleases?: THelmRelease[]
}

export const KubernetesExplorer = ({
  resources,
  helmReleases = [],
}: IKubernetesExplorer) => {
  const entries = useMemo(() => toEntries(resources), [resources])
  const helmEntries = useMemo(() => toHelmEntries(helmReleases), [helmReleases])

  const counts = useMemo(() => {
    const map = new Map<string, number>()
    for (const e of entries) {
      const key = kindKey(e.apiGroup, e.kind)
      map.set(key, (map.get(key) ?? 0) + 1)
    }
    return map
  }, [entries])

  const customKinds = useMemo(() => {
    const map = new Map<string, [string, string]>()
    for (const e of entries) {
      if (!BUILTIN_KINDS.has(e.kind)) {
        map.set(kindKey(e.apiGroup, e.kind), [e.kind, e.apiGroup])
      }
    }
    return [...map.entries()].sort(([, [aKind, aGroup]], [, [bKind, bGroup]]) =>
      aKind.localeCompare(bKind) || aGroup.localeCompare(bGroup)
    )
  }, [entries])

  const namespaces = useMemo(
    () =>
      [
        ...new Set([
          ...entries.map((e) => e.namespace),
          ...helmEntries.map((e) => e.namespace),
        ]),
      ]
        .filter(Boolean)
        .sort(),
    [entries, helmEntries]
  )

  const firstBuiltinEntry = KIND_SECTIONS.flatMap((s) => s.kinds)
    .map((kind) => entries.find((entry) => entry.kind === kind))
    .find((entry) => entry !== undefined)
  const firstKind = firstBuiltinEntry
    ? kindKey(firstBuiltinEntry.apiGroup, firstBuiltinEntry.kind)
    : customKinds.at(0)?.[0] ??
      (helmEntries.length ? HELM_NAV_KIND : kindKey('', 'Deployment'))

  const [selectedKind, setSelectedKind] = useState(firstKind)
  const [namespace, setNamespace] = useState('')
  const [search, setSearch] = useState('')
  const [selectedKey, setSelectedKey] = useState<string | undefined>(undefined)
  const [selectedHelmKey, setSelectedHelmKey] = useState<string | undefined>(
    undefined
  )

  const selected = entries.find((entry) => entryKey(entry) === selectedKey)
  const selectedHelm = helmEntries.find(
    (entry) => helmEntryKey(entry) === selectedHelmKey
  )
  const selectedKindName =
    selectedKind === HELM_NAV_KIND
      ? HELM_NAV_KIND
      : entries.find(
          (entry) => kindKey(entry.apiGroup, entry.kind) === selectedKind
        )?.kind ?? selectedKind.split('/').at(-1) ?? selectedKind

  useEffect(() => {
    const availableKinds = new Set([
      ...entries.map((entry) => kindKey(entry.apiGroup, entry.kind)),
      ...(helmEntries.length ? [HELM_NAV_KIND] : []),
    ])
    if (!availableKinds.has(selectedKind)) {
      setSelectedKind(firstKind)
      setSelectedKey(undefined)
      setSelectedHelmKey(undefined)
    }
  }, [entries, firstKind, helmEntries.length, selectedKind])

  useEffect(() => {
    if (namespace && !namespaces.includes(namespace)) setNamespace('')
  }, [namespace, namespaces])

  const rows = useMemo(
    () =>
      entries
        .filter((e) => kindKey(e.apiGroup, e.kind) === selectedKind)
        .filter((e) => !namespace || e.namespace === namespace)
        .filter(
          (e) =>
            !search || e.name.toLowerCase().includes(search.toLowerCase())
        )
        .sort(
          (a, b) =>
            a.namespace.localeCompare(b.namespace) ||
            a.name.localeCompare(b.name)
        ),
    [entries, selectedKind, namespace, search]
  )

  const helmRows = useMemo(
    () =>
      helmEntries
        .filter((e) => !namespace || e.namespace === namespace)
        .filter(
          (e) =>
            !search || e.name.toLowerCase().includes(search.toLowerCase())
        ),
    [helmEntries, namespace, search]
  )

  const columns = KIND_COLUMNS[selectedKindName] ?? DEFAULT_COLUMNS
  const isHelmView = selectedKind === HELM_NAV_KIND

  const selectKind = (kind: string) => {
    setSelectedKind(kind)
    setSelectedKey(undefined)
    setSelectedHelmKey(undefined)
  }

  const selectEntry = (entry: TK8sEntry) => {
    setSelectedKind(kindKey(entry.apiGroup, entry.kind))
    setSelectedKey(entryKey(entry))
    setSelectedHelmKey(undefined)
  }

  const isSelected = (e: TK8sEntry) => selectedKey === entryKey(e)

  const isHelmSelected = (e: THelmEntry) =>
    selectedHelmKey === helmEntryKey(e)

  return (
    <div className="flex w-full h-full min-h-0 border border-cool-grey-200 dark:border-dark-grey-600 rounded-lg overflow-hidden bg-white dark:bg-dark-grey-900">
      <nav className="w-60 shrink-0 border-r border-cool-grey-200 dark:border-dark-grey-600 overflow-y-auto p-3 flex flex-col gap-4">
        {KIND_SECTIONS.map((section) => {
          const kinds = entries
            .filter((entry) => section.kinds.includes(entry.kind))
            .reduce<[string, string, string][]>((result, entry) => {
              const key = kindKey(entry.apiGroup, entry.kind)
              if (!result.some(([existingKey]) => existingKey === key)) {
                result.push([key, entry.kind, entry.apiGroup])
              }
              return result
            }, [])
            .sort(
              ([, aKind], [, bKind]) =>
                section.kinds.indexOf(aKind) - section.kinds.indexOf(bKind)
            )
          if (!kinds.length) return null
          return (
            <div key={section.heading} className="flex flex-col gap-1">
              <div className="flex items-center gap-1.5 px-2.5">
                <Icon variant={SECTION_ICONS[section.heading]} size={14} />
                <Text variant="label" weight="strong" className="uppercase opacity-60">
                  {section.heading}
                </Text>
              </div>
              {kinds.map(([key, kind, apiGroup]) => (
                <KindNavButton
                  key={key}
                  kind={kind}
                  apiGroup={apiGroup}
                  count={counts.get(key) ?? 0}
                  isActive={selectedKind === key}
                  onClick={() => selectKind(key)}
                />
              ))}
            </div>
          )
        })}
        {helmEntries.length ? (
          <div className="flex flex-col gap-1">
            <div className="flex items-center gap-1.5 px-2.5">
              <Icon variant={SECTION_ICONS['Helm']} size={14} />
              <Text variant="label" weight="strong" className="uppercase opacity-60">
                Helm
              </Text>
            </div>
            <KindNavButton
              kind={HELM_NAV_KIND}
              count={helmEntries.length}
              isActive={isHelmView}
              onClick={() => selectKind(HELM_NAV_KIND)}
            />
          </div>
        ) : null}
        {customKinds.length ? (
          <div className="flex flex-col gap-1">
            <div className="flex items-center gap-1.5 px-2.5">
              <Icon variant={SECTION_ICONS['Custom resources']} size={14} />
              <Text variant="label" weight="strong" className="uppercase opacity-60">
                Custom resources
              </Text>
            </div>
            {customKinds.map(([key, [kind, apiGroup]]) => (
              <KindNavButton
                key={key}
                kind={kind}
                apiGroup={apiGroup}
                count={counts.get(key) ?? 0}
                isActive={selectedKind === key}
                onClick={() => selectKind(key)}
              />
            ))}
          </div>
        ) : null}
      </nav>

      <div className="flex-1 min-w-0 flex flex-col">
        <div className="flex items-center gap-2 p-3 border-b border-cool-grey-200 dark:border-dark-grey-600">
          <NamespaceFilter
            namespaces={namespaces}
            value={namespace}
            onChange={setNamespace}
          />
          <Input
            size="sm"
            placeholder={`Search ${pluralize(selectedKindName).toLowerCase()}`}
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            className="max-w-64"
          />
          <Text variant="subtext" className="ml-auto opacity-60" nowrap>
            {isHelmView
              ? `${helmRows.length} of ${helmEntries.length} releases`
              : `${rows.length} of ${counts.get(selectedKind) ?? 0} ${pluralize(
                  selectedKindName
                ).toLowerCase()}`}
          </Text>
        </div>

        <div className="flex-1 min-h-0 overflow-auto">
          {isHelmView ? (
            helmRows.length ? (
              <table className="w-full border-collapse">
                <thead className="sticky top-0 bg-white dark:bg-dark-grey-900 z-[1]">
                  <tr className="border-b border-cool-grey-200 dark:border-dark-grey-600">
                    {['Name', 'Namespace', 'Revision', 'History', 'Updated', 'Status'].map(
                      (header) => (
                        <th key={header} className="text-left px-3 py-2">
                          <Text
                            variant="label"
                            weight="strong"
                            className="uppercase opacity-60"
                            nowrap
                          >
                            {header}
                          </Text>
                        </th>
                      )
                    )}
                  </tr>
                </thead>
                <tbody>
                  {helmRows.map((row) => (
                    <tr
                      key={`${row.namespace}/${row.name}`}
                      className={cn(
                        'border-b border-cool-grey-100 dark:border-dark-grey-700 cursor-pointer transition-colors',
                        isHelmSelected(row)
                          ? 'bg-cool-grey-100 dark:bg-dark-grey-700'
                          : 'hover:bg-cool-grey-50 dark:hover:bg-dark-grey-800'
                      )}
                      onClick={() =>
                        setSelectedHelmKey(
                          isHelmSelected(row) ? undefined : helmEntryKey(row)
                        )
                      }
                    >
                      <td className="px-3 py-2 max-w-64">
                        <Text variant="body" family="mono" className="truncate">
                          {row.name}
                        </Text>
                      </td>
                      <td className="px-3 py-2">
                        <Text variant="subtext">{row.namespace || '—'}</Text>
                      </td>
                      <td className="px-3 py-2">
                        <Text variant="subtext" family="mono">
                          {row.revision}
                        </Text>
                      </td>
                      <td className="px-3 py-2">
                        <Text variant="subtext" family="mono">
                          {row.history.length}{' '}
                          {row.history.length === 1 ? 'revision' : 'revisions'}
                        </Text>
                      </td>
                      <td className="px-3 py-2">
                        {row.latest?.updated_at ? (
                          <Time
                            variant="subtext"
                            time={row.latest?.updated_at}
                            format="relative"
                          />
                        ) : (
                          <Text variant="subtext">—</Text>
                        )}
                      </td>
                      <td className="px-3 py-2">
                        <Badge size="sm" theme={helmStatusTheme(row.status)}>
                          {row.status}
                        </Badge>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            ) : (
              <EmptyState
                emptyTitle="No releases"
                emptyMessage={
                  search || namespace
                    ? 'No releases match the current filters.'
                    : 'No Helm releases have been captured yet.'
                }
              />
            )
          ) : rows.length ? (
            <table className="w-full border-collapse">
              <thead className="sticky top-0 bg-white dark:bg-dark-grey-900 z-[1]">
                <tr className="border-b border-cool-grey-200 dark:border-dark-grey-600">
                  {columns.map((col) => (
                    <th key={col.header} className="text-left px-3 py-2">
                      <Text variant="label" weight="strong" className="uppercase opacity-60" nowrap>
                        {col.header}
                      </Text>
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {rows.map((row) => (
                  <tr
                    key={entryKey(row)}
                    className={cn(
                      'border-b border-cool-grey-100 dark:border-dark-grey-700 cursor-pointer transition-colors',
                      isSelected(row)
                        ? 'bg-cool-grey-100 dark:bg-dark-grey-700'
                        : 'hover:bg-cool-grey-50 dark:hover:bg-dark-grey-800'
                    )}
                    onClick={() =>
                      setSelectedKey(isSelected(row) ? undefined : entryKey(row))
                    }
                  >
                    {columns.map((col) => (
                      <td key={col.header} className="px-3 py-2 max-w-64">
                        {col.cell(row)}
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          ) : (
            <EmptyState
              emptyTitle={`No ${pluralize(selectedKindName).toLowerCase()}`}
              emptyMessage={
                search || namespace
                  ? 'No resources match the current filters.'
                  : 'No resources of this kind have been captured yet.'
              }
            />
          )}
        </div>
      </div>

      {selected || selectedHelm ? (
        <aside className="w-[400px] shrink-0 border-l border-cool-grey-200 dark:border-dark-grey-600 min-h-0">
          {selectedHelm ? (
            <HelmDetailPane
              entry={selectedHelm}
              onClose={() => setSelectedHelmKey(undefined)}
            />
          ) : selected ? (
            <DetailPane
              entry={selected}
              entries={entries}
              onSelect={selectEntry}
              onClose={() => setSelectedKey(undefined)}
            />
          ) : null}
        </aside>
      ) : null}
    </div>
  )
}
