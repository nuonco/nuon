import { useEffect, useState } from 'react'
import type { Node, NodeProps } from '@xyflow/react'
import { GraphCanvas } from '@/components/branches/graph/GraphCanvas'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { CheckboxFilterDropdown } from '@/components/common/CheckboxFilterDropdown'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { RadioFilterDropdown } from '@/components/common/RadioFilterDropdown'
import { SearchInput } from '@/components/common/SearchInput'
import { Text } from '@/components/common/Text'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { humanize } from '@/utils/string-utils'
import { cn } from '@/utils/classnames'
import {
  groupFleetInstalls,
  labelOptions,
  labelPairs,
  layoutFleet,
  matchesSearch,
  passesFilters,
  trackingLabel,
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

const Detail = ({
  install,
  thisBranch,
  onClose,
}: {
  install: TFleetInstall
  thisBranch: boolean
  onClose: () => void
}) => {
  const labels = labelPairs(install)
  const behind =
    install.commitsBehind === undefined
      ? 'Not reported'
      : install.commitsBehind === 0
        ? 'Up to date'
        : install.commitsBehind === 1
          ? '1 commit behind'
          : `${install.commitsBehind} commits behind`
  return (
    <aside
      className="flex w-80 shrink-0 flex-col gap-4 overflow-y-auto rounded-lg border bg-elevation-2 p-4"
      aria-label={`Install ${install.name}`}
    >
      <div className="flex items-start justify-between gap-2">
        <div className="flex min-w-0 flex-col gap-1">
          <Text family="mono" weight="strong" className="truncate">
            {install.name}
          </Text>
          <Text family="mono" variant="subtext" theme="neutral">
            {install.branchName || 'Not tracking'}
          </Text>
        </div>
        <Button variant="icon" aria-label="Close" onClick={onClose}>
          <Icon variant="XIcon" size={14} />
        </Button>
      </div>
      {thisBranch ? (
        <Badge size="sm" theme="brand" className="w-fit">
          This branch
        </Badge>
      ) : null}
      <div className="flex flex-col gap-2">
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
        <Field label="Tracking" value={behind} />
        <Field
          label="Config drift"
          value={
            install.configDriftCount === 0
              ? 'None'
              : install.configDriftCount === 1
                ? '1 drifted'
                : `${install.configDriftCount} drifted`
          }
        />
      </div>
      <div className="flex flex-col gap-2">
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
    </aside>
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
  const selected = visible.find((install) => install.id === selectedId)
  const options = labelOptions(installs)
  const sections = groupFleetInstalls(visible, currentBranch)
  const layout = layoutFleet(sections)

  useEffect(() => {
    if (selectedId && !visibleKey.split('\0').includes(selectedId)) {
      setSelectedId(undefined)
    }
  }, [selectedId, visibleKey])

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
          onInspect: () => setSelectedId(card.install.id),
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
        <div className="flex min-h-0 flex-1 gap-4">
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
          {selected ? (
            <Detail
              install={selected}
              thisBranch={
                !!currentBranch && selected.branchName === currentBranch
              }
              onClose={() => setSelectedId(undefined)}
            />
          ) : null}
        </div>
      )}
    </div>
  )
}
