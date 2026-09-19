import { useState } from 'react'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { LabelBadge } from '@/components/common/LabelBadge'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Link } from '@/components/common/Link'
import { Menu } from '@/components/common/Menu'
import { SearchInput } from '@/components/common/SearchInput'
import { Status } from '@/components/common/Status'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import { cn } from '@/utils/classnames'
import { humanize } from '@/utils/string-utils'
import type {
  TPlaygroundInstall,
  TActivityEvent,
  TActivityEventType,
  TAppBranchSource,
  TBranchCommitRef,
  TBranchTracking,
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
} from './types'

// ─── Top-level navigation ────────────────────────────────────────────────────

type TTopTab = 'overview' | 'activity' | 'runbooks' | 'resources' | 'operations'

const TOP_TAB_LABELS: Record<TTopTab, string> = {
  overview: 'Overview',
  activity: 'Activity',
  runbooks: 'Runbooks',
  resources: 'Resources',
  operations: 'Operations',
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
}

const InstallPlaygroundHeader = ({ install }: IInstallPlaygroundHeader) => (
  <header className="flex items-start justify-between gap-4 flex-wrap p-4 md:p-6 border-b shrink-0">
    <div className="flex flex-col gap-1.5 min-w-0">
      <div className="flex items-center gap-2 flex-wrap">
        <Text variant="h3" weight="stronger" level={1}>
          {install.name}
        </Text>
        {Object.entries(install.labels).map(([k, v]) => (
          <LabelBadge key={k} size="sm" labelKey={k} labelValue={v} />
        ))}
      </div>
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

    <div className="flex items-start gap-6 flex-wrap shrink-0">
      {install.isManagedByConfig && (
        <LabeledValue label="Managed by">
          <span className="flex items-center gap-1.5">
            <Icon
              variant="FileCodeIcon"
              size={14}
              className="text-cool-grey-400"
            />
            <Text variant="subtext">Install config</Text>
            {install.configFilePath && (
              <Badge size="sm" variant="code" theme="neutral">
                {install.configFilePath}
              </Badge>
            )}
          </span>
        </LabeledValue>
      )}
      <LabeledValue label="App">
        <Link
          href={`/${install.orgId}/apps/${install.appId}`}
          textVariant="subtext"
        >
          {install.appName}
        </Link>
      </LabeledValue>
    </div>
  </header>
)

// ─── Global status strip ──────────────────────────────────────────────────────

interface IGlobalStatusStrip {
  install: TPlaygroundInstall
  onNavigate: (
    tab: TTopTab,
    opts?: { activityType?: string; resourcesTab?: string }
  ) => void
}

const GlobalStatusStrip = ({ install, onNavigate }: IGlobalStatusStrip) => {
  const { activity, resources, componentStatus } = install

  // Updates: in-flight app_branch_run and deploy workflows
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
    <div
      className="flex items-center flex-wrap gap-px px-2 py-1 border-b bg-cool-grey-50 dark:bg-dark-grey-800 shrink-0"
      aria-label="Install status summary"
    >
      <Button
        variant="ghost"
        size="sm"
        onClick={() => onNavigate('activity', { activityType: 'updates' })}
        aria-label={`Updates: ${updatesLabel}. Navigate to activity.`}
      >
        <Icon
          variant="ArrowsClockwiseIcon"
          size={13}
          className="text-cool-grey-400 mr-1"
        />
        <Text as="span" variant="subtext" weight="strong" theme="neutral">
          Updates
        </Text>
        <span className="ml-1.5">
          <Status status={updatesStatus} variant="badge">
            {updatesLabel}
          </Status>
        </span>
      </Button>

      <span
        className="mx-2 h-3.5 w-px bg-cool-grey-200 dark:bg-dark-grey-600"
        aria-hidden="true"
      />

      <Button
        variant="ghost"
        size="sm"
        onClick={() => onNavigate('resources', { resourcesTab: 'components' })}
        aria-label={`Resources: ${resourcesLabel}. Navigate to resources.`}
      >
        <Icon
          variant="CardsIcon"
          size={13}
          className="text-cool-grey-400 mr-1"
        />
        <Text as="span" variant="subtext" weight="strong" theme="neutral">
          Resources
        </Text>
        <span className="ml-1.5">
          <Status status={resourcesStatus} variant="badge">
            {resourcesLabel}
          </Status>
        </span>
      </Button>

      <span
        className="mx-2 h-3.5 w-px bg-cool-grey-200 dark:bg-dark-grey-600"
        aria-hidden="true"
      />

      <Button
        variant="ghost"
        size="sm"
        onClick={() => onNavigate('resources', { resourcesTab: 'components' })}
        aria-label={`Health checks: ${healthLabel}. Navigate to resources.`}
      >
        <Icon
          variant="PulseIcon"
          size={13}
          className="text-cool-grey-400 mr-1"
        />
        <Text as="span" variant="subtext" weight="strong" theme="neutral">
          Health checks
        </Text>
        <span className="ml-1.5">
          <Status status={healthStatus} variant="badge">
            {healthLabel}
          </Status>
        </span>
      </Button>
    </div>
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
  tracking,
}: {
  tracking: TBranchTracking
}) => {
  const { statusValue, label } = TRACKING_STATUS_MAP[tracking.status]
  const repoHref =
    tracking.repo && !tracking.repo.startsWith('http')
      ? `https://github.com/${tracking.repo}`
      : tracking.repo
  const showDirectory =
    tracking.directory &&
    tracking.directory !== '.' &&
    tracking.directory !== '/'

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
          <span className="flex items-center gap-1.5">
            <Icon
              variant="GitBranchIcon"
              size={13}
              className="text-cool-grey-400 shrink-0"
            />
            <Text variant="subtext" family="mono">
              {tracking.targetBranch}
            </Text>
          </span>
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
      <InstallBranchTrackingCard tracking={install.branchTracking} />
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
    if (
      filter.type === 'updates' &&
      e.type !== 'app_branch_run' &&
      e.type !== 'deploy'
    )
      return false
    if (
      filter.type !== 'all' &&
      filter.type !== 'updates' &&
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
    {versions.map((v) => (
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
          <div className="flex items-center gap-3">
            <Status status={v.status} variant="badge" />
            <Time
              time={v.createdAt}
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
          <Status status={sandbox.status} variant="badge" />
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
          <div className="flex items-center gap-3 shrink-0">
            <Status status={cmp.status} variant="badge" />
            <Time
              time={cmp.deployedAt}
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

interface IResourcesTabPanel {
  resources: TPlaygroundInstall['resources']
  initTab?: string
}

const ResourcesTabPanel = ({ resources, initTab }: IResourcesTabPanel) => (
  <Tabs
    tabs={{
      stack: <StackTab versions={resources.stackVersions} />,
      roles: <RolesTab roles={resources.roles} />,
      sandbox: <SandboxTab sandbox={resources.sandbox} />,
      components: <ComponentsTab components={resources.components} />,
      images: <ImagesTab images={resources.images} />,
    }}
    tabLabels={{
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

// ─── Runbooks tab (top-level) ─────────────────────────────────────────────────

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

// ─── Operations tab (Actions only) ────────────────────────────────────────────

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

const OperationsTab = ({
  operations,
}: {
  operations: TPlaygroundInstall['operations']
}) => <ActionsTab actions={operations.actions} />

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
  // Key increments force ResourcesTabPanel to remount (resets Tabs internal state) when navigating from strip
  const [resourcesNav, setResourcesNav] = useState<{
    tab?: string
    key: number
  }>({ key: 0 })

  const TOP_TABS = Object.keys(TOP_TAB_LABELS) as TTopTab[]

  const handleStripNavigate = (
    tab: TTopTab,
    opts?: { activityType?: string; resourcesTab?: string }
  ) => {
    setActiveTab(tab)
    if (tab === 'activity' && opts?.activityType) {
      setActivityFilter({ ...DEFAULT_ACTIVITY_FILTER, type: opts.activityType })
    }
    if (tab === 'resources' && opts?.resourcesTab) {
      setResourcesNav((prev) => ({ tab: opts.resourcesTab, key: prev.key + 1 }))
    }
  }

  return (
    <div
      className={cn(
        'flex flex-col h-full min-h-0 bg-background border rounded-md overflow-hidden',
        className
      )}
    >
      <InstallPlaygroundHeader install={install} />

      <GlobalStatusStrip install={install} onNavigate={handleStripNavigate} />

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
                {key === 'activity' && (
                  <ActivityTab
                    install={install}
                    filter={activityFilter}
                    onFilterChange={setActivityFilter}
                  />
                )}
                {key === 'runbooks' && (
                  <RunbooksTab runbooks={install.operations.runbooks} />
                )}
                {key === 'resources' && (
                  <ResourcesTabPanel
                    key={resourcesNav.key}
                    resources={install.resources}
                    initTab={resourcesNav.tab}
                  />
                )}
                {key === 'operations' && (
                  <OperationsTab operations={install.operations} />
                )}
              </>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
