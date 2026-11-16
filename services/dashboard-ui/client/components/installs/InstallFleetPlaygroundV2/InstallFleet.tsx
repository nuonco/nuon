import { useState } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
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
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { Text, type TTextTheme } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { ToggleButton } from '@/components/common/ToggleButton'
import { GroupLabels } from '@/components/branches/InstallGroupCards/GroupLabels'
import { Panel } from '@/components/surfaces/Panel'
import { InstallStatusAxis } from '@/components/installs/InstallStatusSummary'
import { INSTALL_STATUS_AXES } from '@/components/installs/InstallStatusSummary/InstallStatusSummary'
import { useInstallHref } from '@/hooks/use-install-path'
import type { TAppBranchConfig, TInstall, TInstallStatus } from '@/types'
import { cn } from '@/utils/classnames'
import { getStatusTheme } from '@/utils/status-utils'
import { humanize } from '@/utils/string-utils'
import {
  groupFleet,
  healthFilterValue,
  labelOptions,
  matchesFilters,
  summarizeFleet,
  type TFleetBranch,
  type TFleetGroup,
  type TFleetRunnerFilter,
} from './fleet'
import { HealthFilterDropdown, healthFilterLabel } from './HealthFilterDropdown'

export interface IInstallFleet {
  installs: TInstall[]
  branchConfigs?: TAppBranchConfig[]
  statusById?: Record<string, TInstallStatus>
  isLoading?: boolean
  labelColors?: Record<string, string>
  defaultView?: TFleetView
}

export type TFleetView = 'table' | 'board'

const plural = (count: number, noun: string) =>
  `${count} ${noun}${count === 1 ? '' : 's'}`

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
  isLoading,
}: {
  installs: TInstall[]
  isLoading: boolean
}) => {
  const summary = summarizeFleet(installs)
  const items: { label: string; value: number; theme?: TTextTheme }[] = [
    { label: 'Installs', value: summary.total },
    { label: 'Healthy', value: summary.healthy, theme: 'success' },
    { label: 'Degraded', value: summary.degraded, theme: 'warn' },
    { label: 'Unhealthy', value: summary.unhealthy, theme: 'error' },
    { label: 'Unknown', value: summary.unknown, theme: 'neutral' },
    { label: 'No health signal', value: summary.noSignal, theme: 'neutral' },
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

const BranchHeading = ({ branch }: { branch: TFleetBranch }) => {
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
        {plural(branch.installs.length, 'install')}
      </Text>
    </div>
  )
}

const FleetGroupTable = ({
  group,
  columns,
  labelColors,
}: {
  group: TFleetGroup
  columns: ColumnDef<TInstall>[]
  labelColors?: Record<string, string>
}) => {
  const { visible, viewMore } = useVisibleInstalls(
    group.installs,
    TABLE_VISIBLE_COUNT
  )
  return (
    <div className="flex flex-col gap-3">
      <div className="flex flex-wrap items-center gap-2">
        <GroupName group={group} />
        <GroupSelectorLabels group={group} labelColors={labelColors} />
        <Status variant="badge" status={group.worst} />
        <Text variant="subtext" theme="neutral">
          {plural(group.installs.length, 'install')}
        </Text>
      </div>
      <Table<TInstall> columns={columns} data={visible} enableSearch={false} />
      {viewMore}
    </div>
  )
}

const FleetBranchSection = ({
  branch,
  columns,
  labelColors,
}: {
  branch: TFleetBranch
  columns: ColumnDef<TInstall>[]
  labelColors?: Record<string, string>
}) => (
  <div className="flex flex-col gap-4">
    <BranchHeading branch={branch} />
    <div className="flex flex-col gap-6">
      {branch.groups.map((group) => (
        <FleetGroupTable
          key={group.key}
          group={group}
          columns={columns}
          labelColors={labelColors}
        />
      ))}
    </div>
  </div>
)

const BAR_CLASSES: Record<string, string> = {
  success: 'bg-green-600 dark:bg-green-500',
  error: 'bg-red-600 dark:bg-red-500',
  warn: 'bg-orange-500 dark:bg-orange-400',
  info: 'bg-blue-600 dark:bg-blue-500',
  neutral: 'bg-cool-grey-300 dark:bg-dark-grey-600',
}

const healthBarClass = (install: TInstall) =>
  BAR_CLASSES[getStatusTheme(install.composite_health_status ?? 'unknown')] ??
  BAR_CLASSES.neutral

const InstallBarTooltip = ({ install }: { install: TInstall }) => (
  <div className="flex flex-col gap-0.5 w-44">
    <Text variant="subtext" weight="strong" family="mono">
      {install.name}
    </Text>
    <Text variant="label" theme="neutral">
      Health: {humanize(install.composite_health_status ?? 'unknown')}
    </Text>
  </div>
)

const GroupHealthBars = ({ installs }: { installs: TInstall[] }) => {
  const bars: IHealthBar[] = installs.map((install) => ({
    key: install.id ?? install.name ?? '',
    colorClass: healthBarClass(install),
    ariaLabel: `${install.name}: ${humanize(install.composite_health_status ?? 'unknown')}`,
    tooltip: <InstallBarTooltip install={install} />,
  }))
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
}: {
  group: TFleetGroup
  statusById: Record<string, TInstallStatus>
}) => {
  const [search, setSearch] = useState('')
  const [health, setHealth] = useState<string>()
  const healthOptions = [
    ...new Set(group.installs.map((install) => healthFilterValue(install))),
  ]
  const installs = group.installs.filter((install) =>
    matchesFilters(install, {
      search,
      labels: new Set(),
      health: health ? new Set([health]) : undefined,
    })
  )

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-3">
        <div className="flex flex-wrap items-baseline gap-x-4 gap-y-2">
          <Status status={group.worst} />
          <Text variant="subtext" theme="neutral">
            {plural(group.installs.length, 'install')}
          </Text>
        </div>
        <GroupLabels
          match={
            group.labels
              ? { kind: 'labels', labels: group.labels }
              : { kind: 'default' }
          }
        />
        <GroupHealthBars installs={group.installs} />
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
            id={`fleet-group-${group.key}-health`}
            label="Health"
            options={healthOptions.map((value) => ({
              value,
              label: healthFilterLabel(value),
            }))}
            selected={health}
            onChange={setHealth}
          />
        </div>
        {installs.length ? (
          <ul className="flex flex-col gap-2">
            {installs.map((install) => (
              <li key={install.id} className={INSTALL_CARD_CLASS}>
                <InstallNameCell install={install} />
                <InstallStatusAxes status={statusById[install.id ?? '']} />
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
}: {
  branch: TFleetBranch
  group: TFleetGroup
  statusById: Record<string, TInstallStatus>
  labelColors?: Record<string, string>
}) => (
  <Card className="!p-4 !gap-3 w-64 shrink-0">
    <div className="flex items-center justify-between gap-2">
      <GroupName group={group} />
      <Status variant="badge" status={group.worst} />
    </div>
    {group.labels ? (
      <div className="flex flex-wrap gap-1">
        <GroupSelectorLabels group={group} labelColors={labelColors} />
      </div>
    ) : null}
    <GroupHealthBars installs={group.installs} />
    <div className="flex items-center justify-between gap-2">
      <Text variant="subtext" theme="neutral">
        {plural(group.installs.length, 'install')}
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
        <FleetGroupPanelBody group={group} statusById={statusById} />
      </Panel>
    </div>
  </Card>
)

const FleetLanes = ({
  branches,
  statusById,
  labelColors,
}: {
  branches: TFleetBranch[]
  statusById: Record<string, TInstallStatus>
  labelColors?: Record<string, string>
}) => (
  <div className="flex flex-col gap-8 md:gap-10">
    {branches.map((branch) => (
      <div key={branch.key} className="flex flex-col gap-4">
        <BranchHeading branch={branch} />
        <div className="flex items-stretch gap-4 overflow-x-auto pb-2">
          {branch.groups.map((group) => (
            <FleetGroupCard
              key={group.key}
              branch={branch}
              group={group}
              statusById={statusById}
              labelColors={labelColors}
            />
          ))}
        </div>
      </div>
    ))}
  </div>
)

export const InstallFleet = ({
  installs,
  branchConfigs = [],
  statusById = {},
  isLoading = false,
  labelColors,
  defaultView = 'table',
}: IInstallFleet) => {
  const [search, setSearch] = useState('')
  const [labels, setLabels] = useState<Set<string>>(new Set())
  const [health, setHealth] = useState<Set<string>>(new Set())
  const [runner, setRunner] = useState<TFleetRunnerFilter>()
  const [view, setView] = useState<TFleetView>(defaultView)

  const columns = buildColumns(statusById, labelColors)
  const options = labelOptions(installs)
  const visible = installs.filter((install) =>
    matchesFilters(install, { search, labels, health, runner })
  )
  const branches = groupFleet(visible, branchConfigs)

  const clearFilters = () => {
    setSearch('')
    setLabels(new Set())
    setHealth(new Set())
    setRunner(undefined)
  }

  return (
    <div className="flex flex-col gap-6 md:gap-8">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <FleetSummary installs={installs} isLoading={isLoading} />
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
          <RadioFilterDropdown
            id="fleet-filter-runner"
            label="Runner"
            options={[
              { value: 'active', label: 'Active' },
              { value: 'not-active', label: 'Not active' },
            ]}
            selected={runner}
            onChange={(value) =>
              setRunner(
                value === 'active' || value === 'not-active' ? value : undefined
              )
            }
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
          branches={branches}
          statusById={statusById}
          labelColors={labelColors}
        />
      ) : (
        <div className="flex flex-col gap-8 md:gap-10">
          {branches.map((branch) => (
            <FleetBranchSection
              key={branch.key}
              branch={branch}
              columns={columns}
              labelColors={labelColors}
            />
          ))}
        </div>
      )}
    </div>
  )
}
