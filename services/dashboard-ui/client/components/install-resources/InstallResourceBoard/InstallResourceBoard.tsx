import { useEffect, useRef, useState, type ReactNode } from 'react'
import { useSearchParams } from 'react-router'
import type { Node, NodeProps } from '@xyflow/react'
import { GraphCanvas } from '@/components/branches/graph/GraphCanvas'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { CheckboxFilterDropdown } from '@/components/common/CheckboxFilterDropdown'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { SearchInput } from '@/components/common/SearchInput'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { InstallResourceDetailPanelComponent } from '@/components/install-resources/InstallResourceDetailPanel'
import { resourceStatusFields } from '@/components/install-resources/InstallResourceDetailPanel/resource-details'
import { ResourceKind } from '@/components/install-resources/ResourceKind'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { TInstallResource } from '@/types'
import { cn } from '@/utils/classnames'
import {
  HEALTH_DEGRADED,
  HEALTH_HEALTHY,
  HEALTH_NOT_APPLICABLE,
  HEALTH_PROGRESSING,
  HEALTH_UNHEALTHY,
  HEALTH_UNKNOWN,
  isIdentityOnlyResource,
  reportedHealth,
  resourceIsStale,
  worstHealth,
} from '@/utils/health-utils'
import { layoutResourceKinds, type TResourceBoardItem } from './resource-layout'

export type { TResourceBoardItem } from './resource-layout'

type TResourceNode = Node<
  TResourceBoardItem & {
    matching: boolean
    selected: boolean
    stale: boolean
    onInspect: () => void
  }
>

const ATTENTION_BORDER: Record<string, string> = {
  [HEALTH_UNHEALTHY]: 'bg-red-600 dark:bg-red-500',
  [HEALTH_DEGRADED]: 'bg-orange-600 dark:bg-orange-500',
  [HEALTH_PROGRESSING]: 'bg-blue-600 dark:bg-blue-500',
}

const HEALTH_FILTER_OPTIONS = [
  { value: HEALTH_UNHEALTHY, label: 'Unhealthy' },
  { value: HEALTH_DEGRADED, label: 'Degraded' },
  { value: HEALTH_PROGRESSING, label: 'Progressing' },
  { value: HEALTH_HEALTHY, label: 'Healthy' },
  { value: HEALTH_UNKNOWN, label: 'Unknown' },
  { value: HEALTH_NOT_APPLICABLE, label: 'Not applicable' },
]

const groupAttentionHealth = (resources: TResourceBoardItem[]) => {
  const health = worstHealth(
    resources.map((item) =>
      reportedHealth(item.resource, resourceIsStale(item.resource))
    )
  )
  return ATTENTION_BORDER[health] ? health : undefined
}

const ResourceNode = ({ data }: NodeProps<TResourceNode>) => {
  const fields = resourceStatusFields(data.resource)
  const health = reportedHealth(data.resource, data.stale)
  const attentionBorder = health ? ATTENTION_BORDER[health] : undefined
  return (
    <Button
      variant="ghost"
      className={cn(
        'nodrag nopan relative overflow-hidden !h-52 !w-64 !p-4 !border flex-col !items-start !gap-3 text-left !bg-elevation-2 hover:!bg-elevation-3',
        data.matching &&
          '!bg-primary-500/15 hover:!bg-primary-500/20 ![--tw-outline-style:solid] !outline-1 !outline-primary-500',
        data.selected &&
          '![--tw-outline-style:solid] !outline-1 !outline-primary-500'
      )}
      aria-label={`Inspect ${data.resource.kind} ${data.resource.name}`}
      aria-pressed={data.selected}
      onClick={data.onInspect}
    >
      {attentionBorder ? (
        <span
          aria-hidden
          className={cn('absolute inset-x-0 top-0 h-1', attentionBorder)}
        />
      ) : null}
      {health ? <Status status={health} variant="badge" /> : null}
      <Text
        family="mono"
        weight="strong"
        nowrap
        className="w-full truncate"
        title={data.resource.name}
      >
        {data.resource.name}
      </Text>
      <div className="flex w-full items-center justify-between gap-2">
        <ResourceKind resource={data.resource} compact />
        <div className="flex items-center gap-2">
          {data.selected ? (
            <Text variant="label" theme="brand">
              Selected
            </Text>
          ) : null}
          {data.stale ? (
            <Badge size="sm" theme="warn">
              Stale
            </Badge>
          ) : null}
        </div>
      </div>
      <div className="flex w-full flex-auto flex-col gap-1 min-h-0">
        {fields.length ? (
          fields.slice(0, 3).map(([label, value], index) => (
            <div
              key={`${label}/${index}`}
              className="flex w-full items-start gap-3"
            >
              <Text
                variant="subtext"
                theme="neutral"
                nowrap
                className="min-w-0 flex-auto truncate"
                title={label}
              >
                {label}
              </Text>
              <Text
                family="mono"
                variant="subtext"
                nowrap
                className="min-w-0 max-w-2/3 shrink-0 truncate"
                title={value}
              >
                {/BackOff|ErrImagePull|OOMKilled|^Failed$|^ExitCode:[1-9]/.test(
                  value
                ) ? (
                  <Icon
                    variant="WarningIcon"
                    theme="warn"
                    size={12}
                    className="inline mr-1"
                  />
                ) : null}
                {value}
              </Text>
            </div>
          ))
        ) : (
          <Text variant="subtext" theme="neutral">
            {isIdentityOnlyResource(data.resource)
              ? 'Identity snapshot'
              : 'Status not reported'}
          </Text>
        )}
      </div>
      <Text
        variant="label"
        theme="neutral"
        nowrap
        className="w-full truncate"
        title={data.summary}
      >
        {fields.length > 3 ? `+${fields.length - 3} fields · ` : ''}
        {data.summary}
      </Text>
    </Button>
  )
}

const KindNode = ({
  data,
}: NodeProps<
  Node<{
    resource: TInstallResource
    count: number
    showGroup: boolean
    health?: string
  }>
>) => (
  <div className="flex w-64 flex-col gap-1">
    <div className="flex items-center justify-between gap-2">
      <span className="flex min-w-0 items-center gap-2">
        <Text
          family="mono"
          variant="subtext"
          weight="strong"
          nowrap
          className="min-w-0 truncate"
          title={`${data.resource.kind} · ${data.resource.api_group || 'core'}`}
        >
          {data.resource.kind === 'HorizontalPodAutoscaler'
            ? 'HPA'
            : data.resource.kind === 'PersistentVolumeClaim'
              ? 'PVC'
              : data.resource.kind}
        </Text>
        {data.health ? <Status status={data.health} /> : null}
      </span>
      <Text variant="label" theme="neutral" className="shrink-0">
        {data.count}
      </Text>
    </div>
    {data.showGroup ? (
      <Text
        family="mono"
        variant="label"
        theme="neutral"
        nowrap
        className="truncate"
        title={data.resource.api_group || data.resource.provider}
      >
        {data.resource.api_group || data.resource.provider}
      </Text>
    ) : null}
  </div>
)

const SectionNode = ({
  data,
}: NodeProps<Node<{ title: string; count: number; secondary: boolean }>>) => (
  <div
    className={cn(
      'flex w-full items-center gap-3',
      data.secondary && 'border-t pt-6'
    )}
  >
    <Text variant="subtext" weight="strong">
      {data.title}
    </Text>
    <Text variant="label" theme="neutral">
      {data.count}
    </Text>
  </div>
)

const nodeTypes = {
  resource: ResourceNode,
  kind: KindNode,
  section: SectionNode,
}
const noEdges = []

export const InstallResourceBoard = ({
  items,
  filters,
  toggles,
}: {
  items: TResourceBoardItem[]
  filters?: ReactNode
  toggles?: ReactNode
}) => {
  const { panels, addPanel, updatePanel } = useSurfaces()
  const panelId = useRef<string>()
  const [params, setParams] = useSearchParams()
  const search = params.get('q') || ''
  const legacyKind = params.get('kind') || ''
  const selectedKinds = new Set(params.getAll('resource_kind'))
  const selectedHealth = new Set(params.getAll('resource_health'))
  const [selected, setSelected] = useState<string>()
  const panelVisible = panels.some(
    (panel) => panel.id === panelId.current && panel.isVisible
  )
  const inspected = items.find((item) => item.id === selected)?.resource
  useEffect(() => {
    if (panelId.current && inspected && panelVisible) {
      updatePanel(
        panelId.current,
        <InstallResourceDetailPanelComponent installResource={inspected} />
      )
    }
  }, [inspected, panelVisible, updatePanel])

  const healthScoped = selectedHealth.size
    ? items.filter((item) => selectedHealth.has(item.resource.health || ''))
    : items
  const byKind = new Map<string, TResourceBoardItem[]>()
  for (const item of healthScoped) {
    const key = `${item.resource.provider || ''}/${item.resource.api_group || ''}/${item.resource.kind || ''}`
    byKind.set(key, [...(byKind.get(key) || []), item])
  }
  const kinds = [...byKind.entries()].sort(([a], [b]) => a.localeCompare(b))
  if (legacyKind)
    kinds
      .filter(([, resources]) => resources[0].resource.kind === legacyKind)
      .forEach(([key]) => selectedKinds.add(key))
  const visibleKinds = kinds.filter(
    ([key]) => (!legacyKind && !selectedKinds.size) || selectedKinds.has(key)
  )
  const visibleItems = visibleKinds.flatMap(([, resources]) => resources)
  const needle = search.trim().toLowerCase()
  const matching = visibleItems.filter(({ resource }) =>
    [
      resource.name,
      resource.kind,
      resource.namespace,
      resource.api_group,
      ...resourceStatusFields(resource).flat(),
    ].some((value) => value?.toLowerCase().includes(needle))
  )
  const matchingIds = new Set(matching.map(({ id }) => id))
  const filtering = !!needle
  const layout = layoutResourceKinds(visibleKinds)
  const nodes: Node[] = layout.sections.map((section) => ({
    id: `section-${section.id}`,
    type: 'section',
    position: { x: section.x, y: section.y },
    width: section.width,
    height: section.secondary ? 48 : 24,
    draggable: false,
    selectable: false,
    focusable: false,
    data: section,
  }))

  for (const {
    key,
    resources,
    x,
    y,
    headerHeight,
    additional,
  } of layout.groups) {
    nodes.push({
      id: `kind-${key}`,
      type: 'kind',
      position: { x, y },
      width: 256,
      height: headerHeight,
      draggable: false,
      selectable: false,
      focusable: false,
      data: {
        resource: resources[0].resource,
        count: resources.length,
        showGroup: additional,
        health: groupAttentionHealth(resources),
      },
    })
    resources.forEach((item, index) =>
      nodes.push({
        id: item.id,
        type: 'resource',
        position: { x, y: y + headerHeight + 16 + index * 232 },
        width: 256,
        height: 208,
        draggable: false,
        focusable: false,
        data: {
          ...item,
          matching: filtering && matchingIds.has(item.id),
          stale: resourceIsStale(item.resource),
          selected: panelVisible && selected === item.id,
          onInspect: () => {
            setSelected(item.id)
            panelId.current = addPanel(
              <InstallResourceDetailPanelComponent
                installResource={item.resource}
              />
            )
          },
        },
      })
    )
  }

  const setKinds = (values: Set<string>) =>
    setParams(
      (previous) => {
        previous.delete('kind')
        previous.delete('resource_kind')
        values.forEach((value) => previous.append('resource_kind', value))
        return previous
      },
      { replace: true }
    )
  const setHealth = (values: Set<string>) =>
    setParams(
      (previous) => {
        previous.delete('resource_health')
        values.forEach((value) => previous.append('resource_health', value))
        return previous
      },
      { replace: true }
    )
  const clearFilters = () =>
    setParams(
      (previous) => {
        previous.delete('q')
        previous.delete('kind')
        previous.delete('resource_kind')
        previous.delete('resource_health')
        return previous
      },
      { replace: true }
    )

  return (
    <div className="flex flex-col gap-3">
      <div className="flex min-w-0 flex-wrap items-center gap-2">
        <SearchInput
          aria-label="Find resources in canvas"
          placeholder="Find by name, kind, namespace, or status"
          value={search}
          onChange={(value) =>
            setParams(
              (previous) => {
                if (value) previous.set('q', value)
                else previous.delete('q')
                return previous
              },
              { replace: true }
            )
          }
        />
        <Text
          variant="subtext"
          theme="neutral"
          aria-live="polite"
          className="w-20 shrink-0 tabular-nums"
        >
          {filtering ? `${matching.length} of ${visibleItems.length}` : ''}
        </Text>
        {filters}
        <CheckboxFilterDropdown
          id="resource-health"
          label="Health"
          options={HEALTH_FILTER_OPTIONS}
          selected={selectedHealth}
          onChange={setHealth}
        />
        <CheckboxFilterDropdown
          id="resource-kinds"
          label="Kinds"
          options={kinds.map(([key, resources]) => ({
            value: key,
            label: `${resources[0].resource.kind} · ${resources[0].resource.api_group || 'core'} (${resources.length})`,
          }))}
          selected={selectedKinds}
          onChange={setKinds}
        />
        {toggles}
      </div>
      <div className="flex flex-wrap items-center justify-between gap-2">
        <Text variant="subtext" theme="neutral">
          Grouped by function. Connections are not shown.
        </Text>
        <Text variant="subtext" theme="neutral" aria-live="polite">
          {visibleItems.length} of {items.length} reported resources
        </Text>
      </div>
      <div className="relative">
        <GraphCanvas
          nodes={nodes}
          edges={noEdges}
          nodeTypes={nodeTypes}
          height={640}
          minZoom={0.1}
          maxZoom={1.5}
          fitPadding={0.08}
          fitMinZoom={0.8}
          alignStart
          style={{ background: 'var(--elevation-1)' }}
        />
        {visibleKinds.length ? null : (
          <div className="pointer-events-none absolute inset-0 flex items-center justify-center">
            <EmptyState
              variant="search"
              emptyTitle="No resources match these filters"
              emptyMessage="Clear the filters to see every reported resource."
              action={
                <Button
                  className="pointer-events-auto"
                  variant="secondary"
                  onClick={clearFilters}
                >
                  Clear filters
                </Button>
              }
            />
          </div>
        )}
      </div>
      <Text variant="label" theme="neutral">
        Pan or zoom the canvas. Select a resource to inspect its latest report.
      </Text>
    </div>
  )
}
