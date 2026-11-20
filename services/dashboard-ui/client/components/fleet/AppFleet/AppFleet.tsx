import type { ReactNode } from 'react'
import { AgentPromptBanner } from '@/components/common/AgentPromptBanner'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { DebouncedSearchInput } from '@/components/common/DeboundedSearch'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { LabelBadge } from '@/components/common/LabelBadge'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Text, type TTextTheme } from '@/components/common/Text'
import { ToggleButton } from '@/components/common/ToggleButton'
import { Panel } from '@/components/surfaces/Panel'
import type {
  TAppFleet,
  TAppFleetBranch,
  TAppFleetCounts,
  TAppFleetFilters,
  TAppFleetGroup,
} from '@/types'
import { cn } from '@/utils/classnames'
import { FleetFiltersDropdown } from './FleetFiltersDropdown'
import { FleetInstallsTable } from './FleetInstallsTable'
import { HealthFilterDropdown } from './HealthFilterDropdown'
import { LabelsFilterDropdown } from './LabelsFilterDropdown'
import {
  countText,
  healthSet,
  plural,
  rolloutSegments,
  type TFleetView,
} from './fleet-utils'

export type TRenderGroupInstalls = (
  branch: TAppFleetBranch,
  group: TAppFleetGroup
) => ReactNode

export interface IAppFleet {
  fleet?: TAppFleet
  orgId?: string
  appId?: string
  isLoading?: boolean
  filters: TAppFleetFilters
  isFiltered: boolean
  onFiltersChange: (filters: Partial<TAppFleetFilters>) => void
  onClearFilters: () => void
  view: TFleetView
  onViewChange: (view: TFleetView) => void
  labelOptions?: string[]
  componentChoices?: { value: string; label: string }[]
  labelColors?: Record<string, string>
  renderGroupInstalls: TRenderGroupInstalls
  emptyAction?: ReactNode
}

const ROLLOUT_BAR_CLASS = {
  applied: 'bg-green-600 dark:bg-green-500',
  behind: 'bg-orange-500 dark:bg-orange-400',
  unknown: 'bg-cool-grey-300 dark:bg-dark-grey-600',
} as const

const ROLLOUT_COUNT = {
  applied: { theme: 'success', label: 'on' },
  behind: { theme: 'warn', label: 'not on' },
  unknown: { theme: 'neutral', label: 'not tracked' },
} as const

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

const groupKey = (branch: TAppFleetBranch, group: TAppFleetGroup) =>
  `${branch.id ?? 'no-branch'}-${group.name ?? 'no-group'}`

const RolloutCounts = ({
  counts,
  isFiltered,
}: {
  counts?: TAppFleetCounts
  isFiltered: boolean
}) => {
  const segments = rolloutSegments(counts)
  if (isFiltered || !segments.length)
    return (
      <Text variant="subtext" theme="neutral">
        {countText(counts, isFiltered)}
      </Text>
    )
  return (
    <span className="flex flex-wrap items-center gap-1">
      {segments.map(({ state, count }) => (
        <Badge key={state} theme={ROLLOUT_COUNT[state].theme}>
          {count} {ROLLOUT_COUNT[state].label}
        </Badge>
      ))}
    </span>
  )
}

const RolloutBar = ({ counts }: { counts?: TAppFleetCounts }) => {
  const segments = rolloutSegments(counts)
  return (
    <div
      role="img"
      aria-label={`${counts?.applied ?? 0} applied, ${counts?.behind ?? 0} not applied, ${counts?.unknown ?? 0} not tracked`}
      className="flex h-5 w-full overflow-hidden rounded-sm bg-cool-grey-100 dark:bg-dark-grey-800"
    >
      {segments.map(({ state, count }) => (
        <div
          key={state}
          className={cn('h-full min-w-1', ROLLOUT_BAR_CLASS[state])}
          style={{ flexGrow: count }}
        />
      ))}
    </div>
  )
}

const GroupName = ({ group }: { group: TAppFleetGroup }) =>
  group.name ? (
    <Text variant="subtext" weight="strong" family="mono">
      {group.name}
    </Text>
  ) : (
    <Text variant="subtext" weight="strong">
      No install group
    </Text>
  )

const GroupLabels = ({
  group,
  labelColors,
}: {
  group: TAppFleetGroup
  labelColors?: Record<string, string>
}) => (
  <>
    {Object.entries(group.match_labels ?? {})
      .sort(([a], [b]) => a.localeCompare(b))
      .map(([key, value]) => (
        <LabelBadge
          key={key}
          labelKey={key}
          labelValue={value}
          customColor={labelColors?.[key]}
        />
      ))}
  </>
)

const GroupPanel = ({
  branch,
  group,
  trigger,
  renderGroupInstalls,
}: {
  branch: TAppFleetBranch
  group: TAppFleetGroup
  trigger: string
  renderGroupInstalls: TRenderGroupInstalls
}) => (
  <Panel
    panelKey={`fleet-group-${groupKey(branch, group)}`}
    size="half"
    heading={group.name ?? 'No install group'}
    triggerButton={{ variant: 'ghost', children: trigger }}
  >
    {renderGroupInstalls(branch, group)}
  </Panel>
)

const BranchHeading = ({
  branch,
  orgId,
  appId,
  isFiltered,
}: {
  branch: TAppFleetBranch
  orgId?: string
  appId?: string
  isFiltered: boolean
}) => (
  <div className="flex flex-wrap items-center gap-2">
    {branch.id ? (
      <>
        <Icon variant="GitBranchIcon" size={16} />
        <Text variant="base" weight="strong" family="mono">
          <Link
            href={`/${orgId}/apps/${appId}/branches/${branch.id}`}
            variant="inline"
          >
            {branch.name ?? branch.id}
          </Link>
        </Text>
      </>
    ) : (
      <Text variant="base" weight="strong">
        Not on a branch
      </Text>
    )}
    <RolloutCounts counts={branch.counts} isFiltered={isFiltered} />
  </div>
)

const FleetGroupSection = ({
  branch,
  group,
  orgId,
  appId,
  isFiltered,
  labelColors,
  renderGroupInstalls,
}: {
  branch: TAppFleetBranch
  group: TAppFleetGroup
  orgId?: string
  appId?: string
  isFiltered: boolean
  labelColors?: Record<string, string>
  renderGroupInstalls: TRenderGroupInstalls
}) => (
  <div className="flex flex-col gap-3">
    <div className="flex flex-wrap items-center gap-2">
      <GroupName group={group} />
      <GroupLabels group={group} labelColors={labelColors} />
      <RolloutCounts counts={group.counts} isFiltered={isFiltered} />
    </div>
    {group.installs?.length ? (
      <>
        <FleetInstallsTable
          installs={group.installs}
          orgId={orgId}
          appId={appId}
          labelColors={labelColors}
        />
        {group.has_more ? (
          <div className="flex justify-center">
            <GroupPanel
              branch={branch}
              group={group}
              trigger={`View all ${group.counts?.matched ?? 0}`}
              renderGroupInstalls={renderGroupInstalls}
            />
          </div>
        ) : null}
      </>
    ) : null}
  </div>
)

const FleetGroupCard = ({
  branch,
  group,
  isFiltered,
  renderGroupInstalls,
}: {
  branch: TAppFleetBranch
  group: TAppFleetGroup
  isFiltered: boolean
  renderGroupInstalls: TRenderGroupInstalls
}) => (
  <Card className="!p-4 !gap-3 min-w-0">
    <GroupName group={group} />
    <RolloutBar counts={group.counts} />
    <div className="flex flex-wrap items-center justify-between gap-2">
      <RolloutCounts counts={group.counts} isFiltered={isFiltered} />
      {group.counts?.matched ? (
        <GroupPanel
          branch={branch}
          group={group}
          trigger="View installs"
          renderGroupInstalls={renderGroupInstalls}
        />
      ) : null}
    </div>
  </Card>
)

const FleetSummary = ({
  fleet,
  isLoading,
}: {
  fleet?: TAppFleet
  isLoading?: boolean
}) => {
  const summary = fleet?.summary
  const items: { label: string; value?: number; theme?: TTextTheme }[] = [
    { label: 'Installs', value: summary?.total },
    { label: 'Latest run applied', value: summary?.applied, theme: 'success' },
    { label: 'Latest run not applied', value: summary?.behind, theme: 'warn' },
    {
      label: 'Runner not active',
      value: summary?.runner_not_active,
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
            {item.value ?? 0}
          </Text>
        </LabeledValue>
      ))}
    </div>
  )
}

export const AppFleet = ({
  fleet,
  orgId,
  appId,
  isLoading = false,
  filters,
  isFiltered,
  onFiltersChange,
  onClearFilters,
  view,
  onViewChange,
  labelOptions = [],
  componentChoices = [],
  labelColors,
  renderGroupInstalls,
  emptyAction,
}: IAppFleet) => {
  const total = fleet?.summary?.total ?? 0
  const matched = fleet?.summary?.matched ?? 0
  const branches = (fleet?.branches ?? []).filter(
    (branch) => !isFiltered || (branch.counts?.matched ?? 0) > 0
  )

  return (
    <div className="flex flex-col gap-6 md:gap-8">
      {orgId && appId ? (
        <AgentPromptBanner
          heading="Manage this fleet"
          description="Copy this prompt into an agent connected to the Nuon MCP. It reviews the fleet, then waits before changing anything."
          prompt={fleetPrompt(orgId, appId)}
        />
      ) : null}

      <div className="flex flex-wrap items-start justify-between gap-4">
        <FleetSummary fleet={fleet} isLoading={isLoading} />
        <ToggleButton<TFleetView>
          label="Fleet layout"
          value={view}
          onChange={onViewChange}
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
          <DebouncedSearchInput
            labelClassName="w-full md:w-fit"
            className="w-full md:w-fit"
            placeholder="Search by name or ID..."
          />
          {!isLoading ? (
            <Text variant="subtext" theme="neutral" aria-live="polite">
              {isFiltered
                ? `${matched} of ${plural(total, 'install')}`
                : plural(total, 'install')}
            </Text>
          ) : null}
        </div>
        <div className="flex flex-wrap items-center gap-2">
          {labelOptions.length ? (
            <LabelsFilterDropdown
              options={labelOptions}
              selected={healthSet(filters.labels)}
              onChange={(selected) =>
                onFiltersChange({
                  labels: [...selected].join(',') || undefined,
                })
              }
            />
          ) : null}
          <HealthFilterDropdown
            selected={healthSet(filters.health)}
            onChange={(selected) =>
              onFiltersChange({ health: [...selected].join(',') || undefined })
            }
          />
          <FleetFiltersDropdown
            componentChoices={componentChoices}
            component={filters.component_id}
            onComponentChange={(component_id) =>
              onFiltersChange({ component_id })
            }
            runner={filters.runner_status}
            onRunnerChange={(runner) =>
              onFiltersChange({ runner_status: runner })
            }
            onReset={() =>
              onFiltersChange({
                runner_status: undefined,
                component_id: undefined,
              })
            }
          />
        </div>
      </div>

      {isLoading ? (
        <FleetInstallsTable installs={[]} isLoading />
      ) : !total ? (
        <EmptyState
          variant="table"
          emptyTitle="No installs yet"
          emptyMessage="Installs will appear here once you create one for this app."
          action={emptyAction}
        />
      ) : !matched ? (
        <EmptyState
          variant="search"
          emptyTitle="No matching installs"
          emptyMessage="No installs match the current filters. Clear them to see the whole fleet."
          action={<Button onClick={onClearFilters}>Clear filters</Button>}
        />
      ) : view === 'board' ? (
        <div className="flex flex-col gap-8 md:gap-10">
          {branches.map((branch) => (
            <div key={branch.id ?? 'no-branch'} className="flex flex-col gap-4">
              <BranchHeading
                branch={branch}
                orgId={orgId}
                appId={appId}
                isFiltered={isFiltered}
              />
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
                {(branch.groups ?? []).map((group) => (
                  <FleetGroupCard
                    key={groupKey(branch, group)}
                    branch={branch}
                    group={group}
                    isFiltered={isFiltered}
                    renderGroupInstalls={renderGroupInstalls}
                  />
                ))}
              </div>
            </div>
          ))}
        </div>
      ) : (
        <div className="flex flex-col gap-8 md:gap-10">
          {branches.map((branch) => (
            <div key={branch.id ?? 'no-branch'} className="flex flex-col gap-4">
              <BranchHeading
                branch={branch}
                orgId={orgId}
                appId={appId}
                isFiltered={isFiltered}
              />
              <div className="flex flex-col gap-6">
                {(branch.groups ?? [])
                  .filter(
                    (group) => !isFiltered || (group.counts?.matched ?? 0) > 0
                  )
                  .map((group) => (
                    <FleetGroupSection
                      key={groupKey(branch, group)}
                      branch={branch}
                      group={group}
                      orgId={orgId}
                      appId={appId}
                      isFiltered={isFiltered}
                      labelColors={labelColors}
                      renderGroupInstalls={renderGroupInstalls}
                    />
                  ))}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
