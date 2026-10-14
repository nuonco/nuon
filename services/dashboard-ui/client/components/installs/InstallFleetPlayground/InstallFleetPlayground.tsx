import { useEffect, useRef, useState } from 'react'
import type { Node, NodeProps } from '@xyflow/react'
import { GraphCanvas } from '@/components/branches/graph/GraphCanvas'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { CheckboxFilterDropdown } from '@/components/common/CheckboxFilterDropdown'
import { EmptyState } from '@/components/common/EmptyState'
import { RadioFilterDropdown } from '@/components/common/RadioFilterDropdown'
import { SearchInput } from '@/components/common/SearchInput'
import { Text } from '@/components/common/Text'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { useSurfaces } from '@/hooks/use-surfaces'
import { Modal } from '@/components/surfaces/Modal'
import { humanize } from '@/utils/string-utils'
import { cn } from '@/utils/classnames'
import {
  groupFleetInstalls,
  installResources,
  labelOptions,
  labelPairs,
  layoutFleet,
  layoutInstallResources,
  matchesSearch,
  passesFilters,
  trackingLabel,
  type TFleetResource,
  type TFleetInstall,
  type TFleetRunnerFilter,
} from './fleet'

type TInstallNode = Node<
  TFleetInstall & {
    matching: boolean
    selected: boolean
    thisBranch: boolean
    onInspect: () => void
  }
>

const Field = ({ label, value }: { label: string; value: string }) => (
  <div className="flex w-full items-start gap-3">
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
      {value}
    </Text>
  </div>
)

const InstallNode = ({ data }: NodeProps<TInstallNode>) => {
  const accentClassName =
    data.health === 'healthy'
      ? 'bg-green-500/70 dark:bg-green-500/60'
      : data.health === 'unhealthy'
        ? 'bg-red-500/70 dark:bg-red-500/60'
        : data.health === 'progressing' || data.health === 'degraded'
          ? 'bg-orange-500/70 dark:bg-orange-500/60'
          : undefined
  const runner = data.hasRunner ? humanize(data.runnerStatus) : 'No runner'
  return (
    <Button
      variant="ghost"
      className={cn(
        'nodrag nopan relative !h-52 !w-64 !p-4 !border flex-col !items-start !gap-3 text-left !bg-elevation-2 hover:!bg-elevation-3',
        data.selected &&
          '![--tw-outline-style:solid] !outline-1 !outline-primary-500'
      )}
      aria-label={`Inspect ${data.name}`}
      aria-pressed={data.selected}
      onClick={data.onInspect}
    >
      {accentClassName ? (
        <span
          role="img"
          aria-label={`Health ${data.health}`}
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
        <Text
          family="mono"
          weight="strong"
          nowrap
          className="min-w-0 truncate"
          title={data.name}
        >
          {data.name}
        </Text>
        {data.thisBranch ? (
          <Badge size="sm" theme="brand">
            This branch
          </Badge>
        ) : null}
      </div>
      <Text
        family="mono"
        variant="subtext"
        theme="neutral"
        nowrap
        className="w-full truncate"
        title={data.branchName || 'Not tracking'}
      >
        {data.branchName || 'Not tracking'}
      </Text>
      <div className="flex w-full flex-auto flex-col gap-1 min-h-0">
        <Field label="Lifecycle" value={humanize(data.lifecycle)} />
        <Field label="Runner" value={runner} />
        <Field label="Tracking" value={trackingLabel(data)} />
      </div>
    </Button>
  )
}

const SectionNode = ({
  data,
}: NodeProps<Node<{ title: string; count: number }>>) => (
  <div className="flex w-full items-center gap-3">
    <Text variant="subtext" weight="strong" nowrap className="truncate">
      {data.title}
    </Text>
    <Text variant="label" theme="neutral">
      {data.count}
    </Text>
  </div>
)

const nodeTypes = {
  install: InstallNode,
  section: SectionNode,
}
const noEdges: [] = []

type TResourceNode = Node<TFleetResource>

const resourceAccent = (health?: TFleetInstall['health']) =>
  health === 'healthy'
    ? 'bg-green-500/70 dark:bg-green-500/60'
    : health === 'unhealthy'
      ? 'bg-red-500/70 dark:bg-red-500/60'
      : health === 'progressing' || health === 'degraded'
        ? 'bg-orange-500/70 dark:bg-orange-500/60'
        : undefined

const ResourceNode = ({ data }: NodeProps<TResourceNode>) => {
  const accentClassName = resourceAccent(data.health)
  return (
    <div className="relative flex h-24 w-64 flex-col justify-center gap-1 rounded-lg border bg-elevation-2 p-3">
      {accentClassName ? (
        <span
          role="img"
          aria-label={`Health ${data.health}`}
          className={cn(
            'absolute inset-x-3 top-0 h-1 rounded-b',
            accentClassName
          )}
        />
      ) : null}
      <Text family="mono" variant="label" theme="neutral">
        {data.kind}
      </Text>
      <Text family="mono" weight="strong" nowrap className="truncate">
        {data.name}
      </Text>
      <Text variant="subtext" theme="neutral">
        {data.health ? humanize(data.health) : 'Not reported'}
      </Text>
    </div>
  )
}

const resourceNodeTypes = { resource: ResourceNode }

const trackingValue = (install: TFleetInstall) =>
  install.commitsBehind === undefined
    ? 'Not reported'
    : install.commitsBehind === 0
      ? 'Up to date'
      : install.commitsBehind === 1
        ? '1 commit behind'
        : `${install.commitsBehind} commits behind`

const driftValue = (install: TFleetInstall) =>
  install.configDriftCount === 0
    ? 'None'
    : install.configDriftCount === 1
      ? '1 drifted'
      : `${install.configDriftCount} drifted`

const InstallModal = ({
  install,
  thisBranch,
}: {
  install: TFleetInstall
  thisBranch: boolean
}) => {
  const resources = installResources(install)
  const labels = labelPairs(install)
  const laidOut = layoutInstallResources(resources)
  const nodes: Node[] = laidOut.map(({ resource, x, y }) => ({
    id: resource.id,
    type: 'resource',
    position: { x, y },
    width: 256,
    height: 96,
    draggable: false,
    focusable: false,
    data: resource,
  }))
  return (
    <div className="flex min-h-[28rem] flex-col gap-4 md:flex-row">
      <div className="flex w-full shrink-0 flex-col gap-3 md:w-72">
        <div className="flex flex-wrap items-center gap-2">
          <Text family="mono" variant="subtext" theme="neutral">
            {install.branchName || 'Not tracking'}
          </Text>
          {thisBranch ? (
            <Badge size="sm" theme="brand">
              This branch
            </Badge>
          ) : null}
        </div>
        <Text variant="label" theme="neutral">
          Statuses
        </Text>
        <Field
          label="Health"
          value={install.health ? humanize(install.health) : 'Not reported'}
        />
        <Field
          label="Components"
          value={
            install.componentStatus
              ? humanize(install.componentStatus)
              : 'Not reported'
          }
        />
        <Field label="Lifecycle" value={humanize(install.lifecycle)} />
        <Field
          label="Runner"
          value={
            install.hasRunner ? humanize(install.runnerStatus) : 'No runner'
          }
        />
        <Field label="Tracking" value={trackingValue(install)} />
        <Field label="Config drift" value={driftValue(install)} />
        <Text variant="label" theme="neutral">
          Labels
        </Text>
        {labels.length ? (
          <div className="flex flex-wrap gap-2">
            {labels.map((label) => (
              <Badge key={label} size="sm" variant="code">
                {label}
              </Badge>
            ))}
          </div>
        ) : (
          <Text variant="subtext" theme="neutral">
            No labels
          </Text>
        )}
      </div>
      <div className="min-w-0 flex-1">
        <Text weight="strong" className="mb-2">
          Resources
        </Text>
        <GraphCanvas
          nodes={nodes}
          edges={noEdges}
          nodeTypes={resourceNodeTypes}
          height={Math.max(280, resources.length * 112 + 24)}
          minZoom={0.6}
          maxZoom={1.5}
          fitPadding={0.16}
          alignStart
        />
      </div>
    </div>
  )
}

export const InstallFleetPlayground = ({
  installs,
  currentBranch,
}: {
  installs: TFleetInstall[]
  currentBranch?: string
}) => {
  const [labels, setLabels] = useState<Set<string>>(new Set())
  const [runner, setRunner] = useState<TFleetRunnerFilter>()
  const [needsUpdate, setNeedsUpdate] = useState(false)
  const [query, setQuery] = useState('')
  const [selectedId, setSelectedId] = useState<string>()
  const { addModal, removeModal } = useSurfaces()
  const modalId = useRef<string | undefined>(undefined)

  const filters = { labels, runner, needsUpdate }
  const visible = installs.filter((install) => passesFilters(install, filters))
  const behindCount = installs.filter(
    (install) => install.commitsBehind && install.commitsBehind > 0
  ).length
  const hasActiveFilters =
    labels.size > 0 || !!runner || needsUpdate || query.trim() !== ''
  const needle = query.trim()
  const matchingIds = new Set(
    needle
      ? visible
          .filter((install) => matchesSearch(install, needle))
          .map((install) => install.id)
      : []
  )
  const visibleKey = visible.map((install) => install.id).join('\0')
  const options = labelOptions(installs)
  const sections = groupFleetInstalls(visible, currentBranch)
  const layout = layoutFleet(sections)

  useEffect(() => {
    if (selectedId && !visibleKey.split('\0').includes(selectedId)) {
      setSelectedId(undefined)
      if (modalId.current) {
        removeModal(modalId.current)
        modalId.current = undefined
      }
    }
  }, [selectedId, visibleKey, removeModal])

  const openInstall = (install: TFleetInstall) => {
    if (modalId.current) removeModal(modalId.current)
    setSelectedId(install.id)
    modalId.current = addModal(
      <Modal
        size="full"
        showFooter={false}
        heading={install.name}
        onClose={() => setSelectedId(undefined)}
      >
        <InstallModal
          install={install}
          thisBranch={!!currentBranch && install.branchName === currentBranch}
        />
      </Modal>
    )
  }

  const clearFilters = () => {
    setLabels(new Set())
    setRunner(undefined)
    setNeedsUpdate(false)
    setQuery('')
  }

  const nodes: Node[] = []
  for (const section of layout) {
    nodes.push({
      id: `section-${section.id}`,
      type: 'section',
      position: { x: section.x, y: section.y },
      width: section.width,
      height: 24,
      draggable: false,
      selectable: false,
      focusable: false,
      data: { title: section.title, count: section.count },
    })
    for (const card of section.installs) {
      nodes.push({
        id: card.install.id,
        type: 'install',
        position: { x: card.x, y: card.y },
        width: 256,
        height: 208,
        draggable: false,
        focusable: false,
        data: {
          ...card.install,
          matching: !!needle && matchingIds.has(card.install.id),
          selected: selectedId === card.install.id,
          thisBranch:
            !!currentBranch && card.install.branchName === currentBranch,
          onInspect: () => openInstall(card.install),
        },
      })
    }
  }

  return (
    <div className="flex h-full flex-col gap-3 p-6">
      <SectionHeader
        variant="page"
        title="Fleet"
        description="Installs in this app across branches, grouped by the branch each one tracks."
      />
      <div className="flex min-w-0 flex-wrap items-center gap-2">
        <SearchInput
          aria-label="Search installs"
          placeholder="Search by name, branch, label, or status"
          value={query}
          onChange={setQuery}
          onClear={() => setQuery('')}
        />
        {options.length ? (
          <CheckboxFilterDropdown
            id="fleet-filter-labels"
            label="Labels"
            options={options.map((value) => ({ value, label: value }))}
            selected={labels}
            onChange={setLabels}
          />
        ) : null}
        <RadioFilterDropdown
          id="fleet-filter-runner"
          label="Runner"
          options={[
            { value: 'active', label: 'Runner active' },
            { value: 'inactive', label: 'Runner not active' },
          ]}
          selected={runner}
          onChange={(value) =>
            setRunner(
              value === 'active' || value === 'inactive' ? value : undefined
            )
          }
        />
        <Button
          variant="secondary"
          aria-pressed={needsUpdate}
          disabled={behindCount === 0}
          tooltipProps={
            behindCount === 0
              ? {
                  position: 'bottom',
                  tipContent: 'No installs are behind their tracked branch.',
                }
              : undefined
          }
          onClick={() => setNeedsUpdate((current) => !current)}
          className={cn(
            needsUpdate &&
              '!bg-primary-200 dark:!bg-primary-600/25 !text-primary-800 dark:!text-primary-400 !font-stronger !shadow-none'
          )}
        >
          Needs updating{behindCount ? ` · ${behindCount}` : ''}
        </Button>
        {hasActiveFilters ? (
          <Button variant="ghost" onClick={clearFilters}>
            Clear filters
          </Button>
        ) : null}
        <Text
          variant="subtext"
          theme="neutral"
          className="ml-auto"
          aria-live="polite"
        >
          {needle
            ? `${matchingIds.size === 1 ? '1 match' : `${matchingIds.size} matches`} · `
            : ''}
          {visible.length} of {installs.length}{' '}
          {installs.length === 1 ? 'install' : 'installs'}
        </Text>
      </div>
      {!installs.length ? (
        <EmptyState
          variant="app"
          emptyTitle="No installs yet"
          emptyMessage="Installs will appear here once this app has deployments."
        />
      ) : !visible.length ? (
        <EmptyState
          variant="search"
          emptyTitle="No installs match these filters"
          emptyMessage="Clear the filters to see every install in the app."
          action={
            <Button variant="secondary" onClick={clearFilters}>
              Clear filters
            </Button>
          }
        />
      ) : (
        <div className="min-w-0 flex-1">
          <GraphCanvas
            nodes={nodes}
            edges={noEdges}
            nodeTypes={nodeTypes}
            height={720}
            minZoom={0.35}
            maxZoom={1.5}
            fitPadding={0.12}
            style={{ background: 'var(--elevation-1)' }}
          />
          <Text variant="label" theme="neutral" className="mt-2">
            Pan or zoom the canvas. Select an install to inspect it.
          </Text>
        </div>
      )}
    </div>
  )
}
