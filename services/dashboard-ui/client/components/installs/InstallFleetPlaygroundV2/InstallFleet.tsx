import { useState } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { AgentPromptBanner } from '@/components/common/AgentPromptBanner'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
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
import type {
  TAppBranch,
  TAppBranchConfig,
  TAppBranchRun,
  TInstall,
  TInstallOverview,
  TInstallStatus,
  TWorkflowStepApproval,
} from '@/types'
import { getApprovalType } from '@/utils/approval-utils'
import { cn } from '@/utils/classnames'
import { getApprovalHref, getWorkflowStepTitle } from '@/utils/workflow-utils'
import {
  approvalsByInstall,
  ATTENTION_OPTIONS,
  attentionItems,
  attentionLabel,
  disabledRoles,
  disabledRolesText,
  installOverrides,
  overridesText,
  componentOptions,
  configVersion,
  PENDING_APPROVAL,
  countOnRun,
  groupFleet,
  labelOptions,
  latestConfigIdByBranch,
  matchesFilters,
  rolloutState,
  runsIndex,
  summarizeFleet,
  type TAttentionItem,
  type TFleetBranch,
  type TFleetContext,
  type TFleetGroup,
  type TFleetRunnerFilter,
  type TRolloutState,
} from './fleet'
import { BranchRunPicker, type TFleetRunOption } from './BranchRunPicker'
import { HealthFilterDropdown } from './HealthFilterDropdown'
import { LabelsFilterDropdown } from './LabelsFilterDropdown'
import { FleetFiltersDropdown } from './FleetFiltersDropdown'

export interface IInstallFleet {
  installs: TInstall[]
  branches?: TAppBranch[]
  branchRuns?: Record<string, TAppBranchRun[]>
  branchConfigs?: TAppBranchConfig[]
  statusById?: Record<string, TInstallStatus>
  overviewById?: Record<string, TInstallOverview>
  approvals?: TWorkflowStepApproval[]
  configVersionById?: Record<string, number | undefined>
  isLoading?: boolean
  labelColors?: Record<string, string>
  defaultView?: TFleetView
}

export type TFleetView = 'table' | 'board'

type TFleetMatch = { ids: Set<string>; active: boolean }

type TBranchRollout = Exclude<TRolloutState, 'unknown'>

type TBranchControls = {
  runsFor: (branchId: string) => TFleetRunOption[]
  runByBranch: Record<string, string | undefined>
  onRunChange: (branchId: string, run?: string) => void
  rolloutByBranch: Record<string, TBranchRollout | undefined>
  onRolloutChange: (branchId: string, rollout?: TBranchRollout) => void
}

const rolloutCounts = (installs: TInstall[], context: TFleetContext) => {
  const states = installs.map((install) => rolloutState(install, context))
  return {
    on: states.filter((state) => state === 'applied').length,
    off: states.filter((state) => state === 'behind').length,
  }
}

const fleetPrompt = (orgId: string, appId: string) =>
  `Manage the install fleet for app ${appId} in org ${orgId} with the Nuon MCP.

Start with a read of the fleet. Group installs by branch and report:
- installs not on the latest branch run, and when each install was last upgraded
- installs that are unhealthy or degraded
- installs waiting on approval
- installs whose runner is not active
- installs with a disabled role
- installs with component overrides set

Do not change anything in that first pass. After the summary, ask what to do. Confirm with me before approving a plan, deploying, retrying a step, or deprovisioning.`

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
  const { on, off } = rolloutCounts(group.installs, context)
  if (!on && !off) return plural(total, 'install')
  return `${on} on · ${off} not on`
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

const ROLLOUT_COUNT_THEME: Record<TRolloutState, 'success' | 'warn' | 'neutral'> = {
  applied: 'success',
  behind: 'warn',
  unknown: 'neutral',
}

const ROLLOUT_COUNT_LABEL: Record<TRolloutState, string> = {
  applied: 'on',
  behind: 'not on',
  unknown: 'not tracked',
}

const GroupRolloutCounts = ({
  group,
  match,
  context,
}: {
  group: TFleetGroup
  match: TFleetMatch
  context: TFleetContext
}) => {
  const counts: Record<TRolloutState, number> = {
    applied: 0,
    behind: 0,
    unknown: 0,
  }
  for (const install of group.installs) counts[rolloutState(install, context)]++
  if (match.active || (!counts.applied && !counts.behind))
    return (
      <Text variant="subtext" theme="neutral">
        {groupCountText(group, match, context)}
      </Text>
    )
  const shown = (['applied', 'behind', 'unknown'] as const).filter(
    (state) => counts[state] > 0
  )
  return (
    <span className="flex flex-wrap items-center gap-1">
      {shown.map((state) => (
        <Badge key={state} size="sm" theme={ROLLOUT_COUNT_THEME[state]}>
          {counts[state]} {ROLLOUT_COUNT_LABEL[state]}
        </Badge>
      ))}
    </span>
  )
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
  <div className={cn('flex items-center gap-2', wrap ? 'flex-wrap' : 'w-max')}>
    {INSTALL_STATUS_AXES.map(({ key, label }) => (
      <span key={key} title={label}>
        <InstallStatusAxis axis={status?.[key]} loading={!status} />
      </span>
    ))}
  </div>
)

const Missing = () => <Icon variant="MinusIcon" />

const UpgradeTime = ({ install }: { install: TInstall }) => {
  const time = install.app_config_ref?.applied_config_at
  return time ? (
    <Time time={time} format="relative" variant="subtext" />
  ) : (
    <Missing />
  )
}

const ConfigVersion = ({
  install,
  context,
}: {
  install: TInstall
  context: TFleetContext
}) => {
  const version = configVersion(install, context)
  return version !== undefined ? (
    <Text variant="subtext" family="mono">
      v{version}
    </Text>
  ) : (
    <Missing />
  )
}

const RunnerState = ({ install }: { install: TInstall }) => (
  <Status
    status={install.runner_status ?? 'unknown'}
    title={install.runner_status_description}
  />
)

const ApprovalState = ({
  install,
  context,
}: {
  install: TInstall
  context: TFleetContext
}) => {
  const count = context.approvalsById?.[install.id ?? '']?.length ?? 0
  return count ? (
    <Status status={PENDING_APPROVAL}>
      {count === 1 ? 'Needs approval' : `${count} need approval`}
    </Status>
  ) : (
    <Missing />
  )
}

const RolesState = ({ install }: { install: TInstall }) => {
  const text = disabledRolesText(install)
  return text ? (
    <Status
      status="warning"
      title={disabledRoles(install)
        .map((role) => role.app_role_config?.display_name ?? role.role_id)
        .join(', ')}
    >
      {text}
    </Status>
  ) : (
    <Missing />
  )
}

const OverridesState = ({ install }: { install: TInstall }) => {
  const text = overridesText(install)
  return text ? (
    <Status status="info" title={installOverrides(install).join(', ')}>
      {text}
    </Status>
  ) : (
    <Missing />
  )
}

const InstallFacts = ({
  install,
  context,
}: {
  install: TInstall
  context: TFleetContext
}) => (
  <div className="flex flex-wrap gap-x-6 gap-y-2">
    <LabeledValue label="Config">
      <ConfigVersion install={install} context={context} />
    </LabeledValue>
    <LabeledValue label="Upgraded">
      <UpgradeTime install={install} />
    </LabeledValue>
    <LabeledValue label="Runner">
      <RunnerState install={install} />
    </LabeledValue>
    <LabeledValue label="Approval">
      <ApprovalState install={install} context={context} />
    </LabeledValue>
    <LabeledValue label="Roles">
      <RolesState install={install} />
    </LabeledValue>
    <LabeledValue label="Overrides">
      <OverridesState install={install} />
    </LabeledValue>
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
    id: 'config',
    header: 'Config',
    enableSorting: false,
    cell: (info) => (
      <ConfigVersion install={info.row.original} context={context} />
    ),
  },
  {
    id: 'upgraded',
    header: 'Upgraded',
    enableSorting: false,
    meta: { className: 'w-32' },
    cell: (info) => <UpgradeTime install={info.row.original} />,
  },
  {
    id: 'rollout',
    header: 'Run',
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
    id: 'runner',
    header: 'Runner',
    enableSorting: false,
    cell: (info) => <RunnerState install={info.row.original} />,
  },
  {
    id: 'approval',
    header: 'Approval',
    enableSorting: false,
    cell: (info) => (
      <ApprovalState install={info.row.original} context={context} />
    ),
  },
  {
    id: 'roles',
    header: 'Roles',
    enableSorting: false,
    cell: (info) => <RolesState install={info.row.original} />,
  },
  {
    id: 'overrides',
    header: 'Overrides',
    enableSorting: false,
    cell: (info) => <OverridesState install={info.row.original} />,
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
  context,
  controls,
}: {
  branch: TFleetBranch
  match: TFleetMatch
  context: TFleetContext
  controls: TBranchControls
}) => {
  const appId = branch.installs[0]?.app_id
  const orgId = branch.installs[0]?.org_id
  const branchId = branch.branchId
  const runs = branchId ? controls.runsFor(branchId) : []
  const activeRun =
    runs.find((run) => run.value === controls.runByBranch[branchId ?? '']) ??
    runs[0]
  const { on, off } = rolloutCounts(branch.installs, context)
  const rollout = branchId ? controls.rolloutByBranch[branchId] : undefined
  const toggle = (value: TBranchRollout) => {
    if (branchId)
      controls.onRolloutChange(branchId, rollout === value ? undefined : value)
  }
  return (
    <div className="flex flex-wrap items-center gap-2">
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
      {branchId && runs.length ? (
        <>
          <BranchRunPicker
            id={`fleet-branch-run-${branchId}`}
            runs={runs}
            selected={controls.runByBranch[branchId]}
            onChange={(value) => controls.onRunChange(branchId, value)}
          />
          {activeRun?.createdAt ? (
            <Time
              time={activeRun.createdAt}
              format="relative"
              variant="subtext"
              theme="neutral"
            />
          ) : null}
          <Button
            variant={rollout === 'applied' ? 'secondary' : 'ghost'}
            aria-pressed={rollout === 'applied'}
            onClick={() => toggle('applied')}
          >
            {on} on
          </Button>
          <Button
            variant={rollout === 'behind' ? 'secondary' : 'ghost'}
            aria-pressed={rollout === 'behind'}
            onClick={() => toggle('behind')}
          >
            {off} not on
          </Button>
        </>
      ) : null}
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
  controls,
}: {
  branch: TFleetBranch
  columns: ColumnDef<TInstall>[]
  labelColors?: Record<string, string>
  match: TFleetMatch
  context: TFleetContext
  controls: TBranchControls
}) => (
  <div className="flex flex-col gap-4">
    <BranchHeading
      branch={branch}
      match={match}
      context={context}
      controls={controls}
    />
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

const PROGRESS_ORDER: Record<TRolloutState, number> = {
  applied: 0,
  behind: 1,
  unknown: 2,
}

const GroupRolloutBars = ({
  installs,
  match,
  context,
}: {
  installs: TInstall[]
  match: TFleetMatch
  context: TFleetContext
}) => {
  const bars = [...installs].sort(
    (a, b) =>
      PROGRESS_ORDER[rolloutState(a, context)] -
      PROGRESS_ORDER[rolloutState(b, context)]
  )
  const { on, off } = rolloutCounts(installs, context)
  return (
    <div
      role="img"
      aria-label={`${on} applied, ${off} not applied, ${installs.length - on - off} not tracked`}
      className={cn(
        'flex h-5 overflow-hidden rounded-sm',
        installs.length === 1 ? 'w-12' : 'w-full'
      )}
    >
      {bars.map((install) => {
        const state = rolloutState(install, context)
        return (
          <div
            key={install.id ?? install.name}
            className={cn(
              'h-full min-w-0 flex-1',
              ROLLOUT_BAR_CLASS[state],
              match.active && !isMatch(install, match) && 'opacity-25'
            )}
          />
        )
      })}
    </div>
  )
}

const INSTALL_CARD_CLASS =
  'flex w-full flex-wrap items-center justify-between gap-3 rounded-md border bg-elevation-2 px-4 py-3 shadow-sm'

const FleetGroupPanelBody = ({
  branch,
  group,
  statusById,
  match,
  context,
}: {
  branch: TFleetBranch
  group: TFleetGroup
  statusById: Record<string, TInstallStatus>
  match: TFleetMatch
  context: TFleetContext
}) => {
  const [search, setSearch] = useState('')
  const [rollout, setRollout] = useState<TBranchRollout>()
  const installs = matchedOf(group.installs, match).filter((install) =>
    matchesFilters(
      install,
      {
        search,
        labels: new Set(),
        rolloutByBranch: { [branch.branchId ?? '']: rollout },
      },
      context
    )
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
            label="Run"
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
              <li
                key={install.id}
                className={cn(INSTALL_CARD_CLASS, 'flex-col !items-stretch')}
              >
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <InstallNameCell install={install} />
                  <span className="flex flex-wrap items-center gap-2">
                    <RolloutBadge state={rolloutState(install, context)} />
                    <InstallStatusAxes status={statusById[install.id ?? '']} />
                  </span>
                </div>
                <InstallFacts install={install} context={context} />
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
  match,
  context,
}: {
  branch: TFleetBranch
  group: TFleetGroup
  statusById: Record<string, TInstallStatus>
  match: TFleetMatch
  context: TFleetContext
}) => (
  <Card className="!p-4 !gap-3 min-w-0">
    <GroupName group={group} />
    <GroupRolloutBars
      installs={group.installs}
      match={match}
      context={context}
    />
    <div className="flex flex-wrap items-center justify-between gap-2">
      <GroupRolloutCounts group={group} match={match} context={context} />
      <Panel
        panelKey={`fleet-group-${branch.key}-${group.key}`}
        size="half"
        heading={group.name ?? 'No install group'}
        triggerButton={{
          variant: 'ghost',
          children: 'View installs',
        }}
      >
        <FleetGroupPanelBody
          branch={branch}
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
  match,
  context,
  controls,
}: {
  branches: TFleetBranch[]
  statusById: Record<string, TInstallStatus>
  match: TFleetMatch
  context: TFleetContext
  controls: TBranchControls
}) => (
  <div className="flex flex-col gap-8 md:gap-10">
    {branches.map((branch) => (
      <div key={branch.key} className="flex flex-col gap-4">
        <BranchHeading
          branch={branch}
          match={match}
          context={context}
          controls={controls}
        />
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {branch.groups.map((group) => (
            <FleetGroupCard
              key={group.key}
              branch={branch}
              group={group}
              statusById={statusById}
              match={match}
              context={context}
            />
          ))}
        </div>
      </div>
    ))}
  </div>
)

const attentionSummary = (items: TAttentionItem[]) =>
  ATTENTION_OPTIONS.map((reason) => ({
    reason,
    count: items.filter((item) => item.reasons.includes(reason)).length,
  }))
    .filter(({ count }) => count)
    .map(
      ({ reason, count }) => `${count} ${attentionLabel(reason).toLowerCase()}`
    )
    .join(' · ')

const AttentionRow = ({
  item,
  statusById,
  context,
}: {
  item: TAttentionItem
  statusById: Record<string, TInstallStatus>
  context: TFleetContext
}) => {
  const { install, reasons, approvals } = item
  const health = install.composite_health_status
  return (
    <li className={cn(INSTALL_CARD_CLASS, 'flex-col !items-stretch')}>
      <div className="flex flex-wrap items-start justify-between gap-3">
        <InstallNameCell install={install} />
        <span className="flex flex-wrap items-center gap-2">
          {reasons.map((reason) => (
            <Status key={reason} status={reason} variant="badge">
              {attentionLabel(reason)}
            </Status>
          ))}
        </span>
      </div>
      {reasons.some((reason) => reason === health) &&
      install.composite_health_status_description ? (
        <Text variant="subtext" theme="neutral">
          {install.composite_health_status_description}
        </Text>
      ) : null}
      {approvals.map((approval) => {
        const href = getApprovalHref(install.org_id ?? '', approval)
        const title = getWorkflowStepTitle(approval.workflow_step)
        return (
          <span key={approval.id} className="flex flex-wrap items-center gap-2">
            {href ? (
              <Link href={href} variant="inline">
                {title}
              </Link>
            ) : (
              <Text variant="subtext">{title}</Text>
            )}
            {approval.type ? (
              <Badge size="sm">{getApprovalType(approval.type)}</Badge>
            ) : null}
          </span>
        )
      })}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <span className="flex flex-wrap items-center gap-2">
          {install.app_branch?.name ? (
            <>
              <Icon variant="GitBranchIcon" size={14} />
              <Text variant="subtext" family="mono">
                {install.app_branch.name}
              </Text>
            </>
          ) : null}
          <RolloutBadge state={rolloutState(install, context)} />
        </span>
        <InstallStatusAxes status={statusById[install.id ?? '']} />
      </div>
      <InstallFacts install={install} context={context} />
    </li>
  )
}

const NeedsAttentionPanelBody = ({
  items,
  statusById,
  context,
}: {
  items: TAttentionItem[]
  statusById: Record<string, TInstallStatus>
  context: TFleetContext
}) => {
  const [search, setSearch] = useState('')
  const [reasons, setReasons] = useState<Set<string>>(new Set())
  const needle = search.trim().toLowerCase()
  const shown = items.filter(
    ({ install, reasons: itemReasons }) =>
      (!needle ||
        install.name?.toLowerCase().includes(needle) ||
        install.id?.toLowerCase().includes(needle)) &&
      (!reasons.size || itemReasons.some((reason) => reasons.has(reason)))
  )
  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center gap-2">
        <SearchInput
          labelClassName="flex-1"
          className="w-full md:min-w-0"
          placeholder="Search installs"
          value={search}
          onChange={setSearch}
        />
        <HealthFilterDropdown
          id="fleet-attention-status"
          label="Status"
          options={ATTENTION_OPTIONS}
          optionLabel={attentionLabel}
          selected={reasons}
          onChange={setReasons}
        />
      </div>
      {shown.length ? (
        <ul className="flex flex-col gap-2">
          {shown.map((item) => (
            <AttentionRow
              key={item.install.id}
              item={item}
              statusById={statusById}
              context={context}
            />
          ))}
        </ul>
      ) : (
        <Text variant="subtext" theme="neutral">
          No installs match.
        </Text>
      )}
    </div>
  )
}

const NeedsAttention = ({
  items,
  statusById,
  context,
}: {
  items: TAttentionItem[]
  statusById: Record<string, TInstallStatus>
  context: TFleetContext
}) =>
  items.length ? (
    <Banner theme="warn">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex flex-col gap-0.5">
          <Text weight="strong">
            {plural(items.length, 'install')}{' '}
            {items.length === 1 ? 'needs' : 'need'} attention
          </Text>
          <Text variant="subtext">{attentionSummary(items)}</Text>
        </div>
        <Panel
          panelKey="fleet-needs-attention"
          size="half"
          heading="Needs attention"
          triggerButton={{
            children: 'Review installs',
          }}
        >
          <NeedsAttentionPanelBody
            items={items}
            statusById={statusById}
            context={context}
          />
        </Panel>
      </div>
    </Banner>
  ) : null

export const InstallFleet = ({
  installs,
  branches = [],
  branchRuns = {},
  branchConfigs = [],
  statusById = {},
  overviewById = {},
  approvals = [],
  configVersionById = {},
  isLoading = false,
  labelColors,
  defaultView = 'table',
}: IInstallFleet) => {
  const [search, setSearch] = useState('')
  const [labels, setLabels] = useState<Set<string>>(new Set())
  const [health, setHealth] = useState<Set<string>>(new Set())
  const [runner, setRunner] = useState<TFleetRunnerFilter>()
  const [scope, setScope] = useState<string>()
  const [runByBranch, setRunByBranch] = useState<
    Record<string, string | undefined>
  >({})
  const [rolloutByBranch, setRolloutByBranch] = useState<
    Record<string, TBranchRollout | undefined>
  >({})
  const [view, setView] = useState<TFleetView>(defaultView)

  const context: TFleetContext = {
    latestConfigIdByBranch: latestConfigIdByBranch(branches),
    runsById: runsIndex(branchRuns),
    overviewById,
    runByBranch,
    scope,
    configVersionById,
    approvalsById: approvalsByInstall(approvals),
  }
  const columns = buildColumns(statusById, context, labelColors)
  const options = labelOptions(installs)
  const componentChoices = componentOptions(overviewById)
  const controls: TBranchControls = {
    runsFor: (branchId) =>
      (branchRuns[branchId] ?? []).map((branchRun) => ({
        value: branchRun.id ?? '',
        title:
          branchRun.vcs_connection_commit?.message?.split('\n')[0]?.trim() ||
          'Branch run',
        sha: branchRun.head_sha ?? branchRun.vcs_connection_commit?.sha,
        createdAt: branchRun.created_at,
        count: countOnRun(installs, branchRun.id ?? '', scope, context),
      })),
    runByBranch,
    onRunChange: (branchId, run) =>
      setRunByBranch((current) => ({ ...current, [branchId]: run })),
    rolloutByBranch,
    onRolloutChange: (branchId, rollout) =>
      setRolloutByBranch((current) => ({ ...current, [branchId]: rollout })),
  }
  const visible = installs.filter((install) =>
    matchesFilters(
      install,
      { search, labels, health, runner, rolloutByBranch },
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
      Object.values(rolloutByBranch).some(Boolean),
  }
  const fleet = groupFleet(installs, branchConfigs, context)

  const resetOtherFilters = () => {
    setScope(undefined)
    setRunner(undefined)
  }

  const clearFilters = () => {
    setSearch('')
    setLabels(new Set())
    setHealth(new Set())
    setRolloutByBranch({})
    resetOtherFilters()
  }

  return (
    <div className="flex flex-col gap-6 md:gap-8">
      {!isLoading ? (
        <NeedsAttention
          items={attentionItems(installs, context.approvalsById)}
          statusById={statusById}
          context={context}
        />
      ) : null}
      {installs[0]?.org_id && installs[0]?.app_id ? (
        <AgentPromptBanner
          heading="Manage this fleet"
          description="Copy this prompt into an agent connected to the Nuon MCP. It reviews the fleet, then waits before changing anything."
          prompt={fleetPrompt(installs[0].org_id, installs[0].app_id)}
        />
      ) : null}
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
            <LabelsFilterDropdown
              options={options}
              selected={labels}
              onChange={setLabels}
            />
          ) : null}
          <HealthFilterDropdown selected={health} onChange={setHealth} />
          <FleetFiltersDropdown
            componentChoices={componentChoices}
            component={scope}
            onComponentChange={setScope}
            runner={runner}
            onRunnerChange={setRunner}
            onReset={resetOtherFilters}
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
          match={match}
          context={context}
          controls={controls}
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
              controls={controls}
            />
          ))}
        </div>
      )}
    </div>
  )
}
