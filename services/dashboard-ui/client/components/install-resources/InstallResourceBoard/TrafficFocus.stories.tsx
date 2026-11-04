import { useState } from 'react'
import {
  Handle,
  MarkerType,
  Position,
  type Node,
  type NodeProps,
} from '@xyflow/react'
import { GraphCanvas } from '@/components/branches/graph/GraphCanvas'
import { layoutSequential } from '@/components/branches/graph/layout'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Loading } from '@/components/common/Loading'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { ToggleButton } from '@/components/common/ToggleButton'
import { ResourceKind } from '@/components/install-resources/ResourceKind'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { cn } from '@/utils/classnames'

export default {
  title: 'Features / Installs / Resources / Traffic focus',
}

type TPodMetrics =
  | { state: 'loading' }
  | { state: 'unavailable'; reason: string }
  | {
      state: 'available'
      cpuMillicores: number
      memoryMiB: number
      cpuLimitMillicores?: number
      memoryLimitMiB?: number
      sampledAt: string
      stale: boolean
      windowSeconds: number
    }

const sampledAt = (ageSeconds: number) =>
  new Date(Date.now() - ageSeconds * 1000).toISOString()

type TFixture = {
  id: string
  kind: string
  name: string
  namespace: string
  summary: string
  metrics?: TPodMetrics
}

type TRelationship = {
  source: string
  target: string
  type: 'network' | 'ownership'
}

const fixtures: TFixture[] = [
  {
    id: 'ingress',
    kind: 'Ingress',
    name: 'acme-web',
    namespace: 'default',
    summary: 'acme.example.com → acme-web:80',
  },
  {
    id: 'gateway',
    kind: 'Gateway',
    name: 'acme-public',
    namespace: 'default',
    summary: 'HTTPS listener · port 443',
  },
  {
    id: 'route',
    kind: 'HTTPRoute',
    name: 'acme-web-route',
    namespace: 'default',
    summary: 'Accepted · backend acme-web:80',
  },
  {
    id: 'service',
    kind: 'Service',
    name: 'acme-web',
    namespace: 'default',
    summary: 'Selector app=acme-web · 2 Pods',
  },
  {
    id: 'deployment',
    kind: 'Deployment',
    name: 'acme-web',
    namespace: 'default',
    summary: '2/2 ready',
  },
  {
    id: 'replicaset',
    kind: 'ReplicaSet',
    name: 'acme-web-7d9bf',
    namespace: 'default',
    summary: '2/2 ready',
  },
  {
    id: 'pod-a',
    kind: 'Pod',
    name: 'acme-web-7d9bf-a',
    namespace: 'default',
    summary: 'Running · 1/1 ready',
    metrics: {
      state: 'available',
      cpuMillicores: 125,
      memoryMiB: 180,
      cpuLimitMillicores: 500,
      memoryLimitMiB: 512,
      sampledAt: sampledAt(20),
      stale: false,
      windowSeconds: 30,
    },
  },
  {
    id: 'pod-b',
    kind: 'Pod',
    name: 'acme-web-7d9bf-b',
    namespace: 'default',
    summary: 'Running · 1/1 ready',
    metrics: {
      state: 'available',
      cpuMillicores: 80,
      memoryMiB: 96,
      sampledAt: sampledAt(20),
      stale: false,
      windowSeconds: 30,
    },
  },
  {
    id: 'worker',
    kind: 'Deployment',
    name: 'acme-worker',
    namespace: 'default',
    summary: '1/1 ready',
  },
  {
    id: 'worker-pod',
    kind: 'Pod',
    name: 'acme-worker-56c9-a',
    namespace: 'default',
    summary: 'Running · 1/1 ready',
    metrics: {
      state: 'available',
      cpuMillicores: 300,
      memoryMiB: 220,
      cpuLimitMillicores: 1000,
      memoryLimitMiB: 256,
      sampledAt: sampledAt(360),
      stale: true,
      windowSeconds: 30,
    },
  },
  {
    id: 'other-service',
    kind: 'Service',
    name: 'acme-web',
    namespace: 'staging',
    summary: 'Same name · different namespace',
  },
  {
    id: 'other-pod',
    kind: 'Pod',
    name: 'acme-web-staging-a',
    namespace: 'staging',
    summary: 'Running · 1/1 ready',
    metrics: {
      state: 'unavailable',
      reason: 'Runner access denied',
    },
  },
  {
    id: 'unresolved',
    kind: 'Service',
    name: 'acme-external',
    namespace: 'default',
    summary: 'ExternalName · no workload relationship',
  },
]

const relationships: TRelationship[] = [
  { source: 'ingress', target: 'service', type: 'network' },
  { source: 'gateway', target: 'route', type: 'network' },
  { source: 'route', target: 'service', type: 'network' },
  { source: 'service', target: 'pod-a', type: 'network' },
  { source: 'service', target: 'pod-b', type: 'network' },
  { source: 'deployment', target: 'replicaset', type: 'ownership' },
  { source: 'replicaset', target: 'pod-a', type: 'ownership' },
  { source: 'replicaset', target: 'pod-b', type: 'ownership' },
  { source: 'worker', target: 'worker-pod', type: 'ownership' },
  { source: 'other-service', target: 'other-pod', type: 'network' },
]

const networkKinds = ['Ingress', 'Gateway', 'HTTPRoute', 'Service']

const focusedResources = (selected: string, links: TRelationship[]) => {
  const ids = new Set([selected])
  const pending = [selected]
  while (pending.length) {
    const id = pending.pop()
    for (const link of links) {
      if (link.source === id && !ids.has(link.target)) {
        ids.add(link.target)
        pending.push(link.target)
      }
    }
  }
  const ancestors = [...ids]
  while (ancestors.length) {
    const id = ancestors.pop()
    for (const link of links) {
      if (link.target === id && !ids.has(link.source)) {
        ids.add(link.source)
        ancestors.push(link.source)
      }
    }
  }
  return ids
}

const UsageBar = ({
  label,
  value,
  limit,
  unit,
}: {
  label: string
  value: number
  limit?: number
  unit: string
}) => {
  const percentage =
    limit && limit > 0 ? Math.round((value / limit) * 100) : undefined
  return (
    <div className="grid w-full grid-cols-[3rem_minmax(0,1fr)_auto] items-center gap-x-3 gap-y-1">
      <Text variant="subtext">{label}</Text>
      {percentage !== undefined ? (
        <div
          role="meter"
          aria-label={`${label} percent of limit`}
          aria-valuemin={0}
          aria-valuemax={Math.max(100, percentage)}
          aria-valuenow={percentage}
          className="h-1.5 w-full overflow-hidden rounded-full bg-cool-grey-200 dark:bg-dark-grey-600"
        >
          <div
            className="h-full rounded-full bg-primary-500"
            style={{ width: `${Math.min(100, percentage)}%` }}
          />
        </div>
      ) : (
        <span />
      )}
      <Text
        variant="subtext"
        weight="strong"
        className="text-right tabular-nums"
      >
        {value}
        {unit}
        {percentage !== undefined ? ` (${percentage}%)` : ''}
      </Text>
      <Text
        variant="label"
        theme="neutral"
        className="col-span-2 col-start-2 text-right"
      >
        {percentage !== undefined
          ? `of ${limit}${unit} limit`
          : 'No complete limit'}
      </Text>
    </div>
  )
}

const Metrics = ({ metrics }: { metrics: TPodMetrics }) => {
  if (metrics.state === 'loading') {
    return (
      <div role="status" className="flex items-center gap-2">
        <Loading size={12} />
        <Text variant="label" theme="neutral">
          Fetching Pod metrics…
        </Text>
      </div>
    )
  }
  if (metrics.state === 'unavailable') {
    return (
      <Text
        variant="subtext"
        theme="neutral"
        className="w-full truncate"
        title={metrics.reason}
      >
        {metrics.reason}
      </Text>
    )
  }
  return (
    <div className="flex w-full flex-col gap-3" aria-label="Pod resource usage">
      <UsageBar
        label="CPU"
        value={metrics.cpuMillicores}
        limit={metrics.cpuLimitMillicores}
        unit="m"
      />
      <UsageBar
        label="Memory"
        value={metrics.memoryMiB}
        limit={metrics.memoryLimitMiB}
        unit=" MiB"
      />
    </div>
  )
}

type TFocusNode = Node<
  TFixture & {
    selected: boolean
    dimmed: boolean
    showNamespace: boolean
    onSelect: () => void
  }
>

const FocusCard = ({ data, height }: NodeProps<TFocusNode>) => (
  <Button
    variant="ghost"
    aria-label={`Focus ${data.kind} ${data.namespace}/${data.name}`}
    aria-pressed={data.selected}
    onClick={data.onSelect}
    data-dimmed={data.dimmed}
    style={{ height }}
    tooltipProps={{
      position: 'bottom',
      tipContent: (
        <div className="flex flex-col gap-1">
          <Text variant="label" family="mono">
            {data.kind} · {data.namespace}/{data.name}
          </Text>
          {data.metrics?.state === 'available' ? (
            <>
              <Text variant="label">
                Sampled{' '}
                <Time
                  variant="label"
                  time={data.metrics.sampledAt}
                  format="long-datetime"
                />
              </Text>
              <Text variant="label">
                CPU: {data.metrics.windowSeconds}s average · percentages of
                limits
              </Text>
            </>
          ) : null}
        </div>
      ),
    }}
    className={cn(
      'nodrag nopan !w-80 !p-4 !border hover:!bg-elevation-3 flex-col !items-stretch !justify-start !gap-4 text-left',
      data.dimmed ? '!bg-transparent' : '!bg-elevation-2',
      data.selected && '!outline-2 !outline-primary-500'
    )}
  >
    <Handle
      id="left"
      type="target"
      position={Position.Left}
      isConnectable={false}
      className="!bg-cool-grey-500 dark:!bg-dark-grey-300 !border-0"
    />
    <div className="flex w-full items-start justify-between gap-3">
      <div className="flex min-w-0 flex-col gap-1">
        <div className="flex items-center gap-2">
          <ResourceKind
            resource={{ kind: data.kind, provider: 'kubernetes' }}
            compact
          />
          {data.showNamespace ? (
            <Badge size="xs" variant="code">
              {data.namespace}
            </Badge>
          ) : null}
        </div>
        <Text
          variant="subtext"
          family="mono"
          theme={data.dimmed ? 'neutral' : 'default'}
          weight="strong"
          className="w-full truncate"
          title={data.name}
        >
          {data.name}
        </Text>
      </div>
      <div className="flex shrink-0 flex-col items-end gap-1">
        {data.metrics?.state === 'available' && data.metrics.stale ? (
          <Text
            variant="label"
            theme="warn"
            className="flex items-center gap-1"
          >
            <Icon variant="WarningIcon" size={12} />
            Stale metrics
          </Text>
        ) : data.metrics?.state === 'unavailable' ? (
          <Text variant="label" theme="neutral">
            Metrics unavailable
          </Text>
        ) : data.kind === 'Pod' ? (
          <Status status="healthy">Running</Status>
        ) : null}
        {data.kind === 'Pod' ? (
          <Text variant="label" theme="neutral">
            {data.metrics?.state === 'unavailable' ||
            (data.metrics?.state === 'available' && data.metrics.stale)
              ? data.summary
              : data.summary.replace('Running · ', '')}
          </Text>
        ) : ['Deployment', 'ReplicaSet'].includes(data.kind) ? (
          <Status status="healthy">{data.summary}</Status>
        ) : null}
      </div>
    </div>
    {data.metrics ? (
      <Metrics metrics={data.metrics} />
    ) : !['Deployment', 'ReplicaSet'].includes(data.kind) ? (
      <Text
        variant="subtext"
        theme="neutral"
        className="w-full truncate"
        title={data.summary}
      >
        {data.summary}
      </Text>
    ) : null}
    <Handle
      id="right"
      type="source"
      position={Position.Right}
      isConnectable={false}
      className="!bg-cool-grey-500 dark:!bg-dark-grey-300 !border-0"
    />
  </Button>
)

const nodeTypes = { resource: FocusCard }

const TrafficFocusStory = ({
  initialSelection,
  initialMode = 'highlight',
  initialView = 'ownership',
  podMetricsState,
}: {
  initialSelection?: string
  initialMode?: 'highlight' | 'filter'
  initialView?: TRelationship['type']
  podMetricsState?: 'loading' | 'unavailable'
}) => {
  const [selected, setSelected] = useState(initialSelection)
  const [mode, setMode] = useState(initialMode)
  const [view, setView] = useState(initialView)
  const resources = fixtures.filter(({ kind }) =>
    view === 'network'
      ? networkKinds.includes(kind) || kind === 'Pod'
      : !networkKinds.includes(kind)
  )
  const links = relationships.filter(({ type }) => type === view)
  const focused = selected
    ? focusedResources(selected, links)
    : new Set(resources.map(({ id }) => id))
  const selection = resources.find(({ id }) => id === selected)
  const visible = resources.filter(
    ({ id }) => mode === 'highlight' || focused.has(id)
  )
  const visibleIds = new Set(visible.map(({ id }) => id))
  const pods = resources.filter(
    ({ id, kind }) => focused.has(id) && kind === 'Pod'
  )
  const nodes: TFocusNode[] = visible.map((fixture) => {
    const metrics: TPodMetrics | undefined =
      fixture.id === 'pod-a' && podMetricsState
        ? podMetricsState === 'loading'
          ? { state: 'loading' }
          : { state: 'unavailable', reason: 'Metrics API is not installed' }
        : fixture.metrics
    return {
      id: fixture.id,
      type: 'resource',
      position: { x: 0, y: 0 },
      width: 320,
      height:
        metrics?.state === 'available'
          ? 176
          : ['Deployment', 'ReplicaSet'].includes(fixture.kind)
            ? 88
            : 112,
      draggable: false,
      focusable: false,
      data: {
        ...fixture,
        metrics,
        selected: selected === fixture.id,
        dimmed: !focused.has(fixture.id),
        showNamespace:
          fixture.namespace !== 'default' &&
          resources.some(
            (other) =>
              other.kind === fixture.kind &&
              other.name === fixture.name &&
              other.namespace !== fixture.namespace
          ),
        onSelect: () =>
          setSelected(selected === fixture.id ? undefined : fixture.id),
      },
    }
  })
  const edges = links
    .filter(
      ({ source, target }) => visibleIds.has(source) && visibleIds.has(target)
    )
    .map((link) => {
      const active = focused.has(link.source) && focused.has(link.target)
      const color =
        active && selected ? 'var(--color-primary-500)' : 'var(--foreground)'
      return {
        id: `${link.source}/${link.target}`,
        source: link.source,
        target: link.target,
        sourceHandle: 'right',
        targetHandle: 'left',
        type: 'smoothstep',
        focusable: false,
        style: {
          stroke: color,
          strokeWidth: active && selected ? 2 : 1.5,
          opacity: active ? 0.5 : 0.15,
        },
        markerEnd: { type: MarkerType.ArrowClosed, color },
      }
    })
  const layout = layoutSequential(nodes, edges, {
    nodeWidth: 320,
    minHeight: 88,
    baseHeight: 0,
    rowHeight: 1,
    rowCount: (node) => node.height ?? 88,
    ranksep: 64,
  })

  return (
    <div className="flex flex-col gap-4 p-4">
      <SectionHeader
        title="Resource relationships"
        status={<Badge size="sm">Story prototype</Badge>}
        description={
          <>
            {view === 'ownership'
              ? 'Follow controllers and the resources they own.'
              : 'Follow backend references and Service selectors. Configured relationships, not live traffic.'}{' '}
            Select a resource to focus its path. Pod usage is always visible.
          </>
        }
      />
      <div className="flex flex-wrap items-center gap-3">
        <ToggleButton<TRelationship['type']>
          label="Relationship view"
          value={view}
          onChange={(next) => {
            if (next === view) return
            setView(next)
            if (selection?.kind !== 'Pod') setSelected(undefined)
          }}
          options={[
            { value: 'ownership', label: 'Ownership' },
            { value: 'network', label: 'Network' },
          ]}
        />
        <ToggleButton<'highlight' | 'filter'>
          label="Relationship focus mode"
          value={mode}
          onChange={setMode}
          options={[
            { value: 'highlight', label: 'Highlight related' },
            { value: 'filter', label: 'Only related' },
          ]}
        />
        {selection ? (
          <>
            <Badge variant="code" theme="brand">
              {selection.kind} · {selection.namespace}/{selection.name}
            </Badge>
            <Button
              size="sm"
              variant="ghost"
              onClick={() => setSelected(undefined)}
            >
              Clear focus
            </Button>
          </>
        ) : (
          <Text variant="subtext" theme="neutral">
            {resources.length} resources in this view
          </Text>
        )}
      </div>
      <div role="status">
        <Text variant="subtext" theme="neutral">
          {selection
            ? focused.size > 1
              ? `${focused.size} related resources · ${pods.length} Pod${pods.length === 1 ? '' : 's'} · ${resources.length - focused.size} unrelated resources ${mode === 'filter' ? 'hidden' : 'dimmed'}`
              : 'No relationships found in this view. Switch views or select another resource.'
            : view === 'ownership'
              ? 'Controller ownership only. Network relationships are in the Network view.'
              : 'Backend references and selector matches only. Controllers are in the Ownership view.'}
        </Text>
      </div>
      <GraphCanvas
        key={`${view}/${mode}`}
        nodes={layout.nodes}
        edges={layout.edges}
        nodeTypes={nodeTypes}
        height={650}
        minZoom={0.3}
        fitPadding={0.12}
      />
      <Text variant="label" theme="neutral">
        Relationships and metrics are mocked; no runner requests are made.
        Proposed sampling: 5 min background · 30s while canvas visible · refresh
        stale metrics on Pod click. Percentages are of complete Pod limits, not
        requests.
      </Text>
    </div>
  )
}

export const Explore = () => <TrafficFocusStory />
export const OwnershipHighlight = () => (
  <TrafficFocusStory initialSelection="deployment" />
)
export const Network = () => <TrafficFocusStory initialView="network" />
export const ServiceHighlight = () => (
  <TrafficFocusStory initialSelection="service" initialView="network" />
)
export const IngressFilter = () => (
  <TrafficFocusStory
    initialSelection="ingress"
    initialMode="filter"
    initialView="network"
  />
)
export const GatewayHighlight = () => (
  <TrafficFocusStory initialSelection="gateway" initialView="network" />
)
export const UnresolvedService = () => (
  <TrafficFocusStory initialSelection="unresolved" initialView="network" />
)
export const PodMetricsLoading = () => (
  <TrafficFocusStory initialSelection="pod-a" podMetricsState="loading" />
)
export const PodMetricsUnavailable = () => (
  <TrafficFocusStory initialSelection="pod-a" podMetricsState="unavailable" />
)
export const PodMetrics = () => <TrafficFocusStory initialSelection="pod-a" />
