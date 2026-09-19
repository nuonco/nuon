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
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import { HealthTimelineComponent } from '@/components/install-health/HealthTimeline'
import { cn } from '@/utils/classnames'
import { humanize } from '@/utils/string-utils'
import { SectionNav, type TSectionNavSection } from './SectionNav'
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
  | 'deployments'
  | 'health'
  | 'operations'
  | 'configuration'

const TOP_TAB_LABELS: Record<TTopTab, string> = {
  overview: 'Overview',
  resources: 'Resources',
  deployments: 'Deployments',
  health: 'Health checks',
  operations: 'Operations',
  configuration: 'Configuration',
}

// ─── Deployments filter state ─────────────────────────────────────────────────

type TDeploymentCategory = 'app-branch' | 'install-config' | 'deployment'

type TDeploymentFilter = {
  search: string
  status: string
  category: 'all' | TDeploymentCategory
  component: string
  date: string
}

const DEFAULT_DEPLOYMENT_FILTER: TDeploymentFilter = {
  search: '',
  status: 'all',
  category: 'all',
  component: 'all',
  date: 'all',
}

// ─── Header ──────────────────────────────────────────────────────────────────

interface IInstallPlaygroundHeader {
  install: TPlaygroundInstall
  onNavigate: (
    tab: TTopTab,
    opts?: {
      resourcesTab?: string
      componentId?: string
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
    <nav aria-label="Breadcrumb" className="flex items-center gap-2 min-w-0">
      <Link
        href={`/${install.orgId}`}
        variant="breadcrumb"
        textVariant="subtext"
      >
        {install.orgName}
      </Link>
      <Icon
        variant="CaretRightIcon"
        size={12}
        className="text-cool-grey-400 shrink-0"
      />
      <Link
        href={`/${install.orgId}/installs`}
        variant="breadcrumb"
        textVariant="subtext"
      >
        Installs
      </Link>
      <Icon
        variant="CaretRightIcon"
        size={12}
        className="text-cool-grey-400 shrink-0"
      />
      <Text variant="subtext" weight="strong" className="truncate">
        {install.name}
      </Text>
    </nav>

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
      resourcesTab?: string
      componentId?: string
      configurationTab?: string
    }
  ) => void
}

const InstallStatusCard = ({ install, onNavigate }: IInstallStatusCard) => {
  const { activity, resources } = install

  // Deployment status includes every deploy/config change, but not drift scans.
  const runningUpdates = activity.filter(
    (e) =>
      e.type !== 'drift_scan' &&
      (e.status === 'in-progress' || e.status === 'pending')
  )
  const failedUpdates = activity.filter(
    (e) => e.type !== 'drift_scan' && e.status === 'error'
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
  const healthStatus = install.health.current_health || 'unknown'
  const healthLabel =
    healthStatus === 'active' || healthStatus === 'healthy'
      ? 'Healthy'
      : healthStatus === 'warn' || healthStatus === 'degraded'
        ? 'Degraded'
        : healthStatus === 'error' || healthStatus === 'unhealthy'
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
        onClick={() => onNavigate('deployments')}
        aria-label={`Deployments: ${updatesLabel}. Navigate to deployments.`}
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
        onClick={() => onNavigate('health')}
        aria-label={`Health checks: ${healthLabel}. Navigate to health checks.`}
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

const OverviewTab = ({ install }: { install: TPlaygroundInstall }) => {
  const hasDrift = install.driftedObjects.length > 0

  return (
    <div className="flex flex-col gap-4 p-4">
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

      {install.readme && (
        <Card className="!p-4 !gap-4">
          <Markdown content={install.readme} mode="install" />
        </Card>
      )}
    </div>
  )
}

// ─── Deployments tab ──────────────────────────────────────────────────────────

const DEPLOYMENT_CATEGORY_LABELS: Record<TDeploymentCategory, string> = {
  'app-branch': 'App branch updates',
  'install-config': 'Install config updates',
  deployment: 'Deployments',
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

const deploymentCategory = (
  type: TActivityEventType
): TDeploymentCategory | undefined => {
  switch (type) {
    case 'app_branch_run':
      return 'app-branch'
    case 'config_update':
    case 'inputs_update':
      return 'install-config'
    case 'deploy':
    case 'stack_update':
      return 'deployment'
    case 'drift_scan':
      return undefined
  }
}

const filterDeployments = (
  events: TActivityEvent[],
  filter: TDeploymentFilter
): TActivityEvent[] =>
  events.filter((e) => {
    const category = deploymentCategory(e.type)
    if (!category) return false
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
    if (filter.category !== 'all' && category !== filter.category) return false
    if (filter.component !== 'all' && e.componentName !== filter.component)
      return false
    if (filter.date !== 'all') {
      const maxMs = DATE_FILTER_MS[filter.date] ?? 0
      if (Date.now() - new Date(e.createdAt).getTime() > maxMs) return false
    }
    return true
  })

interface IDeploymentsTab {
  install: TPlaygroundInstall
  filter: TDeploymentFilter
  onFilterChange: (f: TDeploymentFilter) => void
}

const DeploymentsTab = ({
  install,
  filter,
  onFilterChange,
}: IDeploymentsTab) => {
  const set = (patch: Partial<TDeploymentFilter>) =>
    onFilterChange({ ...filter, ...patch })

  const deploymentEvents = install.activity.filter((event) =>
    deploymentCategory(event.type)
  )
  const allComponents = Array.from(
    new Set(deploymentEvents.map((e) => e.componentName).filter(Boolean))
  ) as string[]

  const allStatuses = Array.from(new Set(deploymentEvents.map((e) => e.status)))

  const filtered = filterDeployments(install.activity, filter)

  const hasActiveFilters =
    filter.search !== '' ||
    filter.status !== 'all' ||
    filter.category !== 'all' ||
    filter.component !== 'all' ||
    filter.date !== 'all'

  return (
    <div className="flex flex-col min-h-0">
      {/* Filter bar */}
      <div className="flex items-center flex-wrap gap-2 px-4 py-3 border-b bg-cool-grey-50 dark:bg-dark-grey-800 shrink-0">
        <SearchInput
          aria-label="Search deployments"
          value={filter.search}
          onChange={(v) => set({ search: v })}
          placeholder="Search deployments…"
          labelClassName="flex-1 min-w-44"
        />

        <Dropdown
          id="deployments-filter-status"
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
          id="deployments-filter-category"
          variant="secondary"
          size="sm"
          buttonText={
            filter.category === 'all'
              ? 'Category'
              : DEPLOYMENT_CATEGORY_LABELS[filter.category]
          }
          isActive={filter.category !== 'all'}
        >
          <Menu>
            <Button isMenuButton onClick={() => set({ category: 'all' })}>
              All categories
            </Button>
            {(
              Object.keys(DEPLOYMENT_CATEGORY_LABELS) as TDeploymentCategory[]
            ).map((category) => (
              <Button
                isMenuButton
                key={category}
                onClick={() => set({ category })}
              >
                {DEPLOYMENT_CATEGORY_LABELS[category]}
              </Button>
            ))}
          </Menu>
        </Dropdown>

        {allComponents.length > 0 && (
          <Dropdown
            id="deployments-filter-component"
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
          id="deployments-filter-date"
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
            onClick={() => onFilterChange(DEFAULT_DEPLOYMENT_FILTER)}
          >
            Clear filters
          </Button>
        )}
      </div>

      {/* Event list */}
      <div className="flex flex-col px-4 py-2 overflow-y-auto">
        {filtered.length === 0 ? (
          <div className="flex flex-col items-center gap-2 py-12">
            <Text theme="neutral">No deployments found</Text>
            <Text variant="subtext" theme="neutral">
              Try adjusting your filters.
            </Text>
          </div>
        ) : (
          filtered.map((event) => {
            const caption = sourceCaption(event.source) ?? event.details
            const category = deploymentCategory(event.type)
            const badgeLabel = category
              ? DEPLOYMENT_CATEGORY_LABELS[category]
              : humanize(event.type)

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

const ComponentDetail = ({ component }: { component: TComponentEntry }) => (
  <div className="flex flex-col gap-4 p-4">
    <div className="flex items-start justify-between gap-4 flex-wrap">
      <div className="flex flex-col gap-2 min-w-0">
        <div className="flex items-center gap-2 flex-wrap">
          <Icon variant="CardsIcon" size={16} className="text-cool-grey-400" />
          <Text variant="h3" weight="strong">
            {component.name}
          </Text>
          <Badge size="sm" theme="neutral">
            {humanize(component.type)}
          </Badge>
          <Status status={component.status} variant="badge" />
        </div>
        {component.sha && (
          <Badge size="sm" variant="code" theme="neutral">
            {component.sha.slice(0, 8)}
          </Badge>
        )}
      </div>
      <Button variant="secondary" size="sm">
        Deploy
      </Button>
    </div>

    <div className="flex flex-col gap-2">
      <Text variant="body" weight="strong">
        Latest deploy
      </Text>
      <Card className="!p-4 !gap-0">
        <div className="flex items-center justify-between gap-3 flex-wrap">
          <div className="flex items-center gap-2">
            <Status status={component.status} variant="badge" />
            {component.sha && (
              <Badge size="sm" variant="code" theme="neutral">
                {component.sha.slice(0, 8)}
              </Badge>
            )}
          </div>
          <Time
            time={component.deployedAt}
            format="relative"
            variant="subtext"
            theme="neutral"
          />
        </div>
      </Card>
    </div>

    <div className="flex flex-col gap-2">
      <Text variant="body" weight="strong">
        Health
      </Text>
      <Card className="!p-4 !gap-4">
        <HealthTimelineComponent
          scope="component"
          days={component.health.days}
          daily={component.health.daily}
          uptimePercent={component.health.uptime_percent}
          observedSeconds={component.health.observed_seconds}
          currentHealth={component.health.current_health}
          transitions={component.health.transitions}
          deployBasePath={`/components/${component.id}/deploys`}
        />
      </Card>
    </div>
  </div>
)

const ImageDetail = ({ image }: { image: TImageEntry }) => (
  <div className="flex flex-col gap-4 p-4">
    <div className="flex items-start justify-between gap-4 flex-wrap">
      <div className="flex items-center gap-2 flex-wrap min-w-0">
        <Icon variant="PackageIcon" size={16} className="text-cool-grey-400" />
        <Text variant="h3" weight="strong" family="mono">
          {image.repository}
        </Text>
        <Status status={image.status} variant="badge" />
      </div>
      <Button variant="secondary" size="sm">
        Build image
      </Button>
    </div>
    <Card className="!p-4 !gap-4">
      <div className="flex flex-wrap gap-x-8 gap-y-3">
        <LabeledValue label="Tag">
          <Badge size="sm" variant="code" theme="neutral">
            {image.tag}
          </Badge>
        </LabeledValue>
        {image.sha && (
          <LabeledValue label="Digest">
            <Text variant="subtext" family="mono">
              {image.sha}
            </Text>
          </LabeledValue>
        )}
        <LabeledValue label="Built">
          <Time time={image.builtAt} format="relative" variant="subtext" />
        </LabeledValue>
      </div>
    </Card>
  </div>
)

const formatHealthUptime = (
  uptimePercent?: number,
  observedSeconds?: number
) =>
  (observedSeconds ?? 0) > 0 && uptimePercent !== undefined
    ? `${uptimePercent.toFixed(2)}%`
    : 'No signal'

const HealthChecksTab = ({
  install,
  onSelectComponent,
}: {
  install: TPlaygroundInstall
  onSelectComponent: (componentId: string) => void
}) => (
  <div className="flex flex-col gap-4 p-4">
    <Card className="!p-4 !gap-4">
      <HealthTimelineComponent
        scope="install"
        days={install.health.days}
        daily={install.health.daily}
        uptimePercent={install.health.uptime_percent}
        observedSeconds={install.health.observed_seconds}
        currentHealth={install.health.current_health}
      />
    </Card>
    <Card className="!p-4 !gap-3">
      <Text variant="body" weight="strong">
        Component health
      </Text>
      <div className="flex flex-col divide-y">
        {install.health.components?.map((component) => (
          <Button
            key={component.install_component_id}
            variant="ghost"
            size="sm"
            className="!px-0 w-full justify-between"
            onClick={() =>
              component.component_id &&
              onSelectComponent(component.component_id)
            }
          >
            <Text as="span" variant="subtext">
              {component.component_name}
            </Text>
            <span className="flex items-center gap-3">
              <Status
                status={component.current_health || 'unknown'}
                variant="badge"
              />
              <Text
                as="span"
                variant="subtext"
                theme="neutral"
                className="w-16 text-right"
              >
                {formatHealthUptime(
                  component.uptime_percent,
                  component.observed_seconds
                )}
              </Text>
            </span>
          </Button>
        ))}
      </div>
    </Card>
  </div>
)

interface IResourcesTabPanel {
  install: TPlaygroundInstall
  initTab?: string
  initialComponentId?: string
}

const ResourcesTabPanel = ({
  install,
  initTab,
  initialComponentId,
}: IResourcesTabPanel) => {
  const { resources } = install
  const sections: TSectionNavSection[] = [
    {
      id: 'stack',
      label: 'Stack',
      render: () => <StackTab versions={resources.stackVersions} />,
    },
    {
      id: 'sandbox',
      label: 'Sandbox',
      render: () => <SandboxTab sandbox={resources.sandbox} />,
    },
    {
      id: 'components',
      label: 'Components',
      items: resources.components.map((component) => ({
        id: component.id,
        label: (
          <span className="flex items-center justify-between gap-2 w-full min-w-0">
            <Text as="span" variant="subtext" className="truncate">
              {component.name}
            </Text>
            <Status
              status={component.health.current_health || 'unknown'}
              variant="timeline"
              isWithoutText
              iconSize={12}
            />
          </span>
        ),
      })),
      render: (componentId) => {
        const component = resources.components.find(
          (entry) => entry.id === componentId
        )
        return component ? (
          <ComponentDetail component={component} />
        ) : (
          <div className="p-4">
            <Text variant="subtext" theme="neutral">
              No components configured.
            </Text>
          </div>
        )
      },
    },
    {
      id: 'images',
      label: 'Images',
      items: resources.images.map((image) => ({
        id: image.id,
        label: (
          <Text as="span" variant="subtext" family="mono" className="truncate">
            {image.repository}
          </Text>
        ),
      })),
      render: (imageId) => {
        const image = resources.images.find((entry) => entry.id === imageId)
        return image ? (
          <ImageDetail image={image} />
        ) : (
          <div className="p-4">
            <Text variant="subtext" theme="neutral">
              No images configured.
            </Text>
          </div>
        )
      },
    },
  ]

  return (
    <SectionNav
      sections={sections}
      initSectionId={initTab}
      initItemId={initialComponentId}
      ariaLabel="Resources"
    />
  )
}

// ─── Runbooks (Operations subtab) ─────────────────────────────────────────────

const RunbookDetail = ({ runbook }: { runbook: TRunbookEntry }) => (
  <div className="flex flex-col gap-4 p-4">
    <div className="flex items-start justify-between gap-4 flex-wrap">
      <div className="flex items-center gap-2 min-w-0">
        <Icon
          variant="BookIcon"
          size={16}
          className="text-cool-grey-400 shrink-0"
        />
        <Text variant="h3" weight="strong">
          {runbook.name}
        </Text>
        <Badge size="sm" theme="neutral">
          {runbook.stepCount} {runbook.stepCount === 1 ? 'step' : 'steps'}
        </Badge>
      </div>
      <Button variant="secondary" size="sm">
        Run runbook
      </Button>
    </div>
    <Card className="!p-4 !gap-4">
      <Text variant="subtext" theme="neutral">
        {runbook.description}
      </Text>
      {(runbook.lastRunStatus || runbook.lastRunAt) && (
        <div className="flex items-center gap-3">
          {runbook.lastRunStatus && (
            <Status status={runbook.lastRunStatus} variant="badge" />
          )}
          {runbook.lastRunAt && (
            <Time
              time={runbook.lastRunAt}
              format="relative"
              variant="subtext"
              theme="neutral"
            />
          )}
        </div>
      )}
    </Card>
  </div>
)

// ─── Operations tab (Actions + Runbooks) ──────────────────────────────────────

const ActionDetail = ({ action }: { action: TActionEntry }) => (
  <div className="flex flex-col gap-4 p-4">
    <div className="flex items-start justify-between gap-4 flex-wrap">
      <div className="flex items-center gap-2 min-w-0">
        <Icon
          variant="TerminalWindowIcon"
          size={16}
          className="text-cool-grey-400 shrink-0"
        />
        <Text variant="h3" weight="strong">
          {action.name}
        </Text>
      </div>
      <Button variant="secondary" size="sm">
        Run action
      </Button>
    </div>
    <Card className="!p-4 !gap-4">
      <Text variant="subtext" theme="neutral">
        {action.description}
      </Text>
      {(action.lastRunStatus || action.lastRunAt) && (
        <div className="flex items-center gap-3">
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
      )}
    </Card>
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
}) => {
  const sections: TSectionNavSection[] = [
    {
      id: 'actions',
      label: 'Actions',
      items: operations.actions.map((action) => ({
        id: action.id,
        label: (
          <Text as="span" variant="subtext" className="truncate">
            {action.name}
          </Text>
        ),
      })),
      render: (actionId) => {
        const action = operations.actions.find((entry) => entry.id === actionId)
        return action ? (
          <ActionDetail action={action} />
        ) : (
          <div className="p-4">
            <Text variant="subtext" theme="neutral">
              No actions configured.
            </Text>
          </div>
        )
      },
    },
    {
      id: 'runbooks',
      label: 'Runbooks',
      items: operations.runbooks.map((runbook) => ({
        id: runbook.id,
        label: (
          <Text as="span" variant="subtext" className="truncate">
            {runbook.name}
          </Text>
        ),
      })),
      render: (runbookId) => {
        const runbook = operations.runbooks.find(
          (entry) => entry.id === runbookId
        )
        return runbook ? (
          <RunbookDetail runbook={runbook} />
        ) : (
          <div className="p-4">
            <Text variant="subtext" theme="neutral">
              No runbooks yet. Runbooks will appear here once they are added to
              this app.
            </Text>
          </div>
        )
      },
    },
    {
      id: 'policies',
      label: 'Policies',
      render: () => <PoliciesTab policies={operations.policies} />,
    },
    {
      id: 'runner',
      label: 'Runner',
      render: () => <RunnerTab runner={operations.runner} />,
    },
  ]

  return (
    <SectionNav
      sections={sections}
      initSectionId="actions"
      ariaLabel="Operations"
    />
  )
}

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
}) => {
  const sections: TSectionNavSection[] = [
    {
      id: 'appBranch',
      label: 'App branch',
      render: () => (
        <div className="flex flex-col gap-4 p-4">
          <InstallBranchTrackingCard install={install} />
          <ConfigurationVersionFeed
            versions={install.configuration.appBranchVersions}
            idPrefix="app-branch"
          />
        </div>
      ),
    },
    {
      id: 'inputs',
      label: 'Inputs',
      render: () => (
        <InputsTab
          inputs={install.configuration.inputs}
          versions={install.configuration.inputVersions}
        />
      ),
    },
    {
      id: 'configFile',
      label: 'Config file',
      render: () => (
        <ConfigFileTab
          configFile={install.configuration.configFile}
          versions={install.configuration.configFileVersions}
        />
      ),
    },
    {
      id: 'overrides',
      label: 'Overrides',
      render: () => (
        <OverridesTab overrides={install.configuration.overrides} />
      ),
    },
  ]

  return (
    <SectionNav
      sections={sections}
      initSectionId={initTab}
      ariaLabel="Configuration"
    />
  )
}

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
  const [deploymentFilter, setDeploymentFilter] = useState<TDeploymentFilter>(
    DEFAULT_DEPLOYMENT_FILTER
  )
  // Incrementing keys remount contextual navigation so header cards can select a section.
  const [resourcesNav, setResourcesNav] = useState<{
    tab?: string
    componentId?: string
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
      resourcesTab?: string
      componentId?: string
      configurationTab?: string
    }
  ) => {
    setActiveTab(tab)
    if (tab === 'resources' && opts?.resourcesTab) {
      setResourcesNav((prev) => ({
        tab: opts.resourcesTab,
        componentId: opts.componentId,
        key: prev.key + 1,
      }))
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
                {key === 'overview' && <OverviewTab install={install} />}
                {key === 'deployments' && (
                  <DeploymentsTab
                    install={install}
                    filter={deploymentFilter}
                    onFilterChange={setDeploymentFilter}
                  />
                )}
                {key === 'health' && (
                  <HealthChecksTab
                    install={install}
                    onSelectComponent={(componentId) =>
                      handleStripNavigate('resources', {
                        resourcesTab: 'components',
                        componentId,
                      })
                    }
                  />
                )}
                {key === 'resources' && (
                  <ResourcesTabPanel
                    key={resourcesNav.key}
                    install={install}
                    initTab={resourcesNav.tab}
                    initialComponentId={resourcesNav.componentId}
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
