import { useEffect, useRef, useState } from 'react'
import { useSearchParams } from 'react-router'
import type { Node, NodeProps } from '@xyflow/react'
import { GraphCanvas } from '@/components/branches/graph/GraphCanvas'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { EmptyState } from '@/components/common/EmptyState'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { Icon } from '@/components/common/Icon'
import { SearchInput } from '@/components/common/SearchInput'
import { Text } from '@/components/common/Text'
import { InstallResourceDetailPanelComponent } from '@/components/install-resources/InstallResourceDetailPanel'
import { resourceStatusFields } from '@/components/install-resources/InstallResourceDetailPanel/resource-details'
import { ResourceKind } from '@/components/install-resources/ResourceKind'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { TInstallResource } from '@/types'
import { cn } from '@/utils/classnames'
import { isIdentityOnlyResource } from '@/utils/health-utils'
import { isStaleObservation } from '@/utils/time-utils'
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

const ResourceNode = ({ data }: NodeProps<TResourceNode>) => {
  const fields = resourceStatusFields(data.resource)
  const health =
    !data.stale &&
    !data.resource.removed_from_config &&
    !isIdentityOnlyResource(data.resource)
      ? data.resource.health
      : undefined
  const accentClassName =
    health === 'healthy'
      ? 'bg-green-500/70 dark:bg-green-500/60'
      : health === 'unhealthy'
        ? 'bg-red-500/70 dark:bg-red-500/60'
        : health === 'progressing' || health === 'degraded'
          ? 'bg-orange-500/70 dark:bg-orange-500/60'
          : undefined
  return (
    <Button
      variant="ghost"
      className={cn(
        'nodrag nopan relative !h-52 !w-64 !p-4 !border flex-col !items-start !gap-3 text-left !bg-elevation-2 hover:!bg-elevation-3',
        data.selected &&
          '![--tw-outline-style:solid] !outline-1 !outline-primary-500'
      )}
      aria-label={`Inspect ${data.resource.kind} ${data.resource.name}`}
      aria-pressed={data.selected}
      onClick={data.onInspect}
    >
      {accentClassName ? (
        <span
          role="img"
          aria-label={`Last reported ${health}`}
          className={cn(
            'absolute inset-x-4 top-0 h-1 rounded-b',
            accentClassName
          )}
        />
      ) : null}
      {data.matching && !data.selected ? (
        <span
          className="absolute left-0 top-0 h-2 w-2 rounded-br bg-primary-500"
          aria-label="Search match"
        />
      ) : null}
      <div className="flex w-full items-center justify-between gap-2">
        <ResourceKind resource={data.resource} compact />
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
      <Text
        family="mono"
        weight="strong"
        nowrap
        className="w-full truncate"
        title={data.resource.name}
      >
        {data.resource.name}
      </Text>
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
  Node<{ resource: TInstallResource; count: number; showGroup: boolean }>
>) => (
  <div className="flex w-64 flex-col gap-1">
    <div className="flex items-center justify-between gap-2">
      <Text
        family="mono"
        variant="subtext"
        weight="strong"
        nowrap
        className="truncate"
        title={`${data.resource.kind} · ${data.resource.api_group || 'core'}`}
      >
        {data.resource.kind === 'HorizontalPodAutoscaler'
          ? 'HPA'
          : data.resource.kind === 'PersistentVolumeClaim'
            ? 'PVC'
            : data.resource.kind}
      </Text>
      <Text variant="label" theme="neutral">
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
}: {
  items: TResourceBoardItem[]
}) => {
  const { panels, addPanel, updatePanel } = useSurfaces()
  const panelId = useRef<string>()
  const [params, setParams] = useSearchParams()
  const search = params.get('q') || ''
  const legacyKind = params.get('kind') || ''
  const selectedKinds = new Set(params.getAll('resource_kind'))
  const [kindSearch, setKindSearch] = useState('')
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

  const byKind = new Map<string, TResourceBoardItem[]>()
  for (const item of items) {
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
          stale:
            !isIdentityOnlyResource(item.resource) &&
            isStaleObservation(
              item.resource.observed_at,
              item.resource.stale_after_seconds || undefined
            ),
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
  const toggleKind = (key: string) => {
    const next = new Set(selectedKinds)
    if (next.has(key)) next.delete(key)
    else next.add(key)
    setKinds(next)
  }

  return (
    <div className="flex flex-col gap-3">
      <div className="flex flex-wrap items-center gap-3">
        <Dropdown
          id="resource-kinds"
          closeOnBlur={false}
          buttonText={
            <>
              <Icon variant="FunnelIcon" size={14} />
              Kinds · {selectedKinds.size || 'All'}
            </>
          }
        >
          <div className="flex w-80 flex-col gap-2 p-3">
            <SearchInput
              aria-label="Find a kind"
              placeholder="Find a kind or API group"
              className="w-full !min-w-0"
              labelClassName="w-full"
              value={kindSearch}
              onChange={setKindSearch}
            />
            <div
              className="max-h-72 overflow-y-auto"
              role="group"
              aria-label="Reported resource kinds"
            >
              {kinds
                .filter(([key]) =>
                  key.toLowerCase().includes(kindSearch.toLowerCase())
                )
                .map(([key, resources]) => (
                  <CheckboxInput
                    key={key}
                    aria-label={`${resources[0].resource.kind} ${resources[0].resource.api_group || 'core'} (${resources.length})`}
                    checked={selectedKinds.has(key)}
                    onChange={() => toggleKind(key)}
                    labelProps={{
                      labelTextProps: { className: 'w-full' },
                      labelText: (
                        <span className="flex items-center justify-between gap-2">
                          <span className="flex min-w-0 flex-col gap-1">
                            <Text
                              family="mono"
                              variant="subtext"
                              className="break-words"
                            >
                              {resources[0].resource.kind}
                            </Text>
                            <Text family="mono" variant="label" theme="neutral">
                              {resources[0].resource.api_group || 'core'}
                            </Text>
                          </span>
                          <Text variant="label" theme="neutral">
                            {resources.length}
                          </Text>
                        </span>
                      ),
                    }}
                  />
                ))}
              {!kinds.some(([key]) =>
                key.toLowerCase().includes(kindSearch.toLowerCase())
              ) ? (
                <Text variant="subtext" theme="neutral" className="p-2">
                  No reported kinds match.
                </Text>
              ) : null}
            </div>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => {
                setKinds(new Set())
                setKindSearch('')
              }}
            >
              All kinds
            </Button>
          </div>
        </Dropdown>
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
      </div>
      {selectedKinds.size || legacyKind ? (
        <div
          className="flex flex-wrap items-center gap-2"
          aria-label="Selected kinds"
        >
          {kinds
            .filter(([key]) => selectedKinds.has(key))
            .map(([key, resources]) => (
              <Button
                key={key}
                size="sm"
                onClick={() => toggleKind(key)}
                aria-label={`Remove ${resources[0].resource.kind} ${resources[0].resource.api_group || 'core'} filter`}
              >
                <Text variant="label" family="mono">
                  {resources[0].resource.kind} ·{' '}
                  {resources[0].resource.api_group || 'core'}
                </Text>
                <Icon variant="XIcon" size={12} />
              </Button>
            ))}
          <Button variant="ghost" size="sm" onClick={() => setKinds(new Set())}>
            Clear filters
          </Button>
        </div>
      ) : null}
      <div className="flex flex-wrap items-center justify-between gap-2">
        <Text variant="subtext" theme="neutral">
          Grouped by function. Connections are not shown.
        </Text>
        <Text variant="subtext" theme="neutral" aria-live="polite">
          {visibleItems.length} of {items.length} reported resources
        </Text>
      </div>
      {filtering ? (
        <Text variant="subtext" theme="neutral" aria-live="polite">
          {matching.length} matching{' '}
          {matching.length === 1 ? 'resource' : 'resources'}. Other cards stay
          visible for context.
        </Text>
      ) : null}
      {visibleKinds.length ? (
        <GraphCanvas
          nodes={nodes}
          edges={noEdges}
          nodeTypes={nodeTypes}
          height={640}
          minZoom={0.8}
          maxZoom={1.5}
          fitPadding={0.08}
          alignStart
          style={{ background: 'var(--elevation-1)' }}
        />
      ) : (
        <EmptyState
          emptyTitle="No resources match these filters"
          emptyMessage="Clear the kind filters to browse reported resources."
        />
      )}
      <Text variant="label" theme="neutral">
        Pan or zoom the canvas. Select a resource to inspect its latest report.
      </Text>
    </div>
  )
}
