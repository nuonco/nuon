import { useState } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Button } from '@/components/common/Button'
import { DebouncedSearchInput } from '@/components/common/DeboundedSearch'
import { Dropdown } from '@/components/common/Dropdown'
import { EmptyState } from '@/components/common/EmptyState'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { RadioInput } from '@/components/common/form/RadioInput'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Tooltip } from '@/components/common/Tooltip'
import { InstallResourceDetailPanelButton } from '@/components/install-resources/InstallResourceDetailPanel'
import { RemovedFromAppConfigBadge } from '@/components/installs/RemovedFromAppConfig/RemovedFromAppConfig'
import { Badge } from '@/components/common/Badge'
import type { TInstallResource } from '@/types'
import { cn } from '@/utils/classnames'
import {
  bearsHealthVerdict,
  compareHealthSeverityDesc,
  HEALTH_DEGRADED,
  HEALTH_HEALTHY,
  HEALTH_NOT_APPLICABLE,
  HEALTH_PROGRESSING,
  HEALTH_UNHEALTHY,
  HEALTH_UNKNOWN,
  isFailingHealth,
  worstHealth,
} from '@/utils/health-utils'
import { humanize } from '@/utils/string-utils'
import { isStaleObservation, latestTimestamp } from '@/utils/time-utils'

// Synthetic filter value covering everything that carries no verdict, so the
// summary chips and the dropdown share one filter axis instead of two.
export const NO_SIGNAL_FILTER = 'no-signal'

export const HEALTH_FILTER_OPTIONS = [
  HEALTH_UNHEALTHY,
  HEALTH_DEGRADED,
  HEALTH_PROGRESSING,
  HEALTH_HEALTHY,
  HEALTH_UNKNOWN,
  HEALTH_NOT_APPLICABLE,
  NO_SIGNAL_FILTER,
]

export const HEALTH_SUMMARY_ORDER = [
  HEALTH_UNHEALTHY,
  HEALTH_DEGRADED,
  HEALTH_PROGRESSING,
  HEALTH_HEALTHY,
  NO_SIGNAL_FILTER,
]

// How many rows a group shows before folding the tail away: a component with
// 80 pods otherwise buries every other component on the page. Only the healthy
// tail is ever folded — see visibleRowCount.
const GROUP_PREVIEW_ROWS = 8

// Folding two rows behind a "show more" saves nothing and just adds a click;
// real components cluster at either 1-3 rows or 10+.
const MIN_FOLDED_ROWS = 3

function healthFilterLabel(value: string): string {
  if (value === NO_SIGNAL_FILTER) return 'No signal'
  return humanize(value)
}

// Cloud identity rows (aws/gcp/azure) never bear a verdict — mirrors the
// evaluator's bearsVerdict. Staleness is meaningless for them: they are a
// snapshot of what terraform manages, refreshed at apply time.
export function isIdentityOnlyResource(resource: TInstallResource): boolean {
  return ['aws', 'gcp', 'azure'].includes(resource?.provider || '')
}

export function hasHealthSignal(resource: TInstallResource): boolean {
  return (
    bearsHealthVerdict(resource?.health) &&
    !isIdentityOnlyResource(resource) &&
    !resource?.removed_from_config
  )
}

export function matchesHealthFilter(
  resource: TInstallResource,
  health: string
): boolean {
  if (!health) return true
  if (health === NO_SIGNAL_FILTER) return !hasHealthSignal(resource)
  return resource?.health === health
}

export function matchesResourceSearch(
  resource: TInstallResource,
  search: string
): boolean {
  const needle = search.trim().toLowerCase()
  if (!needle) return true
  return [resource?.name, resource?.kind, resource?.namespace].some((field) =>
    field?.toLowerCase().includes(needle)
  )
}

export function healthFacetCounts(
  resources: TInstallResource[]
): Record<string, number> {
  const counts: Record<string, number> = {}
  resources.forEach((resource) => {
    const key = hasHealthSignal(resource)
      ? resource?.health || HEALTH_UNKNOWN
      : NO_SIGNAL_FILTER
    counts[key] = (counts[key] ?? 0) + 1
  })
  return counts
}

export type TInstallResourceRow = {
  resource: TInstallResource
  kind: string
  namespace: string
  name: string
  health: string
  message: string
  observedAt?: string
  removed?: boolean
  identityOnly?: boolean
  hasSignal: boolean
  staleAfterSeconds?: number
}

export type TInstallResourceGroup = {
  key: string
  heading: string
  rows: TInstallResourceRow[]
  signalRows: TInstallResourceRow[]
  noSignalRows: TInstallResourceRow[]
  worst: string
  failing: number
  live: number
  fullyStale: boolean
  lastReportedAt?: string
  downstreamOf?: string
}

function toInstallResourceRow(resource: TInstallResource): TInstallResourceRow {
  return {
    resource,
    kind: resource?.kind || '',
    namespace: resource?.namespace || '',
    name: resource?.name || '',
    health: resource?.health || HEALTH_UNKNOWN,
    message: resource?.message || '',
    observedAt: resource?.observed_at,
    staleAfterSeconds: resource?.stale_after_seconds || undefined,
    removed: !!resource?.removed_from_config,
    identityOnly: isIdentityOnlyResource(resource),
    hasSignal: hasHealthSignal(resource),
  }
}

// Identity rows have no staleness window and removed rows are historical, so
// neither can tell you whether the component stopped reporting.
function isStaleRow(row: TInstallResourceRow): boolean {
  return (
    !row.removed &&
    !row.identityOnly &&
    isStaleObservation(row.observedAt, row.staleAfterSeconds)
  )
}

// A component whose every probe went quiet is a different failure from a few
// stale rows: the badges below it are all last-known values, so the group says
// so once instead of repeating a chip on every row.
function staleSummary(rows: TInstallResourceRow[]): {
  fullyStale: boolean
  lastReportedAt?: string
} {
  const reporting = rows.filter((row) => !row.removed && !row.identityOnly)
  const stale = reporting.filter(isStaleRow)
  return {
    fullyStale: reporting.length > 0 && stale.length === reporting.length,
    lastReportedAt: latestTimestamp(reporting.map((row) => row.observedAt)),
  }
}

// Rows are sorted worst-first, so keeping at least as many as there are
// non-healthy ones guarantees the fold can only ever hide green rows.
export function visibleRowCount(rows: TInstallResourceRow[]): number {
  const needingAttention = rows.filter(
    (row) => row.health !== HEALTH_HEALTHY
  ).length
  return Math.max(GROUP_PREVIEW_ROWS, needingAttention)
}

function compareRows(a: TInstallResourceRow, b: TInstallResourceRow): number {
  return (
    compareHealthSeverityDesc(a.health, b.health) ||
    a.kind.localeCompare(b.kind) ||
    a.name.localeCompare(b.name)
  )
}

function buildInstallResourceGroups(
  resources: TInstallResource[],
  keyOf: (resource: TInstallResource) => string,
  headingFor: (key: string) => string
): TInstallResourceGroup[] {
  const groups = new Map<string, TInstallResourceRow[]>()

  resources.forEach((resource) => {
    const key = keyOf(resource) || 'unknown'
    groups.set(key, [...(groups.get(key) ?? []), toInstallResourceRow(resource)])
  })

  return Array.from(groups.entries())
    .map(([key, unsorted]) => {
      const rows = [...unsorted].sort(compareRows)
      const signalRows = rows.filter((row) => row.hasSignal)
      return {
        key,
        heading: headingFor(key),
        rows,
        signalRows,
        noSignalRows: rows.filter((row) => !row.hasSignal),
        worst: worstHealth(signalRows.map((row) => row.health)),
        ...failingCounts(rows),
        ...staleSummary(rows),
      }
    })
    .sort(
      (a, b) =>
        groupTier(b) - groupTier(a) ||
        compareHealthSeverityDesc(a.worst, b.worst) ||
        a.heading.localeCompare(b.heading)
    )
}

// A group nobody has heard from outranks a healthy one: its badges are last
// known values, so "we are blind here" needs attention before "all good".
// A live failure still outranks both.
function groupTier(group: { failing: number; fullyStale: boolean }): number {
  if (group.failing > 0) return 2
  return group.fullyStale ? 1 : 0
}

// failingCounts mirrors the server-side fold's exclusions: removed and stale
// rows say nothing about live health, and unknown is the absence of
// information, not a failure. What the group header adds over the per-row
// badges is the count — how many of the live resources are actually bad.
function failingCounts(rows: TInstallResourceRow[]): { failing: number; live: number } {
  const live = rows.filter(
    (row) => !row.removed && !isStaleObservation(row.observedAt, row.staleAfterSeconds)
  )
  const failing = live.filter((row) => isFailingHealth(row.health)).length
  return { failing, live: live.length }
}

export function isSandboxResource(resource: TInstallResource): boolean {
  return resource?.source === 'sandbox'
}

export function groupComponentResources(
  resources: TInstallResource[],
  componentNames: Record<string, string>,
  downstreamOf: Record<string, string> = {}
): TInstallResourceGroup[] {
  return buildInstallResourceGroups(
    resources.filter((resource) => !isSandboxResource(resource)),
    (resource) => resource.install_component_id || 'unknown',
    (key) => componentNames[key] || key
  ).map((group) => ({ ...group, downstreamOf: downstreamOf[group.key] }))
}

export function groupSandboxResources(
  resources: TInstallResource[]
): TInstallResourceGroup[] {
  return buildInstallResourceGroups(
    resources.filter(isSandboxResource),
    (resource) => resource.owner_name || 'unknown',
    (key) => key
  )
}

function buildColumns({
  hideStaleBadge = false,
}: { hideStaleBadge?: boolean } = {}): ColumnDef<TInstallResourceRow>[] {
  return [
  {
    accessorKey: 'kind',
    header: 'Kind',
    cell: (info) => <Text>{info.getValue() as string}</Text>,
  },
  {
    accessorKey: 'namespace',
    header: 'Namespace',
    cell: (info) =>
      info.getValue() ? (
        <Text>{info.getValue() as string}</Text>
      ) : (
        <Icon variant="MinusIcon" />
      ),
  },
  {
    accessorKey: 'name',
    header: 'Name',
    cell: (info) => (
      <Text variant="body" weight="strong">
        {info.getValue() as string}
      </Text>
    ),
  },
  {
    accessorKey: 'health',
    header: 'Health',
    cell: (info) => {
      if (info.row.original?.removed) {
        return <RemovedFromAppConfigBadge kind="probe" />
      }
      const stale = !hideStaleBadge && isStaleRow(info.row.original)
      const badge = <Status variant="badge" status={info.getValue() as string} />
      if (!stale) return badge

      // A green badge on an observation nobody has refreshed is the most
      // misleading thing this table can show, so say so next to it.
      return (
        <span className="flex items-center gap-1.5">
          {badge}
          <Tooltip
            position="top"
            tipContent={
              <Text variant="subtext">
                Last reported longer ago than this check's staleness window, so
                it no longer counts toward the component verdict. A pushed check
                reports only when something pushes it.
              </Text>
            }
          >
            <Badge size="sm" theme="warn">
              Stale
            </Badge>
          </Tooltip>
        </span>
      )
    },
  },
  {
    accessorKey: 'message',
    header: 'Message',
    cell: (info) =>
      info.getValue() ? (
        <Text variant="subtext" theme="neutral" className="line-clamp-2 max-w-[320px]">
          {info.getValue() as string}
        </Text>
      ) : (
        <Icon variant="MinusIcon" />
      ),
    enableSorting: false,
  },
  {
    accessorKey: 'observedAt',
    header: 'Age',
    cell: (info) =>
      info.getValue() ? (
        <Time variant="subtext" time={info.getValue() as string} format="relative" />
      ) : (
        <Icon variant="MinusIcon" />
      ),
  },
  {
    id: 'action',
    enableSorting: false,
    header: '',
    cell: (info) => (
      <InstallResourceDetailPanelButton
        installResource={info.row.original.resource}
      />
    ),
  },
  ]
}

const columns = buildColumns()

interface ISingleSelectFilterDropdown {
  id: string
  label: string
  options: string[]
  value: string
  onChange: (value: string) => void
  formatOption?: (option: string) => string
}

const SingleSelectFilterDropdown = ({
  id,
  label,
  options,
  value,
  onChange,
  formatOption = (option) => option,
}: ISingleSelectFilterDropdown) => {
  if (options.length === 0) return null

  return (
    <Dropdown
      alignment="left"
      id={id}
      buttonText={
        <>
          <Icon variant="FunnelIcon" size={14} />
          {label}
          {value ? ` (${formatOption(value)})` : ''}
        </>
      }
    >
      <Menu className="!p-0 !w-64">
        <form onReset={() => onChange('')}>
          <div className="flex flex-col gap-0.5 max-h-64 overflow-y-auto w-full p-2">
            <RadioInput
              checked={value === ''}
              labelProps={{ labelText: 'All' }}
              name={id}
              onChange={() => onChange('')}
              value=""
            />
            {options.map((option) => (
              <RadioInput
                key={option}
                checked={value === option}
                labelProps={{ labelText: formatOption(option) }}
                name={id}
                onChange={() => onChange(option)}
                value={option}
              />
            ))}
          </div>
          <div className="flex flex-col gap-0.5 px-2 pb-2 w-full">
            <hr />
            <Button className="mt-1" isMenuButton type="reset" variant="ghost">
              Clear
              <Icon variant="XIcon" />
            </Button>
          </div>
        </form>
      </Menu>
    </Dropdown>
  )
}

const HealthSummaryChips = ({
  counts,
  active,
  onChange,
}: {
  counts: Record<string, number>
  active: string
  onChange: (value: string) => void
}) => {
  const present = HEALTH_SUMMARY_ORDER.filter((value) => counts[value])
  if (present.length === 0) return null

  const total = present.reduce((sum, value) => sum + counts[value], 0)

  return (
    <div className="flex flex-wrap items-center gap-2">
      <Text variant="subtext" theme="neutral" nowrap>
        {total} {total === 1 ? 'resource' : 'resources'}
      </Text>
      {present.map((value) => {
        const isActive = active === value
        return (
          <button
            key={value}
            type="button"
            aria-pressed={isActive}
            onClick={() => onChange(isActive ? '' : value)}
            className={cn(
              'flex items-center gap-2 rounded-full border px-2.5 py-1 cursor-pointer transition-all',
              isActive
                ? 'border-primary-400 bg-primary-50 dark:border-primary-500/50 dark:bg-primary-950/40'
                : 'border-cool-grey-300 dark:border-dark-grey-600 hover:bg-black/5 dark:hover:bg-white/5'
            )}
          >
            <Status
              variant="badge"
              className="!border-0 !px-0 !py-0"
              status={value === NO_SIGNAL_FILTER ? HEALTH_UNKNOWN : value}
            >
              {healthFilterLabel(value)}
            </Status>
            <Text variant="subtext" theme="neutral">
              {counts[value]}
            </Text>
          </button>
        )
      })}
    </div>
  )
}

const NoSignalDisclosure = ({
  id,
  label,
  isOpen,
  variant = 'row',
  children,
}: {
  id: string
  label: string
  isOpen: boolean
  variant?: 'row' | 'section'
  children: React.ReactNode
}) => (
  <Expand
    id={id}
    isOpen={isOpen}
    isIconBeforeHeading
    className={cn(
      variant === 'row' && 'border-t border-cool-grey-200 dark:border-dark-grey-700'
    )}
    headerClassName="!px-0 !justify-start"
    heading={
      <span className="flex flex-1 items-center gap-3">
        <Text variant="subtext" theme="neutral" nowrap>
          {label}
        </Text>
        {variant === 'section' ? <hr className="flex-1" /> : null}
      </span>
    }
  >
    <div className="pt-2">{children}</div>
  </Expand>
)

const InstallResourceGroupTable = ({
  group,
  forceExpanded,
  showAllRows = false,
}: {
  group: TInstallResourceGroup
  forceExpanded: boolean
  showAllRows?: boolean
}) => {
  const [isTailExpanded, setIsTailExpanded] = useState(false)

  const rows = showAllRows ? group.rows : group.signalRows
  const previewCount = visibleRowCount(rows)
  const hiddenCount = Math.max(0, rows.length - previewCount)
  const canFoldTail =
    !showAllRows && !forceExpanded && hiddenCount >= MIN_FOLDED_ROWS
  const visibleRows =
    canFoldTail && !isTailExpanded ? rows.slice(0, previewCount) : rows

  const groupColumns = buildColumns({ hideStaleBadge: group.fullyStale })
  const table = (
    <Table<TInstallResourceRow>
      columns={groupColumns}
      data={visibleRows}
      enableSearch={false}
    />
  )

  return (
    <div className="flex flex-col gap-3">
      <div className="sticky top-0 z-10 flex items-center gap-2 bg-background py-2">
        <Text variant="base" weight="strong">
          {group.heading}
        </Text>
        {group.downstreamOf ? (
          <Tooltip
            position="top"
            tipContent={
              <Text variant="subtext">
                This component is failing because its dependency{' '}
                {group.downstreamOf} is unhealthy. Its alert is suppressed —{' '}
                {group.downstreamOf} is the one that notifies.
              </Text>
            }
          >
            <Badge size="sm" theme="warn">
              Downstream of {group.downstreamOf}
            </Badge>
          </Tooltip>
        ) : null}
        {/* The component's own verdict lives on the components tab and is
            debounced; what the header shows is the live roll-up of the rows
            below, with a count once something is actually failing. */}
        {group.fullyStale ? (
          <Tooltip
            position="top"
            tipContent={
              <Text variant="subtext">
                Nothing in this group has reported inside its staleness window,
                so none of it counts toward the component verdict. The badges
                below are the last values seen, not current state.
              </Text>
            }
          >
            <Badge size="sm" theme="warn">
              {group.lastReportedAt ? (
                <>
                  Last reported{' '}
                  <Time
                    as="span"
                    variant="label"
                    time={group.lastReportedAt}
                    format="relative"
                  />
                </>
              ) : (
                'Not reporting'
              )}
            </Badge>
          </Tooltip>
        ) : group.failing > 0 ? (
          <Tooltip
            position="top"
            tipContent={
              <Text variant="subtext">
                Live resources currently degraded or unhealthy in this group —
                not the component's health verdict, which is debounced and can
                lag the rows below.
              </Text>
            }
          >
            <Badge size="sm" theme="error">
              {group.failing} of {group.live} failing
            </Badge>
          </Tooltip>
        ) : (
          <Status variant="badge" status={group.worst} />
        )}
      </div>

      {rows.length > 0 ? table : null}

      {canFoldTail ? (
        <Button
          className="self-start"
          variant="ghost"
          size="sm"
          onClick={() => setIsTailExpanded((expanded) => !expanded)}
        >
          {isTailExpanded
            ? 'Show fewer'
            : `Show ${hiddenCount} more healthy ${
                hiddenCount === 1 ? 'resource' : 'resources'
              }`}
          <Icon variant={isTailExpanded ? 'CaretUpIcon' : 'CaretDownIcon'} />
        </Button>
      ) : null}

      {!showAllRows && group.noSignalRows.length > 0 ? (
        <NoSignalDisclosure
          id={`resource-group-${group.key}-no-signal`}
          isOpen={forceExpanded}
          label={`${group.noSignalRows.length} ${
            group.noSignalRows.length === 1 ? 'resource' : 'resources'
          } with no health signal`}
        >
          <Table<TInstallResourceRow>
            columns={columns}
            data={group.noSignalRows}
            enableSearch={false}
          />
        </NoSignalDisclosure>
      ) : null}
    </div>
  )
}

const ResourceGroupSection = ({
  title,
  groups,
  forceExpanded,
}: {
  title: string
  groups: TInstallResourceGroup[]
  forceExpanded: boolean
}) => {
  const signalGroups = groups.filter((group) => group.signalRows.length > 0)
  const noSignalGroups = groups.filter((group) => group.signalRows.length === 0)

  if (groups.length === 0) return null

  return (
    <div className="flex flex-col gap-4">
      <Text variant="h3" weight="strong">
        {title}
      </Text>
      {signalGroups.length > 0 ? (
        <div className="flex flex-col gap-6 md:gap-8">
          {signalGroups.map((group) => (
            <InstallResourceGroupTable
              key={group.key}
              group={group}
              forceExpanded={forceExpanded}
            />
          ))}
        </div>
      ) : null}
      {noSignalGroups.length > 0 ? (
        <NoSignalDisclosure
          id={`resource-section-${title.toLowerCase()}-no-signal`}
          variant="section"
          isOpen={forceExpanded || signalGroups.length === 0}
          label={`${noSignalGroups.length} ${
            noSignalGroups.length === 1 ? 'group' : 'groups'
          } with no health signal`}
        >
          <div className="flex flex-col gap-6 md:gap-8">
            {noSignalGroups.map((group) => (
              <InstallResourceGroupTable
                key={group.key}
                group={group}
                forceExpanded={forceExpanded}
                showAllRows
              />
            ))}
          </div>
        </NoSignalDisclosure>
      ) : null}
    </div>
  )
}

interface IInstallResourcesTable {
  componentGroups: TInstallResourceGroup[]
  sandboxGroups: TInstallResourceGroup[]
  healthCounts: Record<string, number>
  isLoading: boolean
  kind: string
  namespace: string
  health: string
  search: string
  kindOptions: string[]
  namespaceOptions: string[]
  onKindChange: (value: string) => void
  onNamespaceChange: (value: string) => void
  onHealthChange: (value: string) => void
}

export const InstallResourcesTable = ({
  componentGroups,
  sandboxGroups,
  healthCounts,
  isLoading,
  kind,
  namespace,
  health,
  search,
  kindOptions,
  namespaceOptions,
  onKindChange,
  onNamespaceChange,
  onHealthChange,
}: IInstallResourcesTable) => {
  if (isLoading) {
    return <TableSkeleton columns={columns} skeletonRows={5} />
  }

  const hasResources = componentGroups.length > 0 || sandboxGroups.length > 0
  // A filter is an explicit request to see the matching rows, so it overrides
  // every "tucked away by default" decision below — otherwise filtering to
  // `unknown` renders a page of empty groups.
  const forceExpanded = !!(kind || namespace || health || search)

  return (
    <div className="flex flex-col gap-6 md:gap-8">
      <div className="flex flex-col gap-3">
        <HealthSummaryChips
          counts={healthCounts}
          active={health}
          onChange={onHealthChange}
        />
        <div className="flex flex-wrap items-center gap-3">
          <DebouncedSearchInput
            className="w-full md:w-64"
            labelClassName="w-full md:w-64"
            placeholder="Search resources"
          />
          <SingleSelectFilterDropdown
            id="resource-filter-kind"
            label="Kind"
            options={kindOptions}
            value={kind}
            onChange={onKindChange}
          />
          <SingleSelectFilterDropdown
            id="resource-filter-namespace"
            label="Namespace"
            options={namespaceOptions}
            value={namespace}
            onChange={onNamespaceChange}
          />
          <SingleSelectFilterDropdown
            id="resource-filter-health"
            label="Health"
            options={HEALTH_FILTER_OPTIONS}
            value={health}
            onChange={onHealthChange}
            formatOption={healthFilterLabel}
          />
        </div>
      </div>

      {!hasResources ? (
        <EmptyState
          variant="table"
          emptyTitle={forceExpanded ? 'No matching resources' : 'No resources yet'}
          emptyMessage={
            forceExpanded
              ? 'No resources match the current filters. Clear them to see everything this install manages.'
              : 'Resources will appear here once the runner reports component health for this install.'
          }
        />
      ) : (
        <>
          <ResourceGroupSection
            title="Components"
            groups={componentGroups}
            forceExpanded={forceExpanded}
          />
          <ResourceGroupSection
            title="Sandbox"
            groups={sandboxGroups}
            forceExpanded={forceExpanded}
          />
        </>
      )}
    </div>
  )
}
