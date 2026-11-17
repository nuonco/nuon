import { useState } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { CheckboxFilterDropdown } from '@/components/common/CheckboxFilterDropdown'
import { EmptyState } from '@/components/common/EmptyState'
import { HealthBars, type IHealthBar } from '@/components/common/HealthBars'
import { ID } from '@/components/common/ID'
import { Icon } from '@/components/common/Icon'
import { LabelBadge } from '@/components/common/LabelBadge'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { RadioFilterDropdown } from '@/components/common/RadioFilterDropdown'
import { SearchInput } from '@/components/common/SearchInput'
import { Table } from '@/components/common/Table'
import { Text, type TTextTheme } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { ToggleButton } from '@/components/common/ToggleButton'
import { GroupLabels } from '@/components/branches/InstallGroupCards/GroupLabels'
import { Panel } from '@/components/surfaces/Panel'
import { InstallStatusAxis } from '@/components/installs/InstallStatusSummary'
import { INSTALL_STATUS_AXES } from '@/components/installs/InstallStatusSummary/InstallStatusSummary'
import { useInstallHref } from '@/hooks/use-install-path'
import type {
  TAppBranch,
  TAppBranchConfig,
  TInstall,
  TInstallOverview,
  TInstallStatus,
} from '@/types'
import { cn } from '@/utils/classnames'
import {
  componentOptions,
  groupFleet,
  labelOptions,
  latestConfigIdByBranch,
  matchesFilters,
  rolloutState,
  summarizeFleet,
  type TComponentState,
  type TFleetBranch,
  type TFleetContext,
  type TFleetGroup,
  type TFleetRunnerFilter,
  type TRolloutState,
} from './fleet'
import { FleetFiltersDropdown } from './FleetFiltersDropdown'
import { HealthFilterDropdown } from './HealthFilterDropdown'

export interface IInstallFleet {
  installs: TInstall[]
  branches?: TAppBranch[]
  branchConfigs?: TAppBranchConfig[]
  statusById?: Record<string, TInstallStatus>
  overviewById?: Record<string, TInstallOverview>
  isLoading?: boolean
  labelColors?: Record<string, string>
  defaultView?: TFleetView
}

export type TFleetView = 'table' | 'board'

type TFleetMatch = { ids: Set<string>; active: boolean }

const plural = (count: number, noun: string) =>
  `${count} ${noun}${count === 1 ? '' : 's'}`

const isMatch = (install: TInstall, match: TFleetMatch) =>
  match.ids.has(install.id ?? '')

const matchedOf = (installs: TInstall[], match: TFleetMatch) =>
  installs.filter((install) => isMatch(install, match))

const branchCountText = (branch: TFleetBranch, match: TFleetMatch) =>
  match.active
    ? `${matchedOf(branch.installs, match).length} of ${plural(branch.installs.length, 'install')}`
    : plural(branch.installs.length, 'install')

const groupCountText = (
  group: TFleetGroup,
  match: TFleetMatch,
  context: TFleetContext
) => {
  const total = group.installs.length
  if (match.active)
    return `${matchedOf(group.installs, match).length} of ${total} match`
  const states = group.installs.map((install) => rolloutState(install, context))
  if (states.every((state) => state === 'unknown')) return plural(total, 'install')
  return `${states.filter((state) => state === 'applied').length} of ${total} on latest run`
}

const ROLLOUT_LABEL: Record<TRolloutState, string> = {
  applied: 'Applied',
  behind: 'Not applied',
  unknown: 'Not tracked',
}

const ROLLOUT_BAR_CLASS: Record<TRolloutState, string> = {
  applied: 'bg-green-600 dark:bg-green-500',
  behind: 'bg-orange-500 dark:bg-orange-400',
  unknown: 'bg-cool-grey-300 dark:bg-dark-grey-600',
}

const RolloutBadge = ({ state }: { state: TRolloutState }) =>
  state === 'unknown' ? null : (
    <Badge size="sm" theme={state === 'applied' ? 'success' : 'warn'}>
      {ROLLOUT_LABEL[state]}
    </Badge>
  )

const InstallNameCell = ({ install }: { install: TInstall }) => {
  const installHref = useInstallHref()
  return (
    <span>
      <Text variant="body">
        <Link
          href={installHref({
            orgId: install.org_id,
            appId: install.app_id,
            installId: install.id,
          })}
          variant="inline"
        >
          {install.name}
        </Link>
      </Text>
      <ID>{install.id}</ID>
    </span>
  )
}

const InstallStatusAxes = ({
  status,
  wrap = false,
}: {
  status?: TInstallStatus
  wrap?: boolean
}) => (
  <div
    className={cn('flex items-center gap-2', wrap ? 'flex-wrap' : 'w-max')}
  >
    {INSTALL_STATUS_AXES.map(({ key, label }) => (
      <span key={key} title={label}>
        <InstallStatusAxis axis={status?.[key]} loading={!status} />
      </span>
    ))}
  </div>
)

const buildColumns = (
  statusById: Record<string, TInstallStatus>,
  context: TFleetContext,
  labelColors?: Record<string, string>
): ColumnDef<TInstall>[] => [
  {
    accessorKey: 'name',
    header: 'Install name',
    meta: { className: 'w-1/5' },
    cell: (info) => <InstallNameCell install={info.row.original} />,
  },
  {
    id: 'statuses',
    header: 'Statuses',
    enableSorting: false,
    cell: (info) => (
      <InstallStatusAxes status={statusById[info.row.original.id ?? '']} />
    ),
  },
  {
    id: 'rollout',
    header: 'Latest run',
    enableSorting: false,
    meta: { className: 'w-32' },
    cell: (info) => {
      const state = rolloutState(info.row.original, context)
      return state === 'unknown' ? (
        <Icon variant="MinusIcon" />
      ) : (
        <RolloutBadge state={state} />
      )
    },
  },
  {
    id: 'labels',
    header: 'Labels',
    enableSorting: false,
    cell: (info) => {
      const labels = info.row.original.labels ?? {}
      const keys = Object.keys(labels).sort()
      if (!keys.length) return <Icon variant="MinusIcon" />
      return (
        <span className="flex flex-wrap gap-1">
          {keys.map((key) => (
            <LabelBadge
              key={key}
              size="sm"
              labelKey={key}
              labelValue={labels[key]}
              customColor={labelColors?.[key]}
            />
          ))}
        </span>
      )
    },
  },
  {
    accessorKey: 'updated_at',
    header: 'Updated',
    meta: { className: 'w-32' },
    cell: (info) => (
      <Time
        time={info.row.original.updated_at}
        format="relative"
        variant="subtext"
      />
    ),
  },
]

const FleetSummary = ({
  installs,
  context,
  isLoading,
}: {
  installs: TInstall[]
  context: TFleetContext
  isLoading: boolean
}) => {
  const summary = summarizeFleet(installs, context)
  const items: { label: string; value: number; theme?: TTextTheme }[] = [
    { label: 'Installs', value: summary.total },
    {
      label: 'Latest run applied',
      value: summary.applied,
      theme: 'success',
    },
    {
      label: 'Latest run not applied',
      value: summary.behind,
      theme: 'warn',
    },
    {
      label: 'Runner not active',
      value: summary.runnerNotActive,
      theme: 'warn',
    },
  ]
  return (
    <div className="flex flex-wrap gap-6">
      {items.map((item) => (
        <LabeledValue
          key={item.label}
          label={item.label}
          loading={isLoading}
          loadingWidth={3}
        >
          <Text variant="base" weight="strong" theme={item.theme}>
            {item.value}
          </Text>
        </LabeledValue>
      ))}
    </div>
  )
}

const GroupSelectorLabels = ({
  group,
  labelColors,
}: {
  group: TFleetGroup
  labelColors?: Record<string, string>
}) => (
  <>
    {Object.entries(group.labels ?? {})
      .sort(([a], [b]) => a.localeCompare(b))
      .map(([key, value]) => (
        <LabelBadge
          key={key}
          size="sm"
          labelKey={key}
          labelValue={value}
          customColor={labelColors?.[key]}
        />
      ))}
  </>
)

const TABLE_VISIBLE_COUNT = 5

const useVisibleInstalls = (installs: TInstall[], visibleCount: number) => {
  const [isExpanded, setIsExpanded] = useState(false)
  return {
    visible: isExpanded ? installs : installs.slice(0, visibleCount),
    viewMore:
      installs.length > visibleCount ? (
        <div className="flex justify-center">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setIsExpanded((expanded) => !expanded)}
          >
            {isExpanded
              ? 'Show less'
              : `View more (${installs.length - visibleCount})`}
          </Button>
        </div>
      ) : null,
  }
}

const GroupName = ({ group }: { group: TFleetGroup }) =>
  group.name ? (
    <Text variant="subtext" weight="strong" family="mono">
      {group.name}
    </Text>
  ) : (
    <Text variant="subtext" weight="strong">
      No install group
    </Text>
  )

const BranchHeading = ({
  branch,
  match,
}: {
  branch: TFleetBranch
  match: TFleetMatch
}) => {
  const appId = branch.installs[0]?.app_id
  const orgId = branch.installs[0]?.org_id
  return (
    <div className="flex items-center gap-2">
      {branch.branchId ? (
        <>
          <Icon variant="GitBranchIcon" size={16} />
          <Text variant="base" weight="strong" family="mono">
            <Link
              href={`/${orgId}/apps/${appId}/branches/${branch.branchId}`}
              variant="inline"
            >
              {branch.name}
            </Link>
          </Text>
        </>
      ) : (
        <Text variant="base" weight="strong">
          Not on a branch
        </Text>
      )}
      <Text variant="subtext" theme="neutral">
        {branchCountText(branch, match)}
      </Text>
    </div>
  )
}

const FleetGroupTable = ({
  group,
  columns,
  labelColors,
  match,
  context,
}: {
  group: TFleetGroup
  columns: ColumnDef<TInstall>[]
  labelColors?: Record<string, string>
  match: TFleetMatch
  context: TFleetContext
}) => {
  const matched = matchedOf(group.installs, match)
  const { visible, viewMore } = useVisibleInstalls(matched, TABLE_VISIBLE_COUNT)
  return (
    <div className="flex flex-col gap-3">
      <div className="flex flex-wrap items-center gap-2">
        <GroupName group={group} />
        <GroupSelectorLabels group={group} labelColors={labelColors} />
        <Text variant="subtext" theme="neutral">
          {groupCountText(group, match, context)}
        </Text>
      </div>
      {matched.length ? (
        <>
          <Table<TInstall>
            columns={columns}
            data={visible}
            enableSearch={false}
          />
          {viewMore}
        </>
      ) : null}
    </div>
  )
}

const FleetBranchSection = ({
  branch,
  columns,
  labelColors,
  match,
  context,
}: {
  branch: TFleetBranch
  columns: ColumnDef<TInstall>[]
  labelColors?: Record<string, string>
  match: TFleetMatch
  context: TFleetContext
}) => (
  <div className="flex flex-col gap-4">
    <BranchHeading branch={branch} match={match} />
    <div className="flex flex-col gap-6">
      {branch.groups.map((group) => (
        <FleetGroupTable
          key={group.key}
          group={group}
          columns={columns}
          labelColors={labelColors}
          match={match}
          context={context}
        />
      ))}
    </div>
  </div>
)

const GroupRolloutBars = ({
  installs,
  match,
  context,
}: {
  installs: TInstall[]
  match: TFleetMatch
  context: TFleetContext
}) => {
  const bars: IHealthBar[] = installs.map((install) => {
    const state = rolloutState(install, context)
    return {
      key: install.id ?? install.name ?? '',
      colorClass: cn(
        ROLLOUT_BAR_CLASS[state],
        match.active && !isMatch(install, match) && 'opacity-25'
      ),
      ariaLabel: `${install.name}: ${ROLLOUT_LABEL[state]}`,
      tooltip: (
        <div className="flex flex-col gap-0.5 w-44">
          <Text variant="subtext" weight="strong" family="mono">
            {install.name}
          </Text>
          <Text variant="label" theme="neutral">
            Latest run: {ROLLOUT_LABEL[state]}
          </Text>
        </div>
      ),
    }
  })
  return (
    <HealthBars
      grow
      bars={bars}
      className={installs.length === 1 ? 'w-12' : 'w-full'}
      barClassName="h-5 rounded-xs"
    />
  )
}

const INSTALL_CARD_CLASS =
  'flex w-full flex-wrap items-center justify-between gap-3 rounded-md border bg-elevation-2 px-4 py-3 shadow-sm'

const FleetGroupPanelBody = ({
  group,
  statusById,
  match,
  context,
}: {
  group: TFleetGroup
  statusById: Record<string, TInstallStatus>
  match: TFleetMatch
  context: TFleetContext
}) => {
  const [search, setSearch] = useState('')
  const [rollout, setRollout] = useState<'applied' | 'behind'>()
  const installs = matchedOf(group.installs, match).filter((install) =>
    matchesFilters(install, { search, labels: new Set(), rollout }, context)
  )

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-3">
        <Text variant="subtext" theme="neutral">
          {groupCountText(group, match, context)}
        </Text>
        <GroupLabels
          match={
            group.labels
              ? { kind: 'labels', labels: group.labels }
              : { kind: 'default' }
          }
        />
        <GroupRolloutBars
          installs={group.installs}
          match={match}
          context={context}
        />
      </div>

      <div className="flex flex-col gap-3">
        <div className="flex items-center gap-2">
          <SearchInput
            labelClassName="flex-1"
            className="w-full md:min-w-0"
            placeholder="Search installs"
            value={search}
            onChange={setSearch}
          />
          <RadioFilterDropdown
            id={`fleet-group-${group.key}-rollout`}
            label="Latest run"
            options={[
              { value: 'applied', label: ROLLOUT_LABEL.applied },
              { value: 'behind', label: ROLLOUT_LABEL.behind },
            ]}
            selected={rollout}
            onChange={(value) =>
              setRollout(
                value === 'applied' || value === 'behind' ? value : undefined
              )
            }
          />
        </div>
        {installs.length ? (
          <ul className="flex flex-col gap-2">
            {installs.map((install) => (
              <li key={install.id} className={INSTALL_CARD_CLASS}>
                <InstallNameCell install={install} />
                <span className="flex flex-wrap items-center gap-2">
                  <RolloutBadge state={rolloutState(install, context)} />
                  <InstallStatusAxes status={statusById[install.id ?? '']} />
                </span>
              </li>
            ))}
          </ul>
        ) : (
          <Text variant="subtext" theme="neutral">
            No installs match.
          </Text>
        )}
      </div>
    </div>
  )
}

const FleetGroupCard = ({
  branch,
  group,
  statusById,
  labelColors,
  match,
  context,
}: {
  branch: TFleetBranch
  group: TFleetGroup
  statusById: Record<string, TInstallStatus>
  labelColors?: Record<string, string>
  match: TFleetMatch
  context: TFleetContext
}) => (
  <Card className="!p-4 !gap-3 w-64 shrink-0">
    <GroupName group={group} />
    {group.labels ? (
      <div className="flex flex-wrap gap-1">
        <GroupSelectorLabels group={group} labelColors={labelColors} />
      </div>
    ) : null}
    <GroupRolloutBars installs={group.installs} match={match} context={context} />
    <div className="flex items-center justify-between gap-2">
      <Text variant="subtext" theme="neutral">
        {groupCountText(group, match, context)}
      </Text>
      <Panel
        panelKey={`fleet-group-${branch.key}-${group.key}`}
        size="half"
        heading={group.name ?? 'No install group'}
        triggerButton={{
          variant: 'ghost',
          size: 'sm',
          children: 'View installs',
        }}
      >
        <FleetGroupPanelBody
          group={group}
          statusById={statusById}
          match={match}
          context={context}
        />
      </Panel>
    </div>
  </Card>
)

const FleetLanes = ({
  branches,
  statusById,
  labelColors,
  match,
  context,
}: {
  branches: TFleetBranch[]
  statusById: Record<string, TInstallStatus>
  labelColors?: Record<string, string>
  match: TFleetMatch
  context: TFleetContext
}) => (
  <div className="flex flex-col gap-8 md:gap-10">
    {branches.map((branch) => (
      <div key={branch.key} className="flex flex-col gap-4">
        <BranchHeading branch={branch} match={match} />
        <div className="flex items-stretch gap-4 overflow-x-auto pb-2">
          {branch.groups.map((group) => (
            <FleetGroupCard
              key={group.key}
              branch={branch}
              group={group}
              statusById={statusById}
              labelColors={labelColors}
              match={match}
              context={context}
            />
          ))}
        </div>
      </div>
    ))}
  </div>
)

export const InstallFleet = ({
  installs,
  branches = [],
  branchConfigs = [],
  statusById = {},
  overviewById = {},
  isLoading = false,
  labelColors,
  defaultView = 'table',
}: IInstallFleet) => {
  const [search, setSearch] = useState('')
  const [labels, setLabels] = useState<Set<string>>(new Set())
  const [health, setHealth] = useState<Set<string>>(new Set())
  const [runner, setRunner] = useState<TFleetRunnerFilter>()
  const [rollout, setRollout] = useState<'applied' | 'behind'>()
  const [component, setComponent] = useState<string>()
  const [componentState, setComponentState] = useState<TComponentState>('live')
  const [view, setView] = useState<TFleetView>(defaultView)

  const context: TFleetContext = {
    latestConfigIdByBranch: latestConfigIdByBranch(branches),
    overviewById,
  }
  const columns = buildColumns(statusById, context, labelColors)
  const options = labelOptions(installs)
  const componentChoices = componentOptions(overviewById)
  const visible = installs.filter((install) =>
    matchesFilters(
      install,
      { search, labels, health, runner, rollout, component, componentState },
      context
    )
  )
  const match: TFleetMatch = {
    ids: new Set(visible.map((install) => install.id ?? '')),
    active:
      !!search.trim() ||
      labels.size > 0 ||
      health.size > 0 ||
      !!runner ||
      !!rollout ||
      !!component,
  }
  const fleet = groupFleet(installs, branchConfigs, context)

  const resetMenuFilters = () => {
    setRunner(undefined)
    setRollout(undefined)
    setComponent(undefined)
    setComponentState('live')
  }

  const clearFilters = () => {
    setSearch('')
    setLabels(new Set())
    setHealth(new Set())
    resetMenuFilters()
  }

  return (
    <div className="flex flex-col gap-6 md:gap-8">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <FleetSummary
          installs={installs}
          context={context}
          isLoading={isLoading}
        />
        <ToggleButton<TFleetView>
          label="Fleet layout"
          value={view}
          onChange={setView}
          options={[
            {
              value: 'table',
              label: <Icon variant="TableIcon" size={16} />,
              ariaLabel: 'Table view',
            },
            {
              value: 'board',
              label: <Icon variant="SquaresFourIcon" size={16} />,
              ariaLabel: 'Rollout lanes',
            },
          ]}
        />
      </div>

      <div className="flex flex-row flex-wrap items-center justify-between gap-4">
        <div className="flex flex-wrap items-center gap-4">
          <SearchInput
            aria-label="Search installs"
            labelClassName="w-full md:w-fit"
            className="w-full md:w-fit"
            placeholder="Search by name or ID..."
            value={search}
            onChange={setSearch}
          />
          {!isLoading ? (
            <Text variant="subtext" theme="neutral" aria-live="polite">
              {visible.length === installs.length
                ? plural(installs.length, 'install')
                : `${visible.length} of ${plural(installs.length, 'install')}`}
            </Text>
          ) : null}
        </div>
        <div className="flex flex-wrap items-center gap-2">
          {options.length ? (
            <CheckboxFilterDropdown
              id="fleet-filter-labels"
              label="Labels"
              options={options.map((value) => ({ value, label: value }))}
              selected={labels}
              onChange={setLabels}
            />
          ) : null}
          <HealthFilterDropdown selected={health} onChange={setHealth} />
          <FleetFiltersDropdown
            rollout={rollout}
            onRolloutChange={setRollout}
            componentChoices={componentChoices}
            component={component}
            onComponentChange={(value) => {
              setComponent(value)
              setComponentState('live')
            }}
            componentState={componentState}
            onComponentStateChange={setComponentState}
            runner={runner}
            onRunnerChange={setRunner}
            onReset={resetMenuFilters}
          />
        </div>
      </div>

      {isLoading ? (
        <Table<TInstall>
          columns={columns}
          data={[]}
          enableSearch={false}
          isLoading
        />
      ) : !installs.length ? (
        <EmptyState
          variant="table"
          emptyTitle="No installs yet"
          emptyMessage="Installs will appear here once you create one for this app."
        />
      ) : !visible.length ? (
        <EmptyState
          variant="search"
          emptyTitle="No matching installs"
          emptyMessage="No installs match the current filters. Clear them to see the whole fleet."
          action={<Button onClick={clearFilters}>Clear filters</Button>}
        />
      ) : view === 'board' ? (
        <FleetLanes
          branches={fleet}
          statusById={statusById}
          labelColors={labelColors}
          match={match}
          context={context}
        />
      ) : (
        <div className="flex flex-col gap-8 md:gap-10">
          {fleet.map((branch) => (
            <FleetBranchSection
              key={branch.key}
              branch={branch}
              columns={columns}
              labelColors={labelColors}
              match={match}
              context={context}
            />
          ))}
        </div>
      )}
    </div>
  )
}
