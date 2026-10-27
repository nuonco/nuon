import { useState } from 'react'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { PropertyGrid } from '@/components/common/PropertyGrid'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { ToggleButton } from '@/components/common/ToggleButton'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { cn } from '@/utils/classnames'

export default { title: 'Features / Installs / Resources / Node density' }

type TDimension = 'cpu' | 'memory' | 'storage'
type TResources = Record<TDimension, number>
type TPod = {
  name: string
  group: string
  requests: TResources
  usage: Partial<TResources>
  limits: Partial<TResources>
}
type TDensityNode = {
  name: string
  zone: string
  allocatable: TResources
  pods: TPod[]
}

const dimensions: { value: TDimension; label: string }[] = [
  { value: 'cpu', label: 'CPU' },
  { value: 'memory', label: 'Memory' },
  { value: 'storage', label: 'Storage' },
]

const workloads = {
  'acme-web': {
    requests: { cpu: 1000, memory: 6, storage: 5 },
    limits: { cpu: 2000, memory: 10, storage: 20 },
    color:
      'bg-primary-200 text-primary-950 dark:bg-primary-800 dark:text-primary-100',
    scaling: 'HPA · 2–8 replicas',
  },
  'acme-api': {
    requests: { cpu: 700, memory: 4, storage: 10 },
    limits: { cpu: 2000, memory: 8, storage: 20 },
    color: 'bg-blue-200 text-blue-950 dark:bg-blue-900 dark:text-blue-100',
    scaling: 'MPA · 2–10 replicas · 2–6 GiB memory per replica',
  },
  'acme-worker': {
    requests: { cpu: 500, memory: 2, storage: 15 },
    limits: { cpu: 1500, memory: 6, storage: 30 },
    color: 'bg-green-100 text-green-950 dark:bg-green-900 dark:text-green-100',
    scaling: 'Fixed replica count',
  },
}
type TWorkload = keyof typeof workloads
const daemonGroup = 'DaemonSets'
const groupNames = [...Object.keys(workloads), daemonGroup]
const daemonColor =
  'bg-orange-200 text-orange-950 dark:bg-orange-900 dark:text-orange-100'

const makePods = (
  node: string,
  group: TWorkload,
  requests: TResources,
  usage: [number, number, number?][]
): TPod[] =>
  usage.map(([cpu, memory, storage], index) => ({
    name: `${group}-${node}-${index + 1}`,
    group,
    requests,
    limits: workloads[group].limits,
    usage: { cpu, memory, ...(storage === undefined ? {} : { storage }) },
  }))

const daemonPods = (node: string): TPod[] =>
  [
    ['acme-cni', 0.25, 0.25],
    ['kube-proxy', 0.125, 0.25],
    ['acme-logs', 0.5, 1],
    ['acme-security', 0.5, 1],
    ['node-exporter', 0.125, 0.5],
    ['acme-csi', 0.5, 1],
  ].map(([name, memory, storage]) => ({
    name: `${name}-${node}`,
    group: daemonGroup,
    requests: { cpu: 100, memory: Number(memory), storage: Number(storage) },
    usage: {
      cpu: 35,
      memory: Number(memory) / 2,
      storage: Number(storage) / 2,
    },
    limits: {},
  }))

const pools: { name: string; nodes: TDensityNode[] }[] = [
  {
    name: 'general',
    nodes: [
      {
        name: 'general-01',
        zone: 'zone-a',
        allocatable: { cpu: 7500, memory: 30, storage: 100 },
        pods: [
          ...makePods('01', 'acme-web', workloads['acme-web'].requests, [
            [300, 3, 6],
            [450, 2.5, 4],
          ]),
          ...makePods('01', 'acme-api', { cpu: 700, memory: 4, storage: 10 }, [
            [280, 2, 7],
            [240, 2.5, 6],
          ]),
          ...makePods('01', 'acme-worker', workloads['acme-worker'].requests, [
            [200, 1, 12],
            [250, 1.5, 14],
          ]),
          ...daemonPods('01'),
        ],
      },
      {
        name: 'general-02',
        zone: 'zone-b',
        allocatable: { cpu: 7500, memory: 30, storage: 100 },
        pods: [
          ...makePods('02', 'acme-web', workloads['acme-web'].requests, [
            [250, 3.5, 8],
          ]),
          ...makePods('02', 'acme-api', { cpu: 800, memory: 5, storage: 7.5 }, [
            [300, 3, 5],
            [350, 4, 6],
          ]),
          ...makePods(
            '02',
            'acme-worker',
            { cpu: 600, memory: 3, storage: 20 },
            [
              [200, 1, 18],
              [300, 2, 16],
            ]
          ),
          ...daemonPods('02'),
        ],
      },
      {
        name: 'general-03',
        zone: 'zone-c',
        allocatable: { cpu: 7500, memory: 30, storage: 100 },
        pods: [
          ...makePods('03', 'acme-api', { cpu: 500, memory: 4, storage: 8 }, [
            [200, 2, 6],
            [180, 3, 5],
            [220, 2.5, 4],
          ]),
          ...makePods(
            '03',
            'acme-worker',
            { cpu: 500, memory: 2, storage: 10 },
            [
              [200, 1, 7],
              [180, 1.5],
            ]
          ),
          ...daemonPods('03'),
        ],
      },
    ],
  },
  {
    name: 'compact',
    nodes: [
      {
        name: 'compact-01',
        zone: 'zone-a',
        allocatable: { cpu: 3500, memory: 14, storage: 40 },
        pods: [
          ...makePods(
            'compact-01',
            'acme-api',
            { cpu: 800, memory: 4, storage: 8 },
            [
              [300, 2.5, 6],
              [400, 3, 7],
            ]
          ),
          ...makePods(
            'compact-01',
            'acme-worker',
            { cpu: 1200, memory: 3, storage: 10 },
            [[600, 2, 8]]
          ),
          ...daemonPods('compact-01'),
        ],
      },
      {
        name: 'compact-02',
        zone: 'zone-b',
        allocatable: { cpu: 3500, memory: 14, storage: 40 },
        pods: [
          ...makePods(
            'compact-02',
            'acme-api',
            { cpu: 800, memory: 4, storage: 8 },
            [[350, 2, 5]]
          ),
          ...makePods(
            'compact-02',
            'acme-worker',
            { cpu: 600, memory: 3, storage: 10 },
            [[250, 1.5, 9]]
          ),
          ...daemonPods('compact-02'),
        ],
      },
    ],
  },
]

const number = new Intl.NumberFormat('en-US', { maximumFractionDigits: 2 })
const format = (value: number | undefined, dimension: TDimension) => {
  if (value === undefined) return 'Unavailable'
  if (dimension === 'cpu')
    return value < 1000
      ? `${number.format(value)}m`
      : `${number.format(value / 1000)} CPU`
  return value > 0 && value < 1
    ? `${number.format(value * 1024)} MiB`
    : `${number.format(value)} GiB`
}

const total = (
  pods: TPod[],
  field: 'requests' | 'usage' | 'limits',
  dimension: TDimension
) =>
  pods.some((pod) => pod[field][dimension] === undefined)
    ? undefined
    : pods.reduce((sum, pod) => sum + pod[field][dimension], 0)

const replicaFit = (node: TDensityNode, workload: TWorkload) => {
  const current = node.pods.filter(({ group }) => group === workload).length
  const additional = Math.min(
    ...dimensions.map(({ value }) =>
      Math.floor(
        (node.allocatable[value] - total(node.pods, 'requests', value)) /
          workloads[workload].requests[value]
      )
    )
  )
  return current + Math.max(0, additional)
}

const ResourceProfile = ({ pods }: { pods: TPod[] }) => (
  <PropertyGrid
    aria-label="Resource profile"
    gridTemplate="minmax(0, 1fr) repeat(3, max-content)"
    values={dimensions.map(({ value, label }) => ({
      resource: value === 'storage' ? 'Local storage' : label,
      usage: format(total(pods, 'usage', value), value),
      requests: format(total(pods, 'requests', value), value),
      limits:
        total(pods, 'limits', value) === undefined
          ? pods.every((pod) => pod.limits[value] === undefined)
            ? 'Not set'
            : 'Incomplete'
          : format(total(pods, 'limits', value), value),
    }))}
    columns={[
      {
        key: 'resource',
        header: 'Resource',
        render: (_, item) => <Text variant="subtext">{item.resource}</Text>,
      },
      ...(['usage', 'requests', 'limits'] as const).map((field) => ({
        key: field,
        header:
          field === 'usage'
            ? 'Current'
            : field === 'requests'
              ? 'Request'
              : 'Limit',
        className: '!pl-2 text-right justify-end',
        render: (value: string) => (
          <Text variant="subtext" className="tabular-nums">
            {value}
          </Text>
        ),
      })),
    ]}
  />
)

const DensityNode = ({
  node,
  dimension,
  selectedGroup,
  selectedPod,
  onSelectGroup,
  onSelectPod,
}: {
  node: TDensityNode
  dimension: TDimension
  selectedGroup?: string
  selectedPod?: string
  onSelectGroup: (group: string) => void
  onSelectPod: (pod: string) => void
}) => {
  const requested = total(node.pods, 'requests', dimension)
  const usage = total(node.pods, 'usage', dimension)
  const capacity = node.allocatable[dimension]
  const remaining = capacity - requested
  const groupPods = node.pods.filter(({ group }) => group === selectedGroup)
  const allocations = groupNames
    .map((group) => {
      const pods = node.pods.filter((pod) => pod.group === group)
      return {
        group,
        pods,
        usage: total(pods, 'usage', dimension),
        requests: total(pods, 'requests', dimension),
        limits: total(pods, 'limits', dimension),
      }
    })
    .filter(({ pods }) => pods.length)
  return (
    <Card className="!gap-4 !p-4 min-w-0" aria-label={`Node ${node.name}`}>
      <div className="flex items-center justify-between gap-2">
        <Text family="mono" weight="strong">
          {node.name}
        </Text>
        <Status status="healthy">Ready</Status>
      </div>
      <div className="flex items-center justify-between gap-2">
        <Text variant="subtext" theme="neutral">
          {node.zone}
        </Text>
        <Text variant="subtext" theme="neutral">
          {node.pods.length} Pods
        </Text>
      </div>
      <div className="flex gap-2">
        <div
          aria-hidden
          className="flex w-10 shrink-0 flex-col justify-between text-right"
        >
          {[capacity, capacity / 2, 0].map((value) => (
            <Text key={value} variant="label" theme="neutral">
              {format(value, dimension)}
            </Text>
          ))}
        </div>
        <div
          className="flex h-96 min-w-0 flex-1 flex-col overflow-hidden rounded-md border bg-elevation-1"
          aria-label={`${node.name} ${dimension} allocation`}
        >
          <div
            className="flex min-h-0 shrink-0 items-center justify-center overflow-hidden px-2 text-center"
            style={{ height: `${(remaining / capacity) * 100}%` }}
          >
            {remaining / capacity > 0.08 ? (
              <Text variant="label" theme="neutral">
                {format(remaining, dimension)} unrequested
              </Text>
            ) : null}
          </div>
          {allocations.map(
            ({ group, pods, requests: amount, usage: used, limits }) => {
              const usageLabel =
                used === undefined
                  ? 'Usage unavailable'
                  : `${format(used, dimension)} now`
              const limitLabel =
                limits === undefined
                  ? pods.every((pod) => pod.limits[dimension] === undefined)
                    ? 'No limit set'
                    : 'Incomplete limits'
                  : `${format(limits, dimension)} limit`
              const height = (amount / capacity) * 100
              const expanded = selectedGroup === group
              const color =
                group === daemonGroup ? daemonColor : workloads[group].color
              return (
                <div
                  key={group}
                  data-allocation={group}
                  data-amount={amount}
                  className={cn(
                    'flex min-h-0 shrink-0 flex-col overflow-hidden border-t transition-opacity',
                    color,
                    selectedGroup && !expanded && 'opacity-80'
                  )}
                  style={{ height: `${height}%` }}
                >
                  <Button
                    variant="ghost"
                    aria-label={`${expanded ? 'Collapse' : 'Inspect'} ${group} on ${node.name}`}
                    aria-expanded={expanded}
                    onClick={() => onSelectGroup(group)}
                    aria-description={`${pods.length} Pods · ${usageLabel} · ${format(amount, dimension)} requested · ${limitLabel}`}
                    className={cn(
                      '!w-full !min-h-0 !rounded-none !border-0 !px-2 !py-0 !text-inherit flex-col justify-center !whitespace-normal overflow-hidden focus-visible:!outline-solid focus-visible:!outline-primary-500 focus-visible:!outline-offset-[-2px]',
                      expanded
                        ? height > 10
                          ? '!h-5 shrink-0'
                          : '!h-0 shrink-0'
                        : '!h-full'
                    )}
                  >
                    {expanded || height > 13 ? (
                      <Text
                        variant="label"
                        family={group === daemonGroup ? 'sans' : 'mono'}
                        weight="strong"
                        nowrap
                        className="w-full truncate"
                      >
                        {group}
                        {!expanded && height <= 18
                          ? ` · ${pods.length} Pods`
                          : ''}
                      </Text>
                    ) : (
                      <span className="flex w-full min-w-0 items-center justify-between gap-1">
                        <Text
                          variant="label"
                          family={group === daemonGroup ? 'sans' : 'mono'}
                          nowrap
                          className="truncate"
                        >
                          {group}
                        </Text>
                        <Text
                          variant="label"
                          weight="strong"
                          nowrap
                          className="shrink-0"
                        >
                          {usageLabel}
                        </Text>
                      </span>
                    )}
                    {!expanded && height > 13 ? (
                      <>
                        <Text variant="label" weight="strong" nowrap>
                          {usageLabel}
                        </Text>
                        <Text
                          variant="label"
                          nowrap
                          className="w-full truncate"
                        >
                          {format(amount, dimension)}{' '}
                          {height > 18
                            ? `requested · ${pods.length} ${pods.length === 1 ? 'Pod' : 'Pods'}`
                            : `req · ${limitLabel}`}
                        </Text>
                        {height > 18 ? (
                          <Text variant="label" nowrap>
                            {limitLabel}
                          </Text>
                        ) : null}
                      </>
                    ) : null}
                  </Button>
                  {expanded ? (
                    <div
                      className="flex min-h-0 flex-1"
                      aria-label={`${group} Pods on ${node.name}`}
                    >
                      {pods.map((pod) => (
                        <Button
                          key={pod.name}
                          variant="ghost"
                          onClick={() => onSelectPod(pod.name)}
                          aria-label={`Inspect Pod ${pod.name}`}
                          aria-pressed={selectedPod === pod.name}
                          aria-description={`${format(pod.usage[dimension], dimension)} current usage · ${format(pod.requests[dimension], dimension)} requested · ${pod.limits[dimension] === undefined ? 'No limit set' : `${format(pod.limits[dimension], dimension)} limit`}`}
                          style={{
                            width: `${(pod.requests[dimension] / amount) * 100}%`,
                          }}
                          className={cn(
                            '!h-full !min-w-0 !min-h-0 !rounded-none !border-y-0 !border-r-0 !border-l-current/30 first:!border-l-0 !px-1 !py-0 !text-inherit flex-col !justify-center overflow-hidden !whitespace-normal bg-white/20 dark:bg-black/20 focus-visible:!outline-solid focus-visible:!outline-primary-500 focus-visible:!outline-offset-[-2px]',
                            selectedPod === pod.name &&
                              '!bg-white/50 dark:!bg-black/40 !outline-solid !outline-2 !outline-primary-600 !outline-offset-[-2px] dark:!outline-primary-300'
                          )}
                        >
                          {height > 22 ? (
                            <Text
                              variant="label"
                              family="mono"
                              nowrap
                              className="w-full truncate"
                            >
                              {pod.name}
                            </Text>
                          ) : null}
                          <Text
                            variant="label"
                            weight="strong"
                            nowrap
                            className="w-full truncate"
                          >
                            {pod.usage[dimension] === undefined
                              ? 'Usage unavailable'
                              : `${format(pod.usage[dimension], dimension)} now`}
                          </Text>
                          {height > 13 ? (
                            <Text
                              variant="label"
                              nowrap
                              className="w-full truncate"
                            >
                              {format(pod.requests[dimension], dimension)}{' '}
                              {height > 18
                                ? 'requested'
                                : `req · ${pod.limits[dimension] === undefined ? 'No limit set' : `${format(pod.limits[dimension], dimension)} limit`}`}
                            </Text>
                          ) : null}
                          {height > 18 ? (
                            <Text
                              variant="label"
                              nowrap
                              className="w-full truncate"
                            >
                              {pod.limits[dimension] === undefined
                                ? 'No limit set'
                                : `${format(pod.limits[dimension], dimension)} limit`}
                            </Text>
                          ) : null}
                        </Button>
                      ))}
                    </div>
                  ) : null}
                </div>
              )
            }
          )}
        </div>
      </div>
      <div className="flex flex-col gap-1">
        <Text variant="subtext" weight="strong">
          {format(capacity, dimension)} allocatable
        </Text>
        <Text variant="label" theme="neutral">
          {usage === undefined
            ? 'Usage unavailable'
            : `${format(usage, dimension)} used now`}{' '}
          · {format(requested, dimension)} requested
        </Text>
      </div>
      <PropertyGrid
        aria-label={`${node.name} workload resources`}
        values={allocations}
        gridTemplate="minmax(0, 1.4fr) repeat(3, minmax(0, 1fr))"
        columns={[
          {
            key: 'group',
            header: 'Workload',
            render: (_, { group, pods }) => (
              <span className="flex min-w-0 flex-col gap-1">
                <Text
                  variant="label"
                  family={group === daemonGroup ? 'sans' : 'mono'}
                  nowrap
                  className="truncate"
                >
                  {group}
                </Text>
                <Text variant="label" theme="neutral" nowrap>
                  {pods.length}{' '}
                  {group === daemonGroup
                    ? 'Pods'
                    : pods.length === 1
                      ? 'replica'
                      : 'replicas'}
                </Text>
              </span>
            ),
          },
          ...(['usage', 'requests', 'limits'] as const).map((field) => ({
            key: field,
            header:
              field === 'usage'
                ? 'Current'
                : field === 'requests'
                  ? 'Request'
                  : 'Limit',
            className: '!pl-2 text-right justify-end',
            render: (_, item: (typeof allocations)[number]) => (
              <Text
                variant="label"
                weight={field === 'usage' ? 'strong' : 'normal'}
                className="tabular-nums"
                aria-label={
                  field === 'usage' && item[field] === undefined
                    ? 'Usage unavailable'
                    : undefined
                }
              >
                {item[field] === undefined
                  ? field === 'limits'
                    ? item.pods.every(
                        (pod) => pod.limits[dimension] === undefined
                      )
                      ? 'Not set'
                      : 'Incomplete'
                    : '—'
                  : format(item[field], dimension)}
              </Text>
            ),
          })),
        ]}
      />
      {selectedGroup && selectedGroup !== daemonGroup ? (
        <Text variant="subtext" weight="strong">
          {groupPods.length} running /{' '}
          {replicaFit(node, selectedGroup as TWorkload)} fit
        </Text>
      ) : null}
    </Card>
  )
}

const NodeDensityStory = ({
  initialDimension = 'memory',
  initialGroup,
  storageUnavailable = false,
}: {
  initialDimension?: TDimension
  initialGroup?: string
  storageUnavailable?: boolean
}) => {
  const [poolName, setPoolName] = useState('general')
  const [dimension, setDimension] = useState(initialDimension)
  const [selectedGroup, setSelectedGroup] = useState(initialGroup)
  const [selectedPod, setSelectedPod] = useState<string>()
  const fixture = pools.find(({ name }) => name === poolName)
  const nodes = fixture.nodes.map((node) => ({
    ...node,
    pods: node.pods.map((pod) => ({
      ...pod,
      usage: storageUnavailable
        ? { ...pod.usage, storage: undefined }
        : pod.usage,
    })),
  }))
  const allPods = nodes.flatMap(({ pods }) => pods)
  const groupPods = allPods.filter(({ group }) => group === selectedGroup)
  const pod = groupPods.find(({ name }) => name === selectedPod)
  const inspected = pod ? [pod] : selectedGroup ? groupPods : allPods
  const workload =
    selectedGroup && selectedGroup !== daemonGroup
      ? (selectedGroup as TWorkload)
      : undefined
  const onSelectGroup = (group: string) => {
    setSelectedGroup(selectedGroup === group ? undefined : group)
    setSelectedPod(undefined)
  }
  return (
    <div className="flex flex-col gap-6 p-4" data-testid="node-density-story">
      <SectionHeader
        title="Node density"
        status={<Badge size="sm">Story prototype</Badge>}
        description="See how workloads fill each node. Select a Deployment to split its allocation into Pods."
      />
      <div className="grid gap-6 lg:grid-cols-[10rem_minmax(0,1fr)]">
        <nav
          aria-label="Node pools"
          className="flex flex-wrap content-start gap-2 lg:flex-col"
        >
          <Text variant="subtext" weight="strong" className="w-full">
            Node pools
          </Text>
          {pools.map((pool) => (
            <Button
              key={pool.name}
              variant="ghost"
              aria-pressed={pool.name === poolName}
              onClick={() => {
                setPoolName(pool.name)
                setSelectedGroup(undefined)
                setSelectedPod(undefined)
              }}
              className={cn(
                '!h-auto !min-h-12 !w-full !justify-between !gap-2 !p-3',
                poolName === pool.name &&
                  '!bg-primary-100 dark:!bg-primary-900/40'
              )}
            >
              <Text variant="subtext" family="mono">
                {pool.name}
              </Text>
              <Text variant="label" theme="neutral">
                {pool.nodes.length} nodes
              </Text>
            </Button>
          ))}
        </nav>
        <div className="flex min-w-0 flex-col gap-4">
          <SectionHeader
            title={<span className="font-mono">{poolName}</span>}
            description={
              <>
                {nodes.length} nodes · {nodes.length} zones ·{' '}
                {format(
                  nodes.reduce(
                    (sum, node) => sum + node.allocatable[dimension],
                    0
                  ),
                  dimension
                )}{' '}
                allocatable
              </>
            }
            actions={
              <ToggleButton<TDimension>
                label="Resource dimension"
                value={dimension}
                onChange={setDimension}
                options={dimensions}
              />
            }
          />
          <div className="flex flex-wrap items-center justify-between gap-2">
            <Text variant="subtext" theme="neutral">
              Block area ={' '}
              {dimension === 'storage'
                ? 'node-local ephemeral-storage'
                : dimension === 'cpu'
                  ? 'CPU'
                  : 'memory'}{' '}
              requests. Values marked “now” show current usage.
            </Text>
            {selectedGroup ? (
              <Button
                size="sm"
                onClick={() => {
                  setSelectedGroup(undefined)
                  setSelectedPod(undefined)
                }}
              >
                Show Deployments
              </Button>
            ) : null}
          </div>
          <div className="grid min-w-0 items-start gap-4 2xl:grid-cols-[minmax(0,1fr)_20rem]">
            <div
              className="min-w-0 overflow-x-auto pb-2"
              aria-label="Node allocations"
            >
              <div className="grid grid-cols-[repeat(auto-fit,minmax(16rem,1fr))] items-stretch gap-4">
                {nodes.map((node) => (
                  <DensityNode
                    key={node.name}
                    {...{
                      node,
                      dimension,
                      selectedGroup,
                      selectedPod,
                      onSelectGroup,
                      onSelectPod: setSelectedPod,
                    }}
                  />
                ))}
              </div>
            </div>
            <Card className="!gap-4 !p-4" aria-label="Density inspector">
              <div className="flex min-w-0 flex-col gap-1">
                <Text variant="subtext" theme="neutral">
                  {pod
                    ? 'Pod'
                    : selectedGroup
                      ? selectedGroup === daemonGroup
                        ? 'DaemonSet baseline'
                        : 'Deployment'
                      : 'Pool overview'}
                </Text>
                <Text family="mono" weight="strong" className="break-all">
                  {pod?.name ?? selectedGroup ?? poolName}
                </Text>
                <Text variant="subtext" theme="neutral">
                  {inspected.length} {inspected.length === 1 ? 'Pod' : 'Pods'}
                  {pod
                    ? ''
                    : ` across ${selectedGroup ? nodes.filter(({ pods }) => pods.some(({ group }) => group === selectedGroup)).length : nodes.length} nodes`}
                </Text>
              </div>
              <ResourceProfile pods={inspected} />
              {workload ? (
                <>
                  <hr />
                  <Text variant="subtext">{workloads[workload].scaling}</Text>
                  <Text variant="subtext" weight="strong">
                    {groupPods.length} running ·{' '}
                    {nodes.reduce(
                      (sum, node) => sum + replicaFit(node, workload),
                      0
                    )}{' '}
                    fit on current nodes
                  </Text>
                  <Text variant="label" theme="neutral">
                    Fit uses current template requests across CPU, memory, and
                    local storage, with other workloads held fixed. It is a
                    resource estimate, not a placement guarantee.
                  </Text>
                </>
              ) : null}
              <hr />
              <div className="flex flex-col gap-2">
                <Text variant="subtext" weight="strong">
                  {selectedGroup ? 'Pods' : 'Deployments'}
                </Text>
                {selectedGroup
                  ? groupPods.map((item) => (
                      <Button
                        key={item.name}
                        variant="ghost"
                        onClick={() =>
                          setSelectedPod(
                            item.name === selectedPod ? undefined : item.name
                          )
                        }
                        aria-pressed={item.name === selectedPod}
                        className="!h-auto !w-full !min-w-0 !justify-between !gap-2 !px-0 !py-1"
                      >
                        <Text
                          variant="label"
                          family="mono"
                          className="truncate"
                          title={item.name}
                        >
                          {item.name}
                        </Text>
                        <Text variant="label" className="shrink-0">
                          {format(item.usage[dimension], dimension)}
                          {item.usage[dimension] === undefined ? '' : ' now'}
                        </Text>
                      </Button>
                    ))
                  : groupNames
                      .filter((group) =>
                        allPods.some((pod) => pod.group === group)
                      )
                      .map((group) => (
                        <Button
                          key={group}
                          variant="ghost"
                          onClick={() => onSelectGroup(group)}
                          className="!h-auto !w-full !justify-between !gap-2 !px-0 !py-1"
                        >
                          <Text
                            variant="subtext"
                            family={group === daemonGroup ? 'sans' : 'mono'}
                          >
                            {group}
                          </Text>
                          <Text variant="label">
                            {
                              allPods.filter((pod) => pod.group === group)
                                .length
                            }{' '}
                            Pods
                          </Text>
                        </Button>
                      ))}
              </div>
              {dimension === 'storage' ? (
                <Text variant="label" theme="neutral">
                  Storage blocks show node-local requests, not PVC capacity.
                  Measured usage is shown only when available; incomplete totals
                  are unavailable.
                </Text>
              ) : null}
            </Card>
          </div>
          <div
            className="flex flex-wrap gap-x-4 gap-y-2"
            aria-label="Allocation legend"
          >
            {groupNames
              .filter((group) => allPods.some((pod) => pod.group === group))
              .map((group) => (
                <span key={group} className="flex items-center gap-2">
                  <span
                    className={cn(
                      'h-3 w-3 rounded-sm',
                      group === daemonGroup
                        ? daemonColor
                        : workloads[group].color
                    )}
                  />
                  <Text
                    variant="label"
                    family={group === daemonGroup ? 'sans' : 'mono'}
                  >
                    {group}
                  </Text>
                </span>
              ))}
          </div>
          <Text variant="label" theme="neutral">
            Workload values are totals across replicas on each node. Node
            reservations are already excluded from allocatable. Limits are not
            reserved capacity. All resources and metrics are fixtures; no
            cluster requests are made.
          </Text>
        </div>
      </div>
    </div>
  )
}

export const Overview = () => <NodeDensityStory />
export const CPU = () => <NodeDensityStory initialDimension="cpu" />
export const DeploymentPods = () => <NodeDensityStory initialGroup="acme-web" />
export const Storage = () => <NodeDensityStory initialDimension="storage" />
export const StorageUsageUnavailable = () => (
  <NodeDensityStory initialDimension="storage" storageUnavailable />
)
