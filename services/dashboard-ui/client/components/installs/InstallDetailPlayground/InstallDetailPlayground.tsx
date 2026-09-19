import { useState } from 'react'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Dropdown } from '@/components/common/Dropdown'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { LabelBadge } from '@/components/common/LabelBadge'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Markdown } from '@/components/common/Markdown'
import { Menu } from '@/components/common/Menu'
import { PropertyGrid } from '@/components/common/PropertyGrid'
import { SearchInput } from '@/components/common/SearchInput'
import { Status } from '@/components/common/Status'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import { HealthTimelineComponent } from '@/components/install-health/HealthTimeline'
import { cn } from '@/utils/classnames'
import { humanize } from '@/utils/string-utils'
import type {
  TPlaygroundInstall,
  TActivityEvent,
  TActivityEventType,
  TAppBranchSource,
  TBranchCommitRef,
  TBranchTrackingStatus,
  TLagItem,
  TDriftedObject,
  TStackVersion,
  TRoleEntry,
  TSandboxInfo,
  TComponentEntry,
  TImageEntry,
  TActionEntry,
  TRunbookEntry,
  TInputEntry,
  TConfigFileInfo,
  TConfigurationChange,
  TConfigurationVersion,
  TOverrideEntry,
  TPolicyReportEntry,
  TRunnerInfo,
} from './types'

// ─── Top-level navigation ────────────────────────────────────────────────────

type TTopTab =
  | 'overview'
  | 'resources'
  | 'operations'
  | 'configuration'
  | 'activity'

const TOP_TAB_LABELS: Record<TTopTab, string> = {
  overview: 'Overview',
  resources: 'Resources',
  operations: 'Operations',
  configuration: 'Configuration',
  activity: 'Activity',
}

// ─── Activity filter state ────────────────────────────────────────────────────

type TActivityFilter = {
  search: string
  status: string
  type: string
  component: string
  date: string
}

const DEFAULT_ACTIVITY_FILTER: TActivityFilter = {
  search: '',
  status: 'all',
  type: 'all',
  component: 'all',
  date: 'all',
}

// ─── Header ──────────────────────────────────────────────────────────────────

interface IInstallPlaygroundHeader {
  install: TPlaygroundInstall
  onNavigate: (
    tab: TTopTab,
    opts?: {
      activityType?: string
      resourcesTab?: string
      configurationTab?: string
    }
  ) => void
}

const ConfigurationSummaryRow = ({
  install,
  onNavigate,
}: IInstallPlaygroundHeader) => {
  const branchHref = `/${install.orgId}/apps/${install.appId}/branches/${install.branchTracking.branchId}`

  return (
    <Card className="!p-3 !gap-2 !shadow-none w-full">
      <div className="flex items-center gap-x-6 gap-y-1 flex-wrap">
        <span className="flex items-center gap-1.5">
          <Text as="span" variant="subtext" theme="neutral">
            App
          </Text>
          <Link
            href={`/${install.orgId}/apps/${install.appId}`}
            textVariant="subtext"
          >
            {install.appName}
          </Link>
        </span>

        <span className="flex items-center gap-1.5">
          <Text as="span" variant="subtext" theme="neutral">
            App branch
          </Text>
          <Icon
            variant="GitBranchIcon"
            size={13}
            className="text-cool-grey-400"
          />
          <Link href={branchHref} textVariant="subtext">
            <Text as="span" variant="subtext" family="mono">
              {install.branchTracking.targetBranch}
            </Text>
          </Link>
        </span>
      </div>

      <div className="flex items-center gap-x-6 gap-y-1 flex-wrap">
        <span className="flex items-center gap-1.5">
          <Text as="span" variant="subtext" theme="neutral">
            Managed by
          </Text>
          {install.isManagedByConfig ? (
            <>
              <Icon
                variant="FileCodeIcon"
                size={13}
                className="text-cool-grey-400"
              />
              <Text as="span" variant="subtext">
                Install config
              </Text>
              {install.configFilePath && (
                <Button
                  variant="ghost"
                  size="sm"
                  className="!px-1"
                  onClick={() =>
                    onNavigate('configuration', {
                      configurationTab: 'configFile',
                    })
                  }
                >
                  <Badge size="sm" variant="code" theme="neutral">
                    {install.configFilePath}
                  </Badge>
                </Button>
              )}
            </>
          ) : (
            <Text as="span" variant="subtext">
              Dashboard
            </Text>
          )}
        </span>
      </div>
    </Card>
  )
}

const InstallPlaygroundHeader = ({
  install,
  onNavigate,
}: IInstallPlaygroundHeader) => (
  <header className="flex flex-col gap-4 px-4 pt-8 pb-6 md:px-6 md:pt-10 border-b shrink-0">
    <div className="flex flex-col gap-1.5 min-w-0">
      <Text variant="h3" weight="stronger" level={1}>
        {install.name}
      </Text>
      <div className="flex items-center gap-3 flex-wrap">
        <Text variant="subtext" theme="neutral" family="mono">
          {install.id}
        </Text>
        <Text variant="subtext" theme="info">
          Updated{' '}
          <Time time={install.updatedAt} format="relative" variant="subtext" />
        </Text>
      </div>
    </div>

    <ConfigurationSummaryRow install={install} onNavigate={onNavigate} />

    {/* Status sits on the labels row so it starts below the identity stack. */}
    <div className="flex items-start justify-between gap-6 flex-wrap">
      <div className="flex items-center gap-2 flex-wrap min-w-0 flex-1">
        {Object.entries(install.labels).map(([k, v]) => (
          <LabelBadge key={k} size="sm" labelKey={k} labelValue={v} />
        ))}
      </div>

      <div className="w-full md:w-auto md:min-w-64 shrink-0">
        <InstallStatusCard install={install} onNavigate={onNavigate} />
      </div>
    </div>
  </header>
)

// ─── Header status card ───────────────────────────────────────────────────────

interface IInstallStatusCard {
  install: TPlaygroundInstall
  onNavigate: (
    tab: TTopTab,
    opts?: {
      activityType?: string
      resourcesTab?: string
      configurationTab?: string
    }
  ) => void
}

const InstallStatusCard = ({ install, onNavigate }: IInstallStatusCard) => {
  const { activity, resources, componentStatus } = install

  // Deployments: in-flight app_branch_run and deploy workflows
  const runningUpdates = activity.filter(
    (e) =>
      (e.type === 'app_branch_run' || e.type === 'deploy') &&
      (e.status === 'in-progress' || e.status === 'pending')
  )
  const failedUpdates = activity.filter(
    (e) =>
      (e.type === 'app_branch_run' || e.type === 'deploy') &&
      e.status === 'error'
  )
  const updatesStatus =
    failedUpdates.length > 0
      ? 'error'
      : runningUpdates.length > 0
        ? 'in-progress'
        : 'active'
  const updatesLabel =
    failedUpdates.length > 0
      ? `${failedUpdates.length} ${failedUpdates.length === 1 ? 'failure' : 'failures'}`
      : runningUpdates.length > 0
        ? `${runningUpdates.length} in progress`
        : 'All deployed'

  // Resources: non-image components (images excluded per spec)
  const errorComponents = resources.components.filter(
    (c) => c.status === 'error' || c.status === 'warn'
  )
  const pendingComponents = resources.components.filter(
    (c) => c.status === 'pending' || c.status === 'in-progress'
  )
  const resourcesStatus =
    errorComponents.length > 0
      ? errorComponents[0].status
      : pendingComponents.length > 0
        ? 'pending'
        : 'active'
  const resourcesLabel =
    errorComponents.length > 0
      ? `${errorComponents.length} ${errorComponents.length === 1 ? 'issue' : 'issues'}`
      : pendingComponents.length > 0
        ? `${pendingComponents.length} pending`
        : 'All deployed'

  // Health checks reflect component health, separate from deploy state.
  const healthStatus = componentStatus
  const healthLabel =
    healthStatus === 'active'
      ? 'Healthy'
      : healthStatus === 'warn'
        ? 'Degraded'
        : healthStatus === 'error'
          ? 'Unhealthy'
          : 'Checking'

  return (
    <Card className="!p-4 !gap-2" aria-label="Install status summary">
      <Text variant="body" weight="strong">
        Status
      </Text>
      <Button
        variant="ghost"
        size="sm"
        className="!px-0 w-full justify-between"
        onClick={() => onNavigate('activity', { activityType: 'deployments' })}
        aria-label={`Deployments: ${updatesLabel}. Navigate to activity.`}
      >
        <span className="flex items-center gap-1.5">
          <Icon
            variant="ArrowsClockwiseIcon"
            size={13}
            className="text-cool-grey-400"
          />
          <Text as="span" variant="subtext" weight="strong" theme="neutral">
            Deployments
          </Text>
        </span>
        <Status status={updatesStatus} variant="badge">
          {updatesLabel}
        </Status>
      </Button>

      <Button
        variant="ghost"
        size="sm"
        className="!px-0 w-full justify-between"
        onClick={() => onNavigate('resources', { resourcesTab: 'components' })}
        aria-label={`Resources: ${resourcesLabel}. Navigate to resources.`}
      >
        <span className="flex items-center gap-1.5">
          <Icon variant="CardsIcon" size={13} className="text-cool-grey-400" />
          <Text as="span" variant="subtext" weight="strong" theme="neutral">
            Resources
          </Text>
        </span>
        <Status status={resourcesStatus} variant="badge">
          {resourcesLabel}
        </Status>
      </Button>

      <Button
        variant="ghost"
        size="sm"
        className="!px-0 w-full justify-between"
        onClick={() => onNavigate('resources', { resourcesTab: 'health' })}
        aria-label={`Health checks: ${healthLabel}. Navigate to resources.`}
      >
        <span className="flex items-center gap-1.5">
          <Icon variant="PulseIcon" size={13} className="text-cool-grey-400" />
          <Text as="span" variant="subtext" weight="strong" theme="neutral">
            Health checks
          </Text>
        </span>
        <Status status={healthStatus} variant="badge">
          {healthLabel}
        </Status>
      </Button>
    </Card>
  )
}

// ─── Install branch tracking card ─────────────────────────────────────────────

const TRACKING_STATUS_MAP: Record<
  TBranchTrackingStatus,
  { statusValue: string; label: string }
> = {
  current: { statusValue: 'active', label: 'Current' },
  pending: { statusValue: 'pending', label: 'Pending' },
  updating: { statusValue: 'in-progress', label: 'Updating' },
}

const CommitRef = ({
  commit,
  label,
}: {
  commit?: TBranchCommitRef
  label: string
}) => (
  <div className="flex flex-col gap-1">
    <Text variant="subtext" weight="strong" theme="neutral">
      {label}
    </Text>
    {!commit ? (
      <Text variant="subtext" theme="neutral">
        None
      </Text>
    ) : (
      <div className="flex flex-col gap-1 mt-0.5">
        <span className="flex items-center gap-1.5 flex-wrap">
          <Badge size="sm" variant="code" theme="neutral">
            {commit.sha.slice(0, 8)}
          </Badge>
          {commit.runStatus && (
            <Status status={commit.runStatus} variant="badge" />
          )}
        </span>
        {commit.message && (
          <Text variant="subtext" theme="neutral" className="truncate">
            {commit.message}
          </Text>
        )}
        {(commit.author || commit.createdAt) && (
          <span className="flex items-center gap-1 flex-wrap">
            {commit.author && (
              <Text as="span" variant="subtext" theme="neutral">
                by {commit.author}
              </Text>
            )}
            {commit.author && commit.createdAt && (
              <Text as="span" variant="subtext" theme="neutral">
                ·
              </Text>
            )}
            {commit.createdAt && (
              <Time
                time={commit.createdAt}
                format="relative"
                variant="subtext"
                theme="neutral"
              />
            )}
          </span>
        )}
      </div>
    )}
  </div>
)

const InstallBranchTrackingCard = ({
  install,
}: {
  install: TPlaygroundInstall
}) => {
  const tracking = install.branchTracking
  const { statusValue, label } = TRACKING_STATUS_MAP[tracking.status]
  const repoHref =
    tracking.repo && !tracking.repo.startsWith('http')
      ? `https://github.com/${tracking.repo}`
      : tracking.repo
  const showDirectory =
    tracking.directory &&
    tracking.directory !== '.' &&
    tracking.directory !== '/'
  const branchHref = `/${install.orgId}/apps/${install.appId}/branches/${tracking.branchId}`

  return (
    <Card className="!p-4 !gap-4">
      <div className="flex items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          <Icon
            variant="GitBranchIcon"
            size={14}
            className="text-cool-grey-400"
          />
          <Text variant="body" weight="strong">
            Tracking
          </Text>
        </div>
        <Status status={statusValue} variant="badge">
          {label}
        </Status>
      </div>

      <div className="flex flex-wrap gap-x-8 gap-y-3">
        <LabeledValue label="Target branch">
          <Link href={branchHref} textVariant="subtext">
            <span className="flex items-center gap-1.5">
              <Icon
                variant="GitBranchIcon"
                size={13}
                className="text-cool-grey-400 shrink-0"
              />
              <Text as="span" variant="subtext" family="mono">
                {tracking.targetBranch}
              </Text>
            </span>
          </Link>
        </LabeledValue>
        {tracking.repo && repoHref && (
          <LabeledValue label="Repository">
            <Link href={repoHref} isExternal textVariant="subtext">
              {tracking.repo}
            </Link>
          </LabeledValue>
        )}
        {tracking.gitBranch && (
          <LabeledValue label="Git branch">
            <Text variant="subtext" family="mono">
              {tracking.gitBranch}
            </Text>
          </LabeledValue>
        )}
        {showDirectory && (
          <LabeledValue label="Directory">
            <Text variant="subtext" family="mono">
              {tracking.directory}
            </Text>
          </LabeledValue>
        )}
      </div>

      <hr className="-mx-4" />

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <CommitRef
          label="Expected / latest run"
          commit={tracking.expectedCommit}
        />
        <CommitRef label="Currently applied" commit={tracking.appliedCommit} />
      </div>
    </Card>
  )
}

// ─── Overview tab ─────────────────────────────────────────────────────────────

const LagRow = ({ item }: { item: TLagItem }) => (
  <div className="flex items-center justify-between gap-3 py-1.5 min-w-0">
    <Text variant="subtext" className="truncate">
      {item.name}
    </Text>
    <div className="flex items-center gap-2 shrink-0">
      {item.isCurrent ? (
        <Status status="active" variant="badge">
          Current
        </Status>
      ) : (
        <>
          <Badge size="sm" variant="code" theme="neutral">
            {item.appliedVersion.slice(0, 8)}
          </Badge>
          <Icon
            variant="ArrowRightIcon"
            size={12}
            className="text-cool-grey-400"
          />
          <Badge size="sm" variant="code" theme="warn">
            {item.expectedVersion.slice(0, 8)}
          </Badge>
        </>
      )}
    </div>
  </div>
)

const DriftRow = ({ obj }: { obj: TDriftedObject }) => (
  <div className="flex items-center justify-between gap-3 py-1.5">
    <div className="flex items-center gap-2">
      <Status status="warn" isWithoutText variant="timeline" iconSize={14} />
      <Text variant="subtext">
        {obj.targetType === 'sandbox'
          ? 'Sandbox'
          : (obj.componentName ?? 'Component')}
      </Text>
    </div>
    <Badge size="sm" theme="warn">
      Drift detected
    </Badge>
  </div>
)

const ConfigLagCard = ({ install }: { install: TPlaygroundInstall }) => {
  const { configLag } = install
  const isBranchMoving = install.branchTracking.status !== 'current'

  const hasLag =
    (configLag.stack && !configLag.stack.isCurrent) ||
    (configLag.sandbox && !configLag.sandbox.isCurrent) ||
    configLag.components.some((c) => !c.isCurrent) ||
    configLag.images.some((i) => !i.isCurrent)

  return (
    <Card className="!p-4 !gap-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Icon
            variant="ArrowsClockwiseIcon"
            size={14}
            className="text-cool-grey-400"
          />
          <Text variant="body" weight="strong">
            {isBranchMoving
              ? 'Pending resource changes'
              : 'Config application status'}
          </Text>
        </div>
        {hasLag ? (
          <Badge size="sm" theme={isBranchMoving ? 'neutral' : 'warn'}>
            {isBranchMoving ? 'Branch changing' : 'Lagging'}
          </Badge>
        ) : (
          <Badge size="sm" theme="success">
            All current
          </Badge>
        )}
      </div>

      {isBranchMoving && (
        <Text variant="subtext" theme="neutral">
          Resources will update once the new branch target is applied.
        </Text>
      )}

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-x-8 gap-y-1">
        {configLag.stack && (
          <div className="flex flex-col">
            <Text variant="subtext" weight="strong" theme="neutral">
              Stack
            </Text>
            <LagRow item={configLag.stack} />
          </div>
        )}
        {configLag.sandbox && (
          <div className="flex flex-col">
            <Text variant="subtext" weight="strong" theme="neutral">
              Sandbox
            </Text>
            <LagRow item={configLag.sandbox} />
          </div>
        )}
        {configLag.components.length > 0 && (
          <div className="flex flex-col sm:col-span-2">
            <Text variant="subtext" weight="strong" theme="neutral">
              Components
            </Text>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-x-8">
              {configLag.components.map((c) => (
                <LagRow key={c.name} item={c} />
              ))}
            </div>
          </div>
        )}
        {configLag.images.length > 0 && (
          <div className="flex flex-col sm:col-span-2">
            <Text variant="subtext" weight="strong" theme="neutral">
              Images
            </Text>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-x-8">
              {configLag.images.map((img) => (
                <LagRow key={img.name} item={img} />
              ))}
            </div>
          </div>
        )}
      </div>
    </Card>
  )
}

interface IOverviewTab {
  install: TPlaygroundInstall
  onNavigate: IInstallPlaygroundHeader['onNavigate']
}

const OverviewTab = ({ install, onNavigate }: IOverviewTab) => {
  const hasDrift = install.driftedObjects.length > 0

  return (
    <div className="flex flex-col gap-4 p-4">
      {/* Per-component health lives in Resources -> Health, not here. */}
      <Card className="!p-4 !gap-4">
        <HealthTimelineComponent
          scope="install"
          days={install.health.days}
          daily={install.health.daily}
          uptimePercent={install.health.uptime_percent}
          observedSeconds={install.health.observed_seconds}
          currentHealth={install.health.current_health}
          headerAction={
            <Button
              variant="ghost"
              size="sm"
              className="!px-0"
              onClick={() =>
                onNavigate('resources', { resourcesTab: 'health' })
              }
            >
              <span className="flex items-center gap-1.5">
                <Text as="span" variant="subtext" theme="info">
                  View health
                </Text>
                <Icon variant="ArrowRightIcon" size={13} />
              </span>
            </Button>
          }
        />
      </Card>

      {install.readme && (
        <Card className="!p-4 !gap-4">
          <Markdown content={install.readme} mode="install" />
        </Card>
      )}

      <ConfigLagCard install={install} />

      {/* Infrastructure drift — kept distinct from config lag */}
      <Card className="!p-4 !gap-4">
        <div className="flex items-center gap-2">
          <Icon
            variant="FileDashedIcon"
            size={14}
            className="text-cool-grey-400"
          />
          <Text variant="body" weight="strong">
            Infrastructure drift
          </Text>
        </div>
        {hasDrift ? (
          <div className="flex flex-col divide-y">
            {install.driftedObjects.map((obj) => (
              <DriftRow key={obj.id} obj={obj} />
            ))}
          </div>
        ) : (
          <div className="flex items-center gap-1.5">
            <Status
              status="active"
              isWithoutText
              variant="timeline"
              iconSize={14}
            />
            <Text variant="subtext" theme="neutral">
              No drift detected
            </Text>
          </div>
        )}
      </Card>
    </div>
  )
}

// ─── Activity tab ─────────────────────────────────────────────────────────────

const ACTIVITY_TYPE_LABELS: Record<TActivityEventType, string> = {
  app_branch_run: 'App branch run',
  deploy: 'Deploy',
  config_update: 'Config',
  inputs_update: 'Inputs',
  stack_update: 'Stack',
  drift_scan: 'Drift scan',
}

const DATE_FILTER_LABELS: Record<string, string> = {
  '24h': 'Last 24 hours',
  '7d': 'Last 7 days',
  '30d': 'Last 30 days',
}

const DATE_FILTER_MS: Record<string, number> = {
  '24h': 86_400_000,
  '7d': 604_800_000,
  '30d': 2_592_000_000,
}

const sourceCaption = (source?: TAppBranchSource): string | undefined => {
  if (!source) return undefined
  switch (source.type) {
    case 'push':
      return [
        source.branch,
        source.sha && source.sha.slice(0, 8),
        source.author && `by ${source.author}`,
      ]
        .filter(Boolean)
        .join(' · ')
    case 'pr':
      return [
        source.branch,
        source.prNumber && `PR #${source.prNumber}`,
        source.author && `by ${source.author}`,
      ]
        .filter(Boolean)
        .join(' · ')
    case 'tag':
      return [`tag ${source.tag}`, source.author && `by ${source.author}`]
        .filter(Boolean)
        .join(' · ')
    case 'commit':
      return [
        source.sha && source.sha.slice(0, 8),
        source.author && `by ${source.author}`,
      ]
        .filter(Boolean)
        .join(' · ')
    case 'manual':
      return ['manual trigger', source.author && `by ${source.author}`]
        .filter(Boolean)
        .join(' · ')
  }
}

const filterActivity = (
  events: TActivityEvent[],
  filter: TActivityFilter
): TActivityEvent[] =>
  events.filter((e) => {
    if (filter.search) {
      const q = filter.search.toLowerCase()
      if (
        !e.title.toLowerCase().includes(q) &&
        !e.details?.toLowerCase().includes(q) &&
        !e.componentName?.toLowerCase().includes(q)
      )
        return false
    }
    if (filter.status !== 'all' && e.status !== filter.status) return false
    // 'deployments' is a grouped filter set by the status strip
    if (
      filter.type === 'deployments' &&
      e.type !== 'app_branch_run' &&
      e.type !== 'deploy'
    )
      return false
    if (
      filter.type !== 'all' &&
      filter.type !== 'deployments' &&
      e.type !== filter.type
    )
      return false
    if (filter.component !== 'all' && e.componentName !== filter.component)
      return false
    if (filter.date !== 'all') {
      const maxMs = DATE_FILTER_MS[filter.date] ?? 0
      if (Date.now() - new Date(e.createdAt).getTime() > maxMs) return false
    }
    return true
  })

interface IActivityTab {
  install: TPlaygroundInstall
  filter: TActivityFilter
  onFilterChange: (f: TActivityFilter) => void
}

const ActivityTab = ({ install, filter, onFilterChange }: IActivityTab) => {
  const set = (patch: Partial<TActivityFilter>) =>
    onFilterChange({ ...filter, ...patch })

  const allComponents = Array.from(
    new Set(install.activity.map((e) => e.componentName).filter(Boolean))
  ) as string[]

  const allStatuses = Array.from(new Set(install.activity.map((e) => e.status)))

  const filtered = filterActivity(install.activity, filter)

  const hasActiveFilters =
    filter.search !== '' ||
    filter.status !== 'all' ||
    filter.type !== 'all' ||
    filter.component !== 'all' ||
    filter.date !== 'all'

  return (
    <div className="flex flex-col min-h-0">
      {/* Filter bar */}
      <div className="flex items-center flex-wrap gap-2 px-4 py-3 border-b bg-cool-grey-50 dark:bg-dark-grey-800 shrink-0">
        <SearchInput
          aria-label="Search activity"
          value={filter.search}
          onChange={(v) => set({ search: v })}
          placeholder="Search activity…"
          labelClassName="flex-1 min-w-44"
        />

        <Dropdown
          id="act-filter-status"
          variant="secondary"
          size="sm"
          buttonText={
            filter.status === 'all' ? 'Status' : humanize(filter.status)
          }
          isActive={filter.status !== 'all'}
        >
          <Menu>
            <Button isMenuButton onClick={() => set({ status: 'all' })}>
              All statuses
            </Button>
            {allStatuses.map((s) => (
              <Button isMenuButton key={s} onClick={() => set({ status: s })}>
                {humanize(s)}
              </Button>
            ))}
          </Menu>
        </Dropdown>

        <Dropdown
          id="act-filter-type"
          variant="secondary"
          size="sm"
          buttonText={
            filter.type === 'all'
              ? 'Type'
              : filter.type === 'deployments'
                ? 'Deployments'
                : (ACTIVITY_TYPE_LABELS[filter.type as TActivityEventType] ??
                  humanize(filter.type))
          }
          isActive={filter.type !== 'all'}
        >
          <Menu>
            <Button isMenuButton onClick={() => set({ type: 'all' })}>
              All types
            </Button>
            {(Object.keys(ACTIVITY_TYPE_LABELS) as TActivityEventType[]).map(
              (t) => (
                <Button isMenuButton key={t} onClick={() => set({ type: t })}>
                  {ACTIVITY_TYPE_LABELS[t]}
                </Button>
              )
            )}
          </Menu>
        </Dropdown>

        {allComponents.length > 0 && (
          <Dropdown
            id="act-filter-component"
            variant="secondary"
            size="sm"
            buttonText={
              filter.component === 'all' ? 'Resource' : filter.component
            }
            isActive={filter.component !== 'all'}
          >
            <Menu>
              <Button isMenuButton onClick={() => set({ component: 'all' })}>
                All resources
              </Button>
              {allComponents.map((c) => (
                <Button
                  isMenuButton
                  key={c}
                  onClick={() => set({ component: c })}
                >
                  {c}
                </Button>
              ))}
            </Menu>
          </Dropdown>
        )}

        <Dropdown
          id="act-filter-date"
          variant="secondary"
          size="sm"
          buttonText={
            filter.date === 'all'
              ? 'Date'
              : (DATE_FILTER_LABELS[filter.date] ?? filter.date)
          }
          isActive={filter.date !== 'all'}
        >
          <Menu>
            <Button isMenuButton onClick={() => set({ date: 'all' })}>
              All time
            </Button>
            {Object.entries(DATE_FILTER_LABELS).map(([key, label]) => (
              <Button isMenuButton key={key} onClick={() => set({ date: key })}>
                {label}
              </Button>
            ))}
          </Menu>
        </Dropdown>

        {hasActiveFilters && (
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onFilterChange(DEFAULT_ACTIVITY_FILTER)}
          >
            Clear filters
          </Button>
        )}
      </div>

      {/* Event list */}
      <div className="flex flex-col px-4 py-2 overflow-y-auto">
        {filtered.length === 0 ? (
          <div className="flex flex-col items-center gap-2 py-12">
            <Text theme="neutral">No activity found</Text>
            <Text variant="subtext" theme="neutral">
              Try adjusting your filters.
            </Text>
          </div>
        ) : (
          filtered.map((event) => {
            const caption = sourceCaption(event.source) ?? event.details
            const badgeLabel =
              ACTIVITY_TYPE_LABELS[event.type as TActivityEventType] ??
              humanize(event.type)

            return (
              <TimelineEvent
                key={event.id}
                createdAt={event.createdAt}
                status={event.status}
                title={event.title}
                badge={{ children: badgeLabel, theme: 'neutral', size: 'sm' }}
                caption={caption}
                additionalCaption={
                  event.componentName ? (
                    <Badge size="sm" variant="code" theme="neutral">
                      {event.componentName}
                    </Badge>
                  ) : undefined
                }
              />
            )
          })
        )}
      </div>
    </div>
  )
}

// ─── Resources sub-tabs ───────────────────────────────────────────────────────

const StackTab = ({ versions }: { versions: TStackVersion[] }) => (
  <div className="flex flex-col gap-2 p-4">
    {versions.map((v, index) => (
      <Card key={v.id} className="!p-4 !gap-4">
        <div className="flex items-center justify-between gap-3">
          <div className="flex items-center gap-2">
            <Icon
              variant="StackIcon"
              size={14}
              className="text-cool-grey-400"
            />
            <Badge size="sm" variant="code" theme="neutral">
              {v.version}
            </Badge>
            <Badge size="sm" theme="neutral">
              {humanize(v.planType)}
            </Badge>
          </div>
          <div className="flex items-center gap-3 flex-wrap justify-end">
            <Status status={v.status} variant="badge" />
            <Time
              time={v.createdAt}
              format="relative"
              variant="subtext"
              theme="neutral"
            />
            <Button variant="secondary" size="sm">
              {index === 0 ? 'Reprovision' : 'Deploy'}
            </Button>
          </div>
        </div>
      </Card>
    ))}
  </div>
)

const RolesTab = ({ roles }: { roles: TRoleEntry[] }) => (
  <div className="flex flex-col gap-2 p-4">
    {roles.map((role) => (
      <Card key={role.id} className="!p-4 !gap-0">
        <div className="flex items-center justify-between gap-3">
          <div className="flex items-center gap-2 min-w-0">
            <Icon
              variant="FileLockIcon"
              size={14}
              className="text-cool-grey-400 shrink-0"
            />
            <Text variant="body" family="mono" className="truncate">
              {role.name}
            </Text>
            <Badge size="sm" theme="neutral">
              {humanize(role.type)}
            </Badge>
          </div>
          <Status status={role.status} variant="badge" />
        </div>
      </Card>
    ))}
  </div>
)

const SandboxTab = ({ sandbox }: { sandbox?: TSandboxInfo }) => {
  if (!sandbox) {
    return (
      <div className="p-4">
        <Text variant="subtext" theme="neutral">
          No sandbox configured.
        </Text>
      </div>
    )
  }
  return (
    <div className="p-4">
      <Card className="!p-4 !gap-4">
        <div className="flex items-center justify-between gap-3">
          <div className="flex items-center gap-2">
            <Icon
              variant="ShippingContainerIcon"
              size={14}
              className="text-cool-grey-400"
            />
            <Text variant="body" weight="strong">
              Sandbox
            </Text>
          </div>
          <div className="flex items-center gap-3">
            <Status status={sandbox.status} variant="badge" />
            <Button variant="secondary" size="sm">
              Reprovision
            </Button>
          </div>
        </div>
        <div className="flex flex-wrap gap-x-8 gap-y-3">
          <LabeledValue label="Run type">
            <Badge size="sm" theme="neutral">
              {humanize(sandbox.runType)}
            </Badge>
          </LabeledValue>
          <LabeledValue label="Last run">
            <Time
              time={sandbox.lastRunAt}
              format="relative"
              variant="subtext"
            />
          </LabeledValue>
          {sandbox.workspaceUrl && (
            <LabeledValue label="Workspace">
              <Link
                href={sandbox.workspaceUrl}
                isExternal
                textVariant="subtext"
              >
                View workspace
              </Link>
            </LabeledValue>
          )}
        </div>
      </Card>
    </div>
  )
}

const ComponentsTab = ({ components }: { components: TComponentEntry[] }) => (
  <div className="flex flex-col gap-2 p-4">
    {components.map((cmp) => (
      <Card key={cmp.id} className="!p-4 !gap-0">
        <div className="flex items-center justify-between gap-3">
          <div className="flex items-center gap-2 min-w-0">
            <Icon
              variant="CardsIcon"
              size={14}
              className="text-cool-grey-400 shrink-0"
            />
            <Text variant="body" className="truncate">
              {cmp.name}
            </Text>
            <Badge size="sm" theme="neutral">
              {humanize(cmp.type)}
            </Badge>
            {cmp.sha && (
              <Badge size="sm" variant="code" theme="neutral">
                {cmp.sha.slice(0, 8)}
              </Badge>
            )}
          </div>
          <div className="flex items-center gap-3 shrink-0 flex-wrap justify-end">
            <Status status={cmp.status} variant="badge" />
            <Time
              time={cmp.deployedAt}
              format="relative"
              variant="subtext"
              theme="neutral"
            />
            <Button variant="secondary" size="sm">
              Deploy
            </Button>
          </div>
        </div>
      </Card>
    ))}
  </div>
)

const ImagesTab = ({ images }: { images: TImageEntry[] }) => (
  <div className="flex flex-col gap-2 p-4">
    {images.map((img) => (
      <Card key={img.id} className="!p-4 !gap-0">
        <div className="flex items-center justify-between gap-3">
          <div className="flex items-center gap-2 min-w-0">
            <Icon
              variant="PackageIcon"
              size={14}
              className="text-cool-grey-400 shrink-0"
            />
            <Text variant="body" family="mono" className="truncate">
              {img.repository}
            </Text>
            <Badge size="sm" variant="code" theme="neutral">
              {img.tag}
            </Badge>
          </div>
          <div className="flex items-center gap-3 shrink-0">
            <Status status={img.status} variant="badge" />
            <Time
              time={img.builtAt}
              format="relative"
              variant="subtext"
              theme="neutral"
            />
          </div>
        </div>
      </Card>
    ))}
  </div>
)

const ResourcesHealthTab = ({ install }: { install: TPlaygroundInstall }) => (
  <div className="flex flex-col gap-4 p-4">
    <Card className="!p-4 !gap-4">
      <HealthTimelineComponent
        scope="install"
        days={install.health.days}
        daily={install.health.daily}
        uptimePercent={install.health.uptime_percent}
        observedSeconds={install.health.observed_seconds}
        currentHealth={install.health.current_health}
        components={install.health.components}
        componentBasePath={`/${install.orgId}/installs/${install.id}/components`}
      />
    </Card>
  </div>
)

interface IResourcesTabPanel {
  install: TPlaygroundInstall
  initTab?: string
}

const ResourcesTabPanel = ({ install, initTab }: IResourcesTabPanel) => {
  const { resources } = install

  return (
    <Tabs
      tabs={{
        health: <ResourcesHealthTab install={install} />,
        stack: <StackTab versions={resources.stackVersions} />,
        roles: <RolesTab roles={resources.roles} />,
        sandbox: <SandboxTab sandbox={resources.sandbox} />,
        components: <ComponentsTab components={resources.components} />,
        images: <ImagesTab images={resources.images} />,
      }}
      tabLabels={{
        health: 'Health',
        stack: 'Stack',
        roles: 'Roles',
        sandbox: 'Sandbox',
        components: 'Components',
        images: 'Images',
      }}
      initActiveTab={initTab}
      className="gap-0"
      tabControlsClassName="px-4"
    />
  )
}

// ─── Runbooks (Operations subtab) ─────────────────────────────────────────────

const RunbooksTab = ({ runbooks }: { runbooks: TRunbookEntry[] }) => (
  <div className="flex flex-col gap-2 p-4">
    {runbooks.length === 0 && (
      <Text variant="subtext" theme="neutral">
        No runbooks yet. Runbooks will appear here once they are added to this
        app.
      </Text>
    )}
    {runbooks.map((rb) => (
      <Card key={rb.id} className="!p-4 !gap-4">
        <div className="flex items-center justify-between gap-3">
          <div className="flex items-center gap-2 min-w-0">
            <Icon
              variant="BookIcon"
              size={14}
              className="text-cool-grey-400 shrink-0"
            />
            <Text variant="body" weight="strong" className="truncate">
              {rb.name}
            </Text>
            <Badge size="sm" theme="neutral">
              {rb.stepCount} {rb.stepCount === 1 ? 'step' : 'steps'}
            </Badge>
          </div>
          <div className="flex items-center gap-3 shrink-0">
            {rb.lastRunStatus && (
              <Status status={rb.lastRunStatus} variant="badge" />
            )}
            {rb.lastRunAt && (
              <Time
                time={rb.lastRunAt}
                format="relative"
                variant="subtext"
                theme="neutral"
              />
            )}
          </div>
        </div>
        <Text variant="subtext" theme="neutral">
          {rb.description}
        </Text>
      </Card>
    ))}
  </div>
)

// ─── Operations tab (Actions + Runbooks) ──────────────────────────────────────

const ActionsTab = ({ actions }: { actions: TActionEntry[] }) => (
  <div className="flex flex-col gap-2 p-4">
    {actions.map((action) => (
      <Card key={action.id} className="!p-4 !gap-4">
        <div className="flex items-center justify-between gap-3">
          <div className="flex items-center gap-2 min-w-0">
            <Icon
              variant="TerminalWindowIcon"
              size={14}
              className="text-cool-grey-400 shrink-0"
            />
            <Text variant="body" weight="strong" className="truncate">
              {action.name}
            </Text>
          </div>
          <div className="flex items-center gap-3 shrink-0">
            {action.lastRunStatus && (
              <Status status={action.lastRunStatus} variant="badge" />
            )}
            {action.lastRunAt && (
              <Time
                time={action.lastRunAt}
                format="relative"
                variant="subtext"
                theme="neutral"
              />
            )}
          </div>
        </div>
        <Text variant="subtext" theme="neutral">
          {action.description}
        </Text>
      </Card>
    ))}
  </div>
)

const PoliciesTab = ({ policies }: { policies: TPolicyReportEntry[] }) => (
  <div className="flex flex-col gap-2 p-4">
    {policies.map((policy) => (
      <Card key={policy.id} className="!p-4 !gap-0">
        <div className="flex items-center justify-between gap-3 flex-wrap">
          <div className="flex items-center gap-2 min-w-0">
            <Icon
              variant="ShieldCheckIcon"
              size={14}
              className="text-cool-grey-400 shrink-0"
            />
            <Text variant="body" className="truncate">
              {policy.name}
            </Text>
            <Badge size="sm" variant="code" theme="neutral">
              {policy.componentName}
            </Badge>
          </div>
          <div className="flex items-center gap-3 shrink-0">
            <Status status={policy.status} variant="badge" />
            <Time
              time={policy.evaluatedAt}
              format="relative"
              variant="subtext"
              theme="neutral"
            />
          </div>
        </div>
      </Card>
    ))}
  </div>
)

const RunnerTab = ({ runner }: { runner: TRunnerInfo }) => (
  <div className="flex flex-col gap-4 p-4">
    <Card className="!p-4 !gap-4">
      <div className="flex items-center justify-between gap-3 flex-wrap">
        <div className="flex items-center gap-2">
          <Icon variant="CpuIcon" size={14} className="text-cool-grey-400" />
          <Text variant="body" weight="strong">
            Install runner
          </Text>
          <Badge size="sm" variant="code" theme="neutral">
            {runner.version}
          </Badge>
        </div>
        <div className="flex items-center gap-3">
          <Status status={runner.status} variant="badge" />
          <Button variant="secondary" size="sm">
            Restart runner
          </Button>
        </div>
      </div>
      <Text variant="subtext" family="mono" theme="neutral">
        {runner.id}
      </Text>
    </Card>

    <div className="flex flex-col gap-2">
      <Text variant="body" weight="strong">
        Processes
      </Text>
      {runner.processes.map((process) => (
        <Card key={process.id} className="!p-4 !gap-0">
          <div className="flex items-center justify-between gap-3">
            <Text variant="subtext" family="mono">
              {process.name}
            </Text>
            <div className="flex items-center gap-3">
              <Status status={process.status} variant="badge" />
              <Time
                time={process.startedAt}
                format="relative"
                variant="subtext"
                theme="neutral"
              />
            </div>
          </div>
        </Card>
      ))}
    </div>

    <div className="flex flex-col gap-2">
      <Text variant="body" weight="strong">
        Recent jobs
      </Text>
      {runner.recentJobs.map((job) => (
        <Card key={job.id} className="!p-4 !gap-0">
          <div className="flex items-center justify-between gap-3">
            <Text variant="subtext">{job.name}</Text>
            <div className="flex items-center gap-3">
              <Status status={job.status} variant="badge" />
              <Time
                time={job.createdAt}
                format="relative"
                variant="subtext"
                theme="neutral"
              />
            </div>
          </div>
        </Card>
      ))}
    </div>
  </div>
)

const OperationsTab = ({
  operations,
}: {
  operations: TPlaygroundInstall['operations']
}) => (
  <Tabs
    tabs={{
      actions: <ActionsTab actions={operations.actions} />,
      runbooks: <RunbooksTab runbooks={operations.runbooks} />,
      policies: <PoliciesTab policies={operations.policies} />,
      runner: <RunnerTab runner={operations.runner} />,
    }}
    tabLabels={{
      actions: 'Actions',
      runbooks: 'Runbooks',
      policies: 'Policies',
      runner: 'Runner',
    }}
    className="gap-0"
    tabControlsClassName="px-4"
  />
)

// ─── Configuration sub-tabs ───────────────────────────────────────────────────

const CHANGE_THEME = {
  add: 'success',
  remove: 'error',
  change: 'warn',
} as const

const CHANGE_PREFIX = {
  add: '+',
  remove: '-',
  change: '~',
} as const

const ConfigurationChangeRows = ({
  changes,
}: {
  changes: TConfigurationChange[]
}) => (
  <div className="flex flex-col divide-y border rounded-md overflow-hidden">
    {changes.map((change) => (
      <div
        key={`${change.path}-${change.operation}`}
        className="grid grid-cols-[1rem_minmax(0,1fr)] md:grid-cols-[1rem_minmax(10rem,1fr)_minmax(0,2fr)] items-center gap-3 px-3 py-2"
      >
        <Text
          variant="subtext"
          family="mono"
          weight="strong"
          theme={CHANGE_THEME[change.operation]}
        >
          {CHANGE_PREFIX[change.operation]}
        </Text>
        <Text variant="subtext" family="mono" weight="strong">
          {change.path}
        </Text>
        <div className="col-start-2 md:col-start-auto flex items-center gap-2 min-w-0">
          {change.isRedacted ? (
            <Badge size="sm" theme="neutral">
              Redacted
            </Badge>
          ) : (
            <>
              {change.previousValue !== undefined && (
                <Text
                  variant="subtext"
                  family="mono"
                  theme="neutral"
                  className="truncate"
                >
                  {change.previousValue}
                </Text>
              )}
              {change.previousValue !== undefined &&
                change.nextValue !== undefined && (
                  <Icon
                    variant="ArrowRightIcon"
                    size={12}
                    className="shrink-0 text-cool-grey-400"
                  />
                )}
              {change.nextValue !== undefined && (
                <Text variant="subtext" family="mono" className="truncate">
                  {change.nextValue}
                </Text>
              )}
            </>
          )}
        </div>
      </div>
    ))}
  </div>
)

const ConfigurationVersionFeed = ({
  versions,
  idPrefix,
}: {
  versions: TConfigurationVersion[]
  idPrefix: string
}) => (
  <div className="flex flex-col gap-2">
    <div className="flex items-center justify-between gap-3">
      <Text variant="body" weight="strong">
        Version history
      </Text>
      <Text variant="subtext" theme="neutral">
        Compared with the previous version
      </Text>
    </div>
    {versions.map((version, index) => (
      <Expand
        key={version.id}
        id={`${idPrefix}-${version.id}`}
        isOpen={index === 0}
        className="border rounded-md overflow-hidden"
        headerClassName="px-4 py-3"
        heading={
          <div className="flex items-center justify-between gap-3 w-full min-w-0">
            <div className="flex items-center gap-2 min-w-0">
              <Badge size="sm" variant="code" theme="neutral">
                {version.version}
              </Badge>
              <Text variant="subtext" weight="strong" className="truncate">
                {version.title}
              </Text>
            </div>
            <div className="flex items-center gap-3 shrink-0">
              <Badge size="sm" theme="neutral">
                {version.changes.length}{' '}
                {version.changes.length === 1 ? 'change' : 'changes'}
              </Badge>
              <Time
                time={version.createdAt}
                format="relative"
                variant="subtext"
                theme="neutral"
              />
            </div>
          </div>
        }
      >
        <div className="flex flex-col gap-3 px-4 pb-4 border-t pt-3">
          {(version.actor || version.source) && (
            <div className="flex items-center gap-4 flex-wrap">
              {version.actor && (
                <LabeledValue label="Changed by">
                  <Text variant="subtext">{version.actor}</Text>
                </LabeledValue>
              )}
              {version.source && (
                <LabeledValue label="Source">
                  <Text variant="subtext" family="mono">
                    {version.source}
                  </Text>
                </LabeledValue>
              )}
            </div>
          )}
          <ConfigurationChangeRows changes={version.changes} />
          {version.fileDiff && (
            <CodeBlock language="toml" showCopy>
              {version.fileDiff}
            </CodeBlock>
          )}
        </div>
      </Expand>
    ))}
  </div>
)

const InputsTab = ({
  inputs,
  versions,
}: {
  inputs: TInputEntry[]
  versions: TConfigurationVersion[]
}) => {
  const groups = Array.from(new Set(inputs.map((input) => input.group)))

  return (
    <div className="flex flex-col gap-4 p-4">
      {groups.map((group) => (
        <Card key={group} className="!p-4 !gap-4">
          <Text variant="body" weight="strong">
            {group}
          </Text>
          <PropertyGrid
            values={inputs.filter((input) => input.group === group)}
            align="start"
            columns={[
              {
                key: 'displayName',
                header: 'Input',
                render: (_value, input) => (
                  <div className="flex flex-col min-w-0">
                    <Text variant="subtext">{input.displayName}</Text>
                    <Text variant="subtext" theme="neutral" family="mono">
                      {input.name}
                    </Text>
                  </div>
                ),
              },
              {
                key: 'value',
                header: 'Value',
                render: (_value, input) =>
                  input.isRedacted ? (
                    <Badge size="sm" theme="neutral">
                      Redacted
                    </Badge>
                  ) : (
                    <Text variant="subtext" family="mono" className="break-all">
                      {input.value}
                    </Text>
                  ),
              },
            ]}
          />
        </Card>
      ))}
      <ConfigurationVersionFeed versions={versions} idPrefix="inputs" />
    </div>
  )
}

const ConfigFileTab = ({
  configFile,
  versions,
}: {
  configFile?: TConfigFileInfo
  versions: TConfigurationVersion[]
}) => {
  if (!configFile) {
    return (
      <div className="p-4">
        <Text variant="subtext" theme="neutral">
          This install is managed from the dashboard, so it has no config file.
        </Text>
      </div>
    )
  }

  const repoHref =
    configFile.repo && !configFile.repo.startsWith('http')
      ? `https://github.com/${configFile.repo}`
      : configFile.repo

  return (
    <div className="flex flex-col gap-4 p-4">
      <Card className="!p-4 !gap-4">
        <div className="flex items-center justify-between gap-3">
          <div className="flex items-center gap-2 min-w-0">
            <Icon
              variant="FileCodeIcon"
              size={14}
              className="text-cool-grey-400 shrink-0"
            />
            <Text variant="body" family="mono" className="truncate">
              {configFile.path}
            </Text>
          </div>
          <Badge size="sm" variant="code" theme="neutral">
            {configFile.version}
          </Badge>
        </div>

        <div className="flex flex-wrap gap-x-8 gap-y-3">
          {configFile.repo && repoHref && (
            <LabeledValue label="Repository">
              <Link href={repoHref} isExternal textVariant="subtext">
                {configFile.repo}
              </Link>
            </LabeledValue>
          )}
          {configFile.gitBranch && (
            <LabeledValue label="Git branch">
              <Text variant="subtext" family="mono">
                {configFile.gitBranch}
              </Text>
            </LabeledValue>
          )}
          <LabeledValue label="Last synced">
            <Time
              time={configFile.syncedAt}
              format="relative"
              variant="subtext"
            />
          </LabeledValue>
        </div>

        <CodeBlock language="toml" showCopy>
          {configFile.contents}
        </CodeBlock>
      </Card>
      <ConfigurationVersionFeed versions={versions} idPrefix="config-file" />
    </div>
  )
}

const OverridesTab = ({ overrides }: { overrides: TOverrideEntry[] }) => {
  if (overrides.length === 0) {
    return (
      <div className="p-4">
        <Text variant="subtext" theme="neutral">
          No overrides set. Component overrides let you change one component
          without editing app config.
        </Text>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-2 p-4">
      {overrides.map((override) => (
        <Card key={override.id} className="!p-4 !gap-0">
          <div className="flex items-center justify-between gap-3">
            <div className="flex items-center gap-2 min-w-0">
              <Icon
                variant="FadersIcon"
                size={14}
                className="text-cool-grey-400 shrink-0"
              />
              <Text variant="body" className="truncate">
                {override.componentName}
              </Text>
              <Text variant="subtext" theme="neutral" family="mono">
                {override.inputName}
              </Text>
              <Badge size="sm" variant="code" theme="neutral">
                {override.value}
              </Badge>
            </div>
            <Time
              time={override.updatedAt}
              format="relative"
              variant="subtext"
              theme="neutral"
            />
          </div>
        </Card>
      ))}
    </div>
  )
}

const ConfigurationTabPanel = ({
  install,
  initTab,
}: {
  install: TPlaygroundInstall
  initTab?: string
}) => (
  <Tabs
    tabs={{
      appBranch: (
        <div className="flex flex-col gap-4 p-4">
          <InstallBranchTrackingCard install={install} />
          <ConfigurationVersionFeed
            versions={install.configuration.appBranchVersions}
            idPrefix="app-branch"
          />
        </div>
      ),
      inputs: (
        <InputsTab
          inputs={install.configuration.inputs}
          versions={install.configuration.inputVersions}
        />
      ),
      configFile: (
        <ConfigFileTab
          configFile={install.configuration.configFile}
          versions={install.configuration.configFileVersions}
        />
      ),
      overrides: <OverridesTab overrides={install.configuration.overrides} />,
    }}
    tabLabels={{
      appBranch: 'App branch',
      inputs: 'Inputs',
      configFile: 'Config file',
      overrides: 'Overrides',
    }}
    initActiveTab={initTab}
    className="gap-0"
    tabControlsClassName="px-4"
  />
)

// ─── Main component ───────────────────────────────────────────────────────────

export interface IInstallDetailPlayground {
  install: TPlaygroundInstall
  className?: string
}

export const InstallDetailPlayground = ({
  install,
  className,
}: IInstallDetailPlayground) => {
  const [activeTab, setActiveTab] = useState<TTopTab>('overview')
  const [activityFilter, setActivityFilter] = useState<TActivityFilter>(
    DEFAULT_ACTIVITY_FILTER
  )
  // Incrementing keys remount Tabs so header cards can select a nested tab.
  const [resourcesNav, setResourcesNav] = useState<{
    tab?: string
    key: number
  }>({ key: 0 })
  const [configurationNav, setConfigurationNav] = useState<{
    tab?: string
    key: number
  }>({ key: 0 })

  const TOP_TABS = Object.keys(TOP_TAB_LABELS) as TTopTab[]

  const handleStripNavigate = (
    tab: TTopTab,
    opts?: {
      activityType?: string
      resourcesTab?: string
      configurationTab?: string
    }
  ) => {
    setActiveTab(tab)
    if (tab === 'activity' && opts?.activityType) {
      setActivityFilter({ ...DEFAULT_ACTIVITY_FILTER, type: opts.activityType })
    }
    if (tab === 'resources' && opts?.resourcesTab) {
      setResourcesNav((prev) => ({ tab: opts.resourcesTab, key: prev.key + 1 }))
    }
    if (tab === 'configuration' && opts?.configurationTab) {
      setConfigurationNav((prev) => ({
        tab: opts.configurationTab,
        key: prev.key + 1,
      }))
    }
  }

  return (
    <div
      className={cn(
        'flex flex-col h-full min-h-0 bg-background border rounded-md overflow-hidden',
        className
      )}
    >
      <InstallPlaygroundHeader
        install={install}
        onNavigate={handleStripNavigate}
      />

      {/* Top tab bar */}
      <div
        role="tablist"
        aria-label="Install sections"
        className="flex items-center gap-6 border-b px-4 shrink-0 overflow-x-auto"
      >
        {TOP_TABS.map((key) => (
          <Button
            key={key}
            id={`install-tab-${key}`}
            type="button"
            role="tab"
            aria-controls={`install-tabpanel-${key}`}
            aria-selected={activeTab === key}
            onClick={() => setActiveTab(key)}
            isActive={activeTab === key}
            variant="tab"
          >
            {TOP_TAB_LABELS[key]}
          </Button>
        ))}
      </div>

      {/* Tab panels */}
      <div className="flex-1 min-h-0 overflow-y-auto">
        {TOP_TABS.map((key) => (
          <div
            key={key}
            id={`install-tabpanel-${key}`}
            role="tabpanel"
            aria-labelledby={`install-tab-${key}`}
            hidden={activeTab !== key}
          >
            {activeTab === key && (
              <>
                {key === 'overview' && (
                  <OverviewTab
                    install={install}
                    onNavigate={handleStripNavigate}
                  />
                )}
                {key === 'activity' && (
                  <ActivityTab
                    install={install}
                    filter={activityFilter}
                    onFilterChange={setActivityFilter}
                  />
                )}
                {key === 'resources' && (
                  <ResourcesTabPanel
                    key={resourcesNav.key}
                    install={install}
                    initTab={resourcesNav.tab}
                  />
                )}
                {key === 'operations' && (
                  <OperationsTab operations={install.operations} />
                )}
                {key === 'configuration' && (
                  <ConfigurationTabPanel
                    key={configurationNav.key}
                    install={install}
                    initTab={configurationNav.tab}
                  />
                )}
              </>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
